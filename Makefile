# Copyright (c) 2023 Red Hat, Inc.

##########################################
###### Regional DR Trigger Operator ######
##########################################
default: help

#####################################
###### Image related variables ######
#####################################
IMAGE_BUILDER ?= podman##@ Set a custom image builder if 'podman' is not available
IMAGE_REGISTRY ?= quay.io##@ Set the image registry, defaults to 'quay.io'
IMAGE_NAMESPACE ?= ecosystem-appeng##@ Set the image namespace, defaults to 'ecosystem-appeng'
IMAGE_NAME ?= regional-dr-trigger-operator##@ Set the operator image name, defaults to 'regional-dr-trigger-operator'
IMAGE_TAG ?= $(strip $(shell cat VERSION))##@ Set the operator image tag, defaults to content of the VERSION file

######################################
###### Bundle related variables ######
######################################
BUNDLE_PACKAGE_NAME ?= $(IMAGE_NAME)##@ Set the bundle package name, defaults to IMAGE_NAME
BUNDLE_CHANNELS ?= alpha##@ Set a comma-seperated list of channels the bundle belongs too, defaults to 'alpha'
BUNDLE_DEFAULT_CHANNEL ?= alpha##@ Set the default channel for the bundle, defaults to 'alpha'
BUNDLE_IMAGE_NAME ?= $(IMAGE_NAME)-bundle##@ Set the image name for the bundle, defaults to IMAGE_NAME-bundle
BUNDLE_TARGET_NAMESPACE ?= regional-dr-trigger##@ Set the target namespace for running the bundle, defaults to 'regional-dr-trigger'
BUNDLE_SCORECARD_NAMESPACE ?= $(IMAGE_NAME)-scorecard##@ Set the target namespace for running scorecard tests, defaults to IMAGE_NAME-scorecard

#####################################
###### Tools version variables ######
#####################################
VERSION_CONTROLLER_GEN = v0.14.0
VERSION_OPERATOR_SDK = v1.33.0
VERSION_KUSTOMIZE = v5.3.0
VERSION_GREMLINS = v0.5.0
VERSION_GO_TEST_COVERAGE = v2.8.2
VERSION_GOLANG_CI_LINT = v1.55.2
VERSION_ACTIONLINT = v1.6.26

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
BIN_OPERATOR_SDK ?= $(LOCALBIN)/operator-sdk##@ Set custom 'operator-sdk', if not supplied will install latest in ./bin
BIN_KUSTOMIZE ?= $(LOCALBIN)/kustomize##@ Set custom 'kustomize', if not supplied will install latest in ./bin
BIN_GREMLINS ?= $(LOCALBIN)/gremlins##@ Set custom 'gremlins', if not supplied will install latest in ./bin
BIN_GO_TEST_COVERAGE ?= $(LOCALBIN)/go-test-coverage##@ Set custom 'go-test-coverage', if not supplied will install latest in ./bin
BIN_GOLINTCI ?= $(LOCALBIN)/golangci-lint##@ Set custom 'golangci-lint', if not supplied will install latest in ./bin
BIN_ACTIONLINT ?= $(LOCALBIN)/actionlint##@ Set custom 'actionlint', if not supplied will install latest in ./bin
BIN_AWK ?= awk##@ Set a custom 'awk' binary path if not in PATH
BIN_OC ?= oc##@ Set a custom 'oc' binary path if not in PATH
BIN_GO ?= go##@ Set a custom 'go' binary path if not in PATH (useful for multi versions environment)
BIN_CURL ?= curl##@ Set a custom 'curl' binary path if not in PATH
BIN_YQ ?= yq##@ Set a custom 'yq' binary path if not in PATH

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
-X 'regional-dr-trigger-operator/pkg/version.tag=${IMAGE_TAG}' \
-X 'regional-dr-trigger-operator/pkg/version.commit=${COMMIT_HASH}' \
-X 'regional-dr-trigger-operator/pkg/version.date=${BUILD_DATE}' \
"

#########################
###### Image names ######
#########################
FULL_OPERATOR_IMAGE_NAME = $(strip $(IMAGE_REGISTRY)/$(IMAGE_NAMESPACE)/$(IMAGE_NAME):$(IMAGE_TAG))
FULL_OPERATOR_IMAGE_NAME_UNIQUE = $(FULL_OPERATOR_IMAGE_NAME)_$(COMMIT_HASH)_$(BUILD_TIMESTAMP)
FULL_BUNDLE_IMAGE_NAME = $(strip $(IMAGE_REGISTRY)/$(IMAGE_NAMESPACE)/$(BUNDLE_IMAGE_NAME):$(IMAGE_TAG))
FULL_BUNDLE_IMAGE_NAME_UNIQUE = $(FULL_BUNDLE_IMAGE_NAME)_$(COMMIT_HASH)_$(BUILD_TIMESTAMP)

####################################
###### Build and push project ######
####################################
build/all/image: build/operator/image build/bundle/image ## Build both the operator and bundle images

build/all/image/push: build/operator/image/push build/bundle/image/push ## Build and push both the operator and bundle images

.PHONY: build build/operator
build build/operator: $(LOCALBUILD) ## Build the project as a binary in ./build
	$(BIN_GO) mod tidy
	$(BIN_GO) build $(LDFLAGS) -o $(LOCALBUILD)/rdrtrigger ./main.go

.PHONY: build/operator/image
build/operator/image: ## Build the operator image, customized with IMAGE_REGISTRY, IMAGE_NAMESPACE, IMAGE_NAME, and IMAGE_TAG
	$(IMAGE_BUILDER) build --ignorefile ./.gitignore --tag $(FULL_OPERATOR_IMAGE_NAME) -f ./Containerfile

build/operator/image/push: build/operator/image ## Build and push the operator image, customized with IMAGE_REGISTRY, IMAGE_NAMESPACE, IMAGE_NAME, and IMAGE_TAG
	$(IMAGE_BUILDER) tag $(FULL_OPERATOR_IMAGE_NAME) $(FULL_OPERATOR_IMAGE_NAME_UNIQUE)
	$(IMAGE_BUILDER) push $(FULL_OPERATOR_IMAGE_NAME_UNIQUE)
	$(IMAGE_BUILDER) push $(FULL_OPERATOR_IMAGE_NAME)

.PHONY: build/bundle/image
build/bundle/image: ## Build the bundle image, customized with IMAGE_REGISTRY, IMAGE_NAMESPACE, BUNDLE_IMAGE_NAME, and IMAGE_TAG
	$(IMAGE_BUILDER) build --ignorefile ./.gitignore --tag $(FULL_BUNDLE_IMAGE_NAME) -f ./bundle.Containerfile

build/bundle/image/push: build/bundle/image ## Build and push the bundle image, customized with IMAGE_REGISTRY, IMAGE_NAMESPACE, BUNDLE_IMAGE_NAME, and IMAGE_TAG
	$(IMAGE_BUILDER) tag $(FULL_BUNDLE_IMAGE_NAME) $(FULL_BUNDLE_IMAGE_NAME_UNIQUE)
	$(IMAGE_BUILDER) push $(FULL_BUNDLE_IMAGE_NAME_UNIQUE)
	$(IMAGE_BUILDER) push $(FULL_BUNDLE_IMAGE_NAME)

###########################################
###### Code and Manifests generation ######
###########################################
generate/all: generate/manifests generate/bundle ## Generate both rbac and olm bundle files

.PHONY: generate/manifests
generate/manifests: $(BIN_CONTROLLER_GEN) $(BIN_KUSTOMIZE) ## Generate rbac manifest files
	$(BIN_CONTROLLER_GEN) rbac:roleName=role paths="./pkg/controller/..."

generate/bundle: verify/tools/curl verify/tools/yq $(BIN_OPERATOR_SDK) ## Generate olm bundle
	cp config/default/kustomization.yaml config/default/kustomization.yaml.tmp
	cd config/default && $(BIN_KUSTOMIZE) edit set image rdrtrigger-image=$(FULL_OPERATOR_IMAGE_NAME)
	$(BIN_YQ) -i '.labels[1].pairs."app.kubernetes.io/instance" = "rdrtrigger-$(IMAGE_TAG)"' config/default/kustomization.yaml
	$(BIN_YQ) -i '.labels[1].pairs."app.kubernetes.io/version" = "$(IMAGE_TAG)"' config/default/kustomization.yaml
	$(BIN_KUSTOMIZE) build config/manifests | $(BIN_OPERATOR_SDK) generate bundle --quiet --version $(IMAGE_TAG) \
	--package $(BUNDLE_PACKAGE_NAME) --channels $(BUNDLE_CHANNELS) --default-channel $(BUNDLE_DEFAULT_CHANNEL)
	rm -f ./bundle.Containerfile
	mv ./bundle.Dockerfile ./bundle.Containerfile
	mv config/default/kustomization.yaml.tmp config/default/kustomization.yaml

##############################################
###### Deploy and Undeploy the operator ######
##############################################
operator/deploy: $(BIN_KUSTOMIZE) verify/tools/oc verify/tools/yq ## Deploy the Regional DR Trigger Operator
	cp config/default/kustomization.yaml config/default/kustomization.yaml.tmp
	cd config/default && $(BIN_KUSTOMIZE) edit set image rdrtrigger-image=$(FULL_OPERATOR_IMAGE_NAME)
	$(BIN_YQ) -i '.labels[1].pairs."app.kubernetes.io/instance" = "rdrtrigger-$(IMAGE_TAG)"' config/default/kustomization.yaml
	$(BIN_YQ) -i '.labels[1].pairs."app.kubernetes.io/version" = "$(IMAGE_TAG)"' config/default/kustomization.yaml
	$(BIN_KUSTOMIZE) build config/default | $(BIN_OC) apply -f -
	mv config/default/kustomization.yaml.tmp config/default/kustomization.yaml

operator/undeploy: $(BIN_KUSTOMIZE) verify/tools/oc verify/tools/yq ## Undeploy the Regional DR Trigger Operator
	cp config/default/kustomization.yaml config/default/kustomization.yaml.tmp
	cd config/default && $(BIN_KUSTOMIZE) edit set image rdrtrigger-image=$(FULL_OPERATOR_IMAGE_NAME)
	$(BIN_YQ) -i '.labels[1].pairs."app.kubernetes.io/instance" = "rdrtrigger-$(IMAGE_TAG)"' config/default/kustomization.yaml
	$(BIN_YQ) -i '.labels[1].pairs."app.kubernetes.io/version" = "$(IMAGE_TAG)"' config/default/kustomization.yaml
	$(BIN_KUSTOMIZE) build config/default | $(BIN_OC) delete --ignore-not-found -f -
	mv config/default/kustomization.yaml.tmp config/default/kustomization.yaml

.PHONY: bundle/run
bundle/run: $(BIN_OPERATOR_SDK) verify/tools/oc ## Run the Regional DR Trigger Operator OLM Bundle from image
	-$(BIN_OC) create ns $(BUNDLE_TARGET_NAMESPACE)
	$(BIN_OPERATOR_SDK) run bundle $(FULL_BUNDLE_IMAGE_NAME) -n $(BUNDLE_TARGET_NAMESPACE)

.PHONY: bundle/cleanup
bundle/cleanup: $(BIN_OPERATOR_SDK) ## Cleanup the Regional DR Trigger Operator OLM Bundle package installed
	$(BIN_OPERATOR_SDK) cleanup $(BUNDLE_PACKAGE_NAME) -n $(BUNDLE_TARGET_NAMESPACE)

.PHONY: bundle/cleanup/namespace
bundle/cleanup/namespace: verify/tools/oc ## DELETE the Regional DR Trigger Operator OLM Bundle namespace (BE CAREFUL)
	$(BIN_OC) delete ns $(BUNDLE_TARGET_NAMESPACE)

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

.PHONY: test/bundle
test/bundle: $(BIN_OPERATOR_SDK) verify/tools/oc ## Run Scorecard Bundle Tests (requires connected cluster)
	@ { \
	if $(BIN_OC) create ns $(BUNDLE_SCORECARD_NAMESPACE); then \
		$(BIN_OPERATOR_SDK) scorecard ./bundle -n $(BUNDLE_SCORECARD_NAMESPACE) --pod-security=restricted; \
		$(BIN_OC) delete ns $(BUNDLE_SCORECARD_NAMESPACE); \
	else \
		$(BIN_OPERATOR_SDK) scorecard ./bundle -n $(BUNDLE_SCORECARD_NAMESPACE) --pod-security=restricted; \
	fi \
	}

.PHONY: test/bundle/delete/ns
test/bundle/delete/ns: verify/tools/oc ## DELETE the Scorecard namespace (BE CAREFUL)
	$(BIN_OC) delete ns $(BUNDLE_SCORECARD_NAMESPACE)

###########################
###### Lint codebase ######
###########################
lint/all: lint/code lint/ci lint/containerfile lint/bundle ## Lint the entire project (code, ci, containerfile)

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

.PHONY: lint/bundle
lint/bundle: $(BIN_OPERATOR_SDK) ## Validate OLM bundle
	$(BIN_OPERATOR_SDK) bundle validate ./bundle --select-optional suite=operatorframework

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
			else if($$0 ~ /push/)\
				printf "\t\033[1;36m%-35s \033[0;35m%s\033[0m\n", $$1, $$2 | SORTED;\
			else if($$0 ~ /DELETE/)\
				printf "\t\033[1;36m%-35s \033[0;31m%s\033[0m\n", $$1, $$2 | SORTED;\
			else\
				printf "\t\033[1;36m%-35s \033[0;37m%s\033[0m\n", $$1, $$2 | SORTED; }\
		END { \
			close(SORTED);\
			print "\033[1;32mFurther Information\033[0m";\
			print "\t\033[0;37m* Source code: \033[38;5;26mhttps://github.com/RHEcosystemAppEng/regional-dr-trigger-operator\33[0m"}'\
		$(MAKEFILE_LIST)

####################################
###### Install required tools ######
####################################
$(BIN_KUSTOMIZE): $(LOCALBIN)
	GOBIN=$(LOCALBIN) $(BIN_GO) install sigs.k8s.io/kustomize/kustomize/v5@$(VERSION_KUSTOMIZE)

$(BIN_CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) $(BIN_GO) install sigs.k8s.io/controller-tools/cmd/controller-gen@$(VERSION_CONTROLLER_GEN)

$(BIN_GREMLINS): $(LOCALBIN)
	GOBIN=$(LOCALBIN) $(BIN_GO) install github.com/go-gremlins/gremlins/cmd/gremlins@$(VERSION_GREMLINS)

$(BIN_GO_TEST_COVERAGE): $(LOCALBIN)
	GOBIN=$(LOCALBIN) $(BIN_GO) install github.com/vladopajic/go-test-coverage/v2@$(VERSION_GO_TEST_COVERAGE)

$(BIN_GOLINTCI): $(LOCALBIN)
	GOBIN=$(LOCALBIN) $(BIN_GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(VERSION_GOLANG_CI_LINT)

$(BIN_ACTIONLINT): $(LOCALBIN) # recommendation: manually install shellcheck and verify it's on your PATH, it will be picked up by actionlint
	GOBIN=$(LOCALBIN) $(BIN_GO) install github.com/rhysd/actionlint/cmd/actionlint@$(VERSION_ACTIONLINT)

$(BIN_OPERATOR_SDK): $(LOCALBIN)
	OS=$(shell go env GOOS) && \
	ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(BIN_OPERATOR_SDK) https://github.com/operator-framework/operator-sdk/releases/download/$(VERSION_OPERATOR_SDK)/operator-sdk_$${OS}_$${ARCH}
	chmod +x $(BIN_OPERATOR_SDK)

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

.PHONY: verify/tools/curl
verify/tools/curl:
ifeq (,$(shell which $(BIN_CURL) 2> /dev/null ))
	$(error $(call TOOL_MISSING_ERR_MSG,curl,BIN_CURL))
endif

.PHONY: verify/tools/yq
verify/tools/yq:
ifeq (,$(shell which $(BIN_YQ) 2> /dev/null ))
	$(error $(call TOOL_MISSING_ERR_MSG,yq,BIN_YQ))
endif
