package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func newRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/commands", runCommands())
		r.Get("/status", readStatus())

		r.Route("/{task_id}", func(r chi.Router) {
			r.Post("/cancellation", cancelTask())
		})
	})

	return r
}

func cancelTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("start to cancel task")

		taskId := chi.URLParam(r, "task_id")
		if taskId == "" {
			writeResponse(w, http.StatusBadRequest, "miss task id in url")
			return
		}

		if err := myRobot.CancelTask(taskId); err != nil {
			writeResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeResponse(w, http.StatusNoContent, "")
	}
}

func readStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("start to read status")

		stateBytes, err := json.Marshal(myRobot.CurrentState())
		if err != nil {
			writeResponse(w, http.StatusInternalServerError, "")
			return
		}

		writeResponse(w, http.StatusOK, string(stateBytes))
	}
}

type commandRequest struct {
	Commands string `json:"commands"`
}

type commandResponse struct {
	TaskId string `json:"task_id"`
}

func runCommands() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("start to run commands")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeResponse(w, http.StatusInternalServerError, "failed to read request body")
			return
		}

		command := commandRequest{}
		if err = json.Unmarshal(body, &command); err != nil {
			writeResponse(w, http.StatusInternalServerError, "failed to json unmarshal command")
			return
		}

		taskID, chanState, chanError := myRobot.EnqueueTask(command.Commands)
		go func() {
			for {
				select {
				case state := <-chanState:
					log.Println("current state:", state)
				case err := <-chanError:
					if err != nil {
						log.Println("has error:", err)
					}
				}
			}
		}()

		responseBytes, err := json.Marshal(commandResponse{TaskId: taskID})
		if err != nil {
			writeResponse(w, http.StatusInternalServerError, "failed to json marshal response")
			return
		}
		// task has been accepted successfully
		// but not guarantee succeed eventually
		writeResponse(w, http.StatusAccepted, string(responseBytes))
	}
}

func writeResponse(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	if msg != "" {
		w.Write([]byte(msg))
	}
}
