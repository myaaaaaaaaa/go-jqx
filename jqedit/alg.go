package main

import (
	"strings"
	"sync/atomic"
	"unicode"
	"unicode/utf8"
)

// watcher creates a pair of functions to set and wait for a value to change.
// The 'set' function updates the value and notifies any waiting goroutines, while the 'wait' function blocks until the value is updated.
// This is useful for signaling changes between goroutines.
func watcher[T comparable]() (set func(T), wait func(*T)) {
	ch := make(chan int, 1)
	var storage atomic.Pointer[T]
	storage.Store(new(T))

	set = func(value T) {
		storage.Store(&value)

		// Non-blocking send to signal an update.
		// Because the channel has a buffer size of 1, if a receiver is
		// not ready, this send will either fill the buffer or, if the
		// buffer is already full, the default case will be taken.
		// In either case, any call to wait() is guaranteed to see a
		// signal and then load the latest value from the atomic pointer.
		// This mechanism coalesces multiple rapid updates into a single
		// notification, preventing a backlog of signals and ensuring the
		// waiter always gets the most recent value.
		select {
		case ch <- 0:
		default:
		}
	}

	wait = func(orig *T) {
		value := *orig
		for value == *orig {
			<-ch
			value = *storage.Load()
		}
		*orig = value
	}

	return
}

func truncLines(text string, width int) string {
	const ellipses = "..................."

	text = strings.Map(func(r rune) rune {
		if !unicode.IsPrint(r) {
			switch r {
			case '\n', '\t', ' ':
			default:
				return '.'
			}
		}
		if r == utf8.RuneError {
			return '?'
		}
		return r
	}, text)
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
