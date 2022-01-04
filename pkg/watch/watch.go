package watch

import (
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/fioncat/warden/config"
	"github.com/fioncat/warden/pkg/debug"
	"github.com/fioncat/warden/pkg/pattern"
	"github.com/fsnotify/fsnotify"
)

type Event struct {
	Name string
}

type Watcher struct {
	patterns []*pattern.Pattern

	patternMap map[string][]*pattern.Pattern
	watchSet   map[string]struct{}

	watcher *fsnotify.Watcher

	ignore pattern.Ignore

	notify chan *Event

	pause uint32
}

func Run(cfg *config.Watch) (*Watcher, error) {
	w := new(Watcher)

	w.ignore = pattern.Ignore(cfg.Ignore)
	err := w.ignore.Validate()
	if err != nil {
		return nil, err
	}

	w.patternMap = make(map[string][]*pattern.Pattern)
	w.watchSet = make(map[string]struct{})

	w.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("run watcher: failed to create the watcher: %v", err)
	}

	w.patterns = make([]*pattern.Pattern, len(cfg.Pattern))
	for i, patternStr := range cfg.Pattern {
		p, err := pattern.Parse(patternStr)
		if err != nil {
			return nil, err
		}
		w.patterns[i] = p
	}

	w.notify = make(chan *Event, 500)

	for _, p := range w.patterns {
		err := w.add(p.Dir)
		if err != nil {
			return nil, err
		}
	}

	go w.watch()

	return w, nil
}

func (w *Watcher) Notify() <-chan *Event {
	return w.notify
}

func (w *Watcher) add(dir string) error {
	_, ok := w.watchSet[dir]
	if ok {
		return nil
	}
	ps := w.dirPatterns(dir)
	if len(ps) == 0 {
		return nil
	}

	err := w.watcher.Add(dir)
	if err != nil {
		return fmt.Errorf("add watch for %q: %v", dir, err)
	}
	w.watchSet[dir] = struct{}{}
	debug.Infof("Watch add: %s", dir)

	var rec bool
	for _, p := range ps {
		if p.Recursive {
			rec = true
			break
		}
	}
	if !rec {
		return nil
	}

	es, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("add watch for %q: read dir: %v", dir, err)
	}
	for _, e := range es {
		if !e.IsDir() {
			continue
		}
		subDir := filepath.Join(dir, e.Name())
		if w.ignore.OneMatch(e.Name()) {
			debug.Infof("Watch: Ignore dir: %s", subDir)
			continue
		}
		err := w.add(subDir)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Watcher) Pause() {
	atomic.StoreUint32(&w.pause, 1)
}

func (w *Watcher) Continue() {
	atomic.StoreUint32(&w.pause, 0)
}

func (w *Watcher) onChange(path string) {
	if atomic.LoadUint32(&w.pause) == 1 {
		debug.Infof("Watch: Discard change due to pause: %s", path)
		return
	}
	w.notify <- &Event{Name: path}
}

func (w *Watcher) watch() {
	debug.Info("Begin to watch...")
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				debug.Info("Closing watcher...")
				return
			}
			if event.Op&fsnotify.Chmod == fsnotify.Chmod {
				break
			}
			debug.Infof("Watch: Receive change: %v", event)
			path := event.Name
			dir := filepath.Dir(path)
			name := filepath.Base(path)

			if event.Op&fsnotify.Remove == fsnotify.Remove {
				delete(w.watchSet, path)
				if w.ignore.OneMatch(name) {
					break
				}
				for _, p := range w.dirPatterns(dir) {
					if p.MatchName(name) {
						w.onChange(path)
						break
					}
				}
				break
			}

			stat, err := os.Stat(path)
			if err != nil {
				debug.Error(err, "Failed to get stat for %q", path)
			}
			if stat.IsDir() && event.Op&fsnotify.Create == fsnotify.Create {
				var rec bool
				for _, p := range w.dirPatterns(path) {
					if p.Recursive {
						rec = true
						break
					}
				}
				if rec {
					err := w.add(path)
					if err != nil {
						debug.Error(err, "Failed to add watch")
						break
					}
				}
			}
			if !stat.IsDir() {
				if w.ignore.OneMatch(name) {
					break
				}
				var matched bool
				for _, p := range w.dirPatterns(dir) {
					if p.MatchName(name) {
						matched = true
						break
					}
				}
				if matched {
					w.onChange(path)
				}
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				debug.Info("Closing watcher...")
				return
			}
			debug.Error(err, "Received error from watcher")
		}
	}
}

func (w *Watcher) dirPatterns(dir string) []*pattern.Pattern {
	ps, ok := w.patternMap[dir]
	if ok {
		return ps
	}
	var matches []*pattern.Pattern
	for _, p := range w.patterns {
		if p.MatchDir(dir) {
			matches = append(matches, p)
		}
	}
	w.patternMap[dir] = matches
	return matches
}

func (w *Watcher) Close() error {
	err := w.watcher.Close()
	if err != nil {
		return err
	}
	close(w.notify)
	return nil
}
