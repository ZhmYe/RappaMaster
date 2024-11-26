cd pb
protoc --go_out=. --go-grpc_out=. --proto_path=. service.proto
cd ..
