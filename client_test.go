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
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	if New() == nil {
		t.Fatal("New() return nil")
	}
}

func TestClient_SetHTTPClient(t *testing.T) {
	c := New()
	if c.SetHTTPClient(new(http.Client)) == nil {
		t.Fatal("Client.SetHTTPClient() return nil")
	}
	if c.SetHTTPClient(nil) == nil {
		t.Fatal("Client.SetHTTPClient(nil) return nil")
	}
}

func TestClient_SetTimeout(t *testing.T) {
	c := New()
	if c.SetTimeout(time.Minute) == nil {
		t.Fatal("Client.SetTimeout() return nil")
	}
	if c.SetTimeout(0) == nil {
		t.Fatal("Client.SetTimeout(0) return nil")
	}
}

func TestClient_GetCommonHeaders(t *testing.T) {
	c := New()
	if c.GetCommonHeaders() == nil {
		t.Fatal("Client.GetCommonHeaders() return nil")
	}
}

func TestClient_SetCommonHeader(t *testing.T) {
	c := New()
	if c.SetCommonHeader("Test", "foo") == nil {
		t.Fatal("Client.SetCommonHeader() return nil")
	}
	if c.GetCommonHeaders().Get("Test") != "foo" {
		t.Fatal("Client.SetCommonHeader() failed")
	}
	c.SetCommonHeader("Test", "")
	if c.GetCommonHeaders().Get("Test") != "" {
		t.Fatal("Client.SetCommonHeader() failed")
	}
}

func TestClient_SetCommonHeaders(t *testing.T) {
	c := New()
	h := make(http.Header)
	h.Set("Test", "foo")
	if c.SetCommonHeaders(h) == nil {
		t.Fatal("Client.SetCommonHeaders() return nil")
	}
	if c.GetCommonHeaders().Get("Test") != "foo" {
		t.Fatal("Client.SetCommonHeaders() failed")
	}
	if c.SetCommonHeaders(nil) == nil {
		t.Fatal("Client.SetCommonHeaders() return nil")
	}
}

func TestClient_New(t *testing.T) {
	c := New()
	if c.New("http://test.com/foo") == nil {
		t.Fatal("Client.New() return nil")
	}
}

func TestClient_Do(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "do")
	}))
	defer server.Close()

	c := New()
	r, err := c.Do(server.URL, func(r Request) (Response, error) {
		return r.Send()
	})
	if err == nil {
		if r == nil {
			t.Fatalf("Client.Do() return nil response")
		}
		if got := r.String(); got != "do" {
			t.Fatalf("Client.Do() return %q", got)
		}
	} else {
		t.Fatalf("Client.Do() error: %s", err)
	}
}

func TestClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, r.Method)
	}))
	defer server.Close()

	c := New()

	var r Response
	var err error

	r, err = c.Head(server.URL, url.Values{"a": {"a"}})
	if err == nil {
		if r == nil {
			t.Fatalf("Client.Head() return nil response")
		}
		if got := r.String(); got != "" {
			t.Fatalf("Client.Head() return %q", got)
		}
	} else {
		t.Fatalf("Client.Head() error: %s", err)
	}

	r, err = c.Get(server.URL, url.Values{"b": {"b"}})
	if err == nil {
		if r == nil {
			t.Fatalf("Client.Get() return nil response")
		}
		if got := r.String(); got != "GET" {
			t.Fatalf("Client.Get() return %q", got)
		}
	} else {
		t.Fatalf("Client.Get() error: %s", err)
	}

	r, err = c.Post(server.URL, "foo")
	if err == nil {
		if r == nil {
			t.Fatalf("Client.Post() return nil response")
		}
		if got := r.String(); got != "POST" {
			t.Fatalf("Client.Post() return %q", got)
		}
	} else {
		t.Fatalf("Client.Post() error: %s", err)
	}

	r, err = c.PostJSON(server.URL, map[string]string{"c": "c"})
	if err == nil {
		if r == nil {
			t.Fatalf("Client.PostJSON() return nil response")
		}
		if got := r.String(); got != "POST" {
			t.Fatalf("Client.PostJSON() return %q", got)
		}
	} else {
		t.Fatalf("Client.PostJSON() error: %s", err)
	}

	r, err = c.PostXML(server.URL, []string{"d"})
	if err == nil {
		if r == nil {
			t.Fatalf("Client.PostXML() return nil response")
		}
		if got := r.String(); got != "POST" {
			t.Fatalf("Client.PostXML() return %q", got)
		}
	} else {
		t.Fatalf("Client.PostXML() error: %s", err)
	}

	r, err = c.PostForm(server.URL, url.Values{"e": {"e"}})
	if err == nil {
		if r == nil {
			t.Fatalf("Client.PostForm() return nil response")
		}
		if got := r.String(); got != "POST" {
			t.Fatalf("Client.PostForm() return %q", got)
		}
	} else {
		t.Fatalf("Client.PostForm() error: %s", err)
	}
}