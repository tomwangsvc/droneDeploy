package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

var (
	router = chi.NewRouter()
)

func Test_one_move(t *testing.T) {
	router.Post("/commands", runCommands())

	body, err := json.Marshal(commandRequest{
		Commands: "N",
	})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, "/commands", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if code := rr.Code; code != http.StatusAccepted {
		t.Error("status not match", code)
	}

	// wait for move to be done and check state
	<-time.After(oneMoveConsumptionTime)

	state := myRobot.CurrentState()
	if state.Y != 1 || state.X != 0 {
		t.Error("state not match", state)
	}
}

func Test_mutiple_moves(t *testing.T) {
	router.Post("/commands", runCommands())

	body, err := json.Marshal(commandRequest{
		Commands: "N N N E",
	})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, "/commands", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if code := rr.Code; code != http.StatusAccepted {
		t.Error("status not match", code)
	}

	// wait for move to be done and check state
	<-time.After(oneMoveConsumptionTime * 5)

	state := myRobot.CurrentState()
	if state.Y != 4 || state.X != 1 {
		t.Error("state not match", state)
	}
}

func Test_cancel(t *testing.T) {
	router.Post("/commands", runCommands())

	body, err := json.Marshal(commandRequest{
		Commands: "N N N N N N N N N N",
	})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, "/commands", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if code := rr.Code; code != http.StatusAccepted {
		t.Error("status not match", code)
	}
	var res commandResponse
	if err = json.Unmarshal([]byte(rr.Body.String()), &res); err != nil {
		t.Fatal(err)
	}

	// wait for one move to be complete and cancel task
	<-time.After(oneMoveConsumptionTime * 1)

	req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("/%s/cancellation", res.TaskId), nil)
	if err != nil {
		t.Fatal(err)
	}

	state := myRobot.CurrentState()
	if state.Y != 5 || state.X != 1 {
		t.Error("state not match", state)
	}
}
