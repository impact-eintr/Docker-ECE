package common

import (
	"fmt"
	"runtime"

	log "github.com/sirupsen/logrus"
)

var instance = log.New()

func LogAndErrorf(format string, args interface{}) error {
	instance.Formatter = &log.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
			return "", ""
		},
	}
	instance.Errorf(format, args)
	return fmt.Errorf("%v", args)
}

func LogAndErrorsf(format string, args ...interface{}) error {
	instance.Formatter = &log.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
			return "", ""
		},
	}
	instance.Errorf(format, args...)
	return fmt.Errorf("%v", args...)
}
