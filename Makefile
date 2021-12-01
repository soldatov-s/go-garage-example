SHELL=/bin/bash -o pipefail

# Set envs
ifeq ($(REGISTRY),)
	IMAGE_REGISTRY=nexus.yourdomain.com
	ifneq ($(REGISTRY_POSTFIX),)
		IMAGE_REGISTRY=nexus.yourdomain.com/$(REGISTRY_POSTFIX)
	endif
else
	IMAGE_REGISTRY=$(REGISTRY)
	ifneq ($(REGISTRY_POSTFIX),)
		IMAGE_REGISTRY=$(REGISTRY)/$(REGISTRY_POSTFIX)
	endif
endif

APIDIR=./api

BUILDTIME=$(shell TZ=UTC date -u '+%Y-%m-%dT%H:%M:%SZ' 2>/dev/null)

# PACKAGE name when build on gitlab
PACKAGE=$(CI_SERVER_HOST)/$(CI_PROJECT_NAMESPACE)/$(CI_PROJECT_NAME)
ifeq ($(PACKAGE),//)
	# PACKAGE name when build local
	PACKAGE=$(shell go list -mod vendor 2>/dev/null)
endif

ifneq ($(PACKAGE),)
	PROJECT_NAME=$(shell basename $(PACKAGE))
endif

TAG=$(shell git describe --tags 2>/dev/null; true)
ifeq ($(TAG),)
	GITVER=$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null; true)
else ifneq (,$(findstring -g, $(TAG)))
	GITVER=$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null; true)
else
	GITVER=$(TAG)
endif

HASH=$(shell git rev-parse HEAD 2>/dev/null | cut -c 1-8)
ifeq ($(GITVER), HEAD)
	GITVER=$(shell git name-rev ${HASH} 2>/dev/null | cut -c 10- ; true)
endif

ifneq ($(GITVER),)
	IMAGETAG=$(shell basename $(GITVER) | tr '[:upper:]' '[:lower:]')
endif

DESCRIPTION=$(shell cat .appdesc)

# For case builds docker-image for stage stand
ifneq ($(CI_MERGE_REQUEST_SOURCE_BRANCH_NAME),)
	IMAGETAG=$(shell basename $(CI_MERGE_REQUEST_SOURCE_BRANCH_NAME) | tr '[:upper:]' '[:lower:]')
endif

# Recipes
all: help

.PHONY: init-stable 
init-stable: ## Init stable version
	GO111MODULE=on go mod vendor

.PHONY: build-stable
build-stable: ## Build stable version
	@echo Build ${PACKAGE} ${GITVER} ${BUILDTIME} ${HASH}; \
	GO111MODULE=on go install -mod vendor -ldflags "\
		-X '${PACKAGE}/internal/cmd.version=${GITVER}' \
		-X '${PACKAGE}/internal/cmd.builded=${BUILDTIME}' \
		-X '${PACKAGE}/internal/cmd.hash=${HASH}' \
		-X '${PACKAGE}/internal/cmd.appName=${PROJECT_NAME}' \
		-X '${PACKAGE}/internal/cmd.description=${DESCRIPTION}'" \
		.

.PHONY: run
run: build-stable ## Build and run stable version
	@${GOPATH}/bin/${PROJECT_NAME}

.PHONY: test
test: ## Run tests
	@CGO_ENABLED=0 go test -mod vendor -test.v -cover ./...

bench: ## Run bench
	@CGO_ENABLED=0 go test -mod vendor -bench=. ./...

race: ## Run data race detector
	@CGO_ENABLED=0 go test -mod vendor -race -short ./...

msan: ## Run memory sanitizer
	@CC=clang go test -mod vendor -msan -short ./...

.PHONY: lint
lint: ## Lint the files
	@golangci-lint --version; \
	golangci-lint linters; \
	CGO_ENABLED=0 golangci-lint run -v

.PHONY: docker-build
docker-build: ## Build the docker image
	@echo "docker build --build-arg APP_NAME=${PROJECT_NAME} --build-arg PACKAGE=${PACKAGE} -t ${IMAGE_REGISTRY}/${PROJECT_NAME}:${IMAGETAG} ."
	@docker build --build-arg APP_NAME=${PROJECT_NAME} --build-arg PACKAGE=${PACKAGE} -t ${IMAGE_REGISTRY}/${PROJECT_NAME}:${IMAGETAG} .

.PHONY: docker-build-latest
docker-build-latest: ## Build the docker image with tag latest
	docker build --build-arg APP_NAME=${PROJECT_NAME} --build-arg PACKAGE=${PACKAGE} -t ${IMAGE_REGISTRY}/${PROJECT_NAME}:${IMAGETAG} -t ${IMAGE_REGISTRY}/${PROJECT_NAME}:latest .

.PHONY: docker-push
docker-push: docker-build ## Build and push the docker image
	@docker push ${IMAGE_REGISTRY}/${PROJECT_NAME}:${IMAGETAG}

.PHONY: docker-push-latest
docker-push-latest: docker-build-latest ## Build and push the docker image with tag latest
	@docker push ${IMAGE_REGISTRY}/${PROJECT_NAME}:${IMAGETAG}
	@docker push ${IMAGE_REGISTRY}/${PROJECT_NAME}:latest

.PHONY: docker-compose-build
docker-compose-build: ## Build docker-compose
	@APP_NAME=${PROJECT_NAME} PACKAGE=${PACKAGE} REGISTRY=${IMAGE_REGISTRY} docker-compose build

.PHONY: docker-compose-up
docker-compose-up: docker-compose-build ## Build and run docker-compose
	@APP_NAME=${PROJECT_NAME} PACKAGE=${PACKAGE} REGISTRY=${IMAGE_REGISTRY} docker-compose up

.PHONY: docker-compose-clear
docker-compose-clear: ## Clear docker-compose
	@APP_NAME=${PROJECT_NAME} PACKAGE=${PACKAGE} REGISTRY=${IMAGE_REGISTRY} docker-compose rm -s -f -v
	@APP_NAME=${PROJECT_NAME} PACKAGE=${PACKAGE} REGISTRY=${IMAGE_REGISTRY} docker-compose down -v

.PHONY: gen-mocks
gen-mocks: ## Run generate mocks
	go generate ./...

define gen_swagger_by_api
	oapi-codegen -generate "types" -package "api$(2)" $(1)/swagger.yml -o > $(1)/types.gen.go; \
	oapi-codegen -generate "server" -package "api$(2)" $(1)/swagger.yml -o > $(1)/server.gen.go; \
	oapi-codegen -generate "spec" -package "api$(2)" $(1)/swagger.yml -o > $(1)/spec.gen.go
endef

.PHONY: gen-swagger
gen-swagger: ## Run generate swagger
	@for f in $(shell ls -d ${APIDIR}/*/); do echo swagger for $${f}; $(call gen_swagger_by_api,$$f,$$(basename $$f)); done

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; \
	{printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
		

.PHONY: all build test lint race
