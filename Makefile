.PHONY: help clean fmt vet test test-all check mod

help:    ## Show this help.
	@sed -ne '/@sed/!s/## //p' $(MAKEFILE_LIST)

clean:   ## Cleanup build artifacts including UI assets
	go clean
	rm -rf ./artifacts

build:
	go tool sedgbuild --target x86_64-unknown-linux-gnu --src cmd/hmi $(ARGS)

build-arm:
	go tool sedgbuild --target aarch64-unknown-linux-gnu --src cmd/hmi $(ARGS)

fmt:     ## Run "go fmt" on the entire gortia
	go fmt -C .  $(shell cd . && go list -e ./... | grep -v autogen)

vet:     ## Run "go vet" on the entire gortia
	go vet -C .  $(shell cd . && go list -e ./... | grep -v autogen)

mod:
	go mod tidy

test:    ## Run all the tests in the project except for integration-tests
	go test -C . -v -race -short $(shell cd . && go list -e ./... | grep -v autogen)

test-all:    ## Run all the tests in the project including integration-tests
	go test -C . -tags=integration -v -race $(shell cd . && go list -e ./... | grep -v autogen)

check: fmt vet test  ## Run fmt, vet and test

# For building go localy using sedg go build tool (go tool sedgbuild) follow this steps:
#  - to see the content of the sedgbuild code go to :
#    https://gitlab.solaredge.com/portialinuxdevelopers/infrastructure/cicd/devop-tools/-/blob/master/develop/golang/golang-sedg-build
#  - we install this file like this:
#    GO_TOOL_DIR=$(go env GOTOOLDIR)
#    sudo cp -f ./golang-sedg-build ${GO_TOOL_DIR}/sedgbuild
#  - you also need to copy file from :
#    https://gitlab.solaredge.com/portialinuxdevelopers/infrastructure/cicd/devop-tools/-/tree/master/develop/docker/
#    https://gitlab.solaredge.com/portialinuxdevelopers/infrastructure/cicd/devop-tools/-/tree/master/develop/services/
#     to /usr/local/bin/sedg/
#  - in order to be able to build local controller docker images you also need balena-cli tool:
#     https://github.com/balena-io/balena-cli/blob/master/INSTALL-LINUX.md
