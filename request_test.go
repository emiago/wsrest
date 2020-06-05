package wsrest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type httprequest struct {
	Method   string
	Resource string
	Code     int
}

func TestParseRequest(t *testing.T) {
	b := bytes.NewBuffer([]byte("{}"))

	testcases := []httprequest{
		{Method: "GET", Resource: "/go", Code: http.StatusOK},
		{Method: "POST", Resource: "/java/feature?api=1234", Code: http.StatusBadRequest},
		{Method: "DELETE", Resource: "/python/test?myparam=test", Code: http.StatusAccepted},
		{Method: "PUT", Resource: "/c/more/editors?myparam=test&secondparam=1020", Code: http.StatusNotFound},
		{Method: "ANYTHING", Resource: "/i/could/handle/anything?myparam=test", Code: http.StatusBadGateway},
	}

	for _, req := range testcases {
		r := httptest.NewRequest(req.Method, req.Resource, b)
		m, err := ParseHttpRequest(r)
		require.Nil(t, err)
		m.SetCode(req.Code)

		assert.Equal(t, req.Method, m.GetMethod())
		assert.Equal(t, req.Resource, m.GetResource())
		assert.Equal(t, req.Code, m.GetCode())
	}
}

func TestParseRequestNotJSON(t *testing.T) {
	b := bytes.NewBuffer([]byte("Some string or"))
	r := httptest.NewRequest("GET", "/go", b)
	_, err := ParseHttpRequest(r)
	require.NotNil(t, err)
}

type dummyJSON struct {
	Stringer string
	Integer  int
	Floater  float64
}

func TestRequestSetData(t *testing.T) {
	testdata := dummyJSON{
		Stringer: "string",
		Integer:  755,
		Floater:  3.14,
	}

	m := &Request{
		Method:   "GET",
		Resource: "/golang",
	}

	marshaled, err := json.Marshal(testdata)
	require.Nil(t, err)

	t.Run("Marshal", func(t *testing.T) {
		d, err := json.Marshal(testdata)
		require.Nil(t, err)
		m.SetData(d)

		assert.Equal(t, marshaled, m.Data, "Marshaling not equal")
	})

	t.Run("Unmarshal", func(t *testing.T) {
		out := dummyJSON{}
		err := json.Unmarshal(marshaled, &out)
		require.Nil(t, err)

		assert.Equal(t, testdata, out, "Unmarshaling not equal")
	})
}
