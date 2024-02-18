package runner

import (
	"aggressive-pokes/internal/stats"
	"aggressive-pokes/internal/utils"
	"time"
)

type Runner struct {
	stages   []*stage
	runnable func(reporter stats.Reporter)
}

func NewRunner(runnable func(reporter stats.Reporter)) Runner {
	return Runner{
		runnable: runnable,
	}
}

func (r *Runner) AddStage(qps int, duration time.Duration) {
	r.stages = append(r.stages, newStage(len(r.stages)+1, qps, duration, r.runnable))
}

func (r *Runner) Run() {
	if len(r.stages) == 0 {
		panic("No stages to poke around")
	}

	for _, s := range r.stages {
		s.Run()
	}
	utils.ClearConsole()
	for _, s := range r.stages {
		utils.PrintBoxed("", s.format())
	}

}
