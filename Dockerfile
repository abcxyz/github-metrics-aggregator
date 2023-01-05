# Copyright 2023 The Authors (see AUTHORS file)
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM --platform=$BUILDPLATFORM golang:1.19 AS builder

ENV PORT=8080
ENV CGO_ENABLED=0
ENV GOPROXY=https://proxy.golang.org,direct

WORKDIR /go/src/app
COPY . .

RUN go build \
    -a \
    -trimpath \
    -ldflags "-s -w -extldflags='-static'" \
    -o /go/bin/server \
    ./cmd/github-metrics-aggregator

RUN strip -s /go/bin/server

RUN echo "nobody:*:65534:65534:nobody:/:/bin/false" > /tmp/etc-passwd

# Use a scratch image to host our binary.
FROM scratch
COPY --from=builder /tmp/etc-passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/server /server

USER nobody

EXPOSE 8080
ENTRYPOINT ["/server"]
