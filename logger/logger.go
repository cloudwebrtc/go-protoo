package logger

import (
	log "github.com/pion/ion-log"
)

func init() {
	fixByFile := []string{"asm_amd64.s", "proc.go"}
	fixByFunc := []string{}
	log.Init("debug", fixByFile, fixByFunc)
}

func Infof(format string, v ...interface{}) {
	log.Infof(format, v...)
}

func Debugf(format string, v ...interface{}) {
	log.Debugf(format, v...)
}

func Warnf(format string, v ...interface{}) {
	log.Warnf(format, v...)
}

func Errorf(format string, v ...interface{}) {
	log.Errorf(format, v...)
}
