.PHONY: generate-protobuf
generate-protobuf:
	protoc --go_out=. ./resources/messages/*.proto
