.PHONY: clean log

intsearch: intsearch.c
	$(CC) $(CFLAGS) -o $@ $<

intsearch_trace: intsearch.c
	$(CC) -DPRINT_TRACE -DPRINT_PLAN $(CFLAGS) -o $@ $<

clean:
	rm intsearch

log: intsearch_trace
	./intsearch_trace send more money | tee $@
