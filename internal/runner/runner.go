package runner

import (
	"aggressive-pokes/internal/stats"
	"aggressive-pokes/internal/utils"
	"time"
)

type LoadTest struct {
	stages   []stageRunner
	runnable func(reporter stats.Reporter)
}

func NewLoadTest(runnable func(reporter stats.Reporter)) LoadTest {
	return LoadTest{
		runnable: runnable,
	}
}

func (t *LoadTest) AddQpsStage(qps int, duration time.Duration) {
	t.stages = append(t.stages, newStageQps(len(t.stages)+1, qps, duration, t.runnable))
}

func (t *LoadTest) AddAbsoluteStage(amount, asyncFactor int) {
	t.stages = append(t.stages, newStageAbsolute(len(t.stages)+1, amount, asyncFactor, t.runnable))
}

func (t *LoadTest) Start() {
	if len(t.stages) == 0 {
		panic("No stages to poke around")
	}

	for _, s := range t.stages {
		s.run()
	}
	utils.ClearConsole()
	for _, s := range t.stages {
		utils.PrintBoxed("", s.format())
	}

}
