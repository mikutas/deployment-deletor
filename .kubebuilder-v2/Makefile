
# Image URL to use all building/pushing image targets
IMG ?= ghcr.io/mikutas/deployment-deletor:$(shell git describe --abbrev=0 --tags)
LOCAL_IMG ?= localhost:5000/deployment-deletor:$(shell git describe --tags)
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

build: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

deploy-local: manifests
	cd config/manager && kustomize edit set image controller=${LOCAL_IMG}
	kustomize build config/default | kubectl apply -f -	

delete: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl delete -f -

delete-local: manifests
	cd config/manager && kustomize edit set image controller=${LOCAL_IMG}
	kustomize build config/default | kubectl delete -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

docker-build-local: test
	docker build . -t ${LOCAL_IMG}

# Push the docker image
docker-push:
	echo $(CR_PAT) | docker login ghcr.io -u USERNAME --password-stdin
	docker push ${IMG}

docker-push-local:
	docker push ${LOCAL_IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

REG_NAME="kind-registry"

cluster:
	./kind-with-registry.sh $(REG_NAME)

cluster-off:
	kind delete cluster
	docker stop $(REG_NAME)
