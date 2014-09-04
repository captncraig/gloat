package main

import (
	"github.com/captncraig/gloat"
	"time"
)

func main() {
	test := gloat.NewLoadTest()
	test.RequestsPerSecond = 0
	test.Workers = 10
	test.Duration = 20 * time.Second
	test.F = gloat.HttpGet("http://yahoo.com")
	test.Run()
}
