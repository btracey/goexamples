package optimize

import (
	"encoding/gob"
	"fmt"
	"net"
)

// A RemoteWorker is a worker which concurrently executes an objective function
// over TCP
type RemoteWorker struct {
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

	Port string // Port where things will be sent

	conn net.Conn     // Tcp connection
	enc  *gob.Encoder // Writer stream
	dec  *gob.Decoder // Reader stream
}

func (r *RemoteWorker) Init(read <-chan []float64, write chan<- Ans, fun Objer, quit <-chan bool) {
	r.read = read
	r.write = write
	r.fun = fun
	r.quit = quit

	// Establish TCP connection
	conn, err := net.Dial("tcp", r.Port)
	if err != nil {
		panic(err)
	}
	r.conn = conn

	// Now that the connection is established, serialize the objective function
	// and send it over the wire
	enc := gob.NewEncoder(conn)
	r.dec = gob.NewDecoder(conn)
	err = enc.Encode(&fun)
	if err != nil {
		panic(err)
	}
	r.enc = enc
}

// Run runs the worker
func (w *RemoteWorker) Run() {
	if w.Output {
		fmt.Printf("worker %d launched\n", w.Id)
	}
OuterLoop:
	// Continue looking for function calls to execute until told to quit
	for {
		select {
		case x := <-w.read:
			// Instead of calling the objective function, call it remotely
			err := w.enc.Encode(x)
			if err != nil {
				panic(err)
			}

			// Listen back for the objective value
			var obj float64
			err = w.dec.Decode(&obj)
			if err != nil {
				panic(err)
			}

			w.write <- Ans{Loc: x, Obj: obj}
			if w.Output {
				fmt.Printf("worker %d finished running\n", w.Id)
			}
		case <-w.quit:

			// If read from the quit channel, break out of the loop
			break OuterLoop
		}
	}
	w.conn.Close()
	if w.Output {
		fmt.Printf("worker %d quit\n", w.Id)
	}
}
