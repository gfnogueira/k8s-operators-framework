package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ============================================================================
// MyAppSpec defines the DESIRED state of MyApp
// This is what the user writes in their YAML — the "I want..." part.
// ============================================================================
type MyAppSpec struct {
	// Replicas is the desired number of pod replicas.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10
	// +kubebuilder:default=1
	Replicas int32 `json:"replicas"`

	// Image is the container image to deploy (e.g., "nginx:1.25", "myapp:latest").
	// +kubebuilder:validation:MinLength=1
	Image string `json:"image"`

	// Port is the port the container listens on.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:default=8080
	// +optional
	Port int32 `json:"port,omitempty"`

	// Resources defines CPU/Memory limits (optional, for demo purposes).
	// +optional
	Resources *ResourceSpec `json:"resources,omitempty"`
}

// ResourceSpec defines resource requests/limits for the container.
type ResourceSpec struct {
	// CPULimit e.g. "500m", "1"
	// +optional
	CPULimit string `json:"cpuLimit,omitempty"`
	// MemoryLimit e.g. "128Mi", "1Gi"
	// +optional
	MemoryLimit string `json:"memoryLimit,omitempty"`
}

// ============================================================================
// MyAppStatus defines the OBSERVED state of MyApp
// This is filled by the controller — the "what actually exists" part.
// ============================================================================
type MyAppStatus struct {
	// ReadyReplicas is the number of pods in Ready state.
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// AvailableReplicas is the number of pods available to serve traffic.
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`

	// Phase represents the current lifecycle phase of the MyApp.
	// +kubebuilder:validation:Enum=Pending;Running;Failed
	Phase string `json:"phase,omitempty"`

	// Conditions represent the latest available observations of the MyApp state.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// Phase constants
const (
	PhasePending = "Pending"
	PhaseRunning = "Running"
	PhaseFailed  = "Failed"
)

// Condition types
const (
	// ConditionTypeAvailable indicates the app has minimum availability.
	ConditionTypeAvailable = "Available"
	// ConditionTypeDegraded indicates something is wrong.
	ConditionTypeDegraded = "Degraded"
)

// ============================================================================
// MyApp is the Schema for the myapps API
// This ties Spec + Status together into a full Kubernetes resource.
// ============================================================================

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`
// +kubebuilder:printcolumn:name="Ready",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type MyApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MyAppSpec   `json:"spec,omitempty"`
	Status MyAppStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// MyAppList contains a list of MyApp resources.
type MyAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MyApp `json:"items"`
}
