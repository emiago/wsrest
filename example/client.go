package main

import (
	"fmt"
	"wsrest"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	fmt.Println("Running client")
	c := wsrest.NewClient()
	// c.SetLog(logrus.StandardLogger().WithField("test", "test"))
	if err := c.Connect("ws://127.0.0.1:3000/ws", ReadHandler); err != nil {
		fmt.Println("Fail to connect", err)
		return
	}
	defer c.Close()

	res, err := c.Execute("GET", "/help", "")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(res.Data))
}

func ReadHandler(d []byte) {
	fmt.Println("Got data", string(d))
}
