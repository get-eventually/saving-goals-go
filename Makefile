.PHONY: build-saving-goals-api
build-saving-goals-api:
	CGO_ENABLED=0 go build ./cmd/saving-goals-api

.PHONY: build-saving-goals-consumers
build-saving-goals-consumers:
	CGO_ENABLED=0 go build ./cmd/saving-goals-consumers

.PHONY: build-saving-goals-cli
build-saving-goals-cli:
	CGO_ENABLED=0 go build ./cmd/saving-goals-cli

.PHONY: generate-protobuf
generate-protobuf:
	protoc --go_out=. ./resources/messages/*.proto
