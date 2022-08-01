package main

import (
	"context"
	"fmt"
	"time"
)

func fib(ctx context.Context) <-chan int {
	ch := make(chan int)
	a, b := 0, 1

	go func() {
		defer close(ch)
		for {
			v := a
			a = b
			b = v + b

			select {
			case ch <- v:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	f := fib(ctx)

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	for v := range f {
		fmt.Println(v)
	}
}
