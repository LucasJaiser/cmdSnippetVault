MIGRATE := 'migrate -path internal/storage/sqlite/migrations -database "sqlite://cmdvault.db"'
VERSION := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`
COMMIT  := `git rev-parse --short HEAD 2>/dev/null || echo "none"`
DATE    := `date -u +%Y-%m-%dT%H:%M:%SZ`
LDFLAGS := "-X main.version=" + VERSION + " -X main.commit=" + COMMIT + " -X main.date=" + DATE

lint:
	golangci-lint run ./...

build:
	go build -ldflags "{{LDFLAGS}}" -o ./bin/csv ./cmd/cmdSnippetVault

test:
	go test -race -coverprofile=coverage.out ./...

coverage:
	go tool cover -html=coverage.out

install:
	go install -ldflags "{{LDFLAGS}}" ./cmd/cmdSnippetVault

clean:
	rm -rf ./bin coverage.out

migrate-up:
	{{MIGRATE}} up

migrate-down:
	{{MIGRATE}} down
