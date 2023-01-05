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
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
	@protoc \
		--proto_path=protos \
		--go_out=paths=source_relative:protos \
		pubsub_schemas/event.proto
.PHONY: protoc
