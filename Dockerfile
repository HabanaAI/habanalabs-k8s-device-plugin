# Copyright (c) 2019, HabanaLabs Ltd.  All rights reserved.
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

FROM ubuntu:18.04 as build

RUN apt update && apt install -y --no-install-recommends \
            ca-certificates \
            g++ \
            wget && \
    rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION 1.13
RUN wget -nv -O - https://dl.google.com/go/go${GOLANG_VERSION}.linux-amd64.tar.gz \
    | tar -C /usr/local -xz

ENV GOPATH /opt/habanalabs/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

WORKDIR /opt/habanalabs/go/src/habanalabs-device-plugin
COPY . .

RUN export CGO_LDFLAGS_ALLOW='-Wl,--unresolved-symbols=ignore-in-object-files' && \
    go install -ldflags="-w -s" -v habanalabs-device-plugin

FROM debian:stretch-slim

RUN apt update && apt install -y --no-install-recommends \
            pciutils && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /opt/habanalabs/go/bin/habanalabs-device-plugin /usr/bin/habanalabs-device-plugin

CMD ["habanalabs-device-plugin"]
