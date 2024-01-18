# Copyright (c) 2023 Red Hat, Inc.

########################################################
###### Regional DR OCM-based automated triggering ######
########################################################
default: help

#####################################
###### Image related variables ######
#####################################
IMAGE_BUILDER ?= podman##@ Set a custom image builder if 'podman' is not available
IMAGE_REGISTRY ?= quay.io##@ Set the image registry, defaults to 'quay.io'
IMAGE_NAMESPACE ?= ecosystem-appeng##@ Set the image namespace, defaults to 'ecosystem-appeng'
IMAGE_OPERATOR_NAME ?= regional-dr-trigger-operator##@ Set the operator image name, defaults to 'regional-dr-trigger-operator'
IMAGE_OPERATOR_TAG ?= $(strip $(shell cat VERSION))##@ Set the operator image tag, defaults to content of the VERSION file

##########################################################
###### Create working directories (note .gitignore) ######
##########################################################
LOCALBIN = $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

LOCALBUILD = $(shell pwd)/build
$(LOCALBUILD):
	mkdir -p $(LOCALBUILD)

#####################################
###### Tool binaries variables ######
#####################################
BIN_CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen##@ Set custom 'controller-gen', if not supplied will install latest in ./bin
BIN_KUSTOMIZE ?= $(LOCALBIN)/kustomize##@ Set custom 'kustomize', if not supplied will install latest in ./bin
BIN_GREMLINS ?= $(LOCALBIN)/gremlins##@ Set custom 'gremlins', if not supplied will install latest in ./bin
BIN_GO_TEST_COVERAGE ?= $(LOCALBIN)/go-test-coverage##@ Set custom 'go-test-coverage', if not supplied will install latest in ./bin
BIN_GOLINTCI ?= $(LOCALBIN)/golangci-lint##@ Set custom 'golangci-lint', if not supplied will install latest in ./bin
BIN_ACTIONLINT ?= $(LOCALBIN)/actionlint##@ Set custom 'actionlint', if not supplied will install latest in ./bin
BIN_AWK ?= awk##@ Set a custom 'awk' binary path if not in PATH
BIN_OC ?= oc##@ Set a custom 'oc' binary path if not in PATH
BIN_GO ?= go##@ Set a custom 'go' binary path if not in PATH (useful for multi versions environment)

###############################
###### Various variables ######
###############################
COVERAGE_THRESHOLD ?= 60##@ Set the unit test code coverage threshold, defaults to '60'

#########################
###### Build times ######
#########################
BUILD_DATE = $(strip $(shell date +%FT%T))
BUILD_TIMESTAMP = $(strip $(shell date -d "$(BUILD_DATE)" +%s))

#########################
###### Build flags ######
#########################
COMMIT_HASH = $(strip $(shell git rev-parse --short HEAD))
LDFLAGS=-ldflags="\
-X 'regional-dr-trigger-operator/pkg/version.tag=${IMAGE_OPERATOR_TAG}' \
-X 'regional-dr-trigger-operator/pkg/version.commit=${COMMIT_HASH}' \
-X 'regional-dr-trigger-operator/pkg/version.date=${BUILD_DATE}' \
"

#########################
###### Image names ######
#########################
FULL_OPERATOR_IMAGE_NAME = $(strip $(IMAGE_REGISTRY)/$(IMAGE_NAMESPACE)/$(IMAGE_OPERATOR_NAME):$(IMAGE_OPERATOR_TAG))
FULL_OPERATOR_IMAGE_NAME_UNIQUE = $(FULL_OPERATOR_IMAGE_NAME)_$(COMMIT_HASH)_$(BUILD_TIMESTAMP)

####################################
###### Build and push project ######
####################################
.PHONY: build build/operator
build build/operator: $(LOCALBUILD) ## Build the project as a binary in ./build
	$(BIN_GO) mod tidy
	$(BIN_GO) build $(LDFLAGS) -o $(LOCALBUILD)/rdrtrigger ./main.go

.PHONY: build/operator/image
build/operator/image: ## Build the operator image, customized with IMAGE_REGISTRY, IMAGE_NAMESPACE, IMAGE_OPERATOR_NAME, and IMAGE_OPERATOR_TAG
	$(IMAGE_BUILDER) build --ignorefile ./.gitignore --tag $(FULL_OPERATOR_IMAGE_NAME) -f ./Containerfile

build/operator/image/push: build/operator/image ## Build and push the operator image, customized with IMAGE_REGISTRY, IMAGE_NAMESPACE, IMAGE_OPERATOR_NAME, and IMAGE_OPERATOR_TAG
	$(IMAGE_BUILDER) tag $(FULL_OPERATOR_IMAGE_NAME) $(FULL_OPERATOR_IMAGE_NAME_UNIQUE)
	$(IMAGE_BUILDER) push $(FULL_OPERATOR_IMAGE_NAME_UNIQUE)
	$(IMAGE_BUILDER) push $(FULL_OPERATOR_IMAGE_NAME)

###########################################
###### Code and Manifests generation ######
###########################################
.PHONY: generate/manifests
generate/manifests: $(BIN_CONTROLLER_GEN) ## Generate the manifest files
	$(BIN_CONTROLLER_GEN) rbac:roleName=role paths="./pkg/controller/..."

##############################################
###### Deploy and Undeploy the operator ######
##############################################
deploy/operator: $(BIN_KUSTOMIZE) verify/tools/oc ## Deploy the Regional DR Trigger Operator
	cp config/default/kustomization.yaml config/default/kustomization.yaml.tmp
	cd config/default && $(BIN_KUSTOMIZE) edit set image rdrtrigger-image=$(FULL_OPERATOR_IMAGE_NAME)
	$(BIN_KUSTOMIZE) build config/default | $(BIN_OC) apply -f -
	mv config/default/kustomization.yaml.tmp config/default/kustomization.yaml

undeploy/operator: $(BIN_KUSTOMIZE) verify/tools/oc ## Undeploy the Regional DR Trigger Operator
	cp config/default/kustomization.yaml config/default/kustomization.yaml.tmp
	cd config/default && $(BIN_KUSTOMIZE) edit set image rdrtrigger-image=$(FULL_OPERATOR_IMAGE_NAME)
	$(BIN_KUSTOMIZE) build config/default | $(BIN_OC) delete --ignore-not-found -f -
	mv config/default/kustomization.yaml.tmp config/default/kustomization.yaml

###########################
###### Test codebase ######
###########################
.PHONY: test
test: ## Run all unit tests
	$(BIN_GO) test -v ./...

.PHONY: test/cov
test/cov: $(BIN_GO_TEST_COVERAGE) ## Run all unit tests and print coverage report, use the COVERAGE_THRESHOLD var for setting threshold
	$(BIN_GO) test -failfast -coverprofile=cov.out -v ./...
	$(BIN_GO) tool cover -func=cov.out
	$(BIN_GO) tool cover -html=cov.out -o cov.html
	$(BIN_GO_TEST_COVERAGE) -p cov.out -k 0 -t $(COVERAGE_THRESHOLD)

.PHONY: test/mut
test/mut: $(BIN_GREMLINS) ## Run mutation tests
	$(BIN_GREMLINS) unleash

###########################
###### Lint codebase ######
###########################
lint/all: lint/code lint/ci lint/containerfile ## Lint the entire project (code, ci, containerfile)

.PHONY: lint lint/code
lint lint/code: $(BIN_GOLINTCI) ## Lint the code
	$(BIN_GO) fmt ./...
	$(BIN_GOLINTCI) run

.PHONY: lint/ci
lint/ci: $(BIN_ACTIONLINT) ## Lint the ci
	$(BIN_ACTIONLINT) --verbose

.PHONY: lint/containerfile
lint/containerfile: ## Lint the Containerfile (using Hadolint image, do not use inside a container)
	$(IMAGE_BUILDER) run --rm -i docker.io/hadolint/hadolint:latest < ./Containerfile

################################
###### Display build help ######
################################
help: verify/tools/awk ## Show this help message
	@$(BIN_AWK) 'BEGIN {\
			FS = ".*##@";\
			print "\033[1;31mRegional DR Trigger Operator\033[0m";\
			print "\033[1;32mUsage\033[0m";\
			printf "\t\033[1;37mmake <target> |";\
			printf "\tmake <target> [Variables Set] |";\
            printf "\tmake [Variables Set] <target> |";\
            print "\t[Variables Set] make <target>\033[0m";\
			print "\033[1;32mAvailable Variables\033[0m" }\
		/^(\s|[a-zA-Z_0-9-]|\/)+ \?=.*?##@/ {\
			split($$0,t,"?=");\
			printf "\t\033[1;36m%-35s \033[0;37m%s\033[0m\n",t[1], $$2 | "sort" }'\
		$(MAKEFILE_LIST)
	@$(BIN_AWK) 'BEGIN {\
			FS = ":.*##";\
			SORTED = "sort";\
            print "\033[1;32mAvailable Targets\033[0m"}\
		/^(\s|[a-zA-Z_0-9-]|\/)+:.*?##/ {\
			if($$0 ~ /deploy/)\
				printf "\t\033[1;36m%-35s \033[0;33m%s\033[0m\n", $$1, $$2 | SORTED;\
			else\
				printf "\t\033[1;36m%-35s \033[0;37m%s\033[0m\n", $$1, $$2 | SORTED; }\
		END { \
			close(SORTED);\
			print "\033[1;32mFurther Information\033[0m";\
			print "\t\033[0;37m* Source code: \033[38;5;26mhttps://github.com/RHEcosystemAppEng/multicluster-resiliency-addon\33[0m"}'\
		$(MAKEFILE_LIST)

####################################
###### Install required tools ######
####################################
$(BIN_KUSTOMIZE):
	GOBIN=$(LOCALBIN) $(BIN_GO) install sigs.k8s.io/kustomize/kustomize/v5@latest

$(BIN_CONTROLLER_GEN):
	GOBIN=$(LOCALBIN) $(BIN_GO) install sigs.k8s.io/controller-tools/cmd/controller-gen@latest

$(BIN_GREMLINS):
	GOBIN=$(LOCALBIN) $(BIN_GO) install github.com/go-gremlins/gremlins/cmd/gremlins@latest

$(BIN_GO_TEST_COVERAGE):
	GOBIN=$(LOCALBIN) $(BIN_GO) install github.com/vladopajic/go-test-coverage/v2@latest

$(BIN_GOLINTCI):
	GOBIN=$(LOCALBIN) $(BIN_GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

$(BIN_ACTIONLINT): # recommendation: manually install shellcheck and verify it's on your PATH, it will be picked up by actionlint
	GOBIN=$(LOCALBIN) $(BIN_GO) install github.com/rhysd/actionlint/cmd/actionlint@latest

######################################
###### Verify tools availablity ######
######################################

# member 1 is the missing tool name, member 2 is the name of the variable used for customizing the tool path
TOOL_MISSING_ERR_MSG = Please install '$(1)' or specify a custom path using the '$(2)' variable

.PHONY: verify/tools/awk
verify/tools/awk:
ifeq (,$(shell which $(BIN_AWK) 2> /dev/null ))
	$(error $(call TOOL_MISSING_ERR_MSG,awk,BIN_AWK))
endif

.PHONY: verify/tools/oc
verify/tools/oc:
ifeq (,$(shell which $(BIN_OC) 2> /dev/null ))
	$(error $(call TOOL_MISSING_ERR_MSG,oc,BIN_OC))
endif
