package segmentrouter

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type SegmentRouterTest struct {
	name   string
	router SegmentRouter
	tests  []SegmentRouterTestPath
}

type SegmentRouterTestPath struct {
	path     string
	method   string
	expected RouterResult
	fn       func(*testing.T, SegmentRouterTestPath, http.HandlerFunc, Parameters)
}

var emptyHandler = func(w http.ResponseWriter, r *http.Request) {}

func checkHandler(expectedContent string) func(*testing.T, SegmentRouterTestPath, http.HandlerFunc, Parameters) {
	return func(t *testing.T, testPath SegmentRouterTestPath, handlerFunc http.HandlerFunc, params Parameters) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(testPath.method, testPath.path, nil)

		handlerFunc.ServeHTTP(w, r)

		content := string(w.Body.Bytes())
		if content != expectedContent {
			t.Errorf("Expected content %s, got %s", expectedContent, content)
		}
	}
}

func checkParams(expected map[string]string) func(*testing.T, SegmentRouterTestPath, http.HandlerFunc, Parameters) {
	return func(t *testing.T, testPath SegmentRouterTestPath, handlerFunc http.HandlerFunc, params Parameters) {
		for key, value := range expected {
			if params[key] != value {
				t.Errorf("Expected param %s=%s, got %s", key, value, params[key])
			}
		}
	}
}

var tests = []SegmentRouterTest{
	{
		name:   "Empty router",
		router: SegmentRouter{},
		tests: []SegmentRouterTestPath{
			{
				path:     "/",
				method:   "GET",
				expected: RouterResultPathNotFound,
			},
		},
	},
	{
		name: "Root and about router",
		router: SegmentRouter{
			Segments: []Segment{
				StaticSegment{
					Value: "",
					Handlers: map[string]http.HandlerFunc{
						"GET": emptyHandler,
					},
				},
				StaticSegment{
					Value: "about",
					Handlers: map[string]http.HandlerFunc{
						"GET": emptyHandler,
					},
				},
			},
		},
		tests: []SegmentRouterTestPath{
			{
				path:     "/",
				method:   "GET",
				expected: RouterResultFound,
			},
			{
				path:     "/about",
				method:   "GET",
				expected: RouterResultFound,
			},
			{
				path:     "/contact",
				method:   "GET",
				expected: RouterResultPathNotFound,
			},
		},
	},
	{
		name: "Big router",
		router: SegmentRouter{
			Segments: []Segment{
				StaticSegment{
					Value: "",
					Handlers: map[string]http.HandlerFunc{
						"GET": emptyHandler,
					},
				},
				StaticSegment{
					Value: "collection-read-only",
					Handlers: map[string]http.HandlerFunc{
						"GET": emptyHandler,
					},
					SubSegments: []Segment{
						ParamSegment{
							ParamName: "id",
							Handlers: map[string]http.HandlerFunc{
								"GET": emptyHandler,
							},
							SubSegments: []Segment{
								StaticSegment{
									Value: "subpath",
									Handlers: map[string]http.HandlerFunc{
										"GET": emptyHandler,
									},
								},
							},
						},
					},
				},
				StaticSegment{
					Value: "collection-insertable",
					Handlers: map[string]http.HandlerFunc{
						"GET":  emptyHandler,
						"POST": emptyHandler,
					},
					SubSegments: []Segment{
						ParamSegment{
							ParamName: "id",
							Handlers: map[string]http.HandlerFunc{
								"GET": emptyHandler,
							},
						},
					},
				},
				StaticSegment{
					Value: "collection-updateable",
					Handlers: map[string]http.HandlerFunc{
						"GET": emptyHandler,
					},
					SubSegments: []Segment{
						ParamSegment{
							ParamName: "id",
							Handlers: map[string]http.HandlerFunc{
								"GET":  emptyHandler,
								"POST": emptyHandler,
							},
						},
					},
				},
				StaticSegment{
					Value: "collection-read-write",
					Handlers: map[string]http.HandlerFunc{
						"GET":  emptyHandler,
						"POST": emptyHandler,
					},
					SubSegments: []Segment{
						StaticSegment{
							Value: "any-method",
							Handlers: map[string]http.HandlerFunc{
								"*": emptyHandler,
							},
						},
						ParamSegment{
							ParamName: "id",
							Handlers: map[string]http.HandlerFunc{
								"GET":  emptyHandler,
								"POST": emptyHandler,
							},
						},
					},
				},
			},
		},
		tests: []SegmentRouterTestPath{
			{
				path:     "/",
				method:   "GET",
				expected: RouterResultFound,
			},
			{
				path:     "/",
				method:   "POST",
				expected: RouterResultMethodNotAllowed,
			},
			{
				path:     "/collection-read-only",
				method:   "GET",
				expected: RouterResultFound,
			},
			{
				path:     "/collection-read-only",
				method:   "POST",
				expected: RouterResultMethodNotAllowed,
			},
			{
				path:     "/collection-read-only/123",
				method:   "GET",
				expected: RouterResultFound,
				fn:       checkParams(map[string]string{"id": "123"}),
			},
			{
				path:     "/collection-read-only/123",
				method:   "POST",
				expected: RouterResultMethodNotAllowed,
			},
			{
				path:     "/collection-read-only/234/subpath",
				method:   "GET",
				expected: RouterResultFound,
				fn:       checkParams(map[string]string{"id": "234"}),
			},
			{
				path:     "/collection-insertable",
				method:   "GET",
				expected: RouterResultFound,
			},
			{
				path:     "/collection-insertable",
				method:   "POST",
				expected: RouterResultFound,
			},
			{
				path:     "/collection-insertable/123",
				method:   "GET",
				expected: RouterResultFound,
			},
			{
				path:     "/collection-insertable/123",
				method:   "POST",
				expected: RouterResultMethodNotAllowed,
			},
			{
				path:     "/collection-updateable",
				method:   "GET",
				expected: RouterResultFound,
			},
			{
				path:     "/collection-updateable",
				method:   "POST",
				expected: RouterResultMethodNotAllowed,
			},
			{
				path:     "/collection-updateable/123",
				method:   "GET",
				expected: RouterResultFound,
			},
			{
				path:     "/collection-updateable/123",
				method:   "POST",
				expected: RouterResultFound,
			},
			{
				path:     "/collection-read-write",
				method:   "GET",
				expected: RouterResultFound,
			},
			{
				path:     "/collection-read-write",
				method:   "POST",
				expected: RouterResultFound,
			},
			{
				path:     "/collection-read-write/123",
				method:   "GET",
				expected: RouterResultFound,
			},
			{
				path:     "/collection-read-write/123",
				method:   "POST",
				expected: RouterResultFound,
			},
			{
				path:     "/collection-read-write/123",
				method:   "DELETE",
				expected: RouterResultMethodNotAllowed,
			},
			{
				path:     "/collection-read-write/any-method",
				method:   "GET",
				expected: RouterResultFound,
			},
			{
				path:     "/collection-read-write/any-method",
				method:   "POST",
				expected: RouterResultFound,
			},
			{
				path:     "/collection-read-write/any-method",
				method:   "DELETE",
				expected: RouterResultFound,
			},
		},
	},
	{
		name: "Any method override",
		router: SegmentRouter{
			Segments: []Segment{
				StaticSegment{
					Value: "",
					Handlers: map[string]http.HandlerFunc{
						"*": func(w http.ResponseWriter, r *http.Request) {
							_, _ = w.Write([]byte("Any method"))
						},
						"GET": func(w http.ResponseWriter, r *http.Request) {
							_, _ = w.Write([]byte("GET"))
						},
					},
				},
			},
		},
		tests: []SegmentRouterTestPath{
			{
				path:     "/",
				method:   "GET",
				expected: RouterResultFound,
				fn:       checkHandler("GET"),
			},
			{
				path:     "/",
				method:   "POST",
				expected: RouterResultFound,
				fn:       checkHandler("Any method"),
			},
			{
				path:     "/",
				method:   "DELETE",
				expected: RouterResultFound,
				fn:       checkHandler("Any method"),
			},
		},
	},
}

func TestSegmentRouter_Match(t *testing.T) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, tt := range test.tests {
				t.Run(fmt.Sprintf("%s→%s", tt.method, tt.path), func(t *testing.T) {
					req := httptest.NewRequest(tt.method, tt.path, nil)
					result, handler, params := test.router.Match(req)
					if result != tt.expected {
						t.Errorf("SegmentRouter.Match() = %v, want %v", result, tt.expected)
					}
					if tt.fn != nil {
						tt.fn(t, tt, handler, params)
					}
				})
			}
		})
	}
}

func TestSegmentRouter_ServeHTTP(t *testing.T) {
	router := SegmentRouter{
		Segments: []Segment{
			StaticSegment{
				Value: "",
				Handlers: map[string]http.HandlerFunc{
					"GET": emptyHandler,
				},
			},
		},
		FallbackHandler: func(writer http.ResponseWriter, request *http.Request) {
			http.Error(writer, "Not found", http.StatusNotFound)
		},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)

	router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/not-found", nil)

	router.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code 404, got %d", w.Code)
	}
}
