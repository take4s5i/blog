package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan int)

	go func() {
		ch <- 1
	}()

	time.Sleep(500 * time.Millisecond)
	fmt.Println(<-ch)
	close(ch)
}
