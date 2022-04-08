package main

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
)

func execute(command string, arg ...string) ([]byte, error) {
	cmd := exec.Command(command, arg...)
	var stdOut bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return nil, errors.New(err.Error() + ": " + stderr.String())
	}
	return stdOut.Bytes(), nil
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
	var stdOut bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return errors.New(err.Error() + ": " + stderr.String())
	}
	return nil
}
