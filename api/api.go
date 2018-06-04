/*
 * Copyright (C) 2018 Nikola Trubitsyn
 *
 * This file is part of coderator.
 *
 * coderator is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * coderator is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with coderator.  If not, see <https://www.gnu.org/licenses/>.
 */

package api

import (
	"github.com/gorilla/mux"
	"net/http"
	"log"
	"encoding/json"
	"strconv"
	"fmt"
	"strings"
)

const (
	ErrorNoTasks          = "No tasks found"
	ErrorTaskDoesNotExist = "The specified task does not exist"
	ErrorNoTests          = "No tests found with the specified task id"
	ErrorJobDoesNotExist  = "The specified job does not exist"
	ErrorNoResults        = "No results for the specified task"
)

const (
	PathTasks   = "/tasks"
	PathTask    = "/tasks/{id}"
	PathTests   = "/tasks/{id}/tests"
	PathSolve   = "/tasks/{id}/solve"
	PathQueue   = "/queue/{id}"
	PathResults = "/tasks/{id}/results"
)

type Error struct {
	Error string `json:"error"`
}

type Repository interface {
	AllTasks() ([]Task, error)
	FindTaskById(id uint64) (*Task, error)
	AddTask(task Task)
	FindTestsByTaskId(taskId uint64) ([]Test, error)
}

var database Repository

func Serve(repository Repository, port int) {
	database = repository
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc(PathTasks, tasksEndpoint).Methods("GET")
	router.HandleFunc(PathTask, taskEndpoint).Methods("GET")
	router.HandleFunc(PathTests, taskTestsEndpoint).Methods("GET")
	router.HandleFunc(PathSolve, taskSolveEndpoint).Methods("POST")
	router.HandleFunc(PathQueue, taskSolveQueueEndpoint).Methods("GET")
	router.HandleFunc(PathResults, taskSolveResultsEndpoint).Methods("GET")

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), router))
}

func tasksEndpoint(w http.ResponseWriter, r *http.Request) {
	tasks, err := database.AllTasks()

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{ErrorNoTasks})
		return
	}

	json.NewEncoder(w).Encode(tasks)
}

func taskEndpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars["id"]
	id, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{ErrorTaskDoesNotExist})
		return
	}

	task, err := database.FindTaskById(id)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{ErrorTaskDoesNotExist})
		return
	}

	json.NewEncoder(w).Encode(task)
}

func taskTestsEndpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskIdParam := vars["id"]

	taskId, err := strconv.ParseUint(taskIdParam, 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{ErrorTaskDoesNotExist})
		return
	}

	tests, err := database.FindTestsByTaskId(taskId)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{ErrorNoTests})
		return
	}

	json.NewEncoder(w).Encode(tests)
}

func taskSolveEndpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars["id"]

	id, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{ErrorTaskDoesNotExist})
		return
	}

	task, err := database.FindTaskById(id)

	if err != nil || task == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{ErrorTaskDoesNotExist})
		return
	}

	if IsVerificationQueued(*task) {
		w.Header().Set("Location", strings.Replace(PathQueue, "{id}", idParam, 1))
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	err = r.ParseMultipartForm(32 << 20)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error{"The request Content-Type is not multipart/form-data"})
		return
	}

	source, _, err := r.FormFile("source")

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error{"Could not parse form file"})
		return
	}
	defer source.Close()

	path, err := SaveSolution(*task, source)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	go VerifyTaskSolution(*task, path)

	w.Header().Set("Location", strings.Replace(PathQueue, "{id}", idParam, 1))
	w.WriteHeader(http.StatusAccepted)
}

func taskSolveQueueEndpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars["id"]

	id, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{ErrorJobDoesNotExist})
		return
	}

	task, err := database.FindTaskById(id)

	if err != nil || task == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{ErrorTaskDoesNotExist})
		return
	}

	if HasVerificationCompleted(*task) {
		w.Header().Set("Location", strings.Replace(PathResults, "{id}", idParam, 1))
		w.WriteHeader(http.StatusSeeOther)
	} else {
		if !IsVerificationQueued(*task) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(Error{ErrorJobDoesNotExist})
		} else {
			w.WriteHeader(http.StatusProcessing)
		}
	}
}

func taskSolveResultsEndpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars["id"]

	id, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{ErrorTaskDoesNotExist})
		return
	}

	task, err := database.FindTaskById(id)

	if err != nil || task == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{ErrorTaskDoesNotExist})
		return
	}

	results := GetVerificationResults(*task)

	if results == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{ErrorNoResults})
		return
	}

	json.NewEncoder(w).Encode(results)
}
