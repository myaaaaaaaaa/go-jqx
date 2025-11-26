package main

import (
	"sync/atomic"
)

func debounceQueue[T comparable]() (push func(T), wait func(*T)) {
	ch := make(chan int, 1)
	var payload atomic.Pointer[T]
	payload.Store(new(T))

	push = func(val T) {
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
