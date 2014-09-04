package main

import (
	"github.com/captncraig/gloat"
	"time"
)

func main() {
	test := gloat.NewLoadTest()
	test.RequestsPerSecond = 0
	test.Workers = 4
	test.Duration = 60 * time.Second
	test.F = gloat.HttpGet("http://yahoo.com")
	test.Run()
}
