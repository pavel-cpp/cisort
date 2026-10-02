package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

// Resolver finds the configuration that applies to a directory: the nearest
// FileName in the directory or its parents, or a fallback. It is safe for
// concurrent use and reads every directory at most once.
type Resolver struct {
	fallback *Config
	adjust   func(*Config)

	mu   sync.Mutex
	dirs map[string]resolved
}

type resolved struct {
	cfg *Config
	err error
}

// NewResolver returns a Resolver that uses fallback where no config file
// exists. adjust, if not nil, is applied to every loaded configuration and to
// fallback.
func NewResolver(fallback *Config, adjust func(*Config)) *Resolver {
	if adjust == nil {
		adjust = func(*Config) {}
	}
	adjust(fallback)
	return &Resolver{fallback: fallback, adjust: adjust, dirs: make(map[string]resolved)}
}

// For returns the configuration for files in dir.
func (r *Resolver) For(dir string) (*Config, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	res := r.lookup(dir)
	return res.cfg, res.err
}

func (r *Resolver) lookup(dir string) resolved {
	if res, ok := r.dirs[dir]; ok {
		return res
	}
	var res resolved
	file := filepath.Join(dir, FileName)
	switch _, err := os.Stat(file); {
	case err == nil:
		res.cfg, res.err = Load(file)
		if res.err == nil {
			r.adjust(res.cfg)
		}
	case errors.Is(err, fs.ErrNotExist):
		if parent := filepath.Dir(dir); parent != dir {
			res = r.lookup(parent)
		} else {
			res.cfg = r.fallback
		}
	default:
		res.err = err
	}
	r.dirs[dir] = res
	return res
}
