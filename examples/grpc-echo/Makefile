PROTO_FILE = proto/echo/echo.proto
pb:
	protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative $(PROTO_FILE)
	@echo "done"
