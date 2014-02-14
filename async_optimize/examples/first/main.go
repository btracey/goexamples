package main

import (
	"fmt"
	"math"

	"miso/optimize"
)

// Example is an example objective function
var Example func(x []float64) float64 = func(x []float64) float64 {
	var sum float64
	for i := range x {
		sum += x[i]*math.Sin(x[i]) + x[i]*math.Cos(2*x[i])
	}
	return sum
}

func main() {
	// The := operator means "create a new variable and infer the type". This is
	// equivalent to saying
	// 		// Create a new variable which is a pointer to the type Stupid in the
	//		// optimizer package
	//		var optimizer *optimize.Stupid
	//
	//		optimizer = optimize.Stupid{...}
	optimizer := &optimize.Stupid{
		MaxFunEvals: 30,
		NumDim:      15,
	}
	ans := optimizer.Optimize(Example)
	fmt.Println("Optimization finished\nBest location is", ans.Loc, "\nBest value is ", ans.Obj)
}
