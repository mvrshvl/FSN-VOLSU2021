package notifier

import (
	"bytes"
	"context"
	"fsn"
	"fsn/logger"
	"os"
)

const (
	modified state = iota
	removed

	bash = "bash"
)

type state uint8

type EventController struct {
	Path     string `json:"path"`
	OnCreate string `json:"on_create"`
	OnDelete string `json:"on_delete"`
	OnModify string `json:"on_modify"`

	State state `json:"state"`
}

func NewEventController(info os.FileInfo, path string, onCreate, onDelete, onModify string) EventController {
	st := removed
	if info != nil {
		st = modified
	}

	return EventController{
		OnCreate: onCreate,
		OnDelete: onDelete,
		OnModify: onModify,
		State:    st,
		Path:     path,
	}
}

func (ctrl *EventController) Process(ctx context.Context, newState state) {
	currentState := ctrl.State
	ctrl.State = newState

	var (
		out   *bytes.Buffer
		err   error
		event string
	)

	switch true {
	case currentState == modified && newState == removed:
		out, err = fsn.StartCommand(bash, ctrl.OnDelete)
		event = "ON_DELETE"
	case currentState == modified && newState == modified:
		out, err = fsn.StartCommand(bash, ctrl.OnModify)
		event = "ON_MODIFY"
	case currentState == removed && newState == modified:
		out, err = fsn.StartCommand(bash, ctrl.OnCreate)
		event = "ON_CREATE"
	default:
		return
	}

	if err != nil {
		logger.Errorf(ctx, "path:%s, event: %s, script execute error: %v, out: %s", ctrl.Path, event, err, out)
	} else {
		logger.Infof(ctx, "path:%s, event: %s, successfully execute, out: %s", ctrl.Path, event, out)
	}
}
