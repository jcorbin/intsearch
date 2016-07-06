BENCH_NAME=Plan_topDown
BENCH_TIME=3s

PACKAGES=. ./word ./runnable ./opcode

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
	for pkg in $(PACKAGES); do \
		find $$pkg -maxdepth 1 -type f -name '*.go' -not -name '*_string.go' -not -name '*_test.go' | xargs golint; \
		find $$pkg -maxdepth 1 -type f -name '*.go' -name '*_test.go' | xargs golint; \
	done
	go vet $(PACKAGES)

.PHONY: test
test: lint
	go test -v -bench . -run . $(PACKAGES)

.PHONY: prof_clean
prof_clean:
	rm -rf *.prof.d

.PRECIOUS: %.prof.d/test
%.prof.d/test: %.test
	[ -d $$(dirname $@) ] || mkdir $$(dirname $@)
	ln $< $@

%.prof.d/$(BENCH_NAME): %.prof.d/test
	[ -d $@ ] || mkdir $@
	$< -test.benchtime=$(BENCH_TIME) -test.bench=$(BENCH_NAME) -test.benchmem \
		-test.cpuprofile=$@/cpu \
		-test.memprofile=$@/mem \
		2>&1 | tee $@/log
