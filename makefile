# ex. make test params="-race"
test:
	go test ./... ${params}

# make cover
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	rm coverage.out

genpb:
	protoc --go_out=. --go-grpc_out=. api/pubsub/pubsub.proto 

build:
	go build -o ./bin/app ./cmd/app/main.go 

run: build
	./bin/app