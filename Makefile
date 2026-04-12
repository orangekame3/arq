.PHONY: build test lint fmt vet clean ci fmt-check

build:
	go build -o arq .

test:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

coverage: test
	go tool cover -func=coverage.txt
	go tool cover -html=coverage.txt -o coverage.html

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...
	gofumpt -w .

vet:
	go vet ./...

clean:
	rm -f arq coverage.txt coverage.html

ci: fmt vet test lint fmt-check

fmt-check:
	@test -z "$$(gofumpt -l .)" || (echo "Files need formatting:"; gofumpt -l .; exit 1)
