# set the go project name
BINARY_NAME=fs

# build the go binary
build:
	@go build -o bin/${BINARY_NAME}

# run the go binary
run:
	@bin/${BINARY_NAME}

# build and run the go binary
run-build: build run

# clean up the build artifacts
clean:
	rm -f bin/${BINARY_NAME}