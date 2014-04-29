all: test
#rpc.pb.go:
#	protoc --go_out=. ../src/protobuf/rpc.proto -I ../src/protobuf/
test:  libct.go rpc.pb.go
	go test -v rpc.pb.go libct.go libct_test.go
