package wsrest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
	"wsrest/datastream"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type ConnCloseHandlerFn func(wsc *Conn)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 102400
)

type Conn struct {
	mutex          sync.RWMutex
	C              *websocket.Conn
	W              http.ResponseWriter
	R              *http.Request
	SendCh         chan []byte
	StopCh         chan bool
	Router         Router
	Vars           map[string]interface{}
	Closed         bool
	MaxMessageSize int64
	Log            logrus.StdLogger
	CloseHandlers  []ConnCloseHandlerFn
	marshaler      datastream.Marshaler
}

func (wsc *Conn) Lock() {
	wsc.mutex.Lock()
}

func (wsc *Conn) Unlock() {
	wsc.mutex.Unlock()
}

func (wsc *Conn) RLock() {
	wsc.mutex.RLock()
}

func (wsc *Conn) RUnlock() {
	wsc.mutex.RUnlock()
}

func (wsc *Conn) IsClosed() bool {
	wsc.RLock()
	defer wsc.RUnlock()
	return wsc.Closed
}

func constructConn() *Conn {
	wsc := &Conn{
		C:              nil, //Websocket connecting
		W:              nil, //Http writter
		R:              nil, //Http request
		SendCh:         make(chan []byte),
		StopCh:         make(chan bool),
		Vars:           make(map[string]interface{}),
		MaxMessageSize: 102400,
		Router:         NewRouter(),
		marshaler:      &datastream.JSONMarshaler{},
	}

	return wsc
}

func NewConnWS(w http.ResponseWriter, r *http.Request, router Router) (*Conn, error) {
	wsc := constructConn()
	u, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	wsc.C = u
	wsc.Router = router

	wsc.Log = logrus.New().WithFields(logrus.Fields{
		"conn": fmt.Sprintf("WSCONN[%s]", wsc.GetRemoteAddr()),
	})

	return wsc, nil
}

func NewConnRest(w http.ResponseWriter, r *http.Request, router Router) (*Conn, error) {
	wsc := constructConn()
	wsc.W = w
	wsc.R = r
	wsc.Router = router

	wsc.Log = logrus.New().WithFields(logrus.Fields{
		"conn": fmt.Sprintf("RESTCONN[%s]", wsc.GetRemoteAddr()),
	})

	return wsc, nil
}

func (wsc *Conn) HandleWSConnection() {
	go wsc.writePump()
	wsc.readPump()
}

func (wsc *Conn) HandleRestConnection() {
	r := wsc.R

	m, err := ParseHttpRequest(r)
	if err != nil {
		wsc.Log.Printf("Failed to parse request err=%s\n", err)
		return
	}

	route, found := wsc.Router.Match(m.GetPath(), m.GetMethod())
	if !found {
		wsc.Respond(m, SimpleMsg("Resource not found"), http.StatusNotFound)
		return
	}

	route.Run(wsc, m)
}

func (wsc *Conn) SetLogger(l logrus.FieldLogger) {
	wsc.Log = l
}

func (wsc *Conn) SetVar(name string, value interface{}) {
	wsc.Lock()
	defer wsc.Unlock()
	wsc.Vars[name] = value
}

func (wsc *Conn) GetVar(name string) (value interface{}, exists bool) {
	wsc.RLock()
	defer wsc.RUnlock()
	value, exists = wsc.Vars[name]
	return
}

func (wsc *Conn) DelVar(name string) {
	wsc.Lock()
	defer wsc.Unlock()
	delete(wsc.Vars, name)
}

func (wsc *Conn) AddCloseHandler(fn ConnCloseHandlerFn) {
	wsc.CloseHandlers = append(wsc.CloseHandlers, fn)
}

func (wsc *Conn) RunCloseHandlers() {
	for _, fn := range wsc.CloseHandlers {
		fn(wsc)
	}
}

func (wsc *Conn) WriteJSON(v interface{}) error {
	if wsc.C != nil {
		return wsc.C.WriteJSON(v)
	}

	m, err := json.Marshal(v)
	if err != nil {
		return err
	}

	_, werr := wsc.W.Write(m)
	return werr
}

func (wsc *Conn) WriteWS(data []byte) error {
	select {
	case <-wsc.StopCh:
		return fmt.Errorf("Connection is closed while trying to write data")
	case wsc.SendCh <- data:
	}
	return nil
}

func (wsc *Conn) Respond(m *Request, response interface{}, status int) {
	m.SetCode(status)

	rdata, err := wsc.marshaler.Marshal(response)
	if err != nil {
		wsc.Log.Printf("Failed to marshal response uid=%s err=%s\n", m.GetUID(), err)
		return
	}
	m.SetData(rdata)

	if wsc.C != nil {
		rdata, err := wsc.marshaler.Marshal(m)
		if err != nil {
			wsc.Log.Printf("Failed to marshal request uid=%s err=%s\n", m.GetUID(), err)
			return
		}

		//This will responded by write pump, WE CAN NOT HAVE CONCURENT WRITES
		if err := wsc.WriteWS(rdata); err != nil {
			wsc.Log.Printf("Connection is closed. Failed to response request uid=%s err=%s\n", m.GetUID(), err)
		}
		return
	}

	if wsc.R.Response != nil {
		wsc.R.Response.StatusCode = int(status)
	}

	wsc.W.WriteHeader(int(status))
	if _, err := wsc.W.Write(rdata); err != nil {
		wsc.Log.Printf("err: %s\n", err)
		return
	}
}

//This should be used when there are multiple responses on same  request
func (wsc *Conn) RespondMultiple(m *Request, response interface{}, status int) {
	mcopy := *m
	wsc.Respond(&mcopy, response, status)
}

func (wsc *Conn) GetRemoteAddr() string {
	if wsc.C != nil {
		return wsc.C.RemoteAddr().String()
	}

	return wsc.R.RemoteAddr
}

func (wsc *Conn) Close() {
	wsc.Lock()
	wsc.Closed = true
	wsc.C.Close()
	wsc.Unlock()
	wsc.RunCloseHandlers()
	wsc.Log.Println("Connection is closed")
}

func (wsc *Conn) readPump() {
	defer func() {
		// Before we exit we need to shutdown and write pump channel. Write pump channel should then stopp all seneders
		wsc.Log.Println("Read routine closed. Closing send channel....")
		close(wsc.StopCh) //Close senders
		// close(wsc.sendCh) //Close write pump
	}()
	// wsc.C.SetReadLimit(wsc.MaxMessageSize)
	wsc.C.SetReadDeadline(time.Now().Add(pongWait))
	wsc.C.SetPongHandler(func(string) error { wsc.C.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := wsc.C.ReadMessage()
		if err != nil {
			wsc.Log.Printf("closed upon trying to read message. err=%s Exiting ...\n", err)
			break
		}

		m := &Request{}
		if err := json.Unmarshal(message, m); err != nil {
			wsc.Log.Printf("Unmarshal request failed. err=%s\n", err)
			continue
		}

		path := m.GetPath()
		method := m.GetMethod()
		wsc.Log.Printf("Mathing request. path=%s method=%s\n", path, method)

		route, found := wsc.Router.Match(path, method)
		if !found {
			wsc.Respond(m, SimpleMsg("Resource not found"), http.StatusNotFound)
			continue
		}

		//TODO MAX 8192 go routines
		go route.Run(wsc, m)
	}
}

func (wsc *Conn) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		wsc.C.Close()
		wsc.Log.Println("Write routine closed.")
	}()

	wsc.Log.Println("Write routine started")
	for {
		select {
		case <-wsc.StopCh:
			//we are stoppend any further processing message is not going to be publihsed
			return

		case message, ok := <-wsc.SendCh:
			wsc.C.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				wsc.Log.Println("Send channel is closed, trying to notify ")
				wsc.C.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := wsc.C.WriteMessage(websocket.TextMessage, message); err != nil {
				wsc.Log.Printf("Write err , exiting. err = %s\n", err)
				return
			}

		case <-ticker.C:
			wsc.C.SetWriteDeadline(time.Now().Add(writeWait))
			if err := wsc.C.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
