package main

import (
	"strings"
	"sync/atomic"
)

func watcher[T comparable]() (set func(T), wait func(*T)) {
	ch := make(chan int, 1)
	var payload atomic.Pointer[T]
	payload.Store(new(T))

	set = func(val T) {
		payload.Store(&val)
		select {
		case ch <- 0:
		default:
		}
	}

	wait = func(orig *T) {
		val := *orig
		for val == *orig {
			<-ch
			val = *payload.Load()
		}
		*orig = val
	}

	return
}

func truncLines(text string, width int) string {
	const ellipses = "..................."

	text = strings.ReplaceAll(text, "\t", "        ")
	lines := strings.Split(text, "\n")
	for line := range lines {
		line := &lines[line]
		if len(*line) > width {
			n := max(width-3, 0)
			*line = (*line)[:n] + ellipses[:width-n]
		}
	}
	text = strings.Join(lines, "\n")
	return text
}
