package functions

import (
	"encoding/gob"
	"fmt"
	"math"
	"math/rand"
	"net"
	"time"
)

func init() {
	gob.Register(Varied{})
	gob.Register(&Remote{})
	gob.Register(Example{})
}

type Objer interface {
	Obj([]float64) float64
}

// Example is a simple example objective function
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

// Remote is an objective function who is evaluated remotely
type Remote struct {
	Location string // tcp location of the objective function
	//BufferSize int

	// Function to evaluate.
	Objer
	first bool
	conn  net.Conn
	enc   *gob.Encoder
	dec   *gob.Decoder
	b     []byte
}

func (r *Remote) Init() error {
	conn, err := net.Dial("tcp", r.Location)
	if err != nil {
		return err
	}
	r.conn = conn

	// Now that the connection is established, serialize the objective function
	// and send it over the wire
	enc := gob.NewEncoder(conn)
	r.dec = gob.NewDecoder(conn)
	err = enc.Encode(&r.Objer)
	if err != nil {
		return err
	}
	r.enc = enc

	//r.b = make([]byte, r.BufferSize)
	return nil
}

func (r *Remote) Obj(x []float64) float64 {

	fmt.Println("x is ", x)
	// Serialize and send the location
	err := r.enc.Encode(x)
	if err != nil {
		panic(err)
	}

	// Listen back for the objective value
	var ans float64
	err = r.dec.Decode(&ans)
	if err != nil {
		panic(err)
	}
	return ans
}

func (r *Remote) Result() {
	r.conn.Close()
}

// RemoteReceiver is the other end of the Remote objective function
type RemoteReceiver struct {
	Port string // Where is the request going to

	obj Objer

	conn net.Conn
}

func (r *RemoteReceiver) Do() {
	// Establish the connection and get the objective function
	l, err := net.Listen("tcp", r.Port)
	if err != nil {
		panic(err)
	}
	// Wait for a connection
	conn, err := l.Accept()
	if err != nil {
		panic(err)
	}
	r.conn = conn

	//size := 5000

	/*
		// Wait for the objer
		buf := make([]byte, 5000)
		n, err := conn.Read(buf)
		if err != nil {
			panic(err)
		}
		buf = buf[0:n]
	*/

	// Deserialize the objer
	enc := gob.NewEncoder(conn)
	dec := gob.NewDecoder(conn)
	err = dec.Decode(&r.obj)
	if err != nil {
		panic(err)
	}

	x := make([]float64, 0)
	// Now, continue waiting to receive new locations
	for {
		// Reexpand buffer
		//buf = buf[0:size]

		// Read the new location
		err = dec.Decode(&x)
		if err != nil {
			panic(err)
		}

		fmt.Println("x is ", x)

		ans := r.obj.Obj(x)
		err = enc.Encode(ans)
		if err != nil {
			panic(err)
		}
	}
}
