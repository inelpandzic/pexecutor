package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/inelpandzic/pexecutor/executor"
)

const jsonContentType = "application/json"

// S represents a server
type S struct {
	port   int
	ex     *executor.E
	router *mux.Router

	log *zap.Logger
}

// New creates a new server
func New(port int, ex *executor.E, log *zap.Logger) *S {
	handler := &handler{executor: ex, log: log}

	router := mux.NewRouter()
	router.Methods("POST").Path("/tasks").HandlerFunc(handler.SubmitTasks)
	router.Methods("GET").Path("/tasks/running").HandlerFunc(handler.GetRunningTasks)
	router.Methods("GET").Path("/tasks/pending").HandlerFunc(handler.GetPendingTasks)

	return &S{
		port:   port,
		ex:     ex,
		router: router,
		log:    log,
	}
}

// Serve starts up the server
func (s *S) Serve() error {
	s.log.Sugar().Infof("Server started, listening at port: %d", s.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), s.router)
}

type task struct {
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
}

type handler struct {
	executor *executor.E
	log      *zap.Logger
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
		submitted := h.executor.Submit(&executor.Task{
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
	for _, v := range h.executor.GetRunningTasks() {
		tasks = append(tasks, &task{
			Name:     v.Name,
			Duration: v.Duration / time.Millisecond,
		})
	}

	h.writeResponse(w, tasks)
}

func (h *handler) GetPendingTasks(w http.ResponseWriter, r *http.Request) {
	var tasks []*task
	for _, v := range h.executor.GetPendingTasks() {
		tasks = append(tasks, &task{
			Name:     v.Name,
			Duration: v.Duration / time.Millisecond,
		})
	}

	h.writeResponse(w, tasks)
}

func (h *handler) writeResponse(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		h.log.Sugar().Errorf("Failed writting response: %v", payload)
		http.Error(w, "Failed writing response", http.StatusInternalServerError)
	}
}
