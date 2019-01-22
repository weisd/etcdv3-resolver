GOOS=linux go build -mod=vendor -o bin/server example/helloworld/greeter_server/main.go
GOOS=linux go build -mod=vendor -o bin/client example/helloworld/greeter_client/main.go
docker rmi resolver_test
docker build -t resolver_test .