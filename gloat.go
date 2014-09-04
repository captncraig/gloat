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
	totalOps          int64
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
	fmt.Println("main started")
	time.Sleep(t.Duration)
	fmt.Println("main stopping")
	close(stop)
	waitGroup.Wait()
	fmt.Println("main cooled")
	close(results)
	fmt.Println(t.totalOps)
	fmt.Println(time.Duration(int64(t.totalTime) / t.totalOps))
}
func harvestResults(r <-chan SingleResult, t *LoadTest) {
	for {
		s := <-r
		t.totalOps++
		t.totalTime += s.duration
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
