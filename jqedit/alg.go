package main

import (
	"strings"
	"sync/atomic"
	"unicode"
	"unicode/utf8"
)

func watcher[T comparable]() (set func(T), wait func(*T)) {
	ch := make(chan int, 1)
	var storage atomic.Pointer[T]
	storage.Store(new(T))

	set = func(value T) {
		storage.Store(&value)
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
