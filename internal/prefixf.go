package internal

import "fmt"

// PrefixedF wraps a formatting function and returns a formatting function that
// prefixes the format string passed to it with one of the provided prefix
// strings.  The returned function is stateful such that:
// - the first call to it will use the first prefix
// - the second call will use the second prefix,
// - and so forth, with the last prefixing being used for all subsequent calls;
//   in other wrds there is no wrap-around or re-use of prefixes, the final
//   prefix "sticks"
//
// For example:
//     myf = PrefixedF(log.Printf, "hello", "world", "meh...)
//     myf("such")
//     myf("much")
//     myf("amaze")
//     myf("still amazing")
//
// Would output:
//     hello such
//     world much
//     meh... amaze
//     meh... still amazing
func PrefixedF(
	logf func(string, ...interface{}),
	prefixes ...string,
) func(string, ...interface{}) {
	return (&prefixf{
		logf:     logf,
		prefixes: prefixes,
	}).outf
}

// ElidedF creates a simple PrefixedF where the first prefix is given,
// and all the rest are "...    " with enough spaces to align with the fist
// prefix.
//
// For example:
//     myf := ElidedF(log.Printf, "something")
//     myf("one")
//     myf("two")
//     myf("three")
//
// Would output:
//     something one
//     ...       two
//     ...       three
func ElidedF(
	logf func(string, ...interface{}),
	prefix string,
) func(string, ...interface{}) {
	return PrefixedF(logf, prefix,
		fmt.Sprintf("% -*s", len(prefix), "..."))
}

type prefixf struct {
	logf     func(string, ...interface{})
	prefixes []string
	i        int
}

func (pf *prefixf) outf(format string, args ...interface{}) {
	pfx := pf.prefixes[pf.i]
	if pf.i < len(pf.prefixes)-1 {
		pf.i++
	}
	format = fmt.Sprintf("%s %s", pfx, format)
	pf.logf(format, args...)
}
