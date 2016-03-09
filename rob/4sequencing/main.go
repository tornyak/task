package main

import (
	"fmt"
	"math/rand"
	"time"
)

// Message extends string channel with semaphore channel
type Message struct {
	str  string
	wait chan bool // wait channel is signaller
}

func fanIn(inputs ...<-chan Message) <-chan Message {
	c := make(chan Message)

	for i := range inputs {
		input := inputs[i]
		go func() {
			for {
				c <- <-input
			}
		}()
	}
	return c
}

func boring(msg string) <-chan Message { // Returns receive-only channel of Messages
	c := make(chan Message)
	waitForIt := make(chan bool) // shared between all messages
	go func() {                  // We launch a goroutine from inside the function
		for i := 0; ; i++ {
			c <- Message{fmt.Sprintf("%s %d", msg, i), waitForIt}
			time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
			<-waitForIt
		}
	}()
	return c // Return the channel to the caller
}

func main() {
	c := fanIn(boring("Joe"), boring("Ann"))

	for i := 0; i < 5; i++ {
		msg1 := <-c
		msg2 := <-c
		fmt.Println(msg1.str)
		fmt.Println(msg2.str)
		msg1.wait <- true
		msg2.wait <- true
	}
	fmt.Println("You're boring. I'm leaving")
}
