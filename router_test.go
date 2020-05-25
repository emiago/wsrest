package wsrest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testHandler(wsc *Conn, m *Request) {}

func TestRouteAdd(t *testing.T) {
	router := NewRouter()

	routes := []string{"/go", "/java", "/python"}

	for _, r := range routes {
		router.HandleFunc(r, testHandler)
	}

	assert.Equal(t, len(router.routes), len(routes), "Len routes are not same")
}

func TestRouterMatch(t *testing.T) {
	router := NewRouter()

	routes := []string{
		"/go",
		"/go/is/best",
		"/go/is/very/fast",
		"/java/is/slow",
		"/python",
		"/python/likes/memory",
	}

	for _, r := range routes {
		router.HandleFunc(r, testHandler)
	}

	for _, r := range routes {
		if _, exists := router.Match(r, ""); !exists {
			t.Fatalf("Route %s does not exists", r)
		}
	}
}
