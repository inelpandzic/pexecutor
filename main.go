package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/inelpandzic/pexecutor/executor"
)

const (
	port            = 8080
	jsonContentType = "application/json"

	defaulPoolSize   = 10
	defaultQueueSize = 1000
)

var poolSize int
var queueSize int

func main() {
	flag.IntVar(&poolSize, "pool-size", defaulPoolSize, "Worker pool size")
	flag.IntVar(&queueSize, "queue-size", defaultQueueSize, "Executor task queue size")
    flag.Parse()

	ex := executor.New(poolSize, queueSize)
	go func() {
		ex.Run()
	}()
	defer ex.Close()

	handler := &handler{Executor: ex}

	router := mux.NewRouter()
	router.Methods("POST").Path("/tasks").HandlerFunc(handler.SubmitTasks)
	router.Methods("GET").Path("/tasks/running").HandlerFunc(handler.GetRunningTasks)
	router.Methods("GET").Path("/tasks/pending").HandlerFunc(handler.GetPendingTasks)

	log.Printf("Server started, listening at port: %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
}

type handler struct {
	Executor *executor.E
}

type task struct {
	Name     string `json:"name"`
	Duration int    `json:"duration"`
}

func (h *handler) SubmitTasks(w http.ResponseWriter, r *http.Request) {
	if ct := r.Header.Get("Content-Type"); ct != jsonContentType {
		http.Error(w, fmt.Sprintf("Wrong content type, %s. Must be %s", ct, jsonContentType), http.StatusUnsupportedMediaType)
		return
	}

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed reading request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var tasks []*task

	err = json.Unmarshal(bytes, &tasks)
	if err != nil {
		http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	for _, t := range tasks {
		h.Executor.Submit(&executor.Task{
			Name:     t.Name,
			Duration: time.Duration(t.Duration) * time.Millisecond,
		})
	}


    // TODO: return {"submittedTasks: 23, dublicateTasks: 10"}
	w.WriteHeader(http.StatusOK)
}

func (h *handler) GetRunningTasks(w http.ResponseWriter, r *http.Request) {
    var tasks []*task
    for _, v := range h.Executor.GetRunningTasks()  {
        tasks = append(tasks, &task{
            Name: v.Name,
            Duration: int(v.Duration),
        })
    }

    writeResponse(w, tasks)
}

func (h *handler) GetPendingTasks(w http.ResponseWriter, r *http.Request) {
    var tasks []*task
    for _, v := range h.Executor.GetPendingTasks()  {
        tasks = append(tasks, &task{
            Name: v.Name,
            Duration: int(v.Duration),
        })
    }

    writeResponse(w, tasks)
}


func writeResponse(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Failed writting response: %v", payload)
		http.Error(w, "Failed writing response", http.StatusInternalServerError)
	}
}
