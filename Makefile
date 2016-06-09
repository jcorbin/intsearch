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
	for pkg in; do find $$pkg -maxdepth 1 -type f -name '*.go' -not -name '*_string.go' | xargs golint; done
	go vet $(PACKAGES)

.PHONY: test
test: lint
	go test -v -bench . -run . $(PACKAGES)

%.prof: intsearch.test.%
	[ -d $@ ] || mkdir $@
	$< -test.benchtime=3s -test.bench=Plan_topDown -test.benchmem -test.cpuprofile=$@/cpu -test.memprofile=$@/mem
