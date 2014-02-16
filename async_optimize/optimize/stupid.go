package optimize

import (
	"math"
	"math/rand"
)

// Ans is a struct containing a location and an objective value, and represents
// a result from evaluating an objective function
type Ans struct {
	Loc []float64
	Obj float64
}

// Stupid is an optimizer which finds the objective of the function through iterative
// guess and check.
type Stupid struct {
	// Fields beginning with capital letters are public

	MaxFunEvals int // Maximum number of allowed function evaluations
	NumDim      int // Dimension of the problem

	// Fields beginning with lower-case letters are private
	bestObj float64
	bestLoc []float64
}

// init sets the initial best objective value found to negative infinity and
// allocates memory for the best location
// This is a method definition for the SimpleIterative type. It requires pointer
// receiver
func (stupid *Stupid) init() {
	// In go, when a value is created, it is initialized to its zero value. For
	// float64 types this is 0.
	stupid.bestObj = math.Inf(1)
	// There are  a few special "reference" types in go that need to be allocated
	// with make. A "slice", similar to a dynamic array, is one of them. This
	// creates a slice of nDim doubles
	stupid.bestLoc = make([]float64, stupid.NumDim)
}

func (stupid *Stupid) Optimize(fun func(loc []float64) float64) Ans {
	// Call the initialization
	stupid.init()
	// Create some memory for the new location
	xNext := make([]float64, stupid.NumDim)
	// Guess and check MaxFunEvals number of times
	for i := 0; i < stupid.MaxFunEvals; i++ {
		// Get a new random location
		for j := range xNext {
			xNext[j] = rand.NormFloat64()
		}

		// Evaluate the objective function
		f := fun(xNext)

		// See if it's better, and if so, update the best point
		if f < stupid.bestObj {
			stupid.bestObj = f
			copy(stupid.bestLoc, xNext)
		}
	}
	// Return the best location found
	// The form
	// 		StructType{}
	// is a struct literal. It creates a new value of StructType.
	//		&StructType{}
	// creates a new value of StructType and takes its reference.
	// You can also specify fields in a struct literal.
	return Ans{Loc: stupid.bestLoc, Obj: stupid.bestObj}
}
