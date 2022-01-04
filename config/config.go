package config

import "errors"

type Env map[string]string

var defaultIgnores = []string{
	".git",
}

type Job struct {
	Delay string   `yaml:"delay"`
	Watch *Watch   `yaml:"watch"`
	Build []*Build `yaml:"build"`
	Exec  *Exec    `yaml:"exec"`
	Env   Env      `yaml:"env"`
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
	Ignore  []string `yaml:"ignore"`
	Pattern []string `yaml:"pattern"`
}

type Build struct {
	Exec `yaml:",inline"`
}

type Exec struct {
	Env    Env    `yaml:"env"`
	Script string `yaml:"script"`
}
