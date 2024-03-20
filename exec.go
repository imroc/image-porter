package main

import "os/exec"

func RunCommand(command string, arg ...string) (output []byte, err error) {
	cmd := exec.Command(command, arg...)
	output, err = cmd.CombinedOutput()
	return
}
