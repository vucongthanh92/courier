package utils

import (
	"runtime/debug"

	"github.com/vucongthanh92/go-base-utils/logger"
	"go.uber.org/zap"
)

func SafeGo(f func()) {
	go func() {
		defer HandlePanic()
		// call the provided function
		f()
	}()
}

func HandlePanic() {
	if r := recover(); r != nil {
		logger.Error("Recovered from panic: ", zap.Any("panic", r), zap.String("stack", string(debug.Stack())))
	}
}
