package cmdev

import (
	"os"
	"os/exec"
)

func NewNormalCommand(name string, args, envs []string) exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), envs...)
	return *cmd
}
