# Copyright (c) 2020-2022, HabanaLabs Ltd.  All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

DOCKER ?= docker

include $(CURDIR)/versions.mk


ifeq ($(IMAGE_NAME),)
IMAGE_NAME := $(REGISTRY)/$(APP_NAME)
endif

IMAGE_TAG ?= $(VERSION)-$(MINOR_VERSION)
IMAGE = $(IMAGE_NAME):$(IMAGE_TAG)


.PHONY: build push

## build: build docker image
build: 
	$(DOCKER) build \
	-t $(IMAGE) \
	--build-arg BUILD_REF=$(IMAGE_TAG) \
	--build-arg REGISTRY=$(REGISTRY) \
	--build-arg VERSION="$(VERSION)" \
	--build-arg MINOR_VERSION="$(MINOR_VERSION)" \
	--build-arg DIST="$(DIST)" \
	--build-arg GOLANG_VERSION="$(GOLANG_VERSION)" \
	--build-arg GIT_COMMIT="$(GIT_COMMIT)" \
	--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
	.

## push: push the image to the registry
push:
	$(DOCKER) image push $(IMAGE)
