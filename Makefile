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
