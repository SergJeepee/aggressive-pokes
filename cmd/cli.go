package main

import (
	"aggressive-pokes/internal/ltlogger"
	"aggressive-pokes/internal/runnables"
	"aggressive-pokes/internal/runner"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"time"
)

func main() {
	logger := ltlogger.New(true, "LT Runner", slog.LevelDebug)
	defer recoverLogPanic(logger)

	hs := runnables.HttpRunnableWithSupplier(runnables.NewHttpRequestSupplier(
		logger,
		http.MethodPost,
		"https://localhost:8081/api/search",
		readPayload("internal/fixtures/lvrpl-lo.json"),
		nil,
		nil))

	lt := runner.NewLoadTest()
	lt.AddQpsStage(15, 10*time.Minute, hs)
	lt.AddQpsStage(22, 10*time.Minute, hs)
	lt.AddQpsStage(30, 10*time.Minute, hs)
	lt.Start()
}

func recoverLogPanic(logger ltlogger.Logger) {
	if p := recover(); p != nil {
		logger.Error("Panic!",
			"panicMessage", p,
			"stack", debug.Stack(),
		)
		panic(p)
	}
}

func readPayload(path string) []byte {
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return content
}
