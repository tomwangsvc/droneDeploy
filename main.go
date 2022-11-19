package main

import (
	"net/http"
)

// Create a RESTful API to accept a series of commands to the robot.
// Make sure that the robot doesn't try to move outside the warehouse.
// Create a RESTful API to report the command series's execution status.
// Create a RESTful API cancel the command series.
// The RESTful service should be written in Golang.

var myRobot Robot

func main() {
	ListenAndServe()
}

func init() {
	myRobot = NewRobot()
}

func ListenAndServe() error {
	return http.ListenAndServe(":8080", newRouter())
}
