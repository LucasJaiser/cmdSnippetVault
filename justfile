MIGRATE := 'migrate -path internal/storage/sqlite/migrations -database "sqlite://cmdvault.db"'

lint:
	golangci-lint run ./...

build:
	go build -o ./bin/csv ./cmd/cmdSnipperVault

test:
	go test -race -coverprofile=coverage.out ./...

coverage:
	go tool cover -html=coverage.out

install:
	go install .cmd/cmdSnipperVault

clean:
	rm -rf ./bin coverage.out

migrate-up:
	{{MIGRATE}} up

migrate-down:
	{{MIGRATE}} down
