package main

import (
	"fmt"
	"io/ioutil"
)

func main() {
	dat, err := ioutil.ReadFile("./test_request.http")
	if err != nil {
		fmt.Println("error")
	}
	fmt.Println(string(dat))

}
