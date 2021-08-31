package common

import (
	"os/exec"
)

func Exec(cmd string, args ...string) ([]byte, error) {
	return exec.Command(cmd, args...).CombinedOutput()
}

func Exec2(cmd string, args ...string) (string, error) {
	b, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(b), nil
}
