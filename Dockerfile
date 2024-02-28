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

ARG BASE_IMAGE
FROM ${BASE_IMAGE} as builder

ENV GOLANG_VERSION 1.21.5
RUN wget -nv -O - https://dl.google.com/go/go${GOLANG_VERSION}.linux-amd64.tar.gz \
    | tar -C /usr/local -xz


ENV GOPATH /opt/habanalabs/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

# go-hlml must be download before building the image, since it is hosted in gerrit,
# and gerrit doesn't support go modules in our version. Then it is copied as
# a sibling to the device plugin folder
WORKDIR /opt/habanalabs/go/src/go-hlml

WORKDIR /opt/habanalabs/go/src/habanalabs-device-plugin
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go mod tidy

RUN go build -buildvcs=false -o bin/habanalabs-device-plugin .


ARG BASE_IMAGE
ARG BUILD_DATE
ARG BUILD_REF
FROM ${BASE_IMAGE}

RUN apt update && apt install -y --no-install-recommends \
	pciutils && \
	rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/lib/habanalabs /usr/lib/habanalabs
COPY --from=builder /usr/include/habanalabs /usr/include/habanalabs

RUN echo "/usr/lib/habanalabs/" >> /etc/ld.so.conf.d/habanalabs.conf
RUN ldconfig

COPY --from=builder /opt/habanalabs/go/src/habanalabs-device-plugin/bin/habanalabs-device-plugin /usr/bin/habanalabs-device-plugin
CMD ["habanalabs-device-plugin"]


LABEL   image.created="${BUILD_DATE}" \
        image.revision="${BUILD_REF}" \
        image.title="habana-device-plugin" \
		image.author="Habana Labs Ltd"
