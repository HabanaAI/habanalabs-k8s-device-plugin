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

# option to ovveride by user request(e.g: CD)
IMAGE ?= $(IMAGE_NAME):$(IMAGE_TAG)
BASE_IMAGE ?= ${REGISTRY}/gaudi-docker/${VERSION}/${DIST}/habanalabs/pytorch-installer-2.2.2:${VERSION}-${MINOR_VERSION}

.PHONY: build push

## build: build docker image
build:
	$(DOCKER) build \
	-t $(IMAGE) \
	--build-arg BUILD_REF=$(IMAGE_TAG) \
	--build-arg BASE_IMAGE=$(BASE_IMAGE) \
	--build-arg VERSION="$(VERSION)" \
	--build-arg GOLANG_VERSION="$(GOLANG_VERSION)" \
	--build-arg GIT_COMMIT="$(GIT_COMMIT)" \
	--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
	.

## push: push the image to the registry
push:
	$(DOCKER) image push $(IMAGE)
