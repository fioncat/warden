package config

type Env map[string]string

type Job struct {
	Watch *Watch
	Build []Exec
	Exec  *Exec
}

type Watch struct {
	Ignore  []string
	Pattern []string
}

type Exec struct {
	Env    Env
	Script string
}
