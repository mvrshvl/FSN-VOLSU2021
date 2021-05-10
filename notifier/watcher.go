package notifier

import (
	"context"
	"errors"
	"fsn/logger"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Watcher struct {
	Path            string
	Stopped         bool
	Recursive       int
	ModTime         time.Time
	EventController EventController

	config ConfigController
}

type ConfigController interface {
	Reload(ctx context.Context, modify func())
	Save(ctx context.Context)
}

func NewWatcher(ctx context.Context, recursive int, path, onCreate, onDelete, onModify string) (w *Watcher) {
	w = &Watcher{
		Recursive:       recursive,
		Path:            path,
		EventController: NewEventController(nil, path, onCreate, onDelete, onModify),
	}

	info, err := os.Lstat(path)
	if err != nil {
		logger.Warnf(ctx, "can't get stat %s: %v.", path, err)

		return
	}

	w.ModTime = info.ModTime()
	w.EventController = NewEventController(info, path, onCreate, onDelete, onModify)

	return
}

func (w *Watcher) Process(ctx context.Context, timeout time.Duration, cfg ConfigController) {
	w.config = cfg

	tick := time.NewTicker(timeout)
	defer tick.Stop()

	for range tick.C {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if w.IsStop() {
			logger.Infof(ctx, "%s - stopped", w.Path)
			return
		}

		info, err := os.Lstat(w.Path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				w.EventController.Process(ctx, removed)

				continue
			}

			logger.Warnf(ctx, "can't get stat %s: %v.", w.Path, err)
		}

		lastModify := info.ModTime()

		if w.Recursive >= -1 && info.IsDir() {
			err = filepath.Walk(w.Path, func(path string, info fs.FileInfo, err error) error {
				if (strings.Count(path[len(w.Path):], "/") < w.Recursive || w.Recursive == -1) && info.ModTime().After(lastModify) {
					lastModify = info.ModTime()
				}

				return nil
			})

			if err != nil {
				logger.Warn(ctx, err)
			}
		}

		if !lastModify.Equal(w.ModTime) {
			w.EventController.Process(ctx, modified)
			w.ModTime = lastModify
			w.config.Save(ctx)
		}
	}
}

func (w *Watcher) Stop() {
	w.Stopped = true
}

func (w *Watcher) Run() {
	w.Stopped = false

}

func (w *Watcher) IsStop() bool {
	return w.Stopped
}
