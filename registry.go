package riff

// Registry represents a registry of chunk decoders.
type Registry struct {
	// Map of chunk IDs and decoder Maker functions.
	makers map[uint32]Maker

	// Pool of instantiated chunk decoders that can be reused.
	pool map[uint32][]Chunk

	// Maker for raw chunk decoder.
	raw IDMaker
}

// NewRegistry returns new instance of Registry.
func NewRegistry(raw IDMaker) *Registry {
	return &Registry{
		makers: make(map[uint32]Maker, 4),
		pool:   make(map[uint32][]Chunk, 4),
		raw:    raw,
	}
}

// Has returns true if decoder for given chunk ID is registered.
func (reg *Registry) Has(id uint32) bool {
	_, ok := reg.makers[id]
	return ok
}

// Register registers chunk decoder Maker function.
func (reg *Registry) Register(id uint32, maker Maker) {
	reg.makers[id] = maker
}

// Put chunk decoder back to the pool so it can be reused.
func (reg *Registry) Put(ch Chunk) {
	id := ch.ID()
	if _, ok := reg.pool[id]; !ok {
		reg.pool[id] = make([]Chunk, 0, 4)
	}
	reg.pool[id] = append(reg.pool[id], ch)
}

// Get returns decoder for given ID from the pool. For unknown (not registered)
// chunk ID decoders it returns raw decoder.
func (reg *Registry) Get(id uint32) Chunk {
	ch := reg.GetNoRaw(id)
	if ch == nil {
		ch = reg.raw(id)
	}
	return ch
}

// GetNoRaw returns decoder for given ID from the pool or nil.
func (reg *Registry) GetNoRaw(id uint32) Chunk {
	chs, ok := reg.pool[id]
	cl := len(chs)
	if !ok || cl == 0 {
		if dec, ok := reg.makers[id]; ok {
			return dec()
		}
		return nil
	}
	ch := reg.pool[id][cl-1]
	reg.pool[id] = reg.pool[id][:cl-1]
	return ch
}
