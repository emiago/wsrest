package wsrest

import (
	"fmt"
	"sort"
)

type FastRouter struct {
	routes map[string]*Route
}

func NewRouter() *FastRouter {
	r := &FastRouter{
		routes: make(map[string]*Route),
	}

	return r
}

func (r *FastRouter) HandleFunc(regex string, handler RouteHandlerFn) *Route {
	route := &Route{
		path:    regex,
		handler: handler,
	}

	r.routes[route.path] = route
	return route
}

func (r *FastRouter) String() string {
	result := ""

	sorted := make(RoutesSorted, 0, len(r.routes))
	for _, rt := range r.routes {
		sorted = append(sorted, rt)
	}
	sort.Sort(sort.Reverse(sorted))

	for _, rt := range sorted {
		result += fmt.Sprintf("%s\t%s\n", rt.path, rt.method)
	}
	return result
}

func (r *FastRouter) Match(path string, method string) (*Route, bool) {
	route, exists := r.routes[path]
	if !exists {
		return nil, false
	}

	if route.method != "" && route.method != method {
		return nil, false
	}

	return route, exists
}
