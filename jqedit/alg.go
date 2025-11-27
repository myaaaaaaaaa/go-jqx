package main

import (
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
