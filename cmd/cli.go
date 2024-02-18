package main

import (
	"aggressive-pokes/internal/runnables"
	"aggressive-pokes/internal/runner"
	"os"
	"time"
)

func main() {
	httpRunnable := runnables.HttpRunnable("http://localhost:8081/pixel", readPayload("fixtures/payload.json"))
	r := runner.NewRunner(httpRunnable)
	r.AddStage(100, 3*time.Second)
	r.AddStage(300, 3*time.Second)
	r.AddStage(500, 3*time.Second)
	r.Run()
}

func readPayload(path string) []byte {
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return content
}
