package router

type Mapper interface {
	Get(key string) *Profile
}

type mapper map[string]*Profile

func NewMapper() Mapper {
	m := mapper(make(map[string]*Profile))
	return &m
}

func (m mapper) Get(key string) *Profile {
	return m[key]
}
