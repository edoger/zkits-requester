// Copyright 2021 The ZKits Project Authors.
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

package internal

import (
	"net/http"
)

// Client the default HTTP client instance.
var Client = NewClient()

// NewTransport returns a new *http.Transport instance.
func NewTransport() (t *http.Transport) {
	// Clone is available in go 1.14 or later.
	t = http.DefaultTransport.(*http.Transport).Clone()
	// Default http.Transport.MaxIdleConnsPerHost is http.DefaultMaxIdleConnsPerHost (2).
	t.MaxIdleConnsPerHost = 10
	return t
}

// NewClient returns a new *http.Client instance.
func NewClient() *http.Client {
	return &http.Client{Transport: NewTransport()}
}
