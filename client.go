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
	"net/http"
	"net/url"
	"sync"
	"time"
)

// This is a built-in pool of request objects to provide reusable objects for simple requests.
var requestPool = sync.Pool{New: func() interface{} { return new(request) }}

// The Client interface defines the requester client.
type Client interface {
	// SetHTTPClient sets a private HTTP client instance for the current client.
	// If a nil is given, the built-in default HTTP client is used.
	SetHTTPClient(*http.Client) Client

	// SetTimeout sets a request timeout period for the current client.
	// Setting it to 0 will never time out.
	// This timeout period can be overridden by each request timeout setting.
	SetTimeout(time.Duration) Client

	// GetCommonHeaders returns the common request headers of the current client.
	GetCommonHeaders() http.Header

	// SetCommonHeader sets a common request header of the current client.
	// If the given request header value is an empty string, the corresponding request
	// header will be deleted.
	SetCommonHeader(string, string) Client

	// SetCommonHeaders resets the common request headers of the current client.
	// If nil is given, all common request headers will be deleted.
	SetCommonHeaders(http.Header) Client

	// New returns a new request instance from the given uri.
	New(string) Request

	// Do obtains and initializes the request object from the global request object pool,
	// and will immediately reclaim the request object after the given closure is completed.
	// This method assumes that the request object will not be referenced by space outside the closure.
	Do(string, func(Request) (Response, error)) (Response, error)

	// Head uses the given parameters to send a request and return the Response.
	// This method will send the request using the HEAD method.
	Head(string, url.Values) (Response, error)

	// Get uses the given parameters to send a request and return the Response.
	// This method will send the request using the GET method.
	Get(string, url.Values) (Response, error)

	// Post uses the given parameters to send a request and return the Response.
	// This method will send the request using the POST method.
	Post(string, interface{}) (Response, error)

	// PostJSON uses the given parameters to send a request and return the Response.
	// This method will send the request using the POST method.
	// This method encodes the request body as json data and sends it.
	PostJSON(string, interface{}) (Response, error)

	// PostXML uses the given parameters to send a request and return the Response.
	// This method will send the request using the POST method.
	// This method encodes the request body as xml data and sends it.
	PostXML(string, interface{}) (Response, error)

	// PostForm uses the given parameters to send a request and return the Response.
	// This method will send the request using the POST method.
	// This method encodes the request body as form data (urlencoded) and sends it.
	PostForm(string, url.Values) (Response, error)
}

// The New function creates and returns a new built-in Client instance.
func New() Client {
	return &client{headers: make(http.Header)}
}

// The client type is a built-in implementation of the Client interface.
type client struct {
	http    *http.Client
	timeout time.Duration
	headers http.Header
}

// SetHTTPClient sets a private HTTP client instance for the current client.
// If a nil is given, the built-in default HTTP client is used.
func (c *client) SetHTTPClient(client *http.Client) Client {
	c.http = client
	return c
}

// SetTimeout sets a request timeout period for the current client.
// Setting it to 0 will never time out.
// This timeout period can be overridden by each request timeout setting.
func (c *client) SetTimeout(t time.Duration) Client {
	c.timeout = t
	return c
}

// GetCommonHeaders returns the common request headers of the current client.
func (c *client) GetCommonHeaders() http.Header {
	return c.headers
}

// SetCommonHeader sets a common request header of the current client.
// If the given request header value is an empty string, the corresponding request
// header will be deleted.
func (c *client) SetCommonHeader(key, value string) Client {
	if value == "" {
		if len(c.headers) > 0 {
			c.headers.Del(key)
		}
		return c
	}
	c.headers.Set(key, value)
	return c
}

// SetCommonHeaders resets the common request headers of the current client.
// If nil is given, all common request headers will be deleted.
func (c *client) SetCommonHeaders(headers http.Header) Client {
	if headers == nil {
		c.headers = make(http.Header)
	} else {
		c.headers = headers.Clone()
	}
	return c
}

// New returns a new request instance from the given uri.
func (c *client) New(uri string) Request {
	return &request{client: c, uri: uri}
}

// Do obtains and initializes the request object from the global request object pool,
// and will immediately reclaim the request object after the given closure is completed.
// This method assumes that the request object will not be referenced by space outside the closure.
func (c *client) Do(uri string, f func(Request) (Response, error)) (Response, error) {
	r := requestPool.Get().(*request)
	defer func() { requestPool.Put(r.reset()) }()
	r.client = c
	r.uri = uri
	return f(r)
}

// Head uses the given parameters to send a request and return the Response.
// This method will send the request using the HEAD method.
func (c *client) Head(uri string, qs url.Values) (Response, error) {
	return c.Do(uri, func(r Request) (Response, error) {
		return r.WithQueries(qs).Head()
	})
}

// Get uses the given parameters to send a request and return the Response.
// This method will send the request using the GET method.
func (c *client) Get(uri string, qs url.Values) (Response, error) {
	return c.Do(uri, func(r Request) (Response, error) {
		return r.WithQueries(qs).Get()
	})
}

// Post uses the given parameters to send a request and return the Response.
// This method will send the request using the POST method.
func (c *client) Post(uri string, body interface{}) (Response, error) {
	return c.Do(uri, func(r Request) (Response, error) {
		return r.WithBody(body).Post()
	})
}

// PostJSON uses the given parameters to send a request and return the Response.
// This method will send the request using the POST method.
// This method encodes the request body as json data and sends it.
func (c *client) PostJSON(uri string, body interface{}) (Response, error) {
	return c.Do(uri, func(r Request) (Response, error) {
		return r.WithJSONBody(body).Post()
	})
}

// PostXML uses the given parameters to send a request and return the Response.
// This method will send the request using the POST method.
// This method encodes the request body as xml data and sends it.
func (c *client) PostXML(uri string, body interface{}) (Response, error) {
	return c.Do(uri, func(r Request) (Response, error) {
		return r.WithXMLBody(body).Post()
	})
}

// PostForm uses the given parameters to send a request and return the Response.
// This method will send the request using the POST method.
// This method encodes the request body as form data (urlencoded) and sends it.
func (c *client) PostForm(uri string, body url.Values) (Response, error) {
	return c.Do(uri, func(r Request) (Response, error) {
		return r.WithFormBody(body).Post()
	})
}
