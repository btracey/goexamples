package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"time"

	"github.com/btracey/goexamples/async_optimize/functions"
	"github.com/btracey/goexamples/async_optimize/optimize"
	"github.com/btracey/goexamples/async_optimize/optimize/controller"
)

var port string = ":2000"

func main() {
	nCpu := runtime.NumCPU()
	runtime.GOMAXPROCS(nCpu)
	rand.Seed(time.Now().UnixNano())

	nWorkers := 3

	workers := make([]optimize.Worker, nWorkers)
	for i := range workers {
		workers[i] = &optimize.RemoteWorker{
			Output: true,
			Port:   ":" + strconv.Itoa(2000+i),
			Id:     i,
		}
	}

	control := &controller.AsyncAvoid{}

	objer := functions.Varied{
		Fixed:  time.Second,
		Varied: 3500 * time.Millisecond,
	}

	optimizer := &optimize.Async{
		MaxFunEvals:  25,
		NumDim:       2,
		Workers:      workers,
		PrintReturns: true,

		Controller: control,
	}

	ans, err := optimizer.Optimize(objer)
	if err != nil {
		fmt.Println("Error optimizing ")
	}
	fmt.Println("Optimization finished\nBest location is", ans.Loc, "\nBest value is ", ans.Obj)
}
