1. generate protobuf

protoc --go_out=. --go_opt=paths=source_relative     --go-grpc_out=. --go-grpc_opt=paths=source_relative     remote-build/remote-build.proto

2. run server

go run server/main.go

3. run client

go run client/main.go --name="what ever name you want"

4. run worker

go run worker/main.go --name="what ever name you want"

