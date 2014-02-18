package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/btracey/goexamples/async_optimize/optimize"
)

// Example is an example objective function
var Example func(x []float64) float64 = func(x []float64) float64 {
	var sum float64
	for i := range x {
		inc := 2*x[i]*math.Sin(x[i]) + x[i]*math.Cos(2*x[i])
		inc *= math.Exp(-math.Abs(x[i]) / 5)
		sum += inc
	}
	return sum
}

func main() {
	// Create the optimizer
	// The := operator means "create a new variable and infer the type". This is
	// equivalent to saying
	// 		// Create a new variable which is a pointer to the type Stupid in the
	//		// optimizer package
	//		var optimizer *optimize.Stupid
	//
	//		optimizer = optimize.Stupid{...}
	optimizer := &optimize.Stupid{
		MaxFunEvals: 100000,
		NumDim:      150,
	}

	// Set the random number seed
	rand.Seed(time.Now().UnixNano())
	ans := optimizer.Optimize(Example)
	fmt.Println("Optimization finished\nBest location is", ans.Loc, "\nBest value is ", ans.Obj)
}
