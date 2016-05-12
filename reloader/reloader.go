package reloader

import (
	"log"

	"github.com/matryer/temple"
	"gopkg.in/fsnotify.v1"
)

// Reloader watches for changing files and calls Reload
// on a Temple.
type Reloader struct {
	t        *temple.Temple
	watcher  *fsnotify.Watcher
	stopchan chan struct{}
}

// New creates a new Reloader.
func New(t *temple.Temple) (*Reloader, error) {
	r := &Reloader{t: t}
	var err error
	if r.watcher, err = fsnotify.NewWatcher(); err != nil {
		return nil, err
	}
	go func() {
		for {
			select {
			case event := <-r.watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("temple: changed:", event.Name, "(reloading)")
					r.t.Reload()
				}
			case err := <-r.watcher.Errors:
				log.Println("temple: reloader failed:", err)
			}
		}
	}()

	for _, file := range t.Files() {
		log.Println("temple.reloader: watching", file)
		if err := r.watcher.Add(file); err != nil {
			r.watcher.Close()
			return nil, err
		}
	}

	return r, nil
}

// Close stops watching the files and cleans things up.
func (r *Reloader) Close() error {
	return r.watcher.Close()
}
