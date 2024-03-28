package main

import (
	"github.com/piotrek-r/segmentrouter"
	"net/http"
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
