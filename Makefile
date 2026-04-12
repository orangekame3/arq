.PHONY: build test lint fmt vet clean ci fmt-check

build:
	go build -o arq .

test:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

coverage: test
	go tool cover -func=coverage.txt
	go tool cover -html=coverage.txt -o coverage.html

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -f arq coverage.txt coverage.html

ci: fmt vet test lint fmt-check

fmt-check:
	@test -z "$$(go fmt ./...)" || (echo "Files need formatting"; exit 1)
