package runner

import (
	"aggressive-pokes/internal/stats"
	"aggressive-pokes/internal/utils"
	"aggressive-pokes/internal/worker"
	"context"
	"fmt"
	"runtime"
	"time"
)

type stageState int

const (
	stateInit stageState = iota
	stateRunning
	stateDone
)

type stage struct {
	id        int
	qps       int
	duration  time.Duration
	startTime time.Time
	endTime   time.Time
	interval  time.Duration
	stats     *stats.StageStats
	runnable  func(reporter stats.Reporter)
	state     stageState
}

func newStage(id int, qps int, duration time.Duration, runnable func(reporter stats.Reporter)) *stage {
	if duration.Seconds() < 1 || duration.Minutes() > 60 {
		panic("I don't like this duration")
	}
	return &stage{
		id:       id,
		qps:      qps,
		duration: duration,
		interval: time.Duration(1_000_000/qps) * time.Microsecond,
		runnable: runnable,
		stats:    stats.NewStageStats(),
		state:    stateInit,
	}
}

func (s *stage) format() string {
	switch s.state {
	case stateRunning:
		timeElapsed := time.Since(s.startTime)
		timeLeft := s.duration - timeElapsed

		stageProgress := float64(time.Since(s.startTime)) / float64(s.duration)
		if stageProgress > 1 {
			stageProgress = 0.99
		}
		return fmt.Sprintf("Stage [%v] running, qps: [%v], progress: [%.1f%%], running for: [%v], time left: [%v]",
			s.id,
			s.qps,
			stageProgress*100,
			utils.PrettyDuration(timeElapsed),
			utils.PrettyDuration(timeLeft))
	case stateDone:
		return fmt.Sprintf("Stage [%v] done, qps: [%v], duration: [%v]\n%v\n",
			s.id, s.qps, utils.PrettyDuration(s.endTime.Sub(s.startTime)), s.stats.Format(true))
	default:
		return fmt.Sprintf("Stage [%v], qps: [%v], duration: %v", s.id, s.qps, s.duration)
	}
}

func (s *stage) Run() {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(s.duration))
	defer cancel()

	reporter := stats.NewReporter(s.stats)
	workersFinished := worker.StartWorkers(ctx, reporter, s.qps*5)
	s.runReportRoutine(ctx, 1000*time.Millisecond)
	s.runTaskRoutine(ctx)
	utils.PrintBoxed("", s.format(), "Starting...")

	<-workersFinished
	s.state = stateDone

	utils.PrintBoxed("", s.format())
}
func (s *stage) runTaskRoutine(ctx context.Context) {
	go func() {
		s.startTime = time.Now()
		s.endTime = s.startTime.Add(s.duration)
		ticker := time.Tick(s.interval)
		s.state = stateRunning
		for {
			select {
			case <-ctx.Done():
				//fmt.Printf("Stage #%v task routine done\n", s.id)
				return
			case <-ticker:
				worker.Submit(s.runnable)
			}
		}
	}()
}

func (s *stage) runReportRoutine(ctx context.Context, interval time.Duration) {
	reportStatsTicker := time.Tick(interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				//fmt.Printf("\nStage #%v report routine done\n", s.id)
				return
			case <-reportStatsTicker:
				utils.PrintBoxed(
					s.format(),
					s.stats.Format(false),
					utils.SeparatorLine,
					fmt.Sprintf("Goroutines: %-6v |", runtime.NumGoroutine()),
				)
			}
		}
	}()
}
