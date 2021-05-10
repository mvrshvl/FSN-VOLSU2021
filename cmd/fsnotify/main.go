package main

import (
	"context"
	"flag"
	"fmt"
	"fsn"
	"fsn/config"
	"fsn/logger"
	"fsn/notifier"
	"os"
)

func main() {
	add := flag.String("add", "", "Add Path tracking. Example: fsn -add /some/path.")
	recursive := flag.Int("r", -2, "Recursive directory traversal. Set -1 for max depth. Example: fsn -add /some/path -r 2.")
	onCreate := flag.String("on_create", "", "Path to the script to execute when the create event is triggered. fsn -add /some/path -on_create /path/script.sh.")
	onModify := flag.String("on_modify", "", "Path to the script to execute when the modify event is triggered. fsn -add /some/path -on_modify /path/script.sh.")
	onDelete := flag.String("on_delete", "", "Path to the script to execute when the delete event is triggered. fsn -add /some/path -on_delete /path/script.sh.")

	stop := flag.String("stop", "", "Stop tracking the path. Tracking will be stopped until the fsn run /some/path command is called. Example: fsn -stop /some/path.")
	run := flag.String("run", "", "Resumes tracking the path. The path will be tracked until the following commands are called: -delete, - stop. Example: fsn -run /some/path.")
	delete := flag.String("delete", "", "Removes path tracking. Example: fsn -delete /some/path.")

	service := flag.String("service", "", "Provides some commands: exit - stop service executing, reset - set default config, info - get config app.")

	flag.Parse()

	ctx := logger.NewLogger(context.Background(), os.Stderr)
	cfg := config.LoadConfig(ctx)

	var modify func()

	switch true {
	case len(*service) > 0:
		ServiceCmd(ctx, cfg, *service)

		return
	case len(*add) > 0:
		modify = func() {
			cfg.Watchers[*add] = notifier.NewWatcher(ctx, *recursive, *add, *onCreate, *onDelete, *onModify)
		}
	case len(*stop) > 0:
		w, err := cfg.GetWatcher(*stop)
		if err != nil {
			logger.Errorf(ctx, "running %s failed: %v", *run, err)

			return
		}

		if w.IsStop() {
			return
		}

		modify = func() {
			w.Stop()
		}
	case len(*run) > 0:
		w, err := cfg.GetWatcher(*run)
		if err != nil {
			logger.Errorf(ctx, "running %s failed: %v %v", *run, err, cfg.Watchers)

			return
		}

		if !w.IsStop() {
			return
		}

		modify = func() {
			w.Run()
		}
	case len(*delete) > 0:
		w, err := cfg.GetWatcher(*delete)
		if err != nil {
			logger.Errorf(ctx, "removing %s failed: %v", *run, err)

			return
		}

		if w == nil {
			return
		}

		modify = func() {
			w.Stop()
			cfg.DeleteWatcher(w.Path)
		}
	default:
		logger.Error(ctx, "incorrect command")

		return
	}

	cfg.Reload(ctx, modify)
}

func ServiceCmd(ctx context.Context, cfg *config.Config, cmd string) {
	switch cmd {
	case "exit":
		out, err := fsn.StartCommand("kill", "-s", "SIGTERM", fmt.Sprint(cfg.PID))
		if err != nil {
			logger.Errorf(ctx, "can't send exit signal to service: %v, out: %s", err, out.String())
		}
	case "reset":
		cfg.Reload(ctx, func() {
			cfg = config.NewCfg(ctx)
		})
	case "info":
		list := ""
		for _, watcher := range cfg.Watchers {
			list += fmt.Sprintf("PATH - %v\n\tLast mod: %s\n\tStopped: %v\n\tLast event: %v\n\tRecursive: %v\n\tScript on_create: %s\n\tScript on_modify: %s\n\tScript on_delete: %s\n", watcher.Path, watcher.ModTime.String(), watcher.Stopped, watcher.EventController.State, watcher.Recursive, watcher.EventController.OnCreate, watcher.EventController.OnModify, watcher.EventController.OnDelete)
		}

		logger.Info(ctx, fmt.Sprintf("\nPID: %v\nTIMEOUT: %v\nLIST: \n%s\n(events: 0 - modified, 1 - removed)\n", cfg.PID, cfg.Timeout, list))

	default:
		logger.Errorf(ctx, "unknow command")
	}
}
