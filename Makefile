PACKAGES=.

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
lint: generate
	find $(PACKAGES) -type f -name '*.go' | xargs golint
	go vet $(PACKAGES)

.PHONY: test
test: lint
	go test -v -bench . -run . $(PACKAGES)
