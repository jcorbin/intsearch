.PHONY: generate
generate:
	go generate

.PHONY: build
build: generate
	go build

.PHONY: clean
clean:
	git clean -f -X

.PHONY: lint
lint:
	golint
	go vet

.PHONY: test
test: lint generate
	go test
