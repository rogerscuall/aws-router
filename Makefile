include .envrc

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run/awsrouters/excel: run the cmd/api application
.PHONY: run/awsrouters/excel
run/awsrouters/excel:
	go run *.go excel


# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...
	@echo 'Testing code coverage...'
	go test -cover ./...

## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor

# ==================================================================================== #
# BUILD
# ==================================================================================== #

## build/cmd: build the cmd/api application
.PHONY: build/cmd
build/cmd:
	@echo 'Building cmd/api...'
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/api ./cmd/api
	go build -ldflags='-s' -o=./bin/mac_arm64/api ./cmd/api

# ==================================================================================== #
# RELEASE
# ==================================================================================== #

## release: build and release the application
.PHONY: release
release:
	@echo 'Creating release...'
	goreleaser release --snapshot --clean

## release/init: initialize goreleaser
.PHONY: release/init
release/init:
	@echo 'Initializing the goreleases...'
	goreleaser init

# ==================================================================================== #
# CONTAINER
# ==================================================================================== #


## build/container: build the container image
.PHONY: build/container
build/container:
	@echo 'Building container image...'
	docker build -t registry.presidio.com/arista/arista-avd-cvaas/crispy-enigma:${CE_TAG} .
	docker push registry.presidio.com/arista/arista-avd-cvaas/crispy-enigma:${CE_TAG}