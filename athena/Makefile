.PHONY: build
build:
	../build/go-build.sh . athena

.PHONY: run
run:
	@go get -d -v
	@go build -o app .
	./app
