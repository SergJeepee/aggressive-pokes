package runner

import (
	"aggressive-pokes/internal/stats"
	"aggressive-pokes/internal/utils"
	"time"
)

type LoadTest struct {
	stages []stageRunner
	//runnable func(reporter stats.Reporter)
}

func NewLoadTest() LoadTest {
	return LoadTest{
		//runnable: runnable,
	}
}

func (t *LoadTest) AddQpsStage(qps int, duration time.Duration, runnable func(reporter stats.Reporter)) {
	t.stages = append(t.stages, newStageQps(len(t.stages)+1, qps, duration, runnable))
}

func (t *LoadTest) AddAbsoluteStage(amount, asyncFactor int, runnable func(reporter stats.Reporter)) {
	t.stages = append(t.stages, newStageAbsolute(len(t.stages)+1, amount, asyncFactor, runnable))
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
