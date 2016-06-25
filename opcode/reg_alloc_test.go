package opcode_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jcorbin/intsearch/opcode"
)

func TestRegisterAllocator(t *testing.T) {
	ra := opcode.NewRegisterAllocator(5)

	a := ra.Take("a")
	assert.True(t, ra.Assigned(a))
	assert.Equal(t, "a", ra.LabelOf(a))

	b := ra.Take("b")
	assert.True(t, ra.Assigned(b))
	assert.Equal(t, "b", ra.LabelOf(b))

	assert.NotEqual(t, a, b)

	assert.Equal(t, ra.Free("a"), a)

	c := ra.Take("c")
	assert.Equal(t, a, c)

	d := ra.Take("d")
	assert.NotEqual(t, b, d)
	assert.NotEqual(t, c, d)

	e := ra.Take("e")
	assert.NotEqual(t, b, e)
	assert.NotEqual(t, c, e)
	assert.NotEqual(t, d, e)

	f := ra.Take("f")
	assert.NotEqual(t, b, f)
	assert.NotEqual(t, c, f)
	assert.NotEqual(t, d, f)
	assert.NotEqual(t, e, f)

	g := ra.Take("g")
	assert.Equal(t, b, g)

	ra.Reassign("g", "h")
	assert.Equal(t, "h", ra.LabelOf(g))
}
