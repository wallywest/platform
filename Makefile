.PHONY: build test test-unit

current_dir := $(patsubst %/,%, $(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
REPO_PATH := gitlab.vailsys.com/vail-cloud-services/api
VERSION=$(shell cat $(current_dir)/version/VERSION)
REV := $(shell git rev-parse --short HEAD 2> /dev/null  || echo 'unknown')
BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2> /dev/null  || echo 'unknown')
BUILD_DATE := $(shell date +%Y%m%d-%H:%M:%S)
BUILDFLAGS := -ldflags \
			 " -X $(REPO_PATH)/version.Version=$(VERSION)\
			   -X $(REPO_PATH)/version.Revision=$(REV)\
			   -X $(REPO_PATH)/version.Branch=$(BRANCH)\
			   -X $(REPO_PATH)/version.BuildDate=$(BUILD_DATE)"

GO15VENDOREXPERIMENT=1
SRCS = $(shell git ls-files '*.go')


DOCKERFILE := Dockerfile
DOCKER_FLAGS := docker run --rm -i
DOCKER_IMAGE := platform-dev:$(REV)
DOCKER_RUN_DOCKER := $(DOCKER_FLAGS) "$(DOCKER_IMAGE)"

build:
	docker build ${DOCKER_BUILD_ARGS} -t "$(DOCKER_IMAGE)" -f "$(DOCKERFILE)" .

test:
	ginkgo -r --skipPackage vendor

test-ci: build
	$(DOCKER_RUN_DOCKER) ginkgo -r -skipMeasurements -noColor --skipPackage vendor
