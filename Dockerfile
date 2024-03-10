# Copyright (c) 2022, HabanaLabs Ltd.  All rights reserved.
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

ARG VERSION=1.14.0
ARG MINOR_VERSION=493
ARG DIST=ubuntu22.04
ARG REGISTRY=vault.habana.ai

FROM ${REGISTRY}/gaudi-docker/${VERSION}/${DIST}/habanalabs/pytorch-installer-2.1.1:${VERSION}-${MINOR_VERSION} as builder

RUN apt-get update && \
    apt-get install -y wget make git gcc \
    && \
    rm -rf /var/lib/apt/lists/*

ARG GOLANG_VERSION=1.21.5
RUN set -eux; \
    \
    arch="$(uname -m)"; \
    case "${arch##*-}" in \
        x86_64 | amd64) ARCH='amd64' ;; \
        ppc64el | ppc64le) ARCH='ppc64le' ;; \
        aarch64) ARCH='arm64' ;; \
        *) echo "unsupported architecture" ; exit 1 ;; \
    esac; \
    wget -nv -O - https://storage.googleapis.com/golang/go${GOLANG_VERSION}.linux-${ARCH}.tar.gz \
    | tar -C /usr/local -xz


ENV GOPATH /opt/habanalabs/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

WORKDIR /opt/habanalabs/go/src/habanalabs-device-plugin

COPY . .
RUN go mod tidy

RUN go build -buildvcs=false -o bin/habanalabs-device-plugin .


ARG BUILD_DATE
ARG BUILD_REF

FROM ${REGISTRY}/gaudi-docker/${VERSION}/${DIST}/habanalabs/pytorch-installer-2.1.1:${VERSION}-${MINOR_VERSION}

# Remove Habana libs(compat etc) in favor of libs installed by the NVIDIA driver
RUN apt-get --purge -y autoremove habana*

RUN apt update && apt install -y --no-install-recommends \
	pciutils && \
	rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/lib/habanalabs /usr/lib/habanalabs
COPY --from=builder /usr/include/habanalabs /usr/include/habanalabs
COPY --from=builder /opt/habanalabs/go/src/habanalabs-device-plugin/bin/habanalabs-device-plugin /usr/bin/habanalabs-device-plugin

RUN echo "/usr/lib/habanalabs/" >> /etc/ld.so.conf.d/habanalabs.conf
RUN ldconfig

LABEL   io.k8s.display-name="HABANA Device Plugin" \
        vendor="HABANA" \
        version=${VERSION} \
        image.git-commit="${GIT_COMMIT}" \
        image.created="${BUILD_DATE}" \
        image.revision="${BUILD_REF}" \
        summary="HABANA device plugin for Kubernetes" \
		description="See summary"

CMD ["habanalabs-device-plugin"]
