package gloat

import (
	"fmt"
	"sync"
	"time"
)

type TestResult byte

var waitGroup = sync.WaitGroup{}

const (
	Status_Success TestResult = iota
	Status_Failure
	Status_Timeout
)

type TestFunction func() TestResult

type LoadTest struct {
	RequestsPerSecond int
	Workers           int
	Duration          time.Duration
	F                 TestFunction
	totalTime         time.Duration
	windowTime        time.Duration
	totalOps          int64
	windowOps         int64
}

func NewLoadTest() *LoadTest {
	test := LoadTest{}
	test.RequestsPerSecond = 0
	test.Workers = 1
	test.Duration = 1 * time.Minute
	return &test
}

type SingleResult struct {
	duration time.Duration
	status   TestResult
}

func (t *LoadTest) Run() {
	stop := make(chan struct{})
	doWork := make(chan bool)
	results := make(chan SingleResult)
	for i := 0; i < t.Workers; i++ {
		go work(t.F, doWork, results)
		waitGroup.Add(1)
	}
	go tick(doWork, t.RequestsPerSecond, stop)
	go harvestResults(results, t)
	time.Sleep(t.Duration)
	close(stop)
	waitGroup.Wait()
	fmt.Println("Done")
	close(results)
	fmt.Println(t.totalOps)
	fmt.Println(time.Duration(int64(t.totalTime) / t.totalOps))
}

func harvestResults(r <-chan SingleResult, t *LoadTest) {
	tick := time.NewTicker(1 * time.Second)
	for {
		select {
		case s := <-r:
			t.totalOps++
			t.totalTime += s.duration
			t.windowOps++
			t.windowTime += s.duration
		case <-tick.C:
			fmt.Println(t.windowOps)
			if t.windowOps != 0 {
				fmt.Println(time.Duration(int64(t.windowTime) / t.windowOps))
			} else {
				fmt.Println(0)
			}
			t.windowOps = 0
			t.windowTime = 0
		}
	}
}

func tick(output chan<- bool, maxPerSecond int, stop <-chan struct{}) {
	if maxPerSecond <= 0 {
		for {
			select {
			case <-stop:
				close(output)
				return
			default:
				output <- true
			}
		}
	}
	interval := time.Second / time.Duration(maxPerSecond)
	fmt.Println(interval)
	ticker := time.NewTicker(interval)
	for {

		select {
		case <-stop:
			ticker.Stop()
			close(output)
			return
		case <-ticker.C:
			output <- true
		}
	}
}

func work(f TestFunction, c <-chan bool, r chan<- SingleResult) {
	defer waitGroup.Done()
	for {
		b := <-c
		if !b {
			return
		}
		start := time.Now()
		status := f()
		duration := time.Now().Sub(start)
		r <- SingleResult{duration, status}
	}
}
