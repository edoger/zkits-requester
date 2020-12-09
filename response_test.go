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
	"errors"
	"io/ioutil"
	"net/http"
	"testing"
)

type testErrorReadCloser string

func (s testErrorReadCloser) Read([]byte) (int, error) {
	return 0, errors.New(string(s))
}

func (s testErrorReadCloser) Close() error {
	return nil
}

func TestNewResponse(t *testing.T) {
	r := &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString("body"))}
	if res, err := NewResponse(r, false); err != nil {
		t.Fatalf("NewResponse() error: %s", err)
	} else {
		if res == nil {
			t.Fatal("NewResponse() return nil")
		}
	}

	r.Body = testErrorReadCloser("err")
	if _, err := NewResponse(r, false); err == nil {
		t.Fatal("NewResponse() with ErrorReadCloser body return nil error")
	}
}

func TestResponse(t *testing.T) {
	r, err := NewResponse(&http.Response{
		StatusCode: http.StatusOK,
		Status:     http.StatusText(http.StatusOK),
		Header:     http.Header{"X-Test": {"test"}},
		Body:       ioutil.NopCloser(bytes.NewBufferString("body")),
	}, false)
	if err != nil {
		t.Fatalf("NewResponse() error: %s", err)
	}

	if got := r.Headers().Get("X-Test"); got != "test" {
		t.Fatalf("Response.Headers().Get() return %s", got)
	}
	if got := r.StatusCode(); got != http.StatusOK {
		t.Fatalf("Response.StatusCode() return %d", got)
	}
	if got := r.Status(); got != http.StatusText(http.StatusOK) {
		t.Fatalf("Response.Status() return %s", got)
	}
	if got := string(r.Body()); got != "body" {
		t.Fatalf("Response.Body() return %q", got)
	}
	if got := r.Len(); got != len("body") {
		t.Fatalf("Response.Len() return %d", got)
	}
	if got := r.String(); got != "body" {
		t.Fatalf("Response.Len() return %q", got)
	}
}

func TestResponse_JSON(t *testing.T) {
	r, err := NewResponse(&http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`"body"`))}, false)
	if err != nil {
		t.Fatalf("NewResponse() error: %s", err)
	}

	var got string
	if err := r.JSON(&got); err != nil {
		t.Fatalf("Response.JSON() error: %s", err)
	}
	if got != "body" {
		t.Fatalf("Response.JSON() got: %s", got)
	}
}

func TestResponse_XML(t *testing.T) {
	r, err := NewResponse(&http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`<XML>body</XML>`))}, false)
	if err != nil {
		t.Fatalf("NewResponse() error: %s", err)
	}

	var got string
	if err := r.XML(&got); err != nil {
		t.Fatalf("Response.XML() error: %s", err)
	}
	if got != "body" {
		t.Fatalf("Response.XML() got: %s", got)
	}
}

func TestNewResponseFrom(t *testing.T) {
	var r Response
	r = NewResponseFrom([]byte("foo"), nil, http.StatusOK)
	if r == nil {
		t.Fatal("NewResponseFrom(): return nil")
	}
	if s := r.String(); s != "foo" {
		t.Fatalf("NewResponseFrom(): %s", s)
	}
	if r.Headers() == nil {
		t.Fatal("NewResponseFrom(): nil headers")
	}
	if n := r.StatusCode(); n != http.StatusOK {
		t.Fatalf("NewResponseFrom(): %d", n)
	}
	if s := r.Status(); s != http.StatusText(http.StatusOK) {
		t.Fatalf("NewResponseFrom(): %s", s)
	}

	r = NewResponseFrom([]byte("foo"), nil, http.StatusOK, "Test")
	if s := r.Status(); s != "Test" {
		t.Fatalf("NewResponseFrom(): %s", s)
	}
}

func TestNewResponseFromPanic(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("NewResponseFrom(): no panic")
		}
	}()

	NewResponseFrom([]byte("foo"), nil, http.StatusOK, "Foo", "Bar")
}

func TestNewEmptyResponse(t *testing.T) {
	r := NewEmptyResponse()
	if r == nil {
		t.Fatal("NewEmptyResponse(): return nil")
	}
}

func TestEmptyResponse(t *testing.T) {
	r := NewEmptyResponse()
	if got := r.Status(); got != "" {
		t.Fatalf("NewEmptyResponse().Status(): %s", got)
	}
	if got := r.StatusCode(); got != 0 {
		t.Fatalf("NewEmptyResponse().StatusCode(): %d", got)
	}
	if got := r.Headers(); got == nil || len(got) != 0 {
		t.Fatalf("NewEmptyResponse().Headers(): %v", got)
	}
	if got := r.String(); got != "" {
		t.Fatalf("NewEmptyResponse().String(): %s", got)
	}
	if got := r.Body(); got != nil {
		t.Fatalf("NewEmptyResponse().String(): %s", string(got))
	}
	if got := r.Len(); got != 0 {
		t.Fatalf("NewEmptyResponse().Len(): %d", got)
	}

	var obj struct{}
	if r.JSON(&obj) == nil {
		t.Fatal("NewEmptyResponse().JSON(): nil error")
	}
	if r.XML(&obj) == nil {
		t.Fatal("NewEmptyResponse().XML(): nil error")
	}
}
