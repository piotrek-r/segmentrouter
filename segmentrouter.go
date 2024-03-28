package segmentrouter

import (
	"context"
	"net/http"
	"strings"
)

const RouteNameParam = "__route__"

// CreateHttpHandler creates an http.HandlerFunc that routes requests to the appropriate handler
// based on the request method and path.
func CreateHttpHandler(router SegmentRouter, defaultHandler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ok, handler, params := router.Match(r.Method, r.URL.Path)
		if !ok {
			defaultHandler(w, r)
			return
		}

		handler(w, r.WithContext(context.WithValue(r.Context(), "params", params)))
	}
}

// Parameters is a map of parameters extracted from the request path.
type Parameters map[string]string

// SegmentRouter is the main router struct that contains the segments to match against.
type SegmentRouter struct {
	Segments []Segment
}

// Match matches the request method and path against the segments in the router.
func (r *SegmentRouter) Match(method, path string) (bool, http.HandlerFunc, Parameters) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for _, segment := range r.Segments {
		if ok, handler, params := segment.Match(method, parts[0], parts[1:], Parameters{}); ok {
			return true, handler, params
		}
	}
	return false, nil, Parameters{}
}

// Segment is an interface that is used for different types of segments in the router.
type Segment interface {
	Match(string, string, []string, Parameters) (bool, http.HandlerFunc, Parameters)
}

// StaticSegment is a segment that matches a static path segment.
type StaticSegment struct {
	RouteName   string
	Value       string
	Handlers    map[string]http.HandlerFunc
	SubSegments []Segment
}

// Match matches the request method and path against the static segment.
func (s StaticSegment) Match(method, value string, next []string, params Parameters) (bool, http.HandlerFunc, Parameters) {
	if s.Value == value {
		if len(next) == 0 {
			return matchMethodWithRouteName(s.RouteName, method, s.Handlers, params)
		}
		return matchSubsegments(method, next[0], next[1:], s.SubSegments, params)
	}
	return false, nil, params
}

// ParamSegment is a segment that matches a path segment and puts its value in the parameters.
type ParamSegment struct {
	RouteName   string
	ParamName   string
	Handlers    map[string]http.HandlerFunc
	SubSegments []Segment
}

// Match matches the request method and path against the param segment.
func (s ParamSegment) Match(method, value string, next []string, params Parameters) (bool, http.HandlerFunc, Parameters) {
	if s.ParamName != "" {
		params[s.ParamName] = value
	}
	if len(next) == 0 {
		return matchMethodWithRouteName(s.RouteName, method, s.Handlers, params)
	}
	return matchSubsegments(method, next[0], next[1:], s.SubSegments, params)
}

// matchMethodWithRouteName matches the request method against the handlers by methods using matchMethod and adds the route name to Parameters.
func matchMethodWithRouteName(routeName, method string, handlers map[string]http.HandlerFunc, params Parameters) (bool, http.HandlerFunc, Parameters) {
	ok, handler, params := matchMethod(method, handlers, params)
	if ok {
		params[RouteNameParam] = routeName
	}
	return ok, handler, params
}

// matchMethod matches the request method against the handlers mapped by methods.
func matchMethod(method string, handlers map[string]http.HandlerFunc, params Parameters) (bool, http.HandlerFunc, Parameters) {
	if handler, ok := handlers[method]; ok {
		return true, handler, params
	}
	if handler, ok := handlers["*"]; ok {
		return true, handler, params
	}
	return false, nil, params
}

// matchSubsegments matches the request method and path against the subsegments.
func matchSubsegments(method, value string, next []string, segments []Segment, params Parameters) (bool, http.HandlerFunc, Parameters) {
	for _, segment := range segments {
		if ok, handler, params := segment.Match(method, value, next, params); ok {
			return true, handler, params
		}
	}
	return false, nil, params
}
