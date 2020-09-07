package test

import (
	"testing"
	"time"
)

func TestProcess(t *testing.T) {
	c := make(chan int, 0)
	go func() {
		c <- 1
	}()
	time.Sleep(time.Second * 2)
	t.Log(len(c))
}
