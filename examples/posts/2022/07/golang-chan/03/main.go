package main

import "fmt"

func fib(n int) <-chan int {
	ch := make(chan int)
	a, b := 0, 1

	go func() {
		defer close(ch)
		for {
			v := a
			a = b
			b = v + b
			n--

			if n < 0 {
				return
			}
			ch <- v
		}
	}()

	return ch
}

func main() {
	f := fib(10)

	for v := range f {
		fmt.Println(v)
	}
}
