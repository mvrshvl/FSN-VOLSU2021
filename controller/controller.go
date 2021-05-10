package controller

import (
	"context"
	"fsn/config"
	"fsn/logger"
	"fsn/notifier"
	"reflect"
	"time"
)

type Controller struct {
	watchers map[string]*notifier.Watcher
}

func NewController(cfg *config.Config) *Controller {
	return &Controller{
		watchers: cfg.Watchers,
	}
}

func (ctrl *Controller) Run(ctx context.Context, timeout time.Duration, cfg *config.Config) {
	for _, watcher := range ctrl.watchers {
		if !watcher.IsStop() {
			go watcher.Process(ctx, timeout, cfg)
		}
	}
}

func (ctrl *Controller) Reload(ctx context.Context, newCfg *config.Config) {
	for path, watcher := range ctrl.watchers {
		newWatcher, ok := newCfg.Watchers[path]
		if !ok || (newWatcher.IsStop() && !watcher.IsStop()) {
			watcher.Stop()
			logger.Infof(ctx, "%s stopping...", path)

			continue
		}
	}

	for path, watcher := range newCfg.Watchers {
		oldWatcher, ok := ctrl.watchers[path]
		if !ok || (oldWatcher.IsStop() && !watcher.IsStop()) {
			go watcher.Process(ctx, newCfg.Timeout, newCfg)
			logger.Infof(ctx, "%s - added", path)
		} else if ok && !reflect.DeepEqual(oldWatcher, watcher) {
			oldWatcher.Stop()
			go watcher.Process(ctx, newCfg.Timeout, newCfg)
		}
	}

	ctrl.watchers = newCfg.Watchers
}
