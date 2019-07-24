# Copyright 2019 The MayaData Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

PWD := ${CURDIR}

PACKAGE_NAME = github.com/AmitKumarDas/kgetset
PACKAGE_VERSION ?= $(shell git describe --always --tags)
OS = $(shell uname)

ALL_SRC = $(shell find . -name "*.go" | grep -v -e vendor \
	-e ".*/\..*" \
	-e ".*/_.*" \
	-e ".*/mocks.*" \
	-e ".*/*.pb.go")
ALL_PKGS = $(shell go list $(sort $(dir $(ALL_SRC))) | grep -v vendor)
ALL_PKG_PATHS = $(shell go list -f '{{.Dir}}' ./...)

# External tools required while building this binary or 
# to test source code, artifacts in this project
EXT_TOOLS =\
	github.com/golangci/golangci-lint/cmd/golangci-lint \
	github.com/axw/gocov/gocov \
	github.com/AlekSi/gocov-xml \
	github.com/matm/gocov-html

REGISTRY ?= quay.io/amitkumardas
IMG_NAME ?= kgetset

BUILD_LDFLAGS = -X $(PACKAGE_NAME)/util/build.Hash=$(PACKAGE_VERSION)
GO_FLAGS = -gcflags '-N -l' -ldflags "$(BUILD_LDFLAGS)"

### linux based binary
lbins: $(IMG_NAME).linux

$(IMG_NAME).linux: $(ALL_SRC)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on \
		go build -tags bins $(GO_FLAGS) -o $@ *.go

$(ALL_SRC): ;

### make vendored copy of dependencies
.PHONY: vendor
vendor: go.mod go.sum
	@export GO111MODULES=on go mod vendor

### download modules to local cache
.PHONY: vendor-cache
vendor-cache: go.mod go.sum
	@export GO111MODULES=on go mod download

.PHONY: ext-tools
ext-tools: $(EXT_TOOLS)

### install go based tools
.PHONY: $(EXT_TOOLS)
$(EXT_TOOLS):
	@echo "Installing external tool $@"
	@go get -u $@

### Target to build the docker image
.PHONY: image publish
image:
	docker build -t $(REGISTRY)/$(IMG_NAME):$(PACKAGE_VERSION) -f Dockerfile .
	docker tag $(REGISTRY)/$(IMG_NAME):$(PACKAGE_VERSION) $(IMG_NAME):$(PACKAGE_VERSION)

publish: image
	docker push $(REGISTRY)/$(IMG_NAME):$(PACKAGE_VERSION)

### Targets to lint and test the codebase
.PHONY: test gofmt lint unit-test
test: unit-test

unit-test: 
	@go test

gofmt:
	@go fmt $(ALL_PKG_PATHS)

lint: gofmt
	@echo "---------------------"
	@echo "Running golangci-lint"
	@echo "---------------------"
	@golangci-lint run --disable-all \
		--deadline 5m \
		--enable=misspell \
		--enable=structcheck \
		--enable=golint \
		--enable=deadcode \
		--enable=errcheck \
		--enable=varcheck \
		--enable=goconst \
		--enable=unparam \
		--enable=ineffassign \
		--enable=nakedret \
		--enable=interfacer \
		--enable=misspell \
		--enable=gocyclo \
		--enable=lll \
		--enable=dupl \
		--enable=goimports \
		$(ALL_PKG_PATHS)
