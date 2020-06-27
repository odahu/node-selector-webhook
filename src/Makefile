BUILD_TAG=latest
TAG=
# Example of DOCKER_REGISTRY: nexus.domain.com:443/
DOCKER_REGISTRY=

-include .env
.EXPORT_ALL_VARIABLES:


## lint: Run linters
lint:
	golangci-lint run


## test: Run tests
test:
	gotestsum --junitfile operator-report.xml -- -coverprofile=coverage.txt ./...


## build: Build binary
build:
	go build -o node-selector-webhook ./main.go

## docker-build: Build docker image
docker-build:
	docker build -t odahu/node-selector-webhook:${BUILD_TAG} -f ../containers/Dockerfile .
	docker build . -f ../containers/Dockerfile

## docker-push: Push image to docker
docker-push:
	docker tag odahu/node-selector-webhook:${BUILD_TAG} ${DOCKER_REGISTRY}/odahu/node-selector-webhook:${TAG}
	docker push ${DOCKER_REGISTRY}/odahu/node-selector-webhook:${TAG}