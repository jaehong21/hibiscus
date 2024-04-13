MODULE_NAME := $(shell go list -m)
PROJECT_NAME := hibiscus
PROJECT_PATH := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

TARGET_PATH := $(PROJECT_PATH)/target

VERSION := $(shell git describe --exact-match --tags HEAD 2>/dev/null || git rev-parse --abbrev-ref HEAD)
GIT_SHA := $(shell git rev-parse HEAD)
# GIT_SHA := $(shell git rev-parse --short HEAD)
DATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

GOVERSION := $(shell go version | awk '{print $$3}')
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
LDFLAGS := -ldflags="-X ${MODULE_NAME}/cmd.buildVersion=${VERSION} \
                    -X ${MODULE_NAME}/cmd.buildDate=${DATE} \
                    -X ${MODULE_NAME}/cmd.buildOS=${GOOS} \
                    -X ${MODULE_NAME}/cmd.buildArch=${GOARCH} \
										-X ${MODULE_NAME}/cmd.buildCommit=${GIT_SHA} \
										-X ${MODULE_NAME}/cmd.buildGoVersion=${GOVERSION}"

STAGE ?= development

env:
	@echo "PROJECT_PATH:\t${PROJECT_PATH}"
	@echo "PROJECT_NAME:\t${PROJECT_NAME}"
	@echo "MODULE_NAME:\t${MODULE_NAME}"
	@echo "GOVERSION:\t${GOVERSION}"
	@echo "GOOS:\t\t${GOOS}"
	@echo "GOARCH:\t\t${GOARCH}"
	@echo "STAGE:\t\t${STAGE}"
	@echo "VERSION:\t${VERSION}"

# .SILENT: build
build: 	
	@echo "Building ${PROJECT_NAME}..."
	go mod download

	GOOS=${GOOS} \
	GOARCH=${GOARCH} \
	go build ${LDFLAGS} -o '${TARGET_PATH}/${PROJECT_NAME}.${GOOS}.${GOARCH}'
	
	@echo "Build complete"

dev: build
	@echo "Running ${PROJECT_NAME}..."
	${TARGET_PATH}/${PROJECT_NAME}.${GOOS}.${GOARCH}
	
clean:
	@echo "Cleaning up..."
	rm -rf ${TARGET_PATH}/*
