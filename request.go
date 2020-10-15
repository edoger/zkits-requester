// Copyright 2020 The ZKits Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package requester

import (
	"bytes"
	"context"
	"encoding"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	ErrEmptyRequestURL    = errors.New("empty request url")
	ErrInvalidRequestBody = errors.New("invalid request body")
)

type Request interface {
	WithMethod(string) Request
	WithHeader(string, string) Request
	WithContentType(string) Request
	WithHeaders(http.Header) Request
	WithContext(context.Context) Request
	WithQuery(string, string) Request
	WithQueryValue(string, interface{}) Request
	WithQueries(url.Values) Request
	WithTimeout(time.Duration) Request
	WithBody(interface{}) Request
	WithJSONBody(interface{}) Request
	WithRawJSONBody([]byte) Request
	WithXMLBody(interface{}) Request
	WithRawXMLBody([]byte) Request
	WithFormBody(url.Values) Request
	Head() (Response, error)
	Get() (Response, error)
	Post() (Response, error)
	Send() (Response, error)
}

type request struct {
	client      *client
	uri         string
	method      string
	headers     http.Header
	ctx         context.Context
	query       url.Values
	timeout     time.Duration
	body        interface{}
	bodyEncoder string
	bodyType    string
}

func (r *request) WithMethod(method string) Request {
	r.method = strings.ToUpper(method)
	return r
}

func (r *request) WithHeader(key, value string) Request {
	if value == "" {
		if len(r.headers) > 0 {
			r.headers.Del(key)
		}
		return r
	}
	if r.headers == nil {
		r.headers = make(http.Header, 1)
	}
	r.headers.Set(key, value)
	return r
}

func (r *request) WithContentType(ct string) Request {
	r.WithHeader("Content-Type", ct)
	return r
}

func (r *request) WithHeaders(headers http.Header) Request {
	r.headers = headers
	return r
}

func (r *request) WithContext(ctx context.Context) Request {
	r.ctx = ctx
	return r
}

func (r *request) WithQuery(key, value string) Request {
	if value == "" {
		if len(r.query) > 0 {
			r.query.Del(key)
		}
		return r
	}
	if r.query == nil {
		r.query = make(url.Values, 1)
	}
	r.query.Set(key, value)
	return r
}

func (r *request) WithQueryValue(key string, value interface{}) Request {
	return r.WithQuery(key, toString(value))
}

func (r *request) WithQueries(qs url.Values) Request {
	r.query = qs
	return r
}

func (r *request) WithTimeout(t time.Duration) Request {
	r.timeout = t
	return r
}

func (r *request) WithBody(body interface{}) Request {
	r.body = body
	r.bodyEncoder = ""
	r.bodyType = ""
	return r
}

func (r *request) WithJSONBody(body interface{}) Request {
	r.body = body
	r.bodyEncoder = "json"
	r.bodyType = "application/json"
	return r
}

func (r *request) WithRawJSONBody(body []byte) Request {
	r.body = body
	r.bodyEncoder = ""
	r.bodyType = "application/json"
	return r
}

func (r *request) WithXMLBody(body interface{}) Request {
	r.body = body
	r.bodyEncoder = "xml"
	r.bodyType = "application/xml"
	return r
}

func (r *request) WithRawXMLBody(body []byte) Request {
	r.body = body
	r.bodyEncoder = ""
	r.bodyType = "application/xml"
	return r
}

func (r *request) WithFormBody(body url.Values) Request {
	r.body = body
	r.bodyEncoder = ""
	r.bodyType = "application/x-www-form-urlencoded"
	return r
}

// Post sends the current request and returns the received response.
// This method will send the request using the HEAD method.
func (r *request) Head() (Response, error) {
	return r.SendBy(http.MethodHead)
}

// Post sends the current request and returns the received response.
// This method will send the request using the GET method.
func (r *request) Get() (Response, error) {
	return r.SendBy(http.MethodGet)
}

// Post sends the current request and returns the received response.
// This method will send the request using the POST method.
func (r *request) Post() (Response, error) {
	return r.SendBy(http.MethodPost)
}

// Send sends the current request and returns the received response.
func (r *request) Send() (Response, error) {
	return r.SendBy(r.method)
}

// SendBy sends the current request and returns the received response.
// This method will send the request using the given request method.
func (r *request) SendBy(method string) (Response, error) {
	if r.uri == "" {
		return nil, ErrEmptyRequestURL
	}

	if o, err := r.send(method); err != nil {
		return nil, err
	} else {
		// For HEAD requests, we ignore the response body.
		return NewResponse(o, r.method == http.MethodHead)
	}
}

// The send method sends the current request and returns the received response.
func (r *request) send(method string) (*http.Response, error) {
	ctx := r.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if r.timeout > 0 {
		ctx, _ = context.WithTimeout(ctx, r.timeout)
	} else {
		if r.client.timeout > 0 {
			ctx, _ = context.WithTimeout(ctx, r.timeout)
		}
	}

	body, err := r.makeBodyReader()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, r.uri, body)
	if err != nil {
		return nil, err
	}
	// Add common request headers provided by the client.
	if len(r.client.headers) > 0 {
		for key, values := range r.client.headers {
			req.Header[key] = values
		}
	}
	// Add request headers provided by the current request.
	if len(r.headers) > 0 {
		for key, values := range r.headers {
			req.Header[key] = values
		}
	}
	// If the type of the request body is known, the Content-Type field of the
	// request header is forced to be set.
	if r.bodyType != "" {
		req.Header.Set("Content-Type", r.bodyType)
	}

	if len(r.query) > 0 {
		if req.URL.RawQuery == "" {
			req.URL.RawQuery = r.query.Encode()
		} else {
			// If query parameters are already attached to the target URL,
			// we need to overwrite the set query parameters to the target URL.
			qs, err := url.ParseQuery(req.URL.RawQuery)
			if err != nil {
				return nil, err
			}
			for key, values := range r.query {
				qs[key] = values
			}
			req.URL.RawQuery = qs.Encode()
		}
	}

	if r.client.http == nil {
		return DefaultHTTPClient().Do(req)
	} else {
		return r.client.http.Do(req)
	}
}

// The makeBodyReader method returns the current request body as a io.Reader.
func (r *request) makeBodyReader() (io.Reader, error) {
	if r.body == nil {
		return nil, nil
	}

	if r.bodyEncoder != "" {
		switch r.bodyEncoder {
		case "json":
			if data, err := json.Marshal(r.body); err != nil {
				return nil, err
			} else {
				return bytes.NewReader(data), nil
			}
		case "xml":
			if data, err := xml.Marshal(r.body); err != nil {
				return nil, err
			} else {
				return bytes.NewReader(data), nil
			}
		}
		return nil, fmt.Errorf("invalid request body encoder: %s", r.bodyEncoder)
	}

	switch body := r.body.(type) {
	case string:
		return strings.NewReader(body), nil
	case []byte:
		return bytes.NewReader(body), nil
	case url.Values:
		return strings.NewReader(body.Encode()), nil
	case io.Reader:
		return body, nil
	case fmt.Stringer:
		return strings.NewReader(body.String()), nil
	case encoding.TextMarshaler:
		if data, err := body.MarshalText(); err != nil {
			return nil, err
		} else {
			return bytes.NewReader(data), nil
		}
	}
	return nil, ErrInvalidRequestBody
}

// The reset method resets the current instance.
func (r *request) reset() *request {
	r.client = nil
	r.uri = ""
	r.method = ""
	r.headers = nil
	r.ctx = nil
	r.query = nil
	r.timeout = 0
	r.body = nil
	r.bodyEncoder = ""
	r.bodyType = ""
	return r
}

// The toString function converts the given parameter to a string.
func toString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case fmt.Stringer:
		return v.String()
	case error:
		return v.Error()
	}

	rv := reflect.ValueOf(value)
	for k := rv.Kind(); k == reflect.Ptr || k == reflect.Interface; k = rv.Kind() {
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.String:
		return rv.String()
	case reflect.Int64, reflect.Int, reflect.Int32, reflect.Int16, reflect.Int8:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint64, reflect.Uint, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	case reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 32)
	}
	return fmt.Sprint(value)
}
