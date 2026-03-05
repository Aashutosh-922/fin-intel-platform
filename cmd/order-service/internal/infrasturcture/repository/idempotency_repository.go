package repository

type MemoryRepo struct {
	store map[string]bool
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{
		store: make(map[string]bool),
	}
}

func (r *MemoryRepo) Exists(orderID string) bool {
	return r.store[orderID]
}

func (r *MemoryRepo) Save(orderID string) {
	r.store[orderID] = true
}