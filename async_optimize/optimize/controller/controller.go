package controller

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// C is an interface for generating the next optimization location.
type C interface {
	// Next returns the next point to evaulate in the optimization routine. The
	// location is stored in-place to the first argument
	Next(x []float64)

	// Add adds the result from the function evaluation back into the optimization
	// routine
	Add(loc []float64, obj float64)
}

// Simple is a controller that just guesses a random location
type Simple struct {
}

func (Simple) Next(x []float64) {
	for i := range x {
		x[i] = rand.NormFloat64()
	}
}

func (Simple) Add(loc []float64, obj float64) {
	return
}

func distance(x, y []float64) float64 {
	if len(x) != len(y) {
		panic("length mismatch")
	}
	var sum float64
	for i, val := range x {
		sum += (val - y[i]) * (val - y[i])
	}
	return sum
}

// Avoid guesses a number of random points, and takes the one that is farthest
// away from all of the current locations
type Avoid struct {
	NumGuess int
	locs     [][]float64

	x    []float64
	dist float64
}

func (avoid *Avoid) Init(nDim int) {
	avoid.x = make([]float64, nDim)
}

func (avoid *Avoid) Next(x []float64) {
	newx := make([]float64, len(x))
	avoid.dist = math.Inf(-1)
	for i := 0; i < avoid.NumGuess; i++ {
		for j := range avoid.x {
			avoid.x[j] = rand.NormFloat64()
		}
		if len(avoid.locs) == 0 {
			copy(newx, avoid.x)
			break
		}

		minDist := math.Inf(1)
		// Find the minimum distance to all of the normal points
		for _, oldloc := range avoid.locs {
			dist := distance(avoid.x, oldloc)
			if dist < minDist {
				minDist = dist
			}
		}
		// see if this point is the farthest away so far
		if minDist > avoid.dist {
			avoid.dist = minDist
			copy(newx, avoid.x)
		}
	}
	// Return the best point and add it to the loc
	copy(x, newx)
	avoid.locs = append(avoid.locs, newx)
}

func (avoid *Avoid) Add(loc []float64, obj float64) {
	// Good code should check that the returned x is in the locs
}

// The above avoid type is fine, but the Next step will grow in computational
// cost with time. More generally, it wastes processor time by doing the update
// step sequentially with the iteration. It would be better if the updater could use
// as much time as it could looking for a better location, only stopping when necessary
// Select to the rescue!

// AsyncAvoid acts like avoid, but it searches for a better point concurrently.
// It will continually search for a better point up until the point where one is
// needed
type AsyncAvoid struct {
	Print bool

	locs [][]float64

	x    []float64
	dist float64

	bestloc  []float64
	bestdist float64

	quit     chan bool
	next     chan []float64
	nextback chan []float64
}

// Init initializes the memory and launches the concurrent process
func (avoid *AsyncAvoid) Init(nDim int) {
	avoid.x = make([]float64, nDim)
	avoid.quit = make(chan bool)
	avoid.next = make(chan []float64)
	avoid.nextback = make(chan []float64)
	go avoid.monitor()
}

func (avoid *AsyncAvoid) Add(loc []float64, obj float64) {
	// Good code should check that the returned x is in the locs
}

func (avoid *AsyncAvoid) Result() {
	close(avoid.quit)
}

func (avoid *AsyncAvoid) Next(x []float64) {
	avoid.next <- x
	x = <-avoid.nextback
	xnew := make([]float64, len(x))
	copy(xnew, x)
	avoid.locs = append(avoid.locs, xnew)
}

func (avoid *AsyncAvoid) monitor() {
OuterFor:
	for {
		select {
		case x := <-avoid.next:
			copy(x, avoid.bestloc)
			avoid.nextback <- x
		case <-avoid.quit:
			break OuterFor
		default:
			avoid.search()
		}
	}
}

func (avoid *AsyncAvoid) search() {
	for i := range avoid.x {
		avoid.x[i] = rand.NormFloat64()
	}
	minDist := math.Inf(1)
	for _, loc := range avoid.locs {
		dist := distance(avoid.x, loc)
		if dist < minDist {
			minDist = dist
		}
	}
	if minDist > avoid.bestdist {
		avoid.bestdist = minDist
		copy(avoid.bestloc, avoid.x)
	}
	if avoid.Print {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("done search")
	}
}
