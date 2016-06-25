package opcode

// RegisterAllocator assigns labels to registers with an LRU re-use strategy.
type RegisterAllocator struct {
	max    byte
	assign map[string]byte
	order  []string
}

// NewRegisterAllocator creates a new register allocator for a given number of
// registers.
func NewRegisterAllocator(n byte) *RegisterAllocator {
	ra := &RegisterAllocator{}
	ra.Init(n)
	return ra
}

// Init allocates capacity for a machine with n-many registers.
func (ra *RegisterAllocator) Init(n byte) {
	ra.max = n
	ra.assign = make(map[string]byte, n)
	ra.order = make([]string, 0, n)
}

// Copy copies a register allocator.
func (ra *RegisterAllocator) Copy() {
	order := make([]string, len(ra.order), ra.max)
	copy(order, ra.order)
	ra.order = order

	assign := make(map[string]byte, ra.max)
	for c, b := range ra.assign {
		assign[c] = b
	}
	ra.assign = assign
}

// LabelOf returns the label assigned to a register
func (ra *RegisterAllocator) LabelOf(b byte) string {
	for k, r := range ra.assign {
		if r == b {
			return k
		}
	}
	return ""
}

// Assigned returns true if the register is assigned.
func (ra *RegisterAllocator) Assigned(b byte) bool {
	for _, r := range ra.assign {
		if b == r {
			return true
		}
	}
	return false
}

// Get returns the register assigned to a name, and a bool indicating if it is
// assigned.  If it wasn't assigned, the returned register will be 0.  If it
// was assigned, its place in the LRU ordering is freshened.
func (ra *RegisterAllocator) Get(key string) (r byte, defined bool) {
	r, defined = ra.assign[key]
	if defined {
		ra.order = moveback(key, ra.order)
	}
	return
}

// Take returns the register assigned to a label, possibly after having
// assigned a free register, or stolen the oldest one.
func (ra *RegisterAllocator) Take(key string) byte {
	if r, defined := ra.Get(key); defined {
		return r
	}
	if len(ra.assign) >= int(ra.max) {
		return ra.steal(key)
	}
	for b := byte(1); b <= ra.max; b++ {
		if !ra.Assigned(b) {
			ra.assign[key] = b
			ra.order = append(ra.order, key)
			return b
		}
	}
	panic("should be possible to find a free register when all haven't been used yet")
}

// Reassign freshesns and relabels a register; if oldKey isn't assigned,
// Reassign just calls Take.
func (ra *RegisterAllocator) Reassign(oldKey, newKey string) byte {
	// TODO: inline free, and add a fast path for re-assigning the latest
	// register.
	r := ra.Free(oldKey)
	if r == 0 {
		return ra.Take(newKey)
	}
	ra.assign[newKey] = r
	ra.order = append(ra.order, newKey)
	return r
}

// Free deletes any assignment for the given key.
func (ra *RegisterAllocator) Free(key string) byte {
	r, defined := ra.assign[key]
	if !defined {
		return 0
	}
	delete(ra.assign, key)
	ra.order = without(key, ra.order)
	return r
}

func (ra *RegisterAllocator) steal(newKey string) byte {
	oldKey := ra.order[0]
	r := ra.assign[oldKey]
	if r == 0 {
		panic("unassigned register in RegisterAllocator.order")
	}
	delete(ra.assign, oldKey)
	if len(ra.order) > 1 {
		copy(ra.order, ra.order[1:])
	}
	ra.assign[newKey] = r
	ra.order[len(ra.order)-1] = newKey
	return r
}

func moveback(s string, ss []string) []string {
	i := 0
	for ; i < len(ss); i++ {
		if ss[i] == s {
			goto elide
		}
	}
	return append(ss, s)
elide:
	if j := len(ss) - 1; i < j {
		copy(ss[i:], ss[i+1:])
		ss[j] = s
	}
	return ss
}

func without(s string, ss []string) []string {
	i := 0
	for ; i < len(ss); i++ {
		if ss[i] == s {
			break
		}
	}
	j := i
	i++
	for ; i < len(ss); i++ {
		ss[j] = ss[i]
		j++
	}
	return ss[:j]
}
