package cog

import (
	"fmt"
	"time"
)

type LogType string

const (
	DEBUG LogType = "DEBUG"
	INFO  LogType = "INFO"
	ERROR LogType = "ERROR"
)

var logDebug = false

func SetDebug(debug bool) {
	logDebug = debug
}

func Print(logType LogType, log string) {

	if logType == DEBUG && !logDebug {
		return
	}

	fmt.Println(fmt.Sprintf("[%s]	%s	%s", time.Now().Format("2006-01-02 15:04:05"), logType, log))
}

func PrintPacket(logType LogType, direction int, opCode int16, data []byte) {

	if logType == DEBUG && !logDebug {
		return
	}

	directionTag := "S > C"
	if direction == 1 {
		directionTag = "C > S"
	}

	fmt.Println(fmt.Sprintf("[%s]	%s	%s	%d	\n%v", time.Now().Format("2006-01-02 15:04:05"), logType, directionTag, opCode, data))
}
