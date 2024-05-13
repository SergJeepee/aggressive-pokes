package runnables

import (
	"aggressive-pokes/internal/stats"
)

type Runnable func(reporter stats.Reporter)

func Sequential(runnables ...Runnable) func(reporter stats.Reporter) {
	return func(reporter stats.Reporter) {
		for _, runnable := range runnables {
			runnable(reporter)
		}
	}
}

func Parallel(runnables ...Runnable) func(reporter stats.Reporter) {
	return func(reporter stats.Reporter) {
		for _, runnable := range runnables {
			go runnable(reporter)
		}
	}
}
