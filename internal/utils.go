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
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
)

// ToString converts the given parameter to a string.
func ToString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'g', -1, 32)
	case []byte:
		return string(v)
	case bool:
		return strconv.FormatBool(v)
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

// ForceClose force close the given io.Closer ignore error.
func ForceClose(c io.Closer) {
	_ = c.Close()
}

// ShouldBeRegularFile determines whether the given target is a regular file.
func ShouldBeRegularFile(info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("target %s is not a regular file", info.Name())
	}
	return nil
}

// WriteFormDataFromReader writes the contents of a given io.Reader to upload writer.
func WriteFormDataFromReader(key, name string, w *multipart.Writer, src io.Reader) error {
	fw, err := w.CreateFormFile(key, name)
	if err != nil {
		return err
	}
	// Do we need to disallow uploading empty target?
	// Of course, it's allowed now!
	_, err = io.Copy(fw, src)
	return err
}

// WriteFormDataFromFilePath writes the contents of a given file to upload writer.
func WriteFormDataFromFilePath(key, p string, w *multipart.Writer) error {
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer ForceClose(f)

	return WriteFormDataFromReader(key, filepath.Base(f.Name()), w, f)
}

// WriteFormDataFromFileHeader writes the downstream uploaded file to upload writer.
func WriteFormDataFromFileHeader(key string, p *multipart.FileHeader, w *multipart.Writer) error {
	f, err := p.Open()
	if err != nil {
		return err
	}
	defer ForceClose(f)

	return WriteFormDataFromReader(key, filepath.Base(p.Filename), w, f)
}
