package wsrest

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// type WSWritter struct {
// 	*httptest.ResponseRecorder
// }

// type recorder struct {
// 	httptest.ResponseRecorder
// 	server net.Conn
// }

// // runServer reads the request sent on the connection to the recorder
// // from the websocket.NewDialer.Dial function, and pass it to the recorder.
// // once this is done, the communication is done on the wsConn
// func (r *recorder) runServer(h http.Handler) {
// 	// read from the recorder connection the request sent by the recorder.Dial,
// 	// and use the handler to serve this request.
// 	req, err := http.ReadRequest(bufio.NewReader(r.server))
// 	if err != nil {
// 		return
// 	}
// 	h.ServeHTTP(r, req)
// }

// // Hijack the connection
// func (r *recorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
// 	// return to the recorder the recorder, which is the recorder side of the connection
// 	rw := bufio.NewReadWriter(bufio.NewReader(r.server), bufio.NewWriter(r.server))
// 	return r.server, rw, nil
// }

// // WriteHeader write HTTP header to the client and closes the connection
// func (r *recorder) WriteHeader(code int) {
// 	resp := http.Response{StatusCode: code, Header: r.Header()}
// 	resp.Write(r.server)
// }

// func NewDialer(h http.Handler) *websocket.Dialer {
// 	client, server := net.Pipe()
// 	conn := &recorder{server: server}

// 	// run the runServer in a goroutine, so when the Dial send the request to
// 	// the recorder on the connection, it will be parsed as an HTTPRequest and
// 	// sent to the Handler function.
// 	go conn.runServer(h)

// 	// use the websocket.NewDialer.Dial with the fake net.recorder to communicate with the recorder
// 	// the recorder gets the client which is the client side of the connection
// 	return &websocket.Dialer{NetDial: func(network, addr string) (net.Conn, error) { return client, nil }}
// }

type restHandler struct {
	router Router
}

func (h *restHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logrus.Println("New Rest connection")
	conn, err := NewConnRest(w, r, h.router)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	conn.HandleRestConnection()
}

type wsHandler struct {
	router Router
}

func (h *wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logrus.Println("New WS connection")
	conn, err := NewConnWS(w, r, h.router)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	conn.HandleWSConnection()
}

func TestRequestResponse(t *testing.T) {
	router := NewRouter()
	logrus.SetLevel(logrus.DebugLevel)

	h := http.NewServeMux()
	h.Handle("/", &restHandler{router: router})
	h.Handle("/ws", &wsHandler{router: router})

	response := SimpleMsg("Hello")
	router.HandleFunc("/go", func(c *Conn, m *Request) {
		t.Log("Responding...")
		c.Respond(m, response, http.StatusOK)
	})

	server := httptest.NewServer(h)
	domain := server.URL

	t.Run("StandardHTTP", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", domain+"/go", nil)
		require.Nil(t, err)

		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		require.Nil(t, err)

		require.Equal(t, resp.StatusCode, http.StatusOK)

		res, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		require.Nil(t, err)
		t.Log("Result", string(res))
		assert.Equal(t, string(res), `{"message":"Hello"}`)
	})

	t.Run("OverWebsocket", func(t *testing.T) {
		client := NewClient()
		req, err := NewRequest("GET", domain+"/ws/go", nil)
		require.Nil(t, err)

		// req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		require.Nil(t, err)

		require.Equal(t, resp.Code, http.StatusOK)

		res := resp.Data
		t.Log("Result", string(res))
		assert.Equal(t, string(res), `{"message":"Hello"}`)
	})

	// decode := json.NewDecoder(resp.Body)

	// out := SimpleMessage{}
	// err = decode.Decode(&out)
	// require.Nil(t, err)

	// assert.Equal(t, out, response)

	//Make websocket call
}
