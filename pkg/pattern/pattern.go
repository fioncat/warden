package pattern

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const recursiveSuffix = "..."

type Pattern struct {
	Dir  string
	File string

	Recursive bool
}

func Parse(s string) (*Pattern, error) {
	base := filepath.Base(s)
	dir := filepath.Dir(s)

	p := new(Pattern)
	if strings.HasPrefix(dir, recursiveSuffix) {
		p.Recursive = true
		p.Dir = strings.TrimSuffix(dir, recursiveSuffix)
	} else {
		p.Dir = dir
	}

	var err error
	// We use absolute path to have a better log output.
	p.Dir, err = filepath.Abs(p.Dir)
	if err != nil {
		return nil, fmt.Errorf("get abs path failed: %v", err)
	}

	// Validate if the Dir is a available directory.
	stat, err := os.Stat(p.Dir)
	if err != nil {
		return nil, fmt.Errorf("parse pattern %s: %v", s, err)
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("parse pattern %s: the %q is not "+
			"a directory", s, p.Dir)
	}

	// The base is a glob pattern, we need to validate its
	// correctness before using.
	_, err = filepath.Match(base, "")
	if err != nil {
		return nil, fmt.Errorf("parse pattern %s: the %q is bad "+
			"format: %v", s, base, err)
	}
	p.File = base

	return p, nil
}

func (p *Pattern) MatchName(filename string) bool {
	matched, _ := filepath.Match(p.File, filename)
	return matched
}

func (p *Pattern) MatchDir(dir string) bool {
	if p.Recursive {
		return strings.HasPrefix(dir, p.Dir)
	}
	return dir == p.Dir
}

func (p *Pattern) Equal(o *Pattern) bool {
	return p.Dir == o.Dir && p.File == o.File
}
