package main

import (
	"fmt"
	"wsrest"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	fmt.Println("Running client")
	c, err := wsrest.Dial("ws://127.0.0.1:3000/ws", ReadHandler)
	// c.SetLog(logrus.StandardLogger().WithField("test", "test"))
	if err != nil {
		fmt.Println("Fail to connect", err)
		return
	}
	defer c.Close()

	res, err := c.Get("/help", "")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(res.Data))
}

func ReadHandler(d []byte) {
	fmt.Println("Got data", string(d))
}
