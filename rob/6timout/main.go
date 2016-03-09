package main

import (
	"fmt"
	"math/rand"
	"time"
)

/*
  The time.After function returns a channel that blocks
  the specified duration. After the interval, the channel
  delivers the current time, once
*/

func boring(msg string) <-chan string { // Returns receive-only channel of strings
	c := make(chan string)
	go func() { // We launch a goroutine from inside the function
		rand.Seed(time.Now().Unix())
		for i := 0; ; i++ {
			c <- fmt.Sprintf("%s %d", msg, i)
			time.Sleep(time.Duration(rand.Intn(1.2e3)) * time.Millisecond)
		}
	}()
	return c // Return the channel to the caller
}

func main() {
	c := boring("Joe")
	for {
		select {
		case s := <-c:
			fmt.Println(s)
		case <-time.After(1 * time.Second):
			fmt.Println("You are too slow.")
			return
		}
	}
}
