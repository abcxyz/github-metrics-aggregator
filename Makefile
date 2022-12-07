# build generates the server go binary
build:
	@go build \
		-a \
		-trimpath \
		-ldflags "-s -w -extldflags='-static'" \
		-o ./bin/server \
		./cmd/**/*.go
.PHONY: build
