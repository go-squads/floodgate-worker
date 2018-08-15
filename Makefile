all:

dep:
	@echo "Installing dependencies"
	@dep ensure -v
test: dep
	@echo "Unit testing"
	@go test -v ./...
build: test
	@echo "Building Binaries"
	@go build .
run: build
	@echo "Starting Worker"
	./floodgate-worker aw
clean:
	@rm floodgate-worker
