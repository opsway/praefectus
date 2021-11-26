## help: Show current help
help: Makefile
	@echo "Choose a command run:"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'

## format: Format and simplify all source files
format:
	@printf "Format source files... " && gofmt -s -w . && echo "Done!"

## lint: Check source code by linters
lint:
	@printf "Checking via golangci-lint... " && golangci-lint run ./... && echo "Done!"

## build: Build app binary
build:
	@mkdir -p ./bin
	@echo "Start building production image..." \
    		&& export RELEASE_COMMIT=$(shell git rev-parse --short HEAD)  \
    		&& docker build --tag=praefectus_tmp \
    			--build-arg="RELEASE_BUILD_TIME=$${RELEASE_BUILD_TIME}" \
				--build-arg="RELEASE_VERSION=$${RELEASE_VERSION}" \
				--build-arg="RELEASE_COMMIT=$${RELEASE_COMMIT}" . \
    		&& echo "Done!"
	@export CONTAINER_ID=$$(docker create praefectus_tmp) \
		&& docker cp $${CONTAINER_ID}:/praefectus ./bin/ \
		&& docker rm -f -v $${CONTAINER_ID}
