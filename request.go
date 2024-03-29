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
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/edoger/zkits-requester/internal"
)

var (
	// ErrEmptyRequestURL represents an empty request url error.
	// When sending a request, this error will be returned if the given request url is empty.
	ErrEmptyRequestURL = errors.New("empty request url")

	// ErrInvalidRequestBody indicates an invalid request body error.
	// When sending a request, if the bound request body cannot be recognized, this error is returned.
	ErrInvalidRequestBody = errors.New("invalid request body")

	// ErrEmptyUploadBody represents an empty upload body error.
	// When sending an upload request, this error is returned when the upload data is empty.
	ErrEmptyUploadBody = errors.New("empty upload body")

	// ErrInvalidUploadBody indicates an invalid upload body error.
	// When sending an upload request, this error is returned when the bound upload data is invalid.
	ErrInvalidUploadBody = errors.New("invalid upload body")
)

// Request interface defines client requests.
type Request interface {
	// WithMethod adds the default request method of the current request.
	WithMethod(string) Request

	// WithResponder adds the given responder to the current request.
	WithResponder(Responder) Request

	// WithHeader adds a request header to the current request.
	// If the given request header value is an empty string, the corresponding request
	// header will be deleted.
	WithHeader(string, string) Request

	// WithHeaderValue adds a request header to the current request.
	// This method will automatically convert the given parameter value to a string.
	WithHeaderValue(string, interface{}) Request

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

	// WithFormBodyMap adds the request body key-value map as form (urlencoded).
	// This method will force set "Content-Type" to "application/x-www-form-urlencoded".
	WithFormBodyMap(body map[string]interface{}) Request

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

	// WithFormDataField adds a form data for uploading to the current request.
	// If the given form value is nil, delete the corresponding form key.
	WithFormDataField(string, interface{}) Request

	// WithFormDataFields adds a form data key-value map for uploading to the current request.
	// If the given form value is nil, delete the corresponding form key.
	WithFormDataFields(fields map[string]interface{}) Request

	// WithFormDataFile adds a file target for uploading to the current request.
	// If the given form value is nil, delete the corresponding form key.
	// This method supports adding string paths, opened file descriptors and
	// downstream uploaded files as upload targets.
	WithFormDataFile(string, interface{}) Request

	// WithFormDataFileFromReader adds an upload target to the current request from
	// the given reader and file name.
	// If the given reader is nil, delete the corresponding form key.
	WithFormDataFileFromReader(string, string, io.Reader) Request

	// ClearFormData removes all uploaded form data from the current request.
	// If you need to reuse the current request instance to send multiple upload requests,
	// you should clean up each time before setting new form data (for the added file descriptor,
	// because the position of the cursor is uncertain, it may cause unexpected results ).
	ClearFormData() Request

	// Upload sends the current upload request and receives the response.
	// This method will send the request using the POST method.
	Upload() (Response, error)

	// UploadBy sends the current upload request and receives the response.
	// This method will send the request using the given request method.
	// This method only supports POST method and PUT method.
	UploadBy(string) (Response, error)

	// Clear cleans up the current request instance so that it can be reused.
	// This method will not cut the connection with the client, nor will it
	// change the request url.
	Clear() Request
}

// The request type is a built-in implementation of the Request interface.
type request struct {
	client       *client
	uri          string
	method       string
	headers      http.Header
	ctx          context.Context
	query        url.Values
	timeout      time.Duration
	responder    Responder
	body         interface{}
	bodyFormData map[string][]*formDataValue
	bodyEncoder  string
	bodyType     string
}

// The formDataValue type defines a single upload form data.
type formDataValue struct {
	field string
	file  interface{}
}

// WithMethod adds the default request method of the current request.
func (r *request) WithMethod(method string) Request {
	r.method = method
	return r
}

// WithResponder adds the given responder to the current request.
func (r *request) WithResponder(responder Responder) Request {
	r.responder = responder
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

// WithHeaderValue adds a request header to the current request.
// This method will automatically convert the given parameter value to a string.
func (r *request) WithHeaderValue(key string, value interface{}) Request {
	return r.WithHeader(key, internal.ToString(value))
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
	return r.WithQuery(key, internal.ToString(value))
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

// WithFormBodyMap adds the request body key-value map as form (urlencoded).
// This method will force set "Content-Type" to "application/x-www-form-urlencoded".
func (r *request) WithFormBodyMap(body map[string]interface{}) Request {
	values := make(url.Values, len(body))
	for key, value := range body {
		values.Set(key, internal.ToString(value))
	}
	return r.WithFormBody(values)
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
		return r.fromResponder(o, o.Request.Method == http.MethodHead)
	}
}

// Build the Response instance from the responder.
func (r *request) fromResponder(o *http.Response, noBody bool) (Response, error) {
	if r.responder != nil {
		return r.responder(o, noBody)
	}
	if r.client.responder != nil {
		return r.client.responder(o, noBody)
	}
	return NewResponse(o, noBody)
}

// The send method sends the current request and returns the received response.
func (r *request) send(method string) (*http.Response, error) {
	body, err := r.makeBodyReader()
	if err != nil {
		return nil, err
	}

	ctx, cancel := r.getContext()
	if cancel != nil {
		// The context must be explicitly cancelled after the end of the request,
		// so that the resources of this request can be released quickly.
		defer cancel()
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
		return internal.Client.Do(req)
	} else {
		return r.client.http.Do(req)
	}
}

// Gets the request context.
func (r *request) getContext() (context.Context, context.CancelFunc) {
	ctx := r.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if r.timeout > 0 {
		return context.WithTimeout(ctx, r.timeout)
	}
	if r.client.timeout > 0 {
		return context.WithTimeout(ctx, r.client.timeout)
	}
	return ctx, nil
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

// WithFormDataField adds a form data for uploading to the current request.
// If the given form value is nil, delete the corresponding form key.
func (r *request) WithFormDataField(key string, value interface{}) Request {
	if value == nil {
		return r.withFormData(key, nil)
	}
	return r.withFormData(key, &formDataValue{field: internal.ToString(value)})
}

// WithFormDataFields adds a form data key-value map for uploading to the current request.
// If the given form value is nil, delete the corresponding form key.
func (r *request) WithFormDataFields(fields map[string]interface{}) Request {
	for key, value := range fields {
		r.WithFormDataField(key, value)
	}
	return r
}

// WithFormDataFile adds a file target for uploading to the current request.
// If the given form value is nil, delete the corresponding form key.
// This method supports adding string paths, opened file descriptors and
// downstream uploaded files as upload targets.
func (r *request) WithFormDataFile(key string, value interface{}) Request {
	if value == nil {
		return r.withFormData(key, nil)
	}
	return r.withFormData(key, &formDataValue{file: value})
}

// The formDataFileReader type is used to warp the io.Reader and a file name
// into a file target for upload.
type formDataFileReader struct {
	name   string
	reader io.Reader
}

// WithFormDataFileFromReader adds an upload target to the current request from
// the given reader and file name.
// If the given reader is nil, delete the corresponding form key.
func (r *request) WithFormDataFileFromReader(key string, name string, reader io.Reader) Request {
	if reader == nil {
		return r.withFormData(key, nil)
	}
	return r.withFormData(key, &formDataValue{file: &formDataFileReader{name: name, reader: reader}})
}

// The withFormData method adds an upload form data to the current request.
func (r *request) withFormData(key string, value *formDataValue) *request {
	if value == nil {
		if len(r.bodyFormData) > 0 {
			delete(r.bodyFormData, key)
		}
		return r
	}
	if r.bodyFormData == nil {
		r.bodyFormData = make(map[string][]*formDataValue, 1)
	}
	r.bodyFormData[key] = append(r.bodyFormData[key], value)
	return r
}

// ClearFormData removes all uploaded form data from the current request.
// If you need to reuse the current request instance to send multiple upload requests,
// you should clean up each time before setting new form data (for the added file descriptor,
// because the position of the cursor is uncertain, it may cause unexpected results ).
func (r *request) ClearFormData() Request {
	r.bodyFormData = nil
	return r
}

// Upload sends the current upload request and receives the response.
// This method will send the request using the POST method.
func (r *request) Upload() (Response, error) {
	return r.UploadBy(http.MethodPost)
}

// UploadBy sends the current upload request and receives the response.
// This method will send the request using the given request method.
// This method only supports POST method and PUT method.
func (r *request) UploadBy(method string) (Response, error) {
	if r.uri == "" {
		return nil, ErrEmptyRequestURL
	}

	method = strings.ToUpper(method)
	if method != http.MethodPost && method != http.MethodPut {
		return nil, fmt.Errorf("unsupported upload method: %s", method)
	}
	if len(r.bodyFormData) == 0 {
		return nil, ErrEmptyUploadBody
	}

	b := new(bytes.Buffer)
	w := multipart.NewWriter(b)

	keys := make([]string, 0, len(r.bodyFormData))
	for key := range r.bodyFormData {
		keys = append(keys, key)
	}
	if len(keys) > 1 {
		sort.Strings(keys)
	}

	for i, j := 0, len(keys); i < j; i++ {
		for m, n := 0, len(r.bodyFormData[keys[i]]); m < n; m++ {
			if r.bodyFormData[keys[i]][m].file == nil {
				// This is normal form data.
				if err := w.WriteField(keys[i], r.bodyFormData[keys[i]][m].field); err != nil {
					return nil, err
				}
				continue
			}

			switch v := r.bodyFormData[keys[i]][m].file.(type) {
			case string:
				if err := internal.ShouldBeRegularFile(os.Stat(v)); err != nil {
					return nil, err
				}
				if err := internal.WriteFormDataFromFilePath(keys[i], v, w); err != nil {
					return nil, err
				}
			case *os.File:
				if err := internal.ShouldBeRegularFile(v.Stat()); err != nil {
					return nil, err
				}
				// Here we cannot determine the offset in the file pointed to by the file descriptor,
				// we only read all the file content as upload content.
				// After we finish reading, we do not return the original position of the cursor and
				// close the file descriptor, because we are not sure whether this file is used elsewhere.
				if err := internal.WriteFormDataFromReader(keys[i], filepath.Base(v.Name()), w, v); err != nil {
					return nil, err
				}
			case *multipart.FileHeader:
				// For interim uploads, we read data directly from downstream uploaded files.
				if err := internal.WriteFormDataFromFileHeader(keys[i], v, w); err != nil {
					return nil, err
				}
			case *formDataFileReader:
				if err := internal.WriteFormDataFromReader(keys[i], filepath.Base(v.name), w, v.reader); err != nil {
					return nil, err
				}
			default:
				return nil, ErrInvalidUploadBody
			}
		}
	}

	// It must be closed before the request is sent, otherwise the content of
	// the request body will be incomplete.
	if err := w.Close(); err != nil {
		return nil, err
	}

	r.body = b
	r.bodyEncoder = ""
	r.bodyType = w.FormDataContentType()

	if o, err := r.send(method); err != nil {
		return nil, err
	} else {
		return r.fromResponder(o, false)
	}
}

// Clear cleans up the current request instance so that it can be reused.
// This method will not cut the connection with the client, nor will it
// change the request url.
func (r *request) Clear() Request {
	r.method = ""
	r.headers = nil
	r.ctx = nil
	r.query = nil
	r.timeout = 0
	r.body = nil
	r.bodyEncoder = ""
	r.bodyType = ""
	r.responder = nil

	return r.ClearFormData()
}

// The reset method resets the current instance.
func (r *request) reset() *request {
	r.Clear()
	r.client = nil
	r.uri = ""

	return r
}
