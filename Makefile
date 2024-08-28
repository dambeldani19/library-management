GO_WORKSPACE := ..

protoc:
	protoc  --proto_path=protos protos/*.proto --go_out=.. --go-grpc_out=..
	
docker-up:
	@echo up docker
	@docker build --tag library-managemen .
	@docker-compose up

docker-down:
	@echo down docker
	@docker-compose down
	@docker rmi library-managemen

docker-push:
	@echo docker push
	@docker tag library-management:latest dambeldani19/library-managemen:v1
	@docker push dambeldani19/library-managemen:v1

