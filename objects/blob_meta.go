package objects

func (m Meta) Get(key string) string {
	return m.Meta[key]
}

func (m Meta) Set(key, value string) {
	m.Meta[key] = value
}

func (m Meta) Del(key string) {
	delete(m.Meta, key)
}
