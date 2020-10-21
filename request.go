package wsrest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	uuid "github.com/satori/go.uuid"
)

type IRequest interface {
	GetUID() string
	GetMethod() string
	GetResource() string
	GetData() []byte
	SetData([]byte)
	SetCode(int)
	GetCode() int
}

type Request struct {
	UID      string           `json:"uid"`
	Method   string           `json:"m"`
	Resource string           `json:"r"`
	Code     int              `json:"c"`
	Data     *json.RawMessage `json:"d"`
}

func NewRequest(method string, resource string, data interface{}) (*Request, error) {
	var jsond []byte
	if data != nil {
		switch data.(type) {
		case string:
			dataj := data.(string)
			if dataj == "" {
				dataj = `{}`
			}
			jsond = []byte(dataj)
		default:
			dataj, err := json.Marshal(data)
			if err != nil {
				return nil, err
			}
			jsond = dataj
		}

	}

	raw := json.RawMessage(jsond)

	cr := &Request{
		UID:      uuid.NewV4().String(),
		Method:   method,
		Resource: resource,
		Data:     &raw,
	}

	return cr, nil
}

func ParseHttpRequest(r *http.Request) (*Request, error) {
	data := json.RawMessage{}
	// var data []byte
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil && err != io.EOF {
		return nil, err
	}
	defer r.Body.Close()

	m := &Request{
		Method:   r.Method,
		Resource: r.RequestURI,
		Data:     &data,
	}

	return m, nil
}

type SimpleMessage struct {
	Message string `json:"message"`
}

func SimpleMsg(m string, args ...interface{}) SimpleMessage {
	if len(args) > 0 {
		m = fmt.Sprintf(m, args...)
	}

	return SimpleMessage{
		Message: m,
	}
}

func (cr *Request) GetUID() string {
	return cr.Method
}

func (cr *Request) GetMethod() string {
	return cr.Method
}

func (cr *Request) GetResource() string {
	return cr.Resource
}

func (cr *Request) GetPath() string {
	u, err := url.Parse(cr.Resource)
	if err != nil {
		return cr.Resource
	}

	if len(u.RawPath) > 0 {
		return u.RawPath
	}
	return u.Path
}

func (cr *Request) GetCode() int {
	return cr.Code
}

func (cr *Request) SetCode(c int) {
	cr.Code = c
}

func (cr *Request) GetData() []byte {
	if cr.Data == nil {
		return []byte{}
	}

	return *cr.Data
} 

func (cr *Request) SetData(d []byte) {
	p := json.RawMessage(d)
	cr.Data = &p
}
