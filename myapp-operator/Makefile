IMG ?= myapp-operator:latest

##@ General

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Run go fmt
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: run
run: fmt vet ## Run the operator locally (outside the cluster)
	go run ./main.go

##@ Build

.PHONY: build
build: fmt vet ## Build the operator binary
	go build -o bin/manager main.go

.PHONY: docker-build
docker-build: ## Build the Docker image
	docker build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push the Docker image
	docker push ${IMG}

##@ Deployment

.PHONY: install-crd
install-crd: ## Install CRDs into the cluster
	kubectl apply -f config/crd/myapp-crd.yaml

.PHONY: uninstall-crd
uninstall-crd: ## Uninstall CRDs from the cluster
	kubectl delete -f config/crd/myapp-crd.yaml

.PHONY: deploy
deploy: install-crd ## Deploy operator to the cluster
	kubectl apply -f config/rbac/role.yaml
	kubectl apply -f config/manager/manager.yaml

.PHONY: undeploy
undeploy: ## Remove operator from the cluster
	kubectl delete -f config/manager/manager.yaml
	kubectl delete -f config/rbac/role.yaml
	kubectl delete -f config/crd/myapp-crd.yaml

##@ Samples

.PHONY: sample-nginx
sample-nginx: ## Apply the nginx sample
	kubectl apply -f config/samples/myapp-sample-nginx.yaml

.PHONY: sample-scale
sample-scale: ## Apply the scale test sample
	kubectl apply -f config/samples/myapp-sample-scale-test.yaml

.PHONY: sample-clean
sample-clean: ## Delete all sample MyApps
	kubectl delete myapps --all

.PHONY: watch
watch: ## Watch MyApp resources
	kubectl get myapps -w
