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
	"io"
	"io/ioutil"
	"net/http"
)

// Response interface defines the HTTP response result.
type Response interface {
	// Headers method returns all response headers.
	Headers() http.Header
	// Cookie returns the cookie data of the given name, or an empty string if it cannot be found.
	Cookie(string) string
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

// NewResponse returns a built-in implementation of the Response interface
// from a given http.Response instance.
// If noBody is true, the response body is discarded.
func NewResponse(o *http.Response, noBody bool) (Response, error) {
	defer func() { _ = o.Body.Close() }()
	res := &response{
		code:    o.StatusCode,
		status:  o.Status,
		headers: o.Header.Clone(),
		cookies: o.Cookies(),
	}
	if noBody {
		_, _ = io.Copy(ioutil.Discard, o.Body)
		return res, nil
	}
	if body, err := ioutil.ReadAll(o.Body); err != nil {
		return nil, err
	} else {
		res.body = body
		return res, nil
	}
}

// The response type is a built-in implementation of the Response interface.
type response struct {
	code    int
	status  string
	headers http.Header
	cookies []*http.Cookie
	body    []byte
}

// Headers method returns all response headers.
func (r *response) Headers() http.Header {
	return r.headers
}

// Cookie returns the cookie data of the given name, or an empty string if it cannot be found.
func (r *response) Cookie(name string) string {
	for i, j := 0, len(r.cookies); i < j; i++ {
		if r.cookies[i].Name == name {
			return r.cookies[i].Value
		}
	}
	return ""
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
