package core

type Storage interface {
	Put(b *Block) error
}

type MemoryStorage struct{}

func NewMemorystorage() *MemoryStorage {
	return &MemoryStorage{}
}

func (s *MemoryStorage) Put(b *Block) error {
	return nil
}
