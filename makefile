# ex. make test params="-race"
test:
	go test ./... ${params}

# make cover
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	rm coverage.out