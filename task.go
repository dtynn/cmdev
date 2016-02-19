package cmdev

import (
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

var defaultWaitTime = 3 * time.Second

func NewWatchTask(command exec.Cmd, builds []exec.Cmd, dirs, ignores, watches []string, wait time.Duration) *watchTask {
	if wait <= 0 {
		wait = defaultWaitTime
	}

	return &watchTask{
		dirs:    dirs,
		ignores: ignores,
		watches: watches,

		wait:    wait,
		command: command,
		builds:  builds,

		modifiedTimes: map[string]int64{},

		exit: make(chan bool, 1),
	}
}

type watchTask struct {
	dirs    []string
	ignores []string
	watches []string

	wait    time.Duration
	command exec.Cmd
	builds  []exec.Cmd

	cmd           *exec.Cmd
	modifiedTimes map[string]int64
	scheduleTime  time.Time
	buildMutex    sync.Mutex

	exit chan bool
}

func (this *watchTask) Watch() {
	paths := make([]string, 0)

	for _, dir := range this.dirs {
		walkDir(dir, &paths, this)
	}

	watcher := newWatcher(paths, this)
	defer watcher.Close()

	this.autoBuildAll()
	this.restart()

	<-this.exit
}

func (this *watchTask) Stop() {
	this.exit <- true
}

// build
func (this *watchTask) autoBuildAll() bool {
	this.buildMutex.Lock()
	defer this.buildMutex.Unlock()

	for _, one := range this.builds {
		if err := one.Run(); err != nil {
			logger.Warnf("fail to build: %s", err)
			return false
		}
	}

	return true
}

// command
func (this *watchTask) kill() {
	if this.cmd != nil && this.cmd.Process != nil {
		if err := this.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			logger.Warnf("fail to kill: %s", err)
		}
	}

	return
}

func (this *watchTask) start() {
	cmd := &exec.Cmd{}
	*cmd = this.command

	go func(cmd *exec.Cmd) {
		if err := cmd.Run(); err != nil {
			logger.Warnf("fail to start: %s", err)
		}
	}(cmd)

	this.cmd = cmd
}

func (this *watchTask) restart() {
	this.kill()
	this.start()
}

// watch
func (this *watchTask) shouldWatch(name string) bool {
	abs, _ := filepath.Abs(name)
	if abs == "" {
		return false
	}

	for _, ignore := range this.ignores {
		if strings.HasSuffix(abs, ignore) {
			return false
		}
	}

	for _, watch := range this.watches {
		if strings.HasSuffix(abs, watch) {
			return true
		}
	}

	return false
}
