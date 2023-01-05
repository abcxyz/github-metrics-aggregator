# build generates the server go binary
build:
	@go build \
		-a \
		-trimpath \
		-ldflags "-s -w -extldflags='-static'" \
		-o ./bin/server \
		./cmd/github-metrics-aggregator
.PHONY: build

# protoc generates the protos
protoc:
	@go generate ./protos
.PHONY: protoc
