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
		FallbackHandler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("Not found"))
		},
	}

	_ = http.ListenAndServe(":8080", router)
}
