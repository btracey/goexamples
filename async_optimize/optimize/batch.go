package optimize

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// In go, we can define non-struct types as well. For example, we'll create an
// enum-like type Status which has a base representation of integer

// Status is a type representing the status of the optimizer.
type Status int

// Now, let's define some constants with that type. iota is a built in type
// for defining constants which starts at zero and automatically increments
const (
	Continue    Status = iota // The optimization should continue
	MaxFunEvals               // Maximum number of function evaluations reached
)

// Iterfaces represent a form of duck typing. A type satisfies an interface if it
// has methods which match all of the necessary signitures. So, if I have
// 		func Evaluate(o Objer, loc []float64) float64{
//			return o.Obj(loc)
//		}
// Any type which has a method Status() that is niladic and returns a Status
// can be put into this method.

// Objer is a type for an unconstrained objective function
type Objer interface {
	Obj([]float64) float64
}

// Batch is an optimizer which finds the optimum value through guess and check
// by evaluating the objective function in parallel.
type Batch struct {
	MaxFunEvals    int  // Maximum number of allowed function evaluations
	NumDim         int  // Dimension of the problem
	BatchSize      int  // How many functions to call simultaneously
	PrintBatchTime bool // Display how long it took to run the batch

	// Fields beginning with lower-case letters are private
	nDim    int
	bestObj float64
	bestLoc []float64
}

func (batch *Batch) init() {
	batch.bestObj = math.Inf(1)
	batch.bestLoc = make([]float64, batch.NumDim)
}

// Optimize optimizes the objective function by parallel guess-and-check
func (batch *Batch) Optimize(fun Objer) (Ans, error) {
	// We should do some error handling. error is a built-in type in go which is
	// an interface that has the signature
	//		type error interface{
	//			Error() string
	//		}

	if batch.BatchSize <= 0 {
		return Ans{}, errors.New("batch: BatchSize non-positive")
	}
	if batch.NumDim <= 0 {
		return Ans{}, errors.New("batch: NumDim non-positive")
	}
	if batch.MaxFunEvals <= 0 {
		return Ans{}, errors.New("batch: MaxFunEvals non-positive")
	}

	// Initialize
	batch.init()

	var nFunEvals int
	answers := make([]Ans, batch.BatchSize)

	// You'll see in a moment
	wg := &sync.WaitGroup{}
	// A while loop in go is just for
	for nFunEvals < batch.MaxFunEvals {
		// Evaluate the objective function in parallel

		// The "go" command launches a concurrently executing process. Concurrently
		// executing meaning that it does not communicate back to the main program
		// unless some form of explicit synchronization is provided. Here, we will
		// use the above WaitGroup to do so

		// Tell the wait group that a number of new processes are being launched
		wg.Add(batch.BatchSize)

		// Define our independent function
		f := func(i int) {
			// guess a random location
			x := make([]float64, batch.NumDim)
			for j := range x {
				x[j] = rand.NormFloat64()
			}
			// Evaluate the objective function
			obj := fun.Obj(x)

			// Place the result in the answers struct. Note the scoping rules --
			// the answers struct is not fixed when we define the function. We can
			// edit it as normal. If you want something fixed you can copy and edit it,
			// so this is more powerful than MATLAB style
			answers[i] = Ans{Obj: obj, Loc: x}

			// Tell the waitgroup that the process has finished
			wg.Done()
		}

		startTime := time.Now()
		for i := 0; i < batch.BatchSize; i++ {
			// Launch an concurrently executing call to the function. It launches
			// a new "goroutine" with the function call
			go f(i)

			// Yep, that's it! If the computer has multiple processors the go
			// runtime will map this concurrently executing process onto another
			// processor to execute it in parallel. Note that it needs to be f(i)
			// and not just f() because as the loop iterates i will change. Making
			// i a function argument fixes the value.
		}
		wg.Wait() // The waitgroup pauses this go routine until the counter is zero

		if batch.PrintBatchTime {
			fmt.Println("Parallel runs returned in ", time.Since(startTime))
		}

		// See what our new best point is
		for _, answer := range answers {
			if answer.Obj < batch.bestObj {
				batch.bestObj = answer.Obj
				batch.bestLoc = answer.Loc
			}
		}
		nFunEvals += batch.BatchSize
	}
	// Return the best found value
	return Ans{Loc: batch.bestLoc, Obj: batch.bestObj}, nil
}
