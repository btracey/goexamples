package main

import (
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"time"

	"github.com/btracey/goexamples/async_optimize/optimize"
)

type Example struct{}

func (Example) Obj(x []float64) float64 {
	var sum float64
	for i := range x {
		inc := 2*x[i]*math.Sin(x[i]) + x[i]*math.Cos(2*x[i])
		inc *= math.Exp(-math.Abs(x[i]) / 5)
		sum += inc
	}
	return sum
}

// Varied is an objective function whose runtime is stochastic
type Varied struct {
	Example
	Fixed  time.Duration
	Varied time.Duration
}

func (v Varied) Obj(x []float64) float64 {
	obj := v.Example.Obj(x)
	// rand.Int63n returns a random int64 between 0 and n-1. The time.Duration
	// is a type conversion from an int64 to a time.Duration
	dur := time.Duration(rand.Int63n(int64(v.Varied)))
	time.Sleep(v.Fixed + dur)
	return obj
}

// Racy is an objective function that has a race condition if Obj is called
// multiple times concurrently
type Racy struct {
	i int
}

func (r *Racy) Obj(x []float64) float64 {
	r.i++
	return float64(r.i)
}

func main() {

	nCpu := runtime.NumCPU()
	runtime.GOMAXPROCS(nCpu)

	// Simple case

	objer := Example{}
	optimizer := &optimize.Batch{
		MaxFunEvals: 10000,
		NumDim:      15,
		BatchSize:   nCpu, // Set to be one per Cpu
	}

	/*
		// With a varied objective function
		objer := Varied{
			Fixed:  time.Second,
			Varied: 3500 * time.Millisecond,
		}
		optimizer := &optimize.Batch{
			MaxFunEvals:    100,
			NumDim:         15,
			BatchSize:      nCpu, // Set to be one per Cpu
			PrintBatchTime: true,
		}
		// Note that even though many of the runs take just over 1 second, the whole
		// iteration needs to pause until the last one is finished. This is a waste
		// of many processor hours
	*/
	/*
		// Concurrency is hard. Race conditions can happen and are bad. Fortunately,
		// go has a race detector! go install -race
		objer := &Racy{}
		optimizer := &optimize.Batch{
			MaxFunEvals: 1000,
			NumDim:      15,
			BatchSize:   nCpu, // Set to be one per Cpu
		}
	*/

	ans, err := optimizer.Optimize(objer)
	if err != nil {
		fmt.Println("Error optimizing ")
	}
	fmt.Println("Optimization finished\nBest location is", ans.Loc, "\nBest value is ", ans.Obj)
}
