package fsn

import (
	"bytes"
	"os/exec"
)

const Root = "/etc/fsnotifier"

func StartCommand(name string, arg ...string) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	cmd := exec.Command(name, arg...)
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	if err != nil || cmd.Process == nil {
		return &buf, err
	}

	return &buf, nil
}
