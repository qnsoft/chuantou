package test

import (
	"fmt"
	"testing"
)

// 测试关闭协程

// 使用 for-range
func TestExitRoutine1(t *testing.T) {
	flagChan := make(chan bool)
	go func(flagChan <-chan bool) {
		for x := range flagChan {
			fmt.Println("Process", x)
		}
		fmt.Println("routine exit")
	}(flagChan)

	flagChan <- true
	flagChan <- false
	close(flagChan)

	select {}
}

func TestCleanChan(t *testing.T) {
	numberChan := make(chan int32, 10)
	numberChan <- 1
	numberChan <- 2
	numberChan <- 3

	for n := range numberChan {
		fmt.Println(n, len(numberChan))
		if len(numberChan) == 0 {
			close(numberChan)
		}
	}
}
