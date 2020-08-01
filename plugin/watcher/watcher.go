package watcher

import (
	"bytes"
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/blake2b"

	"github.com/kernelschmelze/pkg/atom"
	"github.com/kernelschmelze/pkg/path"

	"github.com/fsnotify/fsnotify"
)

type cbChanged func(file string)

type Watcher struct {
	Watcher     *fsnotify.Watcher
	running     atom.Bool
	notify      map[string][]cbChanged
	notifyGuard sync.RWMutex
	wg          *sync.WaitGroup
}

var (
	watcher *Watcher
)

func init() {
	watcher = NewWatcher()
}

func NewWatcher() *Watcher {

	w := &Watcher{
		wg:     &sync.WaitGroup{},
		notify: make(map[string][]cbChanged),
	}

	w.Watcher, _ = fsnotify.NewWatcher()

	return w
}

func GetWatcher() *Watcher {

	if watcher == nil {
		watcher = NewWatcher()
	}

	return watcher
}

func Add(file string, fn cbChanged) error {

	watcher := GetWatcher()
	err := watcher.Add(file, fn)
	return err
}

func Remove(file string) {
	watcher := GetWatcher()
	watcher.Remove(file)
}

func Close() {

	watcher := GetWatcher()
	watcher.Stop()
}

func (w *Watcher) Add(file string, fn cbChanged) error {

	var err error

	// normalize file name
	if file, err = utils.ExpandPath(file); err != nil {
		return err
	}

	// add file to watcher
	if err = w.Watcher.Add(file); err != nil {

		// file does not exist
		if !utils.Exists(file) {

			// add folder to watcher
			folder := utils.GetFolder(file)
			if utils.Exists(folder) {
				err = w.Watcher.Add(folder)
			}

		}

		if err != nil {
			return err
		}

	}

	// register callback function to call if file has been changed
	if fn != nil {

		w.notifyGuard.Lock()

		notify, exist := w.notify[file]
		if !exist {
			notify = []cbChanged{}
		}
		notify = append(notify, fn)
		w.notify[file] = notify

		w.notifyGuard.Unlock()

	}

	if !w.running.IsSet() {
		w.Start()

	}

	return nil
}

func (w *Watcher) Remove(file string) {
	w.Watcher.Remove(file)
}

func (w *Watcher) Start() {

	if w.Watcher == nil || w.running.IsSet() {
		return
	}

	w.wg.Add(1)
	w.running.Set(true)

	go func() {

		defer func() {
			w.running.Set(false)
			w.wg.Done()
		}()

		watcher := w.Watcher
		hash := make(map[string][]byte)

		for {

			select {

			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				w.notifyGuard.RLock()
				dispatch, exist := w.notify[event.Name]
				w.notifyGuard.RUnlock()

				notify := event.Op&fsnotify.Write == fsnotify.Write

				if (event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename) && exist {

					if utils.Exists(event.Name) {

						// vim handling: rename, chmod, remove, create

						if err := w.Watcher.Add(event.Name); err == nil {
							notify = true
						}

					} else {

						// add folder to watcher to catch the file again

						folder := utils.GetFolder(event.Name)
						w.Watcher.Add(folder)

					}

				}

				if !notify || !exist {
					continue
				}

				// compare file hash to prevent multiple trigger

				if hasher, err := blake2b.New256(nil); err == nil {
					if f, err := os.OpenFile(event.Name, os.O_RDONLY, 0); err == nil {
						if _, err = io.Copy(hasher, f); err == nil {
							crc := hasher.Sum(nil)
							if oldHash, exist := hash[event.Name]; exist && bytes.Equal(crc, oldHash) {
								f.Close()
								continue
							}
							hash[event.Name] = crc
						}
						f.Close()
					}
				}

				// call registered callback function

				for i := range dispatch {
					cb := dispatch[i]
					if cb != nil {
						cb(event.Name)
					}
				}

			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}

				// prevent high cpu usage on endless loop
				time.Sleep(250 * time.Millisecond)
			}

		}
	}()
}

func (w *Watcher) Stop() {

	if w.Watcher == nil || !w.running.IsSet() {
		return
	}

	w.Watcher.Close()
	w.wg.Wait()

}
