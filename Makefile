build:
	@go build -o bin/tcp-fs

run: build
	@./bin/tcp-fs

test:
	@go test ./... -v