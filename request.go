package wsrest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	uuid "github.com/satori/go.uuid"
)

type HttpCode int

type Messenger interface {
	GetMethod() string
	GetResource() string
	SetCode(HttpCode)
}

type Request struct {
	RequestId string   `json:"requestid"`
	Method    string   `json:"method"`
	Resource  string   `json:"resource"`
	Code      HttpCode `json:"code"`
	Data      []byte   `json:"data,ommitempty"`
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

	cr := &Request{
		RequestId: uuid.NewV4().String(),
		Method:    method,
		Resource:  resource,
		Data:      jsond,
	}

	return cr, nil
}

func ParseHttpRequest(r *http.Request) (*Request, error) {
	data := json.RawMessage{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil && err != io.EOF {
		return nil, err
	}
	defer r.Body.Close()

	m := &Request{
		Method:   r.Method,
		Resource: r.RequestURI,
		Data:     data,
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

func (cr *Request) GetCode() HttpCode {
	return cr.Code
}

func (cr *Request) SetCode(c HttpCode) {
	cr.Code = c
}

func (cr *Request) MarshalData(data interface{}) error {
	d, err := json.Marshal(data)
	if err != nil {
		return err
	}
	cr.Data = d
	return nil
}

func (cr *Request) UnmarshalData(v interface{}) error {
	if cr.Data == nil {
		return fmt.Errorf("Trying to unmarshal emtpy data")
	}

	return json.Unmarshal(cr.Data, v)
}

func (cr *Request) DataToString() string {
	defer func() {
		recover()
	}()

	var s *string = nil
	if err := json.Unmarshal(cr.Data, &s); err != nil {
		return string(cr.Data)
	}
	return *s
}
