package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

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

	logger, _ := zap.NewDevelopment()

	ex := executor.New(poolSize, queueSize, logger)
	go func() {
		ex.Run()
	}()
	defer ex.Close()

	handler := &handler{Executor: ex, Log: logger}

	router := mux.NewRouter()
	router.Methods("POST").Path("/tasks").HandlerFunc(handler.SubmitTasks)
	router.Methods("GET").Path("/tasks/running").HandlerFunc(handler.GetRunningTasks)
	router.Methods("GET").Path("/tasks/pending").HandlerFunc(handler.GetPendingTasks)

	logger.Sugar().Infof("Server started, listening at port: %d", port)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), router); err != nil {
		logger.Sugar().Fatal("Server failed", err)
	}
}

type handler struct {
	Executor *executor.E
	Log      *zap.Logger
}

type task struct {
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
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

	submittedTasks := 0
	dublicateTasks := 0
	for _, t := range tasks {
		submitted := h.Executor.Submit(&executor.Task{
			Name:     t.Name,
			Duration: t.Duration * time.Millisecond,
		})

		if submitted {
			submittedTasks++
		} else {
			dublicateTasks++
		}
	}

	res := &struct {
		RequestedTask  int
		SubmittedTasks int
		DuplicateTasks int
	}{
		RequestedTask:  len(tasks),
		SubmittedTasks: submittedTasks,
		DuplicateTasks: dublicateTasks,
	}
	h.writeResponse(w, res)
}

func (h *handler) GetRunningTasks(w http.ResponseWriter, r *http.Request) {
	var tasks []*task
	for _, v := range h.Executor.GetRunningTasks() {
		tasks = append(tasks, &task{
			Name:     v.Name,
			Duration: v.Duration * time.Millisecond,
		})
	}

	h.writeResponse(w, tasks)
}

func (h *handler) GetPendingTasks(w http.ResponseWriter, r *http.Request) {
	var tasks []*task
	for _, v := range h.Executor.GetPendingTasks() {
		tasks = append(tasks, &task{
			Name:     v.Name,
			Duration: v.Duration * time.Millisecond,
		})
	}

	h.writeResponse(w, tasks)
}

func (h *handler) writeResponse(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		h.Log.Sugar().Errorf("Failed writting response: %v", payload)
		http.Error(w, "Failed writing response", http.StatusInternalServerError)
	}
}
