# ZKits Requester Library #

[![Build Status](https://travis-ci.org/edoger/zkits-requester.svg?branch=master)](https://travis-ci.org/edoger/zkits-requester)
[![Coverage Status](https://coveralls.io/repos/github/edoger/zkits-requester/badge.svg?branch=master)](https://coveralls.io/github/edoger/zkits-requester?branch=master)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/8da10a218dbe4700bcbb409718538fab)](https://www.codacy.com/gh/edoger/zkits-requester/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=edoger/zkits-requester&amp;utm_campaign=Badge_Grade)

## About ##

This package is a library of ZKits project. 
This library provides an efficient and easy-to-use HTTP client.

## Usage ##

 1. Import package.
 
    ```sh
    go get -u -v github.com/edoger/zkits-requester
    ```

 2. Create a runner to run subtasks within the application.

    ```go
    package main
    
    import (
       "fmt"
    
       "github.com/edoger/zkits-requester"
    )
    
    func main() {
       client := requester.Default()
       res, err := client.New("https://test.com").Get()
       if err != nil {
           panic(err)
       }
       fmt.Println(res.String())
       
       // Send parameters.
       res, err = client.New("https://test.com").
           WithQuery("foo", "foo").
           WithJSONBody(map[string]interface{}{
               "key": "value",
           }).Post()
       if err != nil {
           panic(err)
       }
       
       var obj interface{}
       if err = res.JSON(&obj); err != nil {
           // Handle error.
       }
    }
    ```

## License ##

[Apache-2.0](http://www.apache.org/licenses/LICENSE-2.0)
