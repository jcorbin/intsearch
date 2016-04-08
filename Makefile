.PHONY: clean log

intsearch: intsearch.c
	$(CC) $(CFLAGS) -O3 -DMEASURE_TIME -o $@ $<

intsearch_trace: intsearch.c
	$(CC) -DMEASURE_TIME -DPRINT_TRACE -DPRINT_PLAN -DMEASURE_TIME $(CFLAGS) -o $@ $<

clean:
	rm intsearch intsearch_trace

log: intsearch_trace
	rm $@
	./intsearch_trace send more money | tee $@
