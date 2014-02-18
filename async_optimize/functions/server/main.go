package main

import (
	"flag"
	"math/rand"
	"time"

	"github.com/btracey/goexamples/async_optimize/functions"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	var port string
	flag.StringVar(&port, "port", "", "tcp port on which to listen")
	flag.Parse()
	receive := &functions.RemoteReceiver{Port: port}

	receive.Do()
}
