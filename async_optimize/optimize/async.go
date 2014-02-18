package optimize

import (
	"errors"
	"fmt"
	"math"

	"github.com/btracey/goexamples/async_optimize/optimize/controller"
)

// Instead of evaluating in batches, let's instead evaluate them asyncronously
// (when one ends start the next one, don't wait for the whole batch to finish).
// Let's also add in the capability for being more intelligent about selecting
// the next point.

// Channels are the way to communicate between concurrently executing processes
// They can be initialized with
//		c := make(chan Type)
// The form
//		c <- type
// sends the value type on the channel. This should be matched in a different
// goroutine with
// 		t = <- c
// aka read the value from the channel. The go runtime handles this exchange
// happening safely. In a specific goroutine, the t = <-c line will cause the
// goroutine to stop execution until a value can be read from the table. If
// no other goroutine can send on the channel (aka there is a deadlock condition)
// the go runtime will panic.

// A LocalWorker is a worker which concurrently executes an objective function locally
// (within shared memory)
type LocalWorker struct {
	// To help with code legibility and safety, channels can also be read-only
	// <-chan, or write-only chan<-. Channels are always created as being neither,
	// but can be assigned. In this worker struct, the worker is not allowed to
	// send on the read channel, or read from the write channel. This helps prevent
	// communication errors

	read  <-chan []float64 // channel for reading in values to evaluate
	write chan<- Ans       // channel for returning the evaulated objectives

	fun Objer // objective function
	Id  int   // ID of the worker

	Output bool
	quit   <-chan bool // Channel to signal closure of the goroutine upon completion
}

func (l *LocalWorker) Init(read <-chan []float64, write chan<- Ans, fun Objer, quit <-chan bool) {
	l.read = read
	l.write = write
	l.fun = fun
	l.quit = quit
}

// Run runs the worker
func (w *LocalWorker) Run() {
	if w.Output {
		fmt.Printf("worker %d launched\n", w.Id)
	}
OuterLoop:
	// Continue looking for function calls to execute until told to quit
	for {
		// The select statement is a control statement for concurrent programming
		// It is like a switch statement, except it works on channel reads.
		// The goroutine will read from any of the availble channels. If none are
		// available it will go to the default statement, if present, or will
		// wait until one of the channels is available. So, the following code
		// will wait until it can either read from the read channel or it reads
		// from the quit channel.
		select {
		case x := <-w.read:
			// If it gets a case to run, evaluate the objective function and then
			// send the answer back
			obj := w.fun.Obj(x)
			w.write <- Ans{Loc: x, Obj: obj}
			if w.Output {
				fmt.Printf("worker %d finished running\n", w.Id)
			}
		case <-w.quit:
			// If read from the quit channel, break out of the loop
			break OuterLoop
		}
	}
	if w.Output {
		fmt.Printf("worker %d quit\n", w.Id)
	}
}

// A Worker is control device for the concurrent evaluation of an objective function
type Worker interface {
	Init(read <-chan []float64, write chan<- Ans, fun Objer, quit <-chan bool)
	Run() // Launches the process
}

// Async is an optimizer which makes concurrent calls to the objective function
// Assumes the objective function is parallelizable
type Async struct {
	MaxFunEvals   int // Maximum number of allowed function evaluations
	NumDim        int // Dimension of the problem
	numConcurrent int // How many to evaluate concurrently
	PrintReturns  bool

	Controller controller.C // Controller for the next function location to evaluate

	Workers []Worker

	bestObj float64
	bestLoc []float64

	toWorker   chan<- []float64
	fromWorker <-chan Ans
	quitWorker chan bool

	fun Objer
}

func (async *Async) init() {
	async.numConcurrent = len(async.Workers)
	// Allocate memory
	async.bestObj = math.Inf(1)
	async.bestLoc = make([]float64, async.NumDim)

	// Create the communication channels
	toWorker := make(chan []float64)
	fromWorker := make(chan Ans)
	quit := make(chan bool)

	async.toWorker = toWorker
	async.fromWorker = fromWorker
	async.quitWorker = quit

	for _, worker := range async.Workers {
		go func(worker Worker) {
			worker.Init(toWorker, fromWorker, async.fun, quit)

			worker.Run()
		}(worker)
	}
}

func (async *Async) updateBest(ans Ans) {
	if ans.Obj < async.bestObj {
		async.bestObj = ans.Obj
		copy(async.bestLoc, ans.Loc)
	}
}

// Initer is a controller that can requires initialization
type Initer interface {
	Init(nDim int)
}

func (async *Async) Optimize(fun Objer) (Ans, error) {
	if async.NumDim <= 0 {
		return Ans{}, errors.New("async: NumDim non-positive")
	}
	if async.MaxFunEvals <= 0 {
		return Ans{}, errors.New("async: MaxFunEvals non-positive")
	}
	if len(async.Workers) == 0 {
		return Ans{}, errors.New("async: Length of workers is zero")
	}

	async.fun = fun
	async.init()

	// Check if the controller is an initer
	initer, ok := async.Controller.(Initer)
	if ok {
		initer.Init(async.NumDim)
	}

	nDim := async.NumDim

	// Give an initial function to each worker
	for i := 0; i < async.numConcurrent; i++ {
		xnext := make([]float64, nDim)
		async.Controller.Next(xnext)
		async.toWorker <- xnext // The workers are executing concurrently and will read from the channel
	}
	// That's it!
	nFunEvals := async.numConcurrent
	for nFunEvals < async.MaxFunEvals {
		// Wait to read from a solution
		ans := <-async.fromWorker

		async.updateBest(ans)

		// Add the answer to the nexter
		async.Controller.Add(ans.Loc, ans.Obj)
		// Get the next location to evaluate (reuse the memory to avoid allocations)
		xnext := ans.Loc
		async.Controller.Next(xnext)

		// Send the new location to a free worker
		async.toWorker <- xnext
		nFunEvals++
	}
	// Read the final returns from the workers
	for i := 0; i < async.numConcurrent; i++ {
		ans := <-async.fromWorker
		async.Controller.Add(ans.Loc, ans.Obj)
		async.updateBest(ans)
	}
	// The worker goroutines are all still running, so shut them all down.
	// In select, can always read from a closed channel, so this is enough
	close(async.quitWorker)

	xbest := make([]float64, async.NumDim)
	copy(xbest, async.bestLoc)
	return Ans{Loc: xbest, Obj: async.bestObj}, nil
}
