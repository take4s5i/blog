package main

import (
	"fmt"
	"time"
)

func fib(quit <-chan struct{}) <-chan int {
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
			case <-quit:
				// case <-quit: は quit が close された場合も実行される。
				// こう書いておくことで quit を close すると一括で goroutine を終了されられる
				return
			}
		}
	}()

	return ch
}

func main() {
	quit := make(chan struct{})
	f := fib(quit)

	go func() {
		time.Sleep(100 * time.Millisecond)
		close(quit)
	}()

	for v := range f {
		fmt.Println(v)
	}
}
