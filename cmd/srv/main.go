package main
import (
	"net"
	"net/rpc"
	"net/http"
	"log"
	"fmt"
	"github.com/upamanyu/gokv"
)

func main() {
	srv := gokv.MakeGoKVServer()

	rpc.RegisterName("KV", srv)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":12345")
	if e != nil {
		panic(e)
	}

	fmt.Println("Starting server")
	// go http.Serve(l, nil)
	log.Fatal(http.Serve(l, nil))
}
