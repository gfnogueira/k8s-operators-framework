package controllers

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1alpha1 "github.com/nogueira/myapp-operator/api/v1alpha1"
)

const (
	// finalizerName is used to ensure cleanup logic runs before the resource is deleted.
	finalizerName = "app.gfnogueira.com/finalizer"
)

// ============================================================================
// MyAppReconciler reconciles a MyApp object.
//
// This is the HEART of the operator. Every time a MyApp resource is created,
// updated, or deleted, the Reconcile function is called.
//
// The reconciler's job is simple:
//  1. Read the desired state (MyApp spec)
//  2. Read the current state (existing Deployment + Service)
//  3. Make changes to converge current → desired
//
// ============================================================================
type MyAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=app.gfnogueira.com,resources=myapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.gfnogueira.com,resources=myapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=app.gfnogueira.com,resources=myapps/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is the main reconciliation loop.
// It is called every time a MyApp resource changes (create, update, delete).
//
// KEY PRINCIPLE: This function MUST be idempotent.
// Calling it 100 times in a row should produce the same result as calling it once.
func (r *MyAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("=== Reconcile triggered ===", "myapp", req.NamespacedName)

	// ────────────────────────────────────────────────
	// STEP 1: Fetch the MyApp instance
	// ────────────────────────────────────────────────
	myapp := &appv1alpha1.MyApp{}
	if err := r.Get(ctx, req.NamespacedName, myapp); err != nil {
		if errors.IsNotFound(err) {
			// The resource was deleted before we could reconcile.
			// Nothing to do — return without error.
			logger.Info("MyApp resource not found — probably deleted")
			return ctrl.Result{}, nil
		}
		// Real error (network, RBAC, etc.)
		return ctrl.Result{}, err
	}

	// ────────────────────────────────────────────────
	// STEP 2: Handle deletion with Finalizer
	// Finalizers ensure cleanup logic runs BEFORE the resource is removed.
	// Example: you might need to delete external resources (DNS, cloud LB, etc.)
	// ────────────────────────────────────────────────
	if myapp.ObjectMeta.DeletionTimestamp.IsZero() {
		// Resource is NOT being deleted → ensure our finalizer is registered
		if !controllerutil.ContainsFinalizer(myapp, finalizerName) {
			controllerutil.AddFinalizer(myapp, finalizerName)
			if err := r.Update(ctx, myapp); err != nil {
				return ctrl.Result{}, err
			}
			logger.Info("Added finalizer")
		}
	} else {
		// Resource IS being deleted → run cleanup logic
		if controllerutil.ContainsFinalizer(myapp, finalizerName) {
			logger.Info("Running finalizer cleanup logic...")

			// 🔧 PUT YOUR CLEANUP LOGIC HERE
			// Examples: remove external DNS records, deregister from service discovery, etc.

			// Remove the finalizer so K8s can actually delete the resource
			controllerutil.RemoveFinalizer(myapp, finalizerName)
			if err := r.Update(ctx, myapp); err != nil {
				return ctrl.Result{}, err
			}
			logger.Info("Finalizer removed, resource will be deleted")
		}
		return ctrl.Result{}, nil
	}

	// ────────────────────────────────────────────────
	// STEP 3: Reconcile the Deployment
	// Ensure a Deployment exists that matches the MyApp spec.
	// ────────────────────────────────────────────────
	deployment := &appsv1.Deployment{}
	deploymentName := types.NamespacedName{
		Name:      myapp.Name,
		Namespace: myapp.Namespace,
	}

	err := r.Get(ctx, deploymentName, deployment)
	if err != nil && errors.IsNotFound(err) {
		// Deployment doesn't exist → CREATE it
		logger.Info("Creating Deployment", "name", myapp.Name)
		deployment = r.buildDeployment(myapp)

		// SetControllerReference makes MyApp the "owner" of this Deployment.
		// When MyApp is deleted, the Deployment is garbage collected automatically.
		if err := ctrl.SetControllerReference(myapp, deployment, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, deployment); err != nil {
			logger.Error(err, "Failed to create Deployment")
			return ctrl.Result{}, err
		}
		logger.Info("Deployment created successfully")

	} else if err != nil {
		// Real error
		return ctrl.Result{}, err

	} else {
		// Deployment EXISTS → check if it needs updating
		needsUpdate := false

		// Check replicas
		if *deployment.Spec.Replicas != myapp.Spec.Replicas {
			logger.Info("Updating replicas",
				"current", *deployment.Spec.Replicas,
				"desired", myapp.Spec.Replicas,
			)
			deployment.Spec.Replicas = &myapp.Spec.Replicas
			needsUpdate = true
		}

		// Check image
		currentImage := deployment.Spec.Template.Spec.Containers[0].Image
		if currentImage != myapp.Spec.Image {
			logger.Info("Updating image",
				"current", currentImage,
				"desired", myapp.Spec.Image,
			)
			deployment.Spec.Template.Spec.Containers[0].Image = myapp.Spec.Image
			needsUpdate = true
		}

		// Check port
		if myapp.Spec.Port > 0 {
			currentPort := deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort
			if currentPort != myapp.Spec.Port {
				logger.Info("Updating port",
					"current", currentPort,
					"desired", myapp.Spec.Port,
				)
				deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = myapp.Spec.Port
				needsUpdate = true
			}
		}

		if needsUpdate {
			if err := r.Update(ctx, deployment); err != nil {
				logger.Error(err, "Failed to update Deployment")
				return ctrl.Result{}, err
			}
			logger.Info("Deployment updated successfully")
		}
	}

	// ────────────────────────────────────────────────
	// STEP 4: Reconcile the Service
	// Ensure a Service exists to expose the Deployment.
	// ────────────────────────────────────────────────
	service := &corev1.Service{}
	serviceName := types.NamespacedName{
		Name:      myapp.Name + "-svc",
		Namespace: myapp.Namespace,
	}

	err = r.Get(ctx, serviceName, service)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating Service", "name", serviceName.Name)
		service = r.buildService(myapp)
		if err := ctrl.SetControllerReference(myapp, service, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, service); err != nil {
			logger.Error(err, "Failed to create Service")
			return ctrl.Result{}, err
		}
		logger.Info("Service created successfully")
	} else if err != nil {
		return ctrl.Result{}, err
	}

	// ────────────────────────────────────────────────
	// STEP 5: Update Status
	// Report the observed state back to the MyApp status.
	// ────────────────────────────────────────────────
	if err := r.Get(ctx, deploymentName, deployment); err == nil {
		myapp.Status.ReadyReplicas = deployment.Status.ReadyReplicas
		myapp.Status.AvailableReplicas = deployment.Status.AvailableReplicas

		// Determine phase
		if deployment.Status.ReadyReplicas == myapp.Spec.Replicas {
			myapp.Status.Phase = appv1alpha1.PhaseRunning
		} else if deployment.Status.ReadyReplicas > 0 {
			myapp.Status.Phase = appv1alpha1.PhasePending
		} else {
			myapp.Status.Phase = appv1alpha1.PhasePending
		}

		// Set condition
		condition := metav1.Condition{
			Type:               appv1alpha1.ConditionTypeAvailable,
			LastTransitionTime: metav1.Now(),
		}
		if deployment.Status.AvailableReplicas >= myapp.Spec.Replicas {
			condition.Status = metav1.ConditionTrue
			condition.Reason = "MinimumReplicasAvailable"
			condition.Message = fmt.Sprintf("Deployment has %d/%d available replicas",
				deployment.Status.AvailableReplicas, myapp.Spec.Replicas)
		} else {
			condition.Status = metav1.ConditionFalse
			condition.Reason = "ReplicasUnavailable"
			condition.Message = fmt.Sprintf("Deployment has %d/%d available replicas",
				deployment.Status.AvailableReplicas, myapp.Spec.Replicas)
		}
		meta.SetStatusCondition(&myapp.Status.Conditions, condition)

		if err := r.Status().Update(ctx, myapp); err != nil {
			logger.Error(err, "Failed to update MyApp status")
			return ctrl.Result{}, err
		}
		logger.Info("Status updated",
			"ready", myapp.Status.ReadyReplicas,
			"phase", myapp.Status.Phase,
		)
	}

	// Requeue after 30 seconds to keep status in sync
	// (In production, you'd rely more on watches than periodic requeue)
	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// ============================================================================
// buildDeployment constructs a Deployment object from the MyApp spec.
// This is a pure function — no side effects, just data transformation.
// ============================================================================
func (r *MyAppReconciler) buildDeployment(myapp *appv1alpha1.MyApp) *appsv1.Deployment {
	labels := map[string]string{
		"app":                          myapp.Name,
		"app.kubernetes.io/name":       myapp.Name,
		"app.kubernetes.io/managed-by": "myapp-operator",
	}

	port := myapp.Spec.Port
	if port == 0 {
		port = 8080
	}

	replicas := myapp.Spec.Replicas

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      myapp.Name,
			Namespace: myapp.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  myapp.Name,
							Image: myapp.Spec.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: port,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							// Liveness and Readiness probes for production-grade setup
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.FromInt32(port),
									},
								},
								InitialDelaySeconds: 15,
								PeriodSeconds:       20,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.FromInt32(port),
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       10,
							},
						},
					},
				},
			},
		},
	}

	return deployment
}

// ============================================================================
// buildService constructs a ClusterIP Service to expose the Deployment.
// ============================================================================
func (r *MyAppReconciler) buildService(myapp *appv1alpha1.MyApp) *corev1.Service {
	labels := map[string]string{
		"app":                          myapp.Name,
		"app.kubernetes.io/name":       myapp.Name,
		"app.kubernetes.io/managed-by": "myapp-operator",
	}

	port := myapp.Spec.Port
	if port == 0 {
		port = 8080
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      myapp.Name + "-svc",
			Namespace: myapp.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Port:       port,
					TargetPort: intstr.FromInt32(port),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

// ============================================================================
// SetupWithManager registers the controller with the manager.
// It tells controller-runtime:
//   - Watch MyApp resources (primary)
//   - Watch Deployments owned by MyApp (secondary — so we reconcile when
//     the Deployment changes too, not just the MyApp CR)
//
// ============================================================================
func (r *MyAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.MyApp{}).  // Watch MyApp (primary resource)
		Owns(&appsv1.Deployment{}). // Watch Deployments owned by MyApp
		Owns(&corev1.Service{}).    // Watch Services owned by MyApp
		Complete(r)
}
