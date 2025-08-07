package store

type Registry struct {
	Key      string
	Secret   string
	Filename string
}

func newRegistry(secret, filename string) *Registry {
	return &Registry{
		Key:      secret,
		Secret:   secret,
		Filename: filename,
	}
}
