image ?= "artifactory-kfs.habana-labs.com/k8s-docker-dev/device_plugin/habana-device-plugin"
version ?= "test"
base_image ?= "artifactory-kfs.habana-labs.com/docker-local/1.13.0/ubuntu20.04/habanalabs/base-installer:1.13.0-10"


## build: build docker image in ci-cd process
.PHONY: build
build:
	docker build \
	-t $(image):$(version) \
	--build-arg BASE_IMAGE=$(base_image) \
	--build-arg BUILD_REF=$(version) \
	--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
	.

## push-image: push the image to the registry
.PHONY: push-image
push-image:
	docker image push $(image):$(version)