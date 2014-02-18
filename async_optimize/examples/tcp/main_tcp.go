package main

import (
	"fmt"
	"math/rand"
	"runtime"
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

	control := &controller.AsyncAvoid{}

	objer := functions.Varied{
		Fixed:  time.Second,
		Varied: 3500 * time.Millisecond,
	}

	/*
			fun := &functions.Remote{
				Location: port,
				Objer:    objer,
			}

			err := fun.Init()
		if err != nil {
			panic(err)
		}
	*/

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
