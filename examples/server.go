package main

import (
	"encoding/json"
	"github.com/piotrek-r/segmentrouter"
	"log"
	"net/http"
	"strconv"
)

func main() {
	log.Printf("Hello, world!")

	// Router with paths:
	// GET /
	// GET /users
	// GET /users/{id}
	// GET /groups
	// POST /groups
	// GET /groups/{id}
	// any method /groups/{id}/subpath
	router := segmentrouter.SegmentRouter{
		Segments: []segmentrouter.Segment{
			segmentrouter.StaticSegment{
				RouteName: "root",
				Value:     "",
				Handlers: map[string]http.HandlerFunc{
					"GET": handleGetRoot,
				},
			},
			segmentrouter.StaticSegment{
				RouteName: "collection-users",
				Value:     "users",
				Handlers: map[string]http.HandlerFunc{
					"GET": handleGetUsers,
				},
				SubSegments: []segmentrouter.Segment{
					segmentrouter.ParamSegment{
						RouteName: "read-user",
						ParamName: "id",
						Handlers: map[string]http.HandlerFunc{
							"GET": handleGetUser,
						},
					},
				},
			},
			segmentrouter.StaticSegment{
				RouteName: "collection-groups",
				Value:     "groups",
				Handlers: map[string]http.HandlerFunc{
					"GET":  handleGetGroups,
					"POST": handlePostGroups,
				},
				SubSegments: []segmentrouter.Segment{
					segmentrouter.ParamSegment{
						RouteName: "read-group",
						ParamName: "id",
						Handlers: map[string]http.HandlerFunc{
							"GET": handleGetGroup,
						},
						SubSegments: []segmentrouter.Segment{
							segmentrouter.StaticSegment{
								RouteName: "read-group-subpath",
								Value:     "subpath",
								Handlers: map[string]http.HandlerFunc{
									"GET": handleGetGroupSubpath,
								},
							},
						},
					},
				},
			},
		},
		FallbackHandler: send404,
	}

	// Listen
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		panic(err)
	}

	log.Printf("Finished.")
}

func handleGetRoot(w http.ResponseWriter, r *http.Request) {
	sendJson(map[string]string{}, w)
}

func handleGetUsers(w http.ResponseWriter, r *http.Request) {
	sendJson([]map[string]string{{"id": "1", "name": "Alice"}, {"id": "2", "name": "Bob"}}, w)
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	parameters := segmentrouter.GetParametersFromContext(r.Context())

	id := parameters["id"]

	_, err := strconv.Atoi(id)
	if err != nil {
		send404(w, r)
		return
	}

	routeName := parameters[segmentrouter.ParamRouteName]

	sendJson(map[string]string{"id": id, "name": "Alice", "route": routeName}, w)
}

func handleGetGroups(w http.ResponseWriter, r *http.Request) {
	sendJson([]map[string]string{{"id": "1", "name": "Group 1"}, {"id": "2", "name": "Group 2"}}, w)
}

func handlePostGroups(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	sendJson(map[string]string{"id": "3", "name": "Group 3"}, w)
}

func handleGetGroup(w http.ResponseWriter, r *http.Request) {
	parameters := segmentrouter.GetParametersFromContext(r.Context())

	id := parameters["id"]

	_, err := strconv.Atoi(id)
	if err != nil {
		send404(w, r)
		return
	}

	routeName := parameters[segmentrouter.ParamRouteName]

	sendJson(map[string]string{"id": id, "name": "Group 1", "route": routeName}, w)
}

func handleGetGroupSubpath(w http.ResponseWriter, r *http.Request) {
	sendJson(map[string]bool{"subpath": true}, w)
}

func sendJson(data any, w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	output, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	_, _ = w.Write(output)
}

func sendResult(w http.ResponseWriter, r *http.Request) {
	result := segmentrouter.GetRouterResultFromContext(r.Context())

	switch result {
	case segmentrouter.RouterResultMethodNotAllowed:
		send405(w, r)
	case segmentrouter.RouterResultPathNotFound:
	default:
		send404(w, r)
	}
}

func send404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("{\"error\":\"Not found\"}"))
}

func send405(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	_, _ = w.Write([]byte("{\"error\":\"Method not allowed\"}"))
}
