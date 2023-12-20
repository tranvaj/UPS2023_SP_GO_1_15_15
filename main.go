package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/tranvaj/UPS2023_SP_GO/util"
)

func main() {
	var clientId int = 1

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
		player := &util.Player{Conn: &c, ClientId: clientId, TimeSinceLastPing: time.Now()}
		//fmt.Println("Client connected.")
		fmt.Println("Client " + c.RemoteAddr().String() + " connected.")
		clientId++
		//go util.ConnectionCloseHandler(player)
		go util.ProcessClient(c, player)
	}
}
