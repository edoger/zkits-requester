# ZKits Requester Library #

[![ZKits](https://img.shields.io/badge/ZKits-Library-f3c)](https://github.com/edoger/zkits-requester)
[![Build Status](https://travis-ci.org/edoger/zkits-requester.svg?branch=master)](https://travis-ci.org/edoger/zkits-requester)
[![Build status](https://ci.appveyor.com/api/projects/status/0hreadwn4rb9mkw2/branch/master?svg=true)](https://ci.appveyor.com/project/edoger56924/zkits-requester/branch/master)
[![Coverage Status](https://coveralls.io/repos/github/edoger/zkits-requester/badge.svg?branch=master)](https://coveralls.io/github/edoger/zkits-requester?branch=master)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/8da10a218dbe4700bcbb409718538fab)](https://www.codacy.com/gh/edoger/zkits-requester/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=edoger/zkits-requester&amp;utm_campaign=Badge_Grade)
[![Go Report Card](https://goreportcard.com/badge/github.com/edoger/zkits-requester)](https://goreportcard.com/report/github.com/edoger/zkits-requester)
[![Golang Version](https://img.shields.io/badge/golang-1.14+-orange)](https://github.com/edoger/zkits-requester)

## About ##

This package is a library of ZKits project. 
This library provides an efficient and easy-to-use HTTP client.

## Install ##

```sh
go get -u -v github.com/edoger/zkits-requester
```

## Usage ##
```go
package main

import (
    "fmt"

    "github.com/edoger/zkits-requester"
)

func main() {
    client := requester.Default()

    // client.Get("https://test.com", nil)
    res, err := client.New("https://test.com").Get()
    if err != nil {
        panic(err)
    }
    fmt.Println(res.String())

    data := map[string]interface{}{"key": "value"}
    // client.PostJSON("https://test.com", data)
    res, err = client.New("https://test.com").WithJSONBody(data).Post()
    if err != nil {
        panic(err)
    }

    obj := make(map[string]interface{})
    // Bind response json to object.
    if err = res.JSON(&obj); err != nil {
        panic(err)
    }

    // client.UploadFile("https://test.com", "upload", "path/to/file")
    res, err = client.New("https://test.com").WithFormDataFile("upload", "path/to/file").Upload()
    if err != nil {
        panic(err)
    }
    fmt.Println(res.String())
}
```

## License ##

[Apache-2.0](http://www.apache.org/licenses/LICENSE-2.0)
