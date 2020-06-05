package wsrest

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"
	"wsrest/datastream"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type ReadHandler func([]byte)
type ServerCloseHandler func(err error)

type Client struct {
	mutex          sync.RWMutex
	conn           *websocket.Conn
	callbacks      map[string]chan *Request
	RequestTimeout time.Duration
	log            logrus.StdLogger
	fnServerClose  ServerCloseHandler
	closed         chan struct{}
	Marshaler      datastream.Marshaler
	MessageType    int
}

func Dial(wsurl string, eventHandler ReadHandler) (*Client, error) {
	c := &Client{
		conn:           nil,
		RequestTimeout: time.Second * 10,
		callbacks:      make(map[string]chan *Request),
		log:            logrus.New(),
		Marshaler:      &datastream.JSONMarshaler{},
		MessageType:    websocket.TextMessage,
	}

	return c, c.connect(wsurl, eventHandler)
}

func (c *Client) SetLog(l logrus.StdLogger) {
	c.log = l
}

func (c *Client) connect(wsurl string, eventHandler ReadHandler) error {
	u, err := url.Parse(wsurl)
	if err != nil {
		return err
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	c.conn = conn
	c.closed = make(chan struct{})

	if eventHandler == nil {
		eventHandler = func(d []byte) { return }
	}

	go c.readMessage(eventHandler)
	return nil
	// defer conn.Close()
}

func (c *Client) Close() {
	//Because we are closing connection prevent calling server close error
	c.mutex.Lock()
	c.fnServerClose = nil

	err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		c.log.Printf("write close: %s\n0", err)
		return
	}
	c.mutex.Unlock()

	c.log.Println("Closing connection")
	select {
	case <-c.closed:
	case <-time.After(time.Second):
	}

	c.mutex.Lock()
	c.conn.Close()
	c.conn = nil
	c.mutex.Unlock()
}

func (c *Client) SetServerCloseHandler(fn ServerCloseHandler) {
	c.fnServerClose = fn
}

func (c *Client) readMessage(readh ReadHandler) {
	var closeErr error
	defer func() {
		close(c.closed)
		if closeErr != nil {
			if !websocket.IsCloseError(closeErr, websocket.CloseNormalClosure) {
				c.log.Printf("Reading stopped. err=%s\n", closeErr)
			}

			c.mutex.RLock()
			fnServerClose := c.fnServerClose
			c.mutex.RUnlock()
			if fnServerClose != nil {
				//Only call if is Server close
				fnServerClose(closeErr)
			}
		}
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			closeErr = err
			return
		}

		if len(message) == 0 {
			c.log.Println("received empty message")
			continue
		}

		m := &Request{}
		err = json.Unmarshal(message, m)

		if _, ok := err.(*json.UnmarshalTypeError); err != nil && !ok {
			c.log.Printf("UnmarshalTypeError err=%s\n", err)
			continue
		}

		if err == nil {
			if callback, exists := c.getRequestCallback(m.GetUID()); exists {
				callback <- m
				c.removeRequestCallback(m.GetUID())
				continue
			}
		}

		readh(message)
	}
}

func (c *Client) getRequestCallback(RequestId string) (chan *Request, bool) {
	c.mutex.RLock()
	callback, exists := c.callbacks[RequestId]
	c.mutex.RUnlock()
	return callback, exists
}

func (c *Client) addRequestCallback(RequestId string, syncer chan *Request) {
	c.mutex.Lock()
	c.callbacks[RequestId] = syncer
	c.mutex.Unlock()
}

func (c *Client) removeRequestCallback(RequestId string) bool {
	c.mutex.Lock()
	_, exists := c.callbacks[RequestId]
	delete(c.callbacks, RequestId)
	c.mutex.Unlock()
	return exists
}

func (c *Client) exec(cr *Request) error {
	data, err := c.Marshaler.Marshal(cr)
	if err != nil {
		return err
	}

	//Do not allow concurent writes
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conn == nil {
		return fmt.Errorf("No available websocket connection")
	}

	err = c.conn.WriteMessage(c.MessageType, data)
	return err
}

func (c *Client) Do(m *Request) (*Request, error) {
	syncer := make(chan *Request)
	c.addRequestCallback(m.GetUID(), syncer)

	err := c.exec(m)
	if err != nil {
		c.removeRequestCallback(m.GetUID())
		return nil, err
	}

	select {
	case res := <-syncer:
		return res, nil
	case <-time.After(c.RequestTimeout):
		c.removeRequestCallback(m.GetUID())
		return nil, fmt.Errorf("Timeout occured")
	}
}

func (c *Client) Execute(method string, resource string, data interface{}) (*Request, error) {
	m, err := NewRequest(method, resource, data)
	if err != nil {
		return nil, err
	}
	return c.Do(m)
}

func (c *Client) Get(resource string, data interface{}) (*Request, error) {
	return c.Execute("GET", resource, data)
}

func (c *Client) Post(resource string, data interface{}) (*Request, error) {
	return c.Execute("POST", resource, data)
}

func (c *Client) Put(resource string, data interface{}) (*Request, error) {
	return c.Execute("PUT", resource, data)
}

func (c *Client) Delete(resource string, data interface{}) (*Request, error) {
	return c.Execute("DELETE", resource, data)
}
