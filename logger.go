package cmdev

import (
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
)

var logger *log.Entry

func init() {
	logger = log.NewEntry(&log.Logger{
		Handler: text.New(os.Stderr),
		Level:   log.DebugLevel,
	})
}

func SetLogger(l *log.Entry) {
	if l != nil {
		logger = l
	}
}
