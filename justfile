lint:
	golangci-lint run ./...

build:
	go build -o bin/cmdSnipperVault ./cmd/cmdSnipperVault

test:
	go test -race -coverprofile=coverage.out ./...

coverage:
	go tool cover -html=coverage.out

install:
	go install .cmd/cmdSnipperVault

clean:
	rm -rf ./bin coverage.out
