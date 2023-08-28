# Horizon

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:allowDangerousTypes=true"

# TODO
GV="cluster:v1alpha1"
MANIFESTS="cluster:v1alpha1"

# App Version
APP_VERSION = v0.0.1

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
OPENAPI_GEN ?= ${LOCALBIN}/openapi-gen

## Tool Versions
KUSTOMIZE_VERSION ?= v5.0.0
CONTROLLER_TOOLS_VERSION ?= v0.11.3
OPENAPI_TOOLS_VERSION ?= v0.11.3

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"


.PHONY: binary
binary: | hz-apiserver hz-controller-manager; $(info $(M)...Build all of binary.) @ ## Build all of binary.

.PHONY: hz-apiserver
hz-apiserver: ;$(info $(M)...Begin to build hz-apiserver binary.) @ ## Build hz-apiserver.
	go build -o bin/apiserver ./cmd/hz-apiserver;

.PHONY: fmt
fmt: ;$(info $(M)...Begin to run go fmt against code.)
	gofmt -w ./pkg ./cmd ./tools ./api  ./staging 

.PHONY: manifests
manifests: ;$(info $(M)...Begin to generate manifests e.g. CRD, RBAC etc..)
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: ;$(info $(M)...Begin generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations...) 
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: deepcopy
deepcopy: ;$(info $(M)...Begin to deepcopy.) 
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./staging/github.com/sunweiwe/api/...

.PHONY: openapi
openapi: ;$(info $(M)...Begin to openapi.)  @ ## Openapi.
	${OPENAPI_GEN} -O openapi_generated -i github.com/sunweiwe/api/cluster/v1alpha1 -p github.com/sunweiwe/api/cluster/v1alpha1 -h ./hack/boilerplate.go.txt --report-filename ./api/api-rules/violation_exceptions.list  --output-base=staging/
	go run ./tools/crd-doc-gen/main.go
	go run ./tools/doc-gen/main.go

#  Tool

.PHONY: openapi-gen
openapi-gen: ${OPENAPI_GEN}
${OPENAPI_GEN}: ${LOCALBIN}
	GOBIN=$(LOCALBIN) go install k8s.io/kubernetes/kube-openapi/cmd/openapi-gen@latest

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) 
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: kustomize
kustomize: $(KUSTOMIZE) 
$(KUSTOMIZE): $(LOCALBIN)
	@if test -x $(LOCALBIN)/kustomize && ! $(LOCALBIN)/kustomize version | grep -q $(KUSTOMIZE_VERSION); then \
		echo "$(LOCALBIN)/kustomize version is not expected $(KUSTOMIZE_VERSION). Removing it before installing."; \
		rm -rf $(LOCALBIN)/kustomize; \
	fi
	test -s $(LOCALBIN)/kustomize || { curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) --output install_kustomize.sh && bash install_kustomize.sh $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); rm install_kustomize.sh; }
