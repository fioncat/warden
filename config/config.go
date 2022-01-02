package config

import "errors"

type Env map[string]string

var defaultIgnores = []string{
	".git",
}

type Job struct {
	Delay string
	Watch *Watch
	Build []*Exec
	Exec  *Exec
	Env   Env
}

func (job *Job) Normalize() error {
	if job.Watch == nil {
		return errors.New("config: watch can't be empty")
	}

	for _, defaultIgnore := range defaultIgnores {
		var found bool
		for _, ignore := range job.Watch.Ignore {
			if ignore == defaultIgnore {
				found = true
				break
			}
		}
		if !found {
			job.Watch.Ignore = append(job.Watch.Ignore, defaultIgnore)
		}
	}
	return nil
}

type Watch struct {
	Ignore  []string
	Pattern []string
}

type Exec struct {
	Env    Env
	Script string
}
