package main

import (
	"fmt"
	"net"
	"net/http"

	"wsrest"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var router *wsrest.FastRouter
var upgrader = websocket.Upgrader{}

func wsServerStart(address string, port int) {
	router = wsrest.NewRouter()

	//Dump all routes from our router
	router.HandleFunc("/help", func(wsc *wsrest.Conn, m *wsrest.Request) {
		wsc.Respond(m, wsrest.SimpleMsg(router.String()), http.StatusOK)
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/", restHandler)
	mux.HandleFunc("/ws", wsHandler)

	//To listen on tcp4
	l, err := net.Listen("tcp4", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.WithField("addr", address).Info("Server listening")
	logrus.Fatal(http.Serve(l, mux))
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	wsc, err := wsrest.NewConnWS(w, r, router)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer wsc.Close()

	wsc.Log.Info("New ws client connected")
	defer wsc.Log.Info("Client ws disconected")
	wsc.HandleWSConnection()
}

/*
	To accomplish restAPI we will convert http request to our internal ClientRequest object, and then
	we will procedee like we are doing for web socket connection

*/
func restHandler(w http.ResponseWriter, r *http.Request) {
	wsc, err := wsrest.NewConnRest(w, r, router)
	if err != nil {
		logrus.Error(err)
		return
	}

	wsc.Log.Info("New rest client connected")
	defer wsc.Log.Info("Client rest disconected")
	wsc.HandleRestConnection()
}

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	wsServerStart("127.0.0.1", 3000)
}
