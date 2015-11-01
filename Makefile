.PHONY: clean

intsearch: intsearch.c
	$(CC) $(CFLAGS) -o $@ $<

clean:
	rm intsearch
