package main

import (
	"flag"

	"github.com/btracey/goexamples/async_optimize/functions"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "", "tcp port on which to listen")
	flag.Parse()
	receive := &functions.RemoteReceiver{Port: port}

	receive.Do()
}

/*
func main() {
	// Listen on TCP port 2000 on all interfaces.
	l, err := net.Listen("tcp", ":2000")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go func(c net.Conn) {
			// Read and write won't append to the buffer,
			// so need to make it big enough to read
			buf := make([]byte, 512)
			n, err := c.Read(buf)
			if err != nil {
				fmt.Println("Error reading:", err)
			}
			fmt.Println(buf[0:n])
			_, err = c.Write(buf[0:n])
			if err != nil {
				fmt.Println(err)
			}
			c.Close()
		}(conn)
	}
}
*/
