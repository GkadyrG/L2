package main

import (
	"fmt"
	"sync"
	"time"
)

func or(channels ...<-chan interface{}) <-chan interface{} {
	var once sync.Once
	out := make(chan interface{})
	done := make(chan struct{})
	for _, ch := range channels {
		go func(ch <-chan interface{}) {
			for {
				select {
				case v, ok := <-ch:
					if !ok {
						done <- struct{}{}
						once.Do(func() {
							close(done)
							close(out)
						})
						return
					}
					select {
					case out <- v:
					case <-done:
						return
					}
				case <-done:
					return
				}
			}
		}(ch)
	}

	return out
}

func main() {
	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	start := time.Now()
	<-or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)
	fmt.Printf("done after %v", time.Since(start))
}
