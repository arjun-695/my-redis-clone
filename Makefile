run: build
	@./bin/redis-clone
	
build:
	@go build -o bin/redis-clone .
test:
	@go test -v ./...
