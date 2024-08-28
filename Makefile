GO_WORKSPACE := ..

protoc:
	protoc  --proto_path=protos protos/*.proto --go_out=.. --go-grpc_out=..
	
docker-up:
	@echo up docker
	@docker build --tag library-management .
	@docker-compose up

docker-down:
	@echo down docker
	@docker-compose down
	@docker rmi chat-service