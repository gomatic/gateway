export APP_NAME := $(patsubst service-gravity-%,%,$(notdir $(shell pwd)))
DESC :=
PROJECT_URL := "https://github.com/gomatic/$(APP_NAME)"

.PHONY : build linux darwin run container
.PHONY : help report
.DEFAULT_GOAL := help

PREFIX ?= usr/local

# Capture the
export RELEASE := $(shell lsb_release -rs 2>/dev/null)

# Capture the commit and branch
export BRANCH ?= $(shell git rev-parse --symbolic-full-name --abbrev-ref HEAD 2>/dev/null)
export COMMIT_ID := $(shell git log --pretty=format:'%h' -n 1 2>/dev/null)
export COMMIT_TIME := $(shell git show -s --format=%ct 2>/dev/null)
export VERSION := $(COMMIT_TIME)-$(COMMIT_ID)

export STARTD := $(shell pwd)
export THIS := $(abspath $(lastword $(MAKEFILE_LIST)))
export THISD := $(dir $(THIS))

PACKAGE := github.com/gomatic/$(APP_NAME)/vendor/service.VERSION=$(VERSION)

build: $(APP_NAME) ## Make everything

$(APP_NAME) $(GOBIN)/$(APP_NAME): $(DEPENDS) $(SOURCES)
	go vet
	go build -ldflags "-X $(PACKAGE)" -v -o $@

install: $(GOBIN)/$(APP_NAME) ## Install to GOBIN

$(DEPENDS):
	git submodule update --init --recursive

client: cmd/client/client

cmd/client/client: cmd/client/main.go
	cd $(dir $@); go build -ldflags="-s -w -X $(PACKAGE)" -a -installsuffix cgo -o $(notdir $@)

#

linux: service-linux-amd64 ## Compile Linux binary

darwin: service-darwin-amd64 ## Compile Darwin binary

service-linux-amd64 service-darwin-amd64 service-$(GOOS)-$(GOARCH): $(SOURCES)
	CGO_ENABLED=0 go build -ldflags="-s -w -X $(PACKAGE)" -a -installsuffix cgo -o $@

container: ## Create Docker image from Linux binary.
	$(MAKE) linux GOOS=linux GOARCH=amd64
	docker build \
		--build-arg BUILD_BRANCH=$(BRANCH) \
		--build-arg BUILD_TIME=$(shell date +%Y%m%dt%H%M%S) \
		--build-arg BUILD_URL=$(PROJECT_URL)/commit/$(COMMIT_ID) \
		--build-arg BUILD_USER=$(USER) \
		--build-arg BUILD_VERSION=$(VERSION) \
    --tag $(APP_NAME)/$(BRANCH):latest .
	docker tag $(APP_NAME)/$(BRANCH):latest $(APP_NAME)/$(BRANCH):$(VERSION)

up: DAEMON=
up down:
	docker-compose $@ $(DAEMON)


clean:
	rm -f gateway cmd/client/client

help: ## This help.
	@echo $(APP_NAME)
	@echo $(PROJECT_URL)
	@echo Targets:
	@awk 'BEGIN {FS = ":.*?## "} / [#][#] / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)


report: ## Show variables
	@echo STARTD='"$(STARTD)"'
	@echo THIS='"$(THIS)"'
	@echo THISD='"$(THISD)"'
	@echo DEPENDS='"$(DEPENDS)"'
	@echo VERSION='"$(VERSION)"'
