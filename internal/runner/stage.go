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

type stageRunner interface {
	run()
	runTaskRoutine(ctx context.Context)
	runReportRoutine(ctx context.Context, interval time.Duration)
	format() string
}

type baseStage struct {
	id        int
	startTime time.Time
	endTime   time.Time
	stats     *stats.StageStats
	runnable  func(reporter stats.Reporter)
	state     stageState
}

type qpsStage struct {
	baseStage
	qps      int
	duration time.Duration
	interval time.Duration
}

func (s *qpsStage) run() {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(s.duration))
	defer cancel()

	reporter := stats.NewReporter(s.stats)
	workersFinished := worker.StartWorkers(ctx, reporter, s.qps*100)
	s.runReportRoutine(ctx, 1000*time.Millisecond)
	s.runTaskRoutine(ctx)
	utils.PrintBoxed("", s.format(), "Starting...")

	<-workersFinished
	s.state = stateDone

	utils.PrintBoxed("", s.format())
}

func (s *qpsStage) runTaskRoutine(ctx context.Context) {
	go func() {
		s.startTime = time.Now()
		s.endTime = s.startTime.Add(s.duration)
		ticker := time.Tick(s.interval)
		s.state = stateRunning
		for {
			select {
			case <-ctx.Done():
				worker.Cancel()
				//fmt.Printf("Stage #%v task routine done\n", s.id)
				return
			case <-ticker:
				worker.Submit(s.runnable)
			}
		}
	}()
}

func (s *qpsStage) runReportRoutine(ctx context.Context, interval time.Duration) {
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

func (s *qpsStage) format() string {
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

type absoluteStage struct {
	baseStage
	amount      int
	asyncFactor int
}

func (s *absoluteStage) run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reporter := stats.NewReporter(s.stats)
	workersFinished := worker.StartWorkers(ctx, reporter, s.asyncFactor)
	s.runReportRoutine(ctx, 1000*time.Millisecond)
	s.runTaskRoutine(ctx)
	utils.PrintBoxed("", s.format(), "Starting...")

	<-workersFinished
	s.state = stateDone
	s.endTime = time.Now()

	utils.PrintBoxed("", s.format())
}

func (s *absoluteStage) runTaskRoutine(ctx context.Context) {
	go func() {
		s.startTime = time.Now()
		s.state = stateRunning
		for i := 0; i < s.amount; i++ {
			select {
			case <-ctx.Done():
				worker.Cancel()
				//fmt.Printf("Stage #%v task routine done\n", s.id)
				return
			default:
				worker.Submit(s.runnable)
			}
		}
		worker.Cancel()
	}()
}

func (s *absoluteStage) runReportRoutine(ctx context.Context, interval time.Duration) {
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

func (s *absoluteStage) format() string {
	switch s.state {
	case stateRunning:
		timeElapsed := time.Since(s.startTime)
		stageProgress := float64(s.stats.Executed()) / float64(s.amount)
		if stageProgress > 1 {
			stageProgress = 0.99
		}
		timeLeft := time.Duration(float64(timeElapsed.Milliseconds())/stageProgress)*time.Millisecond - timeElapsed
		return fmt.Sprintf("Stage [%v] running, amount: [%v], progress: [%.1f%%], running for: [%v], time left: [%v]",
			s.id,
			s.amount,
			stageProgress*100,
			utils.PrettyDuration(timeElapsed),
			utils.PrettyDuration(timeLeft))
	case stateDone:
		return fmt.Sprintf("Stage [%v] done, amount: [%v], duration: [%v]\n%v\n",
			s.id, s.amount, utils.PrettyDuration(s.endTime.Sub(s.startTime)), s.stats.Format(true))
	default:
		return fmt.Sprintf("Stage [%v], amount: [%v]", s.id, s.amount)
	}
}

func newStageQps(id, qps int, duration time.Duration, runnable func(reporter stats.Reporter)) stageRunner {
	if duration.Seconds() < 1 || duration.Minutes() > 60 {
		panic("Duration should be in range [1s, 60m]")
	}

	return &qpsStage{
		baseStage: baseStage{
			id:       id,
			runnable: runnable,
			stats:    stats.NewStageStats(),
			state:    stateInit,
		},
		qps:      qps,
		duration: duration,
		interval: time.Duration(1_000_000/qps) * time.Microsecond,
	}
}

func newStageAbsolute(id, amount, asyncFactor int, runnable func(reporter stats.Reporter)) stageRunner {
	if amount < 1 || amount > 1_000_000_000 {
		panic("Amount should be in range [1, 1_000_000_000]")
	}
	if asyncFactor < 1 || asyncFactor > worker.MaxWorkerPool {
		panic(fmt.Sprintf("Async factor should be in range [1, %v]", worker.MaxWorkerPool))
	}

	return &absoluteStage{
		baseStage: baseStage{
			id:       id,
			runnable: runnable,
			stats:    stats.NewStageStats(),
			state:    stateInit,
		},
		amount:      amount,
		asyncFactor: asyncFactor,
	}

}
