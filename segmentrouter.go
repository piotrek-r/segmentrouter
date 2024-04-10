package segmentrouter

import (
	"context"
	"net/http"
	"strings"
)

type RouterResult int

const (
	ContextParameters   = "__params__"
	ContextRouterResult = "__router_result__"
	ParamRouteName      = "__route__"
)

const (
	RouterResultFound RouterResult = iota
	RouterResultPathNotFound
	RouterResultMethodNotAllowed
)

// Parameters is a map of parameters extracted from the request path.
type Parameters map[string]string

// SegmentRouter is the main router struct that contains the segments to match against.
type SegmentRouter struct {
	Segments        []Segment
	FallbackHandler http.HandlerFunc
}

// Match matches the request method and path against the segments in the router.
func (sr SegmentRouter) Match(r *http.Request) (RouterResult, http.HandlerFunc, Parameters) {
	path := r.URL.Path
	method := r.Method

	parts := strings.Split(strings.Trim(path, "/"), "/")

	for _, segment := range sr.Segments {
		result, handler, params := segment.Match(method, parts[0], parts[1:], Parameters{})
		switch result {
		case RouterResultFound:
			return result, handler, params
		case RouterResultMethodNotAllowed:
			return result, sr.FallbackHandler, params
		default:
			continue
		}
	}

	return RouterResultPathNotFound, sr.FallbackHandler, Parameters{}
}

func (sr SegmentRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result, handler, params := sr.Match(r)

	ctx := r.Context()
	ctx = context.WithValue(ctx, ContextParameters, params)
	ctx = context.WithValue(ctx, ContextRouterResult, result)

	if handler == nil {
		handler = sr.FallbackHandler
	}
	handler(w, r.WithContext(ctx))
}

// Segment is an interface that is used for different types of segments in the router.
type Segment interface {
	Match(string, string, []string, Parameters) (RouterResult, http.HandlerFunc, Parameters)
}

// StaticSegment is a segment that matches a static path segment.
type StaticSegment struct {
	RouteName   string
	Value       string
	Handlers    map[string]http.HandlerFunc
	SubSegments []Segment
}

// Match matches the request method and path against the static segment.
func (s StaticSegment) Match(method, value string, next []string, params Parameters) (RouterResult, http.HandlerFunc, Parameters) {
	if s.Value == value {
		if len(next) == 0 {
			return matchMethodWithRouteName(s.RouteName, method, s.Handlers, params)
		}
		return matchSubsegments(method, next[0], next[1:], s.SubSegments, params)
	}
	return RouterResultPathNotFound, nil, params
}

// ParamSegment is a segment that matches a path segment and puts its value in the Parameters.
type ParamSegment struct {
	RouteName   string
	ParamName   string
	Handlers    map[string]http.HandlerFunc
	SubSegments []Segment
}

// Match matches the request method and path against the ParamSegment.
func (s ParamSegment) Match(method, value string, next []string, params Parameters) (RouterResult, http.HandlerFunc, Parameters) {
	if s.ParamName != "" {
		params[s.ParamName] = value
	}
	if len(next) == 0 {
		return matchMethodWithRouteName(s.RouteName, method, s.Handlers, params)
	}
	return matchSubsegments(method, next[0], next[1:], s.SubSegments, params)
}

// GetParametersFromContext returns the Parameters from the context.
func GetParametersFromContext(ctx context.Context) Parameters {
	if params, ok := ctx.Value(ContextParameters).(Parameters); ok {
		return params
	}
	return Parameters{}
}

// GetRouterResultFromContext returns the RouterResult from the context.
func GetRouterResultFromContext(ctx context.Context) RouterResult {
	if result, ok := ctx.Value(ContextRouterResult).(RouterResult); ok {
		return result
	}
	return RouterResultPathNotFound
}

// matchMethodWithRouteName matches the request method against the handlers by methods using matchMethod and adds the route name to Parameters.
func matchMethodWithRouteName(routeName, method string, handlers map[string]http.HandlerFunc, params Parameters) (RouterResult, http.HandlerFunc, Parameters) {
	result, handler, params := matchMethod(method, handlers, params)
	if result == RouterResultFound && routeName != "" {
		params[ParamRouteName] = routeName
	}
	return result, handler, params
}

// matchMethod matches the request method against the handlers mapped by methods.
func matchMethod(method string, handlers map[string]http.HandlerFunc, params Parameters) (RouterResult, http.HandlerFunc, Parameters) {
	if handler, ok := handlers[method]; ok {
		return RouterResultFound, handler, params
	}
	if handler, ok := handlers["*"]; ok {
		return RouterResultFound, handler, params
	}
	return RouterResultMethodNotAllowed, nil, params
}

// matchSubsegments matches the request method and path against the subsegments.
func matchSubsegments(method, value string, next []string, segments []Segment, params Parameters) (RouterResult, http.HandlerFunc, Parameters) {
	for _, segment := range segments {
		result, handler, params := segment.Match(method, value, next, params)
		switch result {
		case RouterResultFound, RouterResultMethodNotAllowed:
			return result, handler, params
		default:
			continue
		}
	}
	return RouterResultPathNotFound, nil, params
}
