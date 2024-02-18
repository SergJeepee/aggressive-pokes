package stats

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Reporter struct {
	stats *StageStats
	mx    *sync.Mutex
}

func NewReporter(stats *StageStats) Reporter {
	return Reporter{
		stats: stats,
		mx:    &sync.Mutex{},
	}
}

func (r *Reporter) ReportSuccess(reason string, elapsed time.Duration) {
	r.mx.Lock()
	defer r.mx.Unlock()

	r.stats.totalExecuted++
	bucket := r.stats.metrics[reason]
	bucket.count++
	bucket.elapsed = append(bucket.elapsed, float64(elapsed.Milliseconds()))
	r.stats.metrics[reason] = bucket

}

func (r *Reporter) ReportFailure(reason string, msg string, elapsed time.Duration) {
	r.mx.Lock()
	defer r.mx.Unlock()

	r.stats.totalExecuted++
	bucket := r.stats.metrics[reason]
	bucket.count++
	bucket.msg = append(bucket.msg, msg)
	bucket.elapsed = append(bucket.elapsed, float64(elapsed.Milliseconds()))
	r.stats.metrics[reason] = bucket
}

type StageStats struct {
	totalExecuted int
	metrics       reasonedExecMetrics
	mx            *sync.Mutex
}

func NewStageStats() *StageStats {
	return &StageStats{
		metrics: make(reasonedExecMetrics),
		mx:      &sync.Mutex{},
	}
}

func (s *StageStats) Format(includePercentiles bool) string {
	return fmt.Sprintf("Total: %-11v |\n%v", s.totalExecuted, s.metrics.format(includePercentiles))
}

type reasonedExecMetrics map[string]reasonBucket

func (m *reasonedExecMetrics) format(includePercentiles bool) string {
	var entries []string
	if len(*m) == 0 {
		return "No metrics"
	}
	for k, v := range *m {
		entries = append(entries, fmt.Sprintf("%-18v | %-18v", k, v.Format(includePercentiles)))
	}

	sort.Strings(entries)
	return fmt.Sprintf("%v", strings.Join(entries, "\n"))
}

type reasonBucket struct {
	count   int
	msg     []string
	elapsed []float64 // millis
}

func (b *reasonBucket) Format(includePercentiles bool) string {
	if includePercentiles {
		return fmt.Sprintf("count: %-6v | avg duration: %-10v | percentiles: %v", b.count, b.avgDuration(), b.percentile(50, 90, 99))
	}
	return fmt.Sprintf("count: %-6v | avg duration: %-10v", b.count, b.avgDuration())

}

func (b *reasonBucket) avgDuration() time.Duration {
	if len(b.elapsed) == 0 {
		return time.Duration(0)
	}
	return time.Duration(Avg(b.elapsed)) * time.Millisecond
}

func (b *reasonBucket) percentile(percentile ...float64) map[float64]time.Duration {
	percentiles, err := Percentile(b.elapsed, percentile...)
	if err != nil {
		fmt.Printf("Cannot calcualte percentile: %v", err)
		return nil
	}

	m := make(map[float64]time.Duration)
	for i, p := range percentile {
		m[p] = time.Duration(percentiles[i]) * time.Millisecond
	}
	return m
}
