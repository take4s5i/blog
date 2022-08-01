package main

import (
	"fmt"
	"sync"
)

type MyInt int

func (x MyInt) String() string {
	return fmt.Sprint(int(x))
}

func collatz(n int) <-chan MyInt {
	ch := make(chan MyInt)

	go func() {
		defer close(ch)

		for {
			ch <- MyInt(n)
			switch {
			case n == 1:
				return
			case n%2 == 0:
				n = n / 2
			default:
				n = n*3 + 1
			}
		}
	}()

	return ch
}

func fanIn[T any](cs ...<-chan T) <-chan T {
	var wg sync.WaitGroup
	out := make(chan T)

	wg.Add(len(cs))

	for _, ch := range cs {
		ch := ch

		// 元となる channel から値を受け取り out にわたすだけの goroutine
		// 元channelが close されると wg.Done()する
		go func() {
			defer wg.Done()
			for v := range ch {
				out <- v
			}
		}()
	}

	// すべての元 channel が close されると close(out) する
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func fanOut[T fmt.Stringer](ch <-chan T, n int) {
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		i := i

		// ch から読み取って出力するだけの goroutine
		// ch は goroutine safe なので 複数の goroutine から読み書きしても問題ない
		go func() {
			defer wg.Done()
			for v := range ch {
				fmt.Printf("'%s' from %d\n", v.String(), i)
			}
		}()
	}

	wg.Wait()
}

func main() {
	c := fanIn(collatz(10), collatz(20), collatz(30))
	fanOut(c, 3)
}
