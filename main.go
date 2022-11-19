package main

import (
	"net/http"
)

var myRobot Robot

func main() {
	ListenAndServe()
}

func init() {
	// always initilize my robot for test usage
	myRobot = NewRobot()
}

func ListenAndServe() error {
	return http.ListenAndServe(":8080", newRouter())
}
