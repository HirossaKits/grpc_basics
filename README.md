generate go code

```shell
protoc -I. --go_out=. --go-grpc_out=. proto/*.proto
```
