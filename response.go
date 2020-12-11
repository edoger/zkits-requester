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
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Response interface defines the HTTP response result.
type Response interface {
	fmt.Stringer

	// Headers method returns all response headers.
	Headers() http.Header

	// StatusCode returns the status code of the response.
	StatusCode() int

	// Status returns the status text of the response.
	Status() string

	// Body returns the response body, or nil if there is no response body.
	// For HEAD request, nil is always returned.
	Body() []byte

	// Len returns the length of the response body.
	Len() int

	// JSON binds the response body to the given object as json.
	JSON(interface{}) error

	// XML binds the response body to the given object as xml.
	XML(interface{}) error
}

// Responder defines the Response instance factory.
// In some scenarios, users want to freely control the processing method of http
// response data stream, especially for large response volume (download file).
// At this time, the responder can be specified in the global or single request
// to take over the control of the data stream.
// If noBody is true, the response body is discarded (HEAD request), however, the
// specific processing scheme is decided by users themselves.
type Responder func(r *http.Response, noBody bool) (Response, error)

// NewResponse returns a built-in implementation of the Response interface
// from a given http.Response instance.
// If noBody is true, the response body is discarded.
func NewResponse(o *http.Response, noBody bool) (Response, error) {
	defer func() { _ = o.Body.Close() }()
	res := &response{
		code:    o.StatusCode,
		status:  o.Status,
		headers: o.Header.Clone(),
	}
	if noBody {
		return res, nil
	}
	if body, err := ioutil.ReadAll(o.Body); err != nil {
		return nil, err
	} else {
		res.body = body
		return res, nil
	}
}

// NewResponseFrom creates and returns a Response instance from the given response data.
// This function is used to create a custom Response. Panic if too many status parameters.
// If the status parameter is not given, we will automatically use http.StatusText to convert
// the code to the status string.
func NewResponseFrom(body []byte, headers http.Header, code int, status ...string) Response {
	if len(status) > 1 {
		panic("requester.NewResponseFrom(): too many parameters")
	}

	r := &response{code: code, headers: headers, body: body}
	if len(status) == 0 {
		r.status = http.StatusText(code)
	} else {
		r.status = status[0]
	}
	// Ensure Response.Headers() always returns available values.
	if r.headers == nil {
		r.headers = http.Header{}
	}
	return r
}

// The response type is a built-in implementation of the Response interface.
type response struct {
	code    int
	status  string
	headers http.Header
	body    []byte
}

// Headers method returns all response headers.
func (r *response) Headers() http.Header {
	return r.headers
}

// StatusCode returns the status code of the response.
func (r *response) StatusCode() int {
	return r.code
}

// Status returns the status text of the response.
func (r *response) Status() string {
	return r.status
}

// Body returns the response body, or nil if there is no response body.
func (r *response) Body() []byte {
	return r.body
}

// Len returns the length of the response body.
func (r *response) Len() int {
	return len(r.body)
}

// JSON binds the response body to the given object as json.
func (r *response) JSON(o interface{}) error {
	return json.Unmarshal(r.Body(), o)
}

// XML binds the response body to the given object as xml.
func (r *response) XML(o interface{}) error {
	return xml.Unmarshal(r.Body(), o)
}

// String returns the response body string, or empty string if there is no response body.
func (r *response) String() string {
	return string(r.body)
}

// NewEmptyResponse creates and returns a new empty response instance.
// The returned response instance only implements the Response interface and has no
// meaningful data and behavior.
// This empty response can be used as an embedded object to help users implement their
// own response instances instead of implementing all the Response interface methods.
func NewEmptyResponse() Response {
	return new(emptyResponse)
}

// The emptyResponse type is a built-in implementation of the Response interface.
// This type only implements the Response interface and has no meaningful data or behavior.
type emptyResponse struct{}

// Headers implements the Response interface.
// The method always return empty http.Header.
func (*emptyResponse) Headers() http.Header {
	return http.Header{}
}

// StatusCode implements the Response interface.
// The method always return 0.
func (*emptyResponse) StatusCode() int {
	return 0
}

// Status implements the Response interface.
// The method always return empty string.
func (*emptyResponse) Status() string {
	return ""
}

// Body implements the Response interface.
// The method always return nil.
func (*emptyResponse) Body() []byte {
	return nil
}

// Len implements the Response interface.
// The method always return 0.
func (*emptyResponse) Len() int {
	return 0
}

// JSON implements the Response interface.
// The method always return an error.
func (*emptyResponse) JSON(o interface{}) error {
	return errors.New("requester: empty response")
}

// XML implements the Response interface.
// The method always return an error.
func (*emptyResponse) XML(o interface{}) error {
	return errors.New("requester: empty response")
}

// String implements the Response interface.
// The method always return empty string.
func (*emptyResponse) String() string {
	return ""
}
