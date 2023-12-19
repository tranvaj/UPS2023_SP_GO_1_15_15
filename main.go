package main

import (
	"fmt"
	"net"
	"os"

	"github.com/tranvaj/UPS2023_SP_GO_1_15_15/util"
)

func main() {
	fmt.Println("Starting " + util.ConnType + " server on " + util.ConnHost + ":" + util.ConnPort)
	l, err := net.Listen(util.ConnType, util.ConnHost+":"+util.ConnPort)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("Error connecting:", err.Error())
			return
		}
		fmt.Println("Client connected.")

		fmt.Println("Client " + c.RemoteAddr().String() + " connected.")

		go util.ProcessClient(c)
	}
}
