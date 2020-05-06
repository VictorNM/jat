package jat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
)

// NOTE:
// For simplicity in testing, any error occur when calling
// these methods will cause a panic, where a panic is acceptable.
// This behavior is the same with httptest.NewRequest

// NewRequest is the same with httptest.NewRequest if body is io.Reader
// Otherwise, it will try to marshal body as JSON format
// if an error occur, it will panic
func NewRequest(method, target string, body interface{}) *http.Request {
	return httptest.NewRequest(method, target, toReader(body))
}

func toReader(body interface{}) io.Reader {
	if reader, ok := body.(io.Reader); ok {
		return reader
	}

	b, err := json.Marshal(body)
	if err != nil {
		panic(fmt.Errorf("invalid JSON body: %v, error: %v", body, err))
	}

	return bytes.NewReader(b)
}

// RequestWrapper wraps *httpRequest for building with fluent interface
// Example:
// r := Wrap(GET("/api/users"))).
// 		AddQuery("type", "code").
//		Unwrap()
// fmt.Println(r.URL.RequestURI())
// Output: "/api/users?type=code"
type RequestWrapper struct {
	Request *http.Request
}

// Wrap wraps *httpRequest and returns a *RequestWrapper
// for building with fluent interface
func Wrap(r *http.Request) *RequestWrapper {
	return &RequestWrapper{Request: r}
}

// Unwrap return the wrapped request
// It similar to using rw.Request directly
// but will log the final Request method and URL for debug
func (rw *RequestWrapper) Unwrap() *http.Request {
	// TODO: replace with user custom logger
	log.Printf("[%s] %s\n", rw.Request.Method, rw.Request.URL)
	return rw.Request
}

// ===== method ====

func GET(target string) *http.Request {
	return NewRequest(http.MethodGet, target, nil)
}

func WrapGET(target string) *RequestWrapper {
	return Wrap(GET(target))
}

func POST(target string, body interface{}) *http.Request {
	return NewRequest(http.MethodPost, target, body)
}

func WrapPOST(target string, body interface{}) *RequestWrapper {
	return Wrap(POST(target, body))
}

func PUT(target string, body interface{}) *http.Request {
	return NewRequest(http.MethodPut, target, body)
}

func WrapPUT(target string, body interface{}) *RequestWrapper {
	return Wrap(PUT(target, body))
}

func PATCH(target string, body interface{}) *http.Request {
	return NewRequest(http.MethodPatch, target, body)
}

func WrapPATCH(target string, body interface{}) *RequestWrapper {
	return Wrap(PATCH(target, body))
}

func DELETE(target string, body interface{}) *http.Request {
	return NewRequest(http.MethodDelete, target, body)
}

func WrapDELETE(target string, body interface{}) *RequestWrapper {
	return Wrap(DELETE(target, body))
}

// ===== body =====

// WithBody replaces the current body of the request with new body
// Also replace the ContentLength because the body has been changed
func WithBody(r *http.Request, body interface{}) {
	req := NewRequest(r.Method, r.URL.String(), body)

	r.Body = req.Body
	r.ContentLength = req.ContentLength
}

func (rw *RequestWrapper) WithBody(body interface{}) *RequestWrapper {
	WithBody(rw.Request, body)

	return rw
}

// ===== path params =====
// TODO: Maybe also support custom function for matching param name
// that user can define their own
// Example: /users/{id}, /users/{:id}, /users/_id ...

func WithParam(r *http.Request, param map[string]interface{}) {
	for k, v := range param {
		SetParam(r, k, v)
	}
}

func (rw *RequestWrapper) WithParam(param map[string]interface{}) *RequestWrapper {
	WithParam(rw.Request, param)

	return rw
}

func SetParam(r *http.Request, key string, value interface{}) {
	// key should be a valid identifier, if not, panic
	validID := regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	if !validID.MatchString(key) {
		panic(fmt.Errorf("param key should be a valid identifier %v", key))
	}

	expr := `:` + key + `\b`

	re, err := regexp.Compile(expr)
	if err != nil {
		panic(fmt.Errorf("compile regex failed, may be key %q contain invalid regex %v", key, err))
	}

	r.URL.Path = re.ReplaceAllStringFunc(r.URL.Path, func(s string) string {
		return fmt.Sprint(value)
	})
}

func (rw *RequestWrapper) SetParam(key string, value interface{}) *RequestWrapper {
	SetParam(rw.Request, key, value)

	return rw
}

// ===== query ====

// Add adds the value to key. It appends to any existing
// values associated with key.
func AddQuery(r *http.Request, key string, value interface{}) {
	q := r.URL.Query()

	q.Add(key, fmt.Sprint(value))
	r.URL.RawQuery = q.Encode()
}

func (rw *RequestWrapper) AddQuery(key string, value interface{}) *RequestWrapper {
	AddQuery(rw.Request, key, value)

	return rw
}

// Set sets the key to value. It replaces any existing
// values.
func SetQuery(r *http.Request, key string, value interface{}) {
	q := r.URL.Query()

	q.Set(key, fmt.Sprint(value))
	r.URL.RawQuery = q.Encode()
}

func (rw *RequestWrapper) SetQuery(key string, value interface{}) *RequestWrapper {
	SetQuery(rw.Request, key, value)

	return rw
}

// WithQuery replaces the current query of the request with new query
func WithQuery(r *http.Request, query map[string][]interface{}) {
	q := url.Values{}
	for key, values := range query {
		for _, value := range values {
			q.Add(key, fmt.Sprint(value))
		}
	}

	r.URL.RawQuery = q.Encode()
}

func (rw *RequestWrapper) WithQuery(query map[string][]interface{}) *RequestWrapper {
	WithQuery(rw.Request, query)

	return rw
}

// WithQueryString replaces the current query of the request with new query
func WithQueryString(r *http.Request, query string) {
	q, err := url.ParseQuery(query)
	if err != nil {
		panic(fmt.Errorf("parse query failed %v", err))
	}

	r.URL.RawQuery = q.Encode()
}

func (rw *RequestWrapper) WithQueryString(query string) *RequestWrapper {
	WithQueryString(rw.Request, query)

	return rw
}

// WithQueryValues replaces the current query of the request with new query
func WithQueryValues(r *http.Request, query url.Values) {
	r.URL.RawQuery = query.Encode()
}

func (rw *RequestWrapper) WithQueryValues(query url.Values) *RequestWrapper {
	WithQueryValues(rw.Request, query)

	return rw
}

// ===== header =====

// AddHeader adds the key, value pair to the header of the request.
// See: http.Header.Add
func AddHeader(r *http.Request, key, value string) {
	r.Header.Add(key, value)
}

func (rw *RequestWrapper) AddHeader(key, value string) *RequestWrapper {
	AddHeader(rw.Request, key, value)
	return rw
}

// SetHeader sets the request's header entries associated with key to the
// single element value. It replaces any existing values
// associated with key.
// See: http.Header.Set
func SetHeader(r *http.Request, key, value string) {
	r.Header.Set(key, value)
}

func (rw *RequestWrapper) SetHeader(key, value string) *RequestWrapper {
	SetHeader(rw.Request, key, value)
	return rw
}

// SetBasicAuth sets the request's Authorization header to use HTTP
// Basic Authentication with the provided username and password.
// See: http.Request.SetBasicAuth
func (rw *RequestWrapper) SetBasicAuth(username, password string) *RequestWrapper {
	rw.Request.SetBasicAuth(username, password)

	return rw
}

// SetBasicAuth sets the request's Authorization header to use HTTP
// Bearer Authentication with the provided token.
func SetBearerAuth(r *http.Request, token string) {
	r.Header.Set("Authorization", "Bearer "+token)
}

func (rw *RequestWrapper) SetBearerAuth(token string) *RequestWrapper {
	SetBearerAuth(rw.Request, token)

	return rw
}

func (rw *RequestWrapper) AddCookie(c *http.Cookie) *RequestWrapper {
	rw.Request.AddCookie(c)

	return rw
}
