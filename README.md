generate go code

```shell
protoc -I. --go_out=. --go-grpc_out=. proto/*.proto
```

test

```
gotests -template_dir template -all client/main.go
```

ssl

```
brew install mkcert
mkcert -install
mkcert -CAROOT
mkcert localhost
```
