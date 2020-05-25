package wsrest

// "golang.org/x/net/websocket"

type RouteHandlerFn func(*Conn, *Request)

type Router interface {
	Match(path string, method string) (*Route, bool)
}

type Route struct {
	method  string
	path    string
	handler RouteHandlerFn
}

func (r *Route) Method(m string) *Route {
	r.method = m
	return r
}

func (r *Route) Run(wsc *Conn, m *Request) {
	r.handler(wsc, m)
}

type RoutesSorted []*Route

func (a RoutesSorted) Len() int           { return len(a) }
func (a RoutesSorted) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a RoutesSorted) Less(i, j int) bool { return a[i].path > a[j].path }
