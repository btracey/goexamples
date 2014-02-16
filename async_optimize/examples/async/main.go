package main

import (
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"time"

	"github.com/btracey/goexamples/async_optimize/optimize"
	"github.com/btracey/goexamples/async_optimize/optimize/controller"
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

func main() {

	nCpu := runtime.NumCPU()
	runtime.GOMAXPROCS(nCpu)
	rand.Seed(time.Now().UnixNano())

	// Set up the optimization run

	//control := controller.Simple{}
	//control := &controller.Avoid{NumGuess: 100}
	control := &controller.AsyncAvoid{}

	objer := Varied{
		Fixed:  time.Second,
		Varied: 3500 * time.Millisecond,
	}
	optimizer := &optimize.Async{
		MaxFunEvals:   25,
		NumDim:        2,
		NumConcurrent: nCpu - 1,
		PrintReturns:  true,

		Controller: control,
	}

	ans, err := optimizer.Optimize(objer)
	if err != nil {
		fmt.Println("Error optimizing ")
	}
	fmt.Println("Optimization finished\nBest location is", ans.Loc, "\nBest value is ", ans.Obj)
}
