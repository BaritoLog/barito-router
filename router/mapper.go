package router

type Mapper interface {
	Get(key string) *Item
}

type mapper map[string]*Item

func NewMapper() Mapper {
	m := mapper(make(map[string]*Item))
	return &m
}

func (m mapper) Get(key string) *Item {
	return m[key]
}
