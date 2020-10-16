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
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

type testFmtStringer string

func (s testFmtStringer) String() string {
	return string(s)
}

type testEncodingTextMarshaler string

func (s testEncodingTextMarshaler) MarshalText() ([]byte, error) {
	return []byte(s), nil
}

type testErrorEncodingTextMarshaler string

func (s testErrorEncodingTextMarshaler) MarshalText() ([]byte, error) {
	return nil, errors.New(string(s))
}

func TestRequest(t *testing.T) {
	var handler func(*http.Request) string
	clear := func() { handler = nil }
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handler != nil {
			_, _ = io.WriteString(w, handler(r))
		}
	}))
	defer server.Close()

	c := New()
	c.SetTimeout(time.Hour)
	c.SetHTTPClient(NewDefaultHTTPClient())
	c.SetCommonHeader("X-Common", "common")

	_, _ = c.Do(server.URL+"?test=ok", func(req Request) (Response, error) {
		defer clear()

		method := ""
		headers := make(http.Header)
		queries := make(url.Values)
		body := ""

		handler = func(h *http.Request) string {
			method = h.Method
			for k, vs := range h.Header {
				for _, s := range vs {
					headers.Add(k, s)
				}
			}
			for k, vs := range h.URL.Query() {
				for _, s := range vs {
					queries.Add(k, s)
				}
			}
			if data, err := ioutil.ReadAll(h.Body); err != nil {
				body = "read body error: " + err.Error()
			} else {
				body = string(data)
			}
			return "ok"
		}

		if req.WithMethod(http.MethodPost) == nil {
			t.Fatal("Request.WithMethod() return nil")
		}
		if req.WithMethod("") == nil {
			t.Fatal("Request.WithMethod() return nil")
		}

		if req.WithHeader("X-Foo", "test") == nil {
			t.Fatal("Request.WithHeader() return nil")
		}
		if req.WithHeader("X-Foo", "") == nil {
			t.Fatal("Request.WithHeader() return nil")
		}
		if req.WithContentType("application/test") == nil {
			t.Fatal("Request.WithContentType() return nil")
		}
		if req.WithHeaders(http.Header{"X-Test": {"test"}}) == nil {
			t.Fatal("Request.WithHeaders() return nil")
		}
		if req.WithHeaders(nil) == nil {
			t.Fatal("Request.WithHeaders(nil) return nil")
		}

		if req.WithContext(context.TODO()) == nil {
			t.Fatal("Request.WithContext() return nil")
		}

		if req.WithQuery("foo", "foo") == nil {
			t.Fatal("Request.WithQuery() return nil")
		}
		if req.WithQueryValue("foo", 1) == nil {
			t.Fatal("Request.WithQueryValue() return nil")
		}
		if req.WithQueries(url.Values{}) == nil {
			t.Fatal("Request.WithQueries() return nil")
		}

		if req.WithTimeout(time.Minute) == nil {
			t.Fatal("Request.WithTimeout() return nil")
		}
		if req.WithTimeout(0) == nil {
			t.Fatal("Request.WithTimeout(0) return nil")
		}

		if req.WithBody("test") == nil {
			t.Fatal("Request.WithBody() return nil")
		}
		if req.WithJSONBody([]string{"test"}) == nil {
			t.Fatal("Request.WithJSONBody() return nil")
		}
		if req.WithRawJSONBody([]byte("{}")) == nil {
			t.Fatal("Request.WithRawJSONBody() return nil")
		}
		if req.WithXMLBody([]string{"test"}) == nil {
			t.Fatal("Request.WithXMLBody() return nil")
		}
		if req.WithRawXMLBody([]byte("<XML></XML>")) == nil {
			t.Fatal("Request.WithRawXMLBody() return nil")
		}
		if req.WithFormBody(url.Values{}) == nil {
			t.Fatal("Request.WithFormBody() return nil")
		}

		// check
		req.WithMethod(http.MethodPut)
		req.WithHeader("X-Test", "foo")
		req.WithQuery("foo", "foo")
		req.WithBody("TestBody")

		if res, err := req.Send(); err != nil {
			t.Fatalf("Request.Send() error: %s", err)
		} else {
			if got := res.String(); got != "ok" {
				t.Fatalf("Request.Send() return: %s", got)
			}
		}

		if method != http.MethodPut {
			t.Fatalf("Request.Send(): method got %s", method)
		}
		for k, v := range map[string]string{"X-Common": "common", "X-Test": "foo"} {
			if got := headers.Get(k); got != v {
				t.Fatalf("Request.Send(): header got %s = %s", k, got)
			}
		}
		for k, v := range map[string]string{"test": "ok", "foo": "foo"} {
			if got := queries.Get(k); got != v {
				t.Fatalf("Request.Send(): query got %s = %s", k, got)
			}
		}
		if body != "TestBody" {
			t.Fatalf("Request.Send(): body got %s", body)
		}
		return nil, nil
	})

	_, _ = c.Do(server.URL, func(req Request) (Response, error) {
		defer clear()

		handler = func(h *http.Request) string {
			if data, err := ioutil.ReadAll(h.Body); err != nil {
				return "read body error: " + err.Error()
			} else {
				return string(data)
			}
		}

		req.WithTimeout(time.Hour)

		if res, err := req.WithBody("string").Post(); err != nil {
			t.Fatalf("Request.Send() error: %s", err)
		} else {
			if got := res.String(); got != "string" {
				t.Fatalf("Request.Send() body got: %s", got)
			}
		}

		if res, err := req.WithBody([]byte("bytes")).Post(); err != nil {
			t.Fatalf("Request.Send() error: %s", err)
		} else {
			if got := res.String(); got != "bytes" {
				t.Fatalf("Request.Send() body got: %s", got)
			}
		}

		if res, err := req.WithBody(url.Values{"url": {"values"}}).Post(); err != nil {
			t.Fatalf("Request.Send() error: %s", err)
		} else {
			if got := res.String(); got != "url=values" {
				t.Fatalf("Request.Send() body got: %s", got)
			}
		}

		if res, err := req.WithBody(strings.NewReader("reader")).Post(); err != nil {
			t.Fatalf("Request.Send() error: %s", err)
		} else {
			if got := res.String(); got != "reader" {
				t.Fatalf("Request.Send() body got: %s", got)
			}
		}

		if res, err := req.WithBody(testFmtStringer("stringer")).Post(); err != nil {
			t.Fatalf("Request.Send() error: %s", err)
		} else {
			if got := res.String(); got != "stringer" {
				t.Fatalf("Request.Send() body got: %s", got)
			}
		}

		if res, err := req.WithBody(testEncodingTextMarshaler("text-marshaler")).Post(); err != nil {
			t.Fatalf("Request.Send() error: %s", err)
		} else {
			if got := res.String(); got != "text-marshaler" {
				t.Fatalf("Request.Send() body got: %s", got)
			}
		}

		if _, err := req.WithBody(100).Post(); err == nil {
			t.Fatal("Request.Send() with invalid body return nil error")
		} else {
			if err != ErrInvalidRequestBody {
				t.Fatalf("Request.Send() with invalid body return error: %s", err)
			}
		}

		if _, err := req.WithBody(testErrorEncodingTextMarshaler("text-marshaler")).Post(); err == nil {
			t.Fatal("Request.Send() with ErrorEncodingTextMarshaler return nil error")
		}
		if _, err := req.WithJSONBody(testErrorEncodingTextMarshaler("text-marshaler")).Post(); err == nil {
			t.Fatal("Request.Send() with ErrorEncodingTextMarshaler return nil error")
		}
		if _, err := req.WithXMLBody(testErrorEncodingTextMarshaler("text-marshaler")).Post(); err == nil {
			t.Fatal("Request.Send() with ErrorEncodingTextMarshaler return nil error")
		}

		return nil, nil
	})

	if _, err := c.New("").Send(); err == nil {
		t.Fatal("Request.Send() with empty uri return nil error")
	} else {
		if err != ErrEmptyRequestURL {
			t.Fatalf("Request.Send() with empty uri return error: %s", err)
		}
	}

	if _, err := c.New("::////").Send(); err == nil {
		t.Fatal("Request.Send() with invalid uri return nil error")
	}

	_, _ = c.Do(server.URL, func(req Request) (Response, error) {
		defer clear()

		return nil, nil
	})
}

func TestToString(t *testing.T) {
	s := "s"
	n := 1

	items := []struct {
		Give interface{}
		Want string
	}{
		{"s", "s"},
		{1, "1"},
		{1.5, "1.5"},
		{uint(1), "1"},
		{float32(1.5), "1.5"},
		{[]byte("b"), "b"},
		{errors.New("err"), "err"},
		{testFmtStringer("s"), "s"},
		{true, "true"},
		{[]string{"s", "s"}, fmt.Sprint([]string{"s", "s"})},
		{s, "s"},
		{&s, "s"},
		{n, "1"},
		{&n, "1"},
		{nil, ""},
	}

	for i, item := range items {
		if got := toString(item.Give); got != item.Want {
			t.Fatalf("ToString() [%d] want %q got %q", i, item.Want, got)
		}
	}
}
