package cmdev

import (
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/fsnotify.v1"
)

const modTimeNotFound int64 = -1

func newWatcher(paths []string, task *watchTask) *fsnotify.Watcher {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Errorf("fail to create watcher: %s", err)
		os.Exit(2)
	}

	go handleEvents(watcher, task)

	for _, p := range paths {
		if err := watcher.Add(p); err != nil {
			logger.Errorf("fail to watch path %s : %s", p, err)
			os.Exit(2)
		}
		logger.Infof("tracing %s", p)
	}

	return watcher
}

func handleEvents(watcher *fsnotify.Watcher, task *watchTask) {
	for {
		select {
		case event := <-watcher.Events:
			if !task.shouldWatch(event.Name) {
				continue
			}

			// focus on create event
			if checkEventOp(event, fsnotify.Rename) {
				task.modifiedTimes[event.Name] = -1
				logger.Debugf("ignore rename event: %s", event)
				continue
			}

			if checkEventOp(event, fsnotify.Remove) {
				task.modifiedTimes[event.Name] = modTimeNotFound
			} else {
				mt := getFileModTime(event.Name)

				// fail to get mod time
				if mt == modTimeNotFound {
					continue
				}

				// skip event
				if last := task.modifiedTimes[event.Name]; last == mt {
					logger.Debugf("skip event: %s", event)
					continue
				}

				logger.Infof("event: %s", event)
				task.modifiedTimes[event.Name] = mt

				// wait before autobuild util there is no file change.
				go func(task *watchTask) {
					task.scheduleTime = time.Now().Add(task.wait)

					for {
						time.Sleep(task.scheduleTime.Sub(time.Now()))
						if time.Now().After(task.scheduleTime) {
							break
						}
						return
					}

					// build
					if success := task.autoBuildAll(); !success {
						return
					}

					// run
					task.restart()

				}(task)

			}

		case err := <-watcher.Errors:
			logger.Warnf("watcher error: %s ", err)
		}
	}
}

func checkEventOp(event fsnotify.Event, op fsnotify.Op) bool {
	return event.Op&op != 0
}

// getFileModTime retuens unix timestamp of `os.File.ModTime` by given path.
func getFileModTime(path string) int64 {
	f, err := os.Open(path)
	if err != nil {
		logger.Warn(err.Error())
		return modTimeNotFound
	}

	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		logger.Warn(err.Error())
		return modTimeNotFound
	}

	return fi.ModTime().Unix()
}

func walkDir(dir string, paths *[]string, task *watchTask) {
	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}

	useDir := false
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() == true && fileInfo.Name()[0] != '.' {
			walkDir(dir+"/"+fileInfo.Name(), paths, task)
			continue
		}

		if useDir == true {
			continue
		}

		if !fileInfo.IsDir() && task.shouldWatch(fileInfo.Name()) {
			*paths = append(*paths, dir)
			useDir = true
		}
	}

	return
}
