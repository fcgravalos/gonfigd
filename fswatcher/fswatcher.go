package fswatcher

import (
	"context"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fcgravalos/gonfigd/kv"
	"github.com/fcgravalos/gonfigd/pubsub"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
)

var excludedFileExtensions map[string]struct{}

func init() {
	excludedFileExtensions = make(map[string]struct{})
	for _, ext := range []string{".swp", ".swx", ".~", ".tmp"} {
		excludedFileExtensions[ext] = struct{}{}
	}
}

type fsWatcher struct {
	watcher  *fsnotify.Watcher
	registry *registry
	kv       kv.KV
	ps       pubsub.PubSub
	log      zerolog.Logger
}

type registry struct {
	sync.Mutex
	r map[string]struct{}
}

func (r *registry) register(path string) {
	r.Lock()
	r.r[path] = struct{}{}
	r.Unlock()
}

func (r *registry) isRegistered(path string) bool {
	_, ok := r.r[path]
	return ok
}

func (r *registry) unregister(path string) {
	r.Lock()
	delete(r.r, path)
	r.Unlock()
}

func isValidFileName(name string) bool {
	// Discard hidden files
	if match, _ := filepath.Match("\\.*", name); match {
		return false
	}
	if _, ok := excludedFileExtensions[filepath.Ext(name)]; ok {
		return false
	}
	return true
}

func isValidFile(name string) bool {
	fi, err := os.Stat(name)
	if err != nil {
		return false
	}
	return fi.Mode().IsRegular() && isValidFileName(filepath.Base(name))
}

func (fsw *fsWatcher) upsertFileOnDb(path string) (bool, error) {
	changed := false
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return changed, err
	}

	v, err := fsw.kv.Get(path)
	if (v != nil && v.MD5() != fmt.Sprintf("%x", md5.Sum(data))) || err != nil {
		newVal, _ := kv.NewValue(data)
		err = fsw.kv.Put(path, newVal)
		if err != nil {
			return changed, err
		}
		changed = true
	}
	return changed, nil
}

func (fsw *fsWatcher) walk(path string, fi os.FileInfo, err error) error {
	if fi.Mode().IsDir() && !fsw.registry.isRegistered(path) {
		abs, _ := filepath.Abs(path)
		fsw.registry.register(path)
		fsw.log.Info().Msgf("adding watcher for %s", abs)
		return fsw.watcher.Add(path)

		// If it's a regular file, we don't have to set a fsnotify watch but we check if it's stored in db
	} else if isValidFile(path) {
		if _, err := fsw.kv.Get(path); err != nil {
			fsw.log.Warn().Msgf("missing file %s in kv, inserting and creating event", path)
			return fsw.createEventHandler(path)
		}
	}
	return nil
}

func (fsw *fsWatcher) publishEvent(config string, evType pubsub.EventType) error {
	ev := pubsub.NewEvent(evType, config)
	if !fsw.ps.TopicExists(config) {
		err := fsw.ps.CreateTopic(config)
		if err != nil {
			return err
		}
	}
	return fsw.ps.Publish(config, ev)
}

func (fsw *fsWatcher) createOrWriteEventHandler(name string, evType pubsub.EventType) error {
	changed, err := fsw.upsertFileOnDb(name)
	if err != nil {
		return err
	}
	if changed {
		return fsw.publishEvent(name, evType)
	}
	return nil
}

func (fsw *fsWatcher) createEventHandler(name string) error {
	return fsw.createOrWriteEventHandler(name, pubsub.ConfigCreated)
}

func (fsw *fsWatcher) writeEventHandler(name string) error {
	return fsw.createOrWriteEventHandler(name, pubsub.ConfigUpdated)
}

func (fsw *fsWatcher) removeEventHandler(name string) error {
	if fsw.registry.isRegistered(name) {
		if err := fsw.watcher.Remove(name); err != nil {
			return err
		}
		fsw.registry.unregister(name)
	}
	err := fsw.kv.Delete(name)
	if err != nil {
		return err
	}
	return fsw.publishEvent(name, pubsub.ConfigDeleted)
}

func (fsw *fsWatcher) routeEvent(ev fsnotify.Event) {
	evOp := ev.Op.String()
	var err error
	switch evOp {
	case "CREATE":
		if isValidFile(ev.Name) {
			err = fsw.createEventHandler(ev.Name)
		}
		break
	case "WRITE":
		if isValidFile(ev.Name) {
			err = fsw.writeEventHandler(ev.Name)
		}
		break
	case "REMOVE":
		err = fsw.removeEventHandler(ev.Name)
		break
	default:
		err = nil
		break
	}

	if err != nil {
		fsw.log.Error().Msgf("error while handling %s event for %s: %v", evOp, ev.Name, err)
	}
}

// Start creates a new fsWatcher
// It will return an error if it's not able to create a *fsnotify.Watcer
func Start(ctx context.Context, root string, fwalkInterval time.Duration, kv kv.KV, ps pubsub.PubSub, logger zerolog.Logger) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error().Msgf("failed to create new fsnotify watcher: %v", err)
		return err
	}
	defer watcher.Close()
	registry := &registry{
		r: map[string]struct{}{},
	}
	fsw := &fsWatcher{watcher: watcher, registry: registry, kv: kv, ps: ps, log: logger}

	stopCh := make(chan struct{}, 1)
	go func(ctx context.Context, root string, fsw *fsWatcher, stopCh chan struct{}) {
		for {
			select {
			case <-time.After(fwalkInterval):
				fsw.log.Debug().Msgf("walking %s directory", root)
				filepath.Walk(root, fsw.walk)
				break
			case <-ctx.Done():
				stopCh <- struct{}{}
				fsw.log.Info().Msgf("fswatcher walker goroutine stopped")
				return
			}
		}
	}(ctx, root, fsw, stopCh)

	for {
		select {
		case ev := <-fsw.watcher.Events:
			go fsw.routeEvent(ev)
		case err := <-fsw.watcher.Errors:
			fsw.log.Error().Msgf("error watching for filesystem changes: %v\n", err)
		case <-ctx.Done():
			<-stopCh
			fsw.log.Info().Msgf("fswatcher stopped")
			return nil
		}
	}

}
