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
	// ErrEmptyRequestURL represents an empty request url error.
	// When sending a request, this error will be returned if the given request url is empty.
	ErrEmptyRequestURL = errors.New("empty request url")

	// ErrInvalidRequestBody indicates an invalid request body error.
	// When sending a request, if the bound request body cannot be recognized, this error is returned.
	ErrInvalidRequestBody = errors.New("invalid request body")
)

// Request interface defines client requests.
type Request interface {
	// WithMethod adds the default request method of the current request.
	WithMethod(string) Request

	// WithHeader adds a request header to the current request.
	// If the given request header value is an empty string, the corresponding request
	// header will be deleted.
	WithHeader(string, string) Request

	// WithContentType adds the ContentType header information of the current request.
	WithContentType(string) Request

	// WithHeaders adds and replaces some headers to the current request.
	WithHeaders(http.Header) Request

	// WithContext adds a context to the current request.
	// If the given context is nil, context.Background() is automatically used.
	WithContext(context.Context) Request

	// WithQuery adds a query parameter to the current request.
	WithQuery(string, string) Request

	// WithQueryValue adds a query parameter to the current request.
	// This method will automatically convert the given parameter value to a string.
	WithQueryValue(string, interface{}) Request

	// WithQueries adds and replaces some query parameters to the current request.
	WithQueries(url.Values) Request

	// WithTimeout adds a timeout for the current request.
	// If the given timeout period is zero, the client's timeout setting is used.
	WithTimeout(time.Duration) Request

	// WithBody adds request body to the current request.
	WithBody(interface{}) Request

	// WithJSONBody adds the request body as json.
	// This method will force set "Content-Type" to "application/json".
	WithJSONBody(interface{}) Request

	// WithRawJSONBody adds the request body as raw json.
	// This method will force set "Content-Type" to "application/json".
	WithRawJSONBody([]byte) Request

	// WithXMLBody adds the request body as xml.
	// This method will force set "Content-Type" to "application/xml".
	WithXMLBody(interface{}) Request

	// WithRawXMLBody adds the request body as raw xml.
	// This method will force set "Content-Type" to "application/xml".
	WithRawXMLBody([]byte) Request

	// WithFormBody adds the request body as form (urlencoded).
	// This method will force set "Content-Type" to "application/x-www-form-urlencoded".
	WithFormBody(url.Values) Request

	// Head sends the current request and returns the received response.
	// This method will send the request using the HEAD method.
	Head() (Response, error)

	// Get sends the current request and returns the received response.
	// This method will send the request using the GET method.
	Get() (Response, error)

	// Post sends the current request and returns the received response.
	// This method will send the request using the POST method.
	Post() (Response, error)

	// Send sends the current request and returns the received response.
	Send() (Response, error)

	// SendBy sends the current request and returns the received response.
	// This method will send the request using the given request method.
	SendBy(string) (Response, error)
}

// The request type is a built-in implementation of the Request interface.
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

// WithMethod adds the default request method of the current request.
func (r *request) WithMethod(method string) Request {
	r.method = method
	return r
}

// WithHeader adds a request header to the current request.
// If the given request header value is an empty string, the corresponding request
// header will be deleted.
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

// WithContentType adds the ContentType header information of the current request.
func (r *request) WithContentType(ct string) Request {
	r.WithHeader("Content-Type", ct)
	return r
}

// WithHeaders adds and replaces some headers to the current request.
func (r *request) WithHeaders(headers http.Header) Request {
	r.headers = headers
	return r
}

// WithContext adds a context to the current request.
// If the given context is nil, context.Background() is automatically used.
func (r *request) WithContext(ctx context.Context) Request {
	r.ctx = ctx
	return r
}

// WithQuery adds a query parameter to the current request.
func (r *request) WithQuery(key, value string) Request {
	if r.query == nil {
		r.query = make(url.Values, 1)
	}
	r.query.Set(key, value)
	return r
}

// WithQueryValue adds a query parameter to the current request.
// This method will automatically convert the given parameter value to a string.
func (r *request) WithQueryValue(key string, value interface{}) Request {
	return r.WithQuery(key, toString(value))
}

// WithQueries adds and replaces some query parameters to the current request.
func (r *request) WithQueries(qs url.Values) Request {
	r.query = qs
	return r
}

// WithTimeout adds a timeout for the current request.
// If the given timeout period is zero, the client's timeout setting is used.
func (r *request) WithTimeout(t time.Duration) Request {
	r.timeout = t
	return r
}

// WithBody adds request body to the current request.
func (r *request) WithBody(body interface{}) Request {
	r.body = body
	r.bodyEncoder = ""
	r.bodyType = ""
	return r
}

// WithJSONBody adds the request body as json.
// This method will force set "Content-Type" to "application/json".
func (r *request) WithJSONBody(body interface{}) Request {
	r.body = body
	r.bodyEncoder = "json"
	r.bodyType = "application/json"
	return r
}

// WithRawJSONBody adds the request body as raw json.
// This method will force set "Content-Type" to "application/json".
func (r *request) WithRawJSONBody(body []byte) Request {
	r.body = body
	r.bodyEncoder = ""
	r.bodyType = "application/json"
	return r
}

// WithXMLBody adds the request body as xml.
// This method will force set "Content-Type" to "application/xml".
func (r *request) WithXMLBody(body interface{}) Request {
	r.body = body
	r.bodyEncoder = "xml"
	r.bodyType = "application/xml"
	return r
}

// WithRawXMLBody adds the request body as raw xml.
// This method will force set "Content-Type" to "application/xml".
func (r *request) WithRawXMLBody(body []byte) Request {
	r.body = body
	r.bodyEncoder = ""
	r.bodyType = "application/xml"
	return r
}

// WithFormBody adds the request body as form (urlencoded).
// This method will force set "Content-Type" to "application/x-www-form-urlencoded".
func (r *request) WithFormBody(body url.Values) Request {
	r.body = body
	r.bodyEncoder = ""
	r.bodyType = "application/x-www-form-urlencoded"
	return r
}

// Head sends the current request and returns the received response.
// This method will send the request using the HEAD method.
func (r *request) Head() (Response, error) {
	return r.SendBy(http.MethodHead)
}

// Get sends the current request and returns the received response.
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

	if o, err := r.send(strings.ToUpper(method)); err != nil {
		return nil, err
	} else {
		// For HEAD requests, we ignore the response body.
		return NewResponse(o, o.Request.Method == http.MethodHead)
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
			ctx, _ = context.WithTimeout(ctx, r.client.timeout)
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
			qs := req.URL.Query()
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
		// In theory, this error will never be returned.
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
