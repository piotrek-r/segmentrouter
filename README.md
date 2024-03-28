# Segment Router #

This is a simple Go router that allows you to define routes and their handlers. Routing is based on the path of the request URL. It treats the URL path as a series of segments separated by slashes (`/`).

New segment types can be defined by implementing the `Segment` interface. New types will be added to this library in the future as well.

This router cannot build paths from route names or parameters yet. This might be added in the future.

See examples in the [`examples`](./examples) directory.

The library doesn't have any dependencies outside the standard library. The library is in an early stage of development and have limited functionality.

## Basic Usage ##

```go
package main

import (
	"net/http"
	"segmentrouter"
)

func main() {
	router := segmentrouter.SegmentRouter{
		Segments: []segmentrouter.Segment{
			segmentrouter.StaticSegment{
				RouteName: "root",
				Value:     "",
				Handlers: map[string]http.HandlerFunc{
					"GET": func(w http.ResponseWriter, r *http.Request) {
						_, _ = w.Write([]byte("Hello, world!"))
					},
				},
			},
		},
	}

	_ = http.ListenAndServe(":8080", segmentrouter.CreateHttpHandler(router, func(w http.ResponseWriter, r *http.Request) {
		// fallback handler
	}))
}
```

## Segment types ##

Common properties of all segments:

- `RouteName` (string): The name of the route. This is used to identify the route in the router.
- `Handlers` (map[string]http.HandlerFunc): A map of HTTP methods to handlers. The key is the HTTP method (e.g. `"GET"`, `"POST"`, etc.) and the value is the handler function. For key `"*"` the handler is used for all methods that do not have a specific handler.
- `SubSegments` ([]Segment): A list of sub-segments that are children of this segment. This is used to define nested routes.

### `StaticSegment` ###

A `StaticSegment` is a segment that matches a specific string. It is defined by the `Value` field and must be an exact match to the segment in the URL path.

 Properties:

- `Value` (string): The value that this segment must match.

### `ParamSegment` ###

A `ParamSegment` is a segment that matches any string. The value of the segment is added to a map of parameters that is passed to the context of the request.

Properties:

- `ParamName` (string): The name of the parameter that is added to the map.
