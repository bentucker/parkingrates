PROTOFILES=$(wildcard proto/*.proto)
GOPROTOFILES=$(patsubst %.proto,gen%.pb.go,$(PROTOFILES))

default: build

$(GOPROTOFILES): $(PROTOFILES)
	@mkdir -p genproto docs
	protoc \
		-I proto \
		-I vendor/github.com/gogo/googleapis \
		-I vendor \
		--doc_out=docs --doc_opt=html,index.html \
		--grpc-gateway_out=logtostderr=true:genproto \
		--gogo_out=plugins=grpc,\
Mgoogle/protobuf/types/timestamp.proto=github.com/gogo/protobuf/types,\
Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api,\
Mgoogle/rpc/status.proto=github.com/gogo/googleapis/google/rpc:genproto \
proto/rates.proto

build: $(GOPROTOFILES)
	go build .

clean:
	@rm -rf genproto docs parkingrates

test:
	go test ./...

deps:
	@go get -u github.com/golang/dep/cmd/dep
	@go get -u github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc
	@go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	@go get -u github.com/gogo/protobuf/protoc-gen-gogo
	@dep ensure


.PHONY: build test clean