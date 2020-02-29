package store

var _ Store = &store{}

type Store interface {
	GetClustersByNamespace(string) error
}

type store struct{}

func NewStore() *store {
	return &store{}
}

func (s *store) GetClustersByNamespace(namespace string) error {
	return nil
}
