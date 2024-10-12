package debuglib

import (
	"encoding/json"
	"fmt"
	"runtime"
)

func PrintJSON(val interface{}) {
	res, err := json.MarshalIndent(val, "> ", "  ")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(res))
	}
}

func FileLine() (s string) {
	_, fileName, fileLine, ok := runtime.Caller(1)
	if ok {
		s = fmt.Sprintf("%s:%d", fileName, fileLine)
	}
	return
}

func FileLineError(message string, args ...interface{}) error {
	message = fmt.Sprintf(message, args...)
	_, fileName, fileLine, ok := runtime.Caller(1)
	if ok {
		return fmt.Errorf("[%s:%d] %s", fileName, fileLine, message)
	}
	return fmt.Errorf(message, args...)
}

func FileLinePrintf(message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)
	_, fileName, fileLine, ok := runtime.Caller(1)
	if ok {
		fmt.Printf("[%s:%d] %s", fileName, fileLine, message)
	}
	fmt.Printf(message, args...)
}
