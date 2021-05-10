package config

import (
	"context"
	"encoding/json"
	"fmt"
	"fsn"
	"fsn/logger"
	"fsn/notifier"
	"github.com/mcuadros/go-defaults"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"
)

type Config struct {
	Timeout  time.Duration                `default:"100ms",json:"timeout"`
	Watchers map[string]*notifier.Watcher `json:"watchers"`
	PID      int                          `json:"pid"`
	mtx      sync.Mutex
}

func NewCfg(ctx context.Context) *Config {
	cfg := new(Config)
	defaults.SetDefaults(cfg)
	cfg.Watchers = make(map[string]*notifier.Watcher)

	cfg.PID = os.Getpid()
	cfg.Save(ctx)

	return cfg
}

func LoadConfig(ctx context.Context) *Config {
	cfg := new(Config)
	defaults.SetDefaults(cfg)

	cfgJSON, err := ioutil.ReadFile(GetFilepath())
	if err != nil {
		logger.Warn(ctx, "config not found, get defaults", err)

		return NewCfg(ctx)
	}

	err = json.Unmarshal(cfgJSON, cfg)
	if err != nil {
		logger.Warn(ctx, "can't unmarshall config file, get defaults", err)

		return NewCfg(ctx)
	}

	return cfg
}

func (cfg *Config) Save(ctx context.Context) {
	f, err := os.OpenFile(GetFilepath(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)

	if err != nil {
		logger.Error(ctx, "config saving error:", err)

		return
	}

	cfgJSON, err := json.Marshal(cfg)
	if err != nil {
		logger.Error(ctx, "can't marshal config:", err)

		return
	}

	_, err = f.Write(cfgJSON)
	if err != nil {
		logger.Error(ctx, "can't write config file:", err)

		return
	}
}

func (cfg *Config) DeleteWatcher(path string) {
	cfg.mtx.Lock()
	defer cfg.mtx.Unlock()

	delete(cfg.Watchers, path)
}

func (cfg *Config) Reload(ctx context.Context, modify func()) {
	modify()
	cfg.Save(ctx)
	out, err := fsn.StartCommand("kill", "-s", "SIGHUP", fmt.Sprint(cfg.PID))
	if err != nil {
		logger.Errorf(ctx, "can't send reload signal to service: %v, out: %s", err, out.String())
	}
}

func (cfg *Config) GetWatcher(key string) (*notifier.Watcher, error) {
	w, ok := cfg.Watchers[key]
	if !ok {
		return nil, fmt.Errorf("watcher does not exist")
	}

	return w, nil
}

func GetFilepath() string {
	return path.Join(fsn.Root, "config.json")
}
