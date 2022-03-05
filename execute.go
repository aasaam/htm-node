package main

import (
	"os/exec"
	"strings"
)

func execute(command string, arg ...string) ([]byte, error) {
	out, err := exec.Command(command, arg...).Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

func executeString(command string, arg ...string) (string, error) {
	out, err := exec.Command(command, arg...).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func executeMany(arg ...string) error {
	cmd := exec.Command("/bin/sh", "-c", strings.Join(arg, " && "))
	return cmd.Run()
}
