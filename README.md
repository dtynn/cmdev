#### cmdev

##### install

`go get -v github.com/dtynn/cmdev`

##### demo

```
package main

import (
	"os/exec"

	"github.com/dtynn/cmdev"
)

func main() {
	cmd := cmdev.NewNormalCommand("echo", []string{"hello", ",", "cmdev"}, nil)
	build1 := cmdev.NewNormalCommand("echo", []string{"build", "1"}, nil)
	build2 := cmdev.NewNormalCommand("echo", []string{"build", "2"}, nil)
	build3 := cmdev.NewNormalCommand("echo", []string{"build", "3"}, nil)
	build4 := cmdev.NewNormalCommand("echo", []string{"$CMDEV_TEST"}, []string{"CMDEV_TEST=cmd_dev_test"})

	task := cmdev.NewWatchTask(cmd,
		[]exec.Cmd{build1, build2, build3, build4},
		[]string{"./"},
		[]string{"a.txt", "a.md"},
		[]string{".json", ".txt", ".md"},
		0)

	task.Watch()
}

```