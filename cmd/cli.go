package main

import (
	"aggressive-pokes/internal/runnables"
	"aggressive-pokes/internal/runner"
	"os"
	"time"
)

func main() {
	httpRunnable := runnables.HttpRunnable("http://localhost:8081/pixel", readPayload("internal/fixtures/payload.json"))
	r := runner.NewLoadTest(httpRunnable)
	r.AddQpsStage(1000, 10*time.Second)
	//r.AddAbsoluteStage(2000, 50)
	r.Start()
}

func readPayload(path string) []byte {
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return content
}
