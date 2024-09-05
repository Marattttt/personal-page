#! /bin/sh

protoc --go_out=./internal/handlers/grpcgen/ --go_opt=paths=source_relative \
	--go-grpc_out=./internal/handlers/grpcgen/ --go-grpc_opt=paths=source_relative \
	--proto_path=../proto ../proto/gorunner.proto
