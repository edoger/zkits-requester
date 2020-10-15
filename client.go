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

var requestPool = sync.Pool{New: func() interface{} { return new(request) }}

type Client interface {
	SetHTTPClient(*http.Client) Client
	SetTimeout(time.Duration) Client
	GetCommonHeaders() http.Header
	SetCommonHeader(key, value string) Client
	SetCommonHeaders(headers http.Header) Client
	New(string) Request
	Do(string, func(Request) (Response, error)) (Response, error)
	Head(string, url.Values) (Response, error)
	Get(string, url.Values) (Response, error)
	Post(string, interface{}) (Response, error)
	PostJSON(string, interface{}) (Response, error)
	PostXML(string, interface{}) (Response, error)
	PostForm(string, url.Values) (Response, error)
}

func New() Client {
	return &client{headers: make(http.Header)}
}

type client struct {
	http    *http.Client
	timeout time.Duration
	headers http.Header
}

func (c *client) SetHTTPClient(client *http.Client) Client {
	c.http = client
	return c
}

func (c *client) SetTimeout(t time.Duration) Client {
	c.timeout = t
	return c
}

func (c *client) GetCommonHeaders() http.Header {
	return c.headers
}

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

func (c *client) SetCommonHeaders(headers http.Header) Client {
	if headers == nil {
		c.headers = make(http.Header)
	} else {
		c.headers = headers.Clone()
	}
	return c
}

func (c *client) New(uri string) Request {
	return &request{client: c, uri: uri}
}

func (c *client) Do(uri string, f func(Request) (Response, error)) (Response, error) {
	r := requestPool.Get().(*request)
	defer func() { requestPool.Put(r.reset()) }()
	r.client = c
	r.uri = uri
	return f(r)
}

func (c *client) Head(uri string, qs url.Values) (Response, error) {
	return c.Do(uri, func(r Request) (Response, error) {
		return r.WithQueries(qs).Head()
	})
}

func (c *client) Get(uri string, qs url.Values) (Response, error) {
	return c.Do(uri, func(r Request) (Response, error) {
		return r.WithQueries(qs).Get()
	})
}

func (c *client) Post(uri string, body interface{}) (Response, error) {
	return c.Do(uri, func(r Request) (Response, error) {
		return r.WithBody(body).Post()
	})
}

func (c *client) PostJSON(uri string, body interface{}) (Response, error) {
	return c.Do(uri, func(r Request) (Response, error) {
		return r.WithJSONBody(body).Post()
	})
}

func (c *client) PostXML(uri string, body interface{}) (Response, error) {
	return c.Do(uri, func(r Request) (Response, error) {
		return r.WithXMLBody(body).Post()
	})
}

func (c *client) PostForm(uri string, body url.Values) (Response, error) {
	return c.Do(uri, func(r Request) (Response, error) {
		return r.WithFormBody(body).Post()
	})
}
