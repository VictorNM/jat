package jat_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/victornm/jat"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestMethod(t *testing.T) {
	tests := []struct {
		f func() *http.Request

		wantedMethod string
		wantedURI    string
	}{
		{
			f: func() *http.Request {
				return jat.WrapGET("/users").Unwrap()
			},

			wantedMethod: http.MethodGet,
			wantedURI:    "/users",
		},

		{
			f: func() *http.Request {
				return jat.WrapPOST("/users", nil).Unwrap()
			},

			wantedMethod: http.MethodPost,
			wantedURI:    "/users",
		},

		{
			f: func() *http.Request {
				return jat.WrapPUT("/users", nil).Unwrap()
			},

			wantedMethod: http.MethodPut,
			wantedURI:    "/users",
		},

		{
			f: func() *http.Request {
				return jat.WrapPATCH("/users", nil).Unwrap()
			},

			wantedMethod: http.MethodPatch,
			wantedURI:    "/users",
		},

		{
			f: func() *http.Request {
				return jat.WrapDELETE("/users", nil).Unwrap()
			},

			wantedMethod: http.MethodDelete,
			wantedURI:    "/users",
		},
	}

	for _, test := range tests {
		name := fmt.Sprintf("%s %s", test.wantedMethod, test.wantedURI)
		t.Run(name, func(t *testing.T) {
			req := test.f()

			assert.Equal(t, test.wantedMethod, req.Method)
			assert.Equal(t, test.wantedURI, req.URL.RequestURI())
		})
	}
}

func TestBodyJSON(t *testing.T) {
	tests := map[string]struct {
		f func(wrapper *jat.RequestWrapper)

		wanted string
	}{
		"body from map": {
			f: func(wrapper *jat.RequestWrapper) {
				wrapper.WithBody(map[string]string{
					"email": "foo@bar.com",
				})
			},

			wanted: `{"email": "foo@bar.com"}`,
		},

		"body from io.Reader": {
			f: func(wrapper *jat.RequestWrapper) {
				m := map[string]string{
					"email": "foo@bar.com",
				}

				b, _ := json.Marshal(m)

				wrapper.WithBody(bytes.NewReader(b))
			},

			wanted: `{"email": "foo@bar.com"}`,
		},

		"body from struct": {
			f: func(wrapper *jat.RequestWrapper) {
				body := struct {
					Email string `json:"email"`
				}{
					Email: "foo@bar.com",
				}

				wrapper.WithBody(body)
			},

			wanted: `{"email": "foo@bar.com"}`,
		},

		"body from array": {
			f: func(wrapper *jat.RequestWrapper) {
				body := []string{"hello", "world"}

				wrapper.WithBody(body)
			},

			wanted: `["hello", "world"]`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			wrapper := jat.WrapPOST("/", nil)

			test.f(wrapper)
			req := wrapper.Unwrap()

			b, err := ioutil.ReadAll(req.Body)
			assert.NoError(t, err)

			assert.JSONEq(t, test.wanted, string(b))
		})
	}
}

func TestHeader(t *testing.T) {
	target := "/api/ping"

	t.Run("add header", func(t *testing.T) {
		req := jat.WrapGET(target).
			AddHeader("Host", "localhost:3000").
			Unwrap()

		wanted := http.Header{
			"Host": []string{"localhost:3000"},
		}

		if !reflect.DeepEqual(req.Header, wanted) {
			t.Error("unexpected header")
		}
	})

	t.Run("add header twice", func(t *testing.T) {
		req := jat.WrapGET(target).
			AddHeader("Host", "localhost:3000").
			AddHeader("Host", "localhost:8080").
			Unwrap()

		wanted := http.Header{
			"Host": []string{"localhost:3000", "localhost:8080"},
		}

		if !reflect.DeepEqual(req.Header, wanted) {
			t.Error("unexpected header")
		}
	})

	t.Run("set header twice", func(t *testing.T) {
		req := jat.WrapGET(target).
			AddHeader("Host", "localhost:3000").
			AddHeader("Host", "localhost:8080").
			SetHeader("Host", "localhost:5432").
			Unwrap()

		wanted := http.Header{
			"Host": []string{"localhost:5432"},
		}

		if !reflect.DeepEqual(req.Header, wanted) {
			t.Error("unexpected header")
		}
	})

	t.Run("set basic auth", func(t *testing.T) {
		username, password := "foo", "bar"
		req := jat.WrapGET(target).
			SetBasicAuth(username, password).
			Unwrap()

		gotU, gotP, ok := req.BasicAuth()

		assert.True(t, ok)
		assert.Equal(t, username, gotU)
		assert.Equal(t, password, gotP)
	})

	t.Run("add bearer auth", func(t *testing.T) {
		token := "6eTUFP4HNhvvIwz5nNiL"
		req := jat.WrapGET(target).
			SetBearerAuth(token).
			Unwrap()

		gotToken := req.Header.Get("Authorization")

		assert.Equal(t, "Bearer "+token, gotToken)
	})

	t.Run("add cookie", func(t *testing.T) {
		username, password := "foo", "bar"
		req := jat.WrapGET(target).
			AddCookie(&http.Cookie{Name: username, Value: password}).
			Unwrap()

		gotC, err := req.Cookie(username)

		assert.NoError(t, err)
		assert.Equal(t, password, gotC.Value)
	})
}

func TestQuery(t *testing.T) {
	tests := map[string]struct {
		initURI string

		f func(wrapper *jat.RequestWrapper)

		wanted string // Note: query should be sorted by key
	}{
		"add query": {
			initURI: "/api/ping",
			f: func(wrapper *jat.RequestWrapper) {
				wrapper.AddQuery("type", "code")
			},

			wanted: "/api/ping?type=code",
		},

		"add query integer": {
			initURI: "/api/ping",
			f: func(wrapper *jat.RequestWrapper) {
				wrapper.AddQuery("num", 1)
			},

			wanted: "/api/ping?num=1",
		},

		"add two queries": {
			initURI: "/api/ping",
			f: func(wrapper *jat.RequestWrapper) {
				wrapper.
					AddQuery("type", "code").
					AddQuery("provider", "google")
			},

			wanted: "/api/ping?provider=google&type=code",
		},

		"add query twice": {
			initURI: "/api/ping",
			f: func(wrapper *jat.RequestWrapper) {
				wrapper.
					AddQuery("type", "code").
					AddQuery("type", "token")
			},

			wanted: "/api/ping?type=code&type=token",
		},

		"set query": {
			initURI: "/api/ping",
			f: func(wrapper *jat.RequestWrapper) {
				wrapper.
					AddQuery("type", "code").
					AddQuery("provider", "google").
					SetQuery("type", "money")
			},

			wanted: "/api/ping?provider=google&type=money",
		},

		"query from string": {
			initURI: "/api/ping",
			f: func(wrapper *jat.RequestWrapper) {
				wrapper.WithQueryString("provider=google&type=money")
			},

			wanted: "/api/ping?provider=google&type=money",
		},

		"query from map": {
			initURI: "/api/ping",
			f: func(wrapper *jat.RequestWrapper) {
				wrapper.WithQuery(map[string][]interface{}{
					"provider": {"google"},
					"int":      {1},
					"float":    {12.34},
					"negative": {-5},
				})
			},

			wanted: "/api/ping?float=12.34&int=1&negative=-5&provider=google",
		},

		"query from url.Values": {
			initURI: "localhost:3000/api/ping",
			f: func(wrapper *jat.RequestWrapper) {
				wrapper.WithQueryValues(url.Values{
					"provider": {"google"},
					"type":     {"money"},
				})
			},

			wanted: "localhost:3000/api/ping?provider=google&type=money",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			wrapper := jat.WrapGET(test.initURI)
			test.f(wrapper)

			req := wrapper.Unwrap()

			got := req.URL.String()
			if test.wanted != got {
				t.Errorf("unexpected query: wanted %q but got %q", test.wanted, got)
			}
		})
	}
}

func TestParam(t *testing.T) {
	tests := map[string]struct {
		template string
		param    map[string]interface{}

		wantedPath string
	}{
		"int param": {
			template: "/users/:id",
			param: map[string]interface{}{
				"id": 1,
			},

			wantedPath: "/users/1",
		},

		"string param": {
			template: "/users/:id/courses/:course_name",
			param: map[string]interface{}{
				"id":          1,
				"course_name": "cs50",
			},

			wantedPath: "/users/1/courses/cs50",
		},

		"more param than template": {
			template: "/users/:id",
			param: map[string]interface{}{
				"id":          1,
				"course_name": "cs50",
			},

			wantedPath: "/users/1",
		},

		"less param than template": {
			template: "/users/:id/courses/:course_name",
			param: map[string]interface{}{
				"id":          1,
			},

			wantedPath: "/users/1/courses/:course_name",
		},

		"similar key name": {
			template: "/users/:id/:id_string",
			param: map[string]interface{}{
				"id":         1,
				"id_string": "cs50",
			},

			wantedPath: "/users/1/cs50",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			req := jat.WrapGET(test.template).
				WithParam(test.param).
				Unwrap()

			assert.Equal(t, test.wantedPath, req.URL.Path)
		})
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			rw := jat.WrapGET(test.template)
			for k, v := range test.param {
				rw.SetParam(k, v)
			}

			req := rw.Unwrap()

			assert.Equal(t, test.wantedPath, req.URL.Path)
		})
	}
}