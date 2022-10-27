pb: 
	protoc -I. --go_out=. --go-grpc_out=. proto/*.proto
.PHONY: pb