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

FROM golang:1.15 as builder

ENV GOPATH /opt/habanalabs/go

WORKDIR /opt/habanalabs/go/src/habanalabs-device-plugin

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY vendor/github.com/HabanaAI/gohlml .
COPY habanalabs.go .
COPY hlml.go .
COPY main.go .
COPY server.go .
COPY watcher.go .

RUN go install

FROM debian:stretch-slim

RUN apt update && apt install -y --no-install-recommends \
            pciutils && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /opt/habanalabs/go/bin/habanalabs-k8s-device-plugin /usr/bin/habanalabs-device-plugin

CMD ["habanalabs-device-plugin"]
