package main

import (
	"fmt"
	"sync"
)

func startSend() <-chan string {
	var wg sync.WaitGroup
	ch := make(chan string)
	nSender := 2
	nSend := 3

	wg.Add(nSender)

	// nSender 個の goroutine を起動
	for n := 0; n < nSender; n++ {
		n := n
		go func() {
			defer wg.Done()

			// nSend 回のメッセージを ch に送信
			for v := 0; v < nSend; v++ {
				ch <- fmt.Sprintf("%d from sender %d", v, n)
			}
		}()
	}

	// すべての sender goroutine が終了したら close(ch) する
	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch
}

func startReceive(ch <-chan string) {
	var wg sync.WaitGroup
	nReceiver := 2

	wg.Add(nReceiver)

	// nReceiver 個の goroutine を起動
	for n := 0; n < nReceiver; n++ {
		n := n
		go func() {
			defer wg.Done()

			for v := range ch {
				fmt.Printf("receive '%s' at receiver %d\n", v, n)
			}
		}()
	}

	wg.Wait()
}

func main() {
	ch := startSend()
	startReceive(ch)
}
