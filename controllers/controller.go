// Package controllers contains the todo backend controllers
package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"sort"
	"strconv"
	"todo-rest-backend/models"
)

const BackendHostUrl string = ":8080"

// Run does the running of the web server
func Run(enablePersistence bool) {
	if enablePersistence {
		models.EnableFilePersistence()
	} else {
		models.DisableFilePersistence()
	}

	models.Initialize()

	fmt.Println("Backend running at:", BackendHostUrl)
	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/todos", TodosGet)
	router.GET("/todos/:id", TodoGetById)
	router.POST("/todos", TodoPost)
	router.PUT("/todos/:id", TodoPut)
	router.DELETE("/todos/:id", TodoDelete)
	router.DELETE("/todos", DeleteAllTodos)

	err := http.ListenAndServe(BackendHostUrl, router)
	log.Fatal(err)
}

// Index Handler for the index action
// GET /
func Index(writer http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	writer.WriteHeader(http.StatusOK)
	_, err := fmt.Fprint(writer, "Welcome to the Todo REST API!\n")
	if err != nil {
		panic(err)
	}
}

// TodosGet Handler for the todos get action
// GET /todos
func TodosGet(writer http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	var todos []models.Todo
	for _, todo := range models.TodoStore() {
		todos = append(todos, todo)
	}

	sortedTodos := sortTodosAfterIdAscending(todos)
	response := models.JsonDataResponse{Data: sortedTodos}
	writer.Header().Set("Content-Type", "application/json; charset=UTF-8")
	writer.WriteHeader(http.StatusOK)
	err := json.NewEncoder(writer).Encode(response)
	if err != nil {
		panic(err)
	}
}

func sortTodosAfterIdAscending(todos []models.Todo) []models.Todo {
	sort.Slice(todos, func(i, j int) bool {
		leftValueAsInt, _ := strconv.Atoi(todos[i].Id)
		rightValueAsInt, _ := strconv.Atoi(todos[j].Id)
		return leftValueAsInt < rightValueAsInt
	})

	return todos
}

// TodoGetById Handler for a todo get by id action
func TodoGetById(writer http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	// Get todo id from url parameters
	id := params.ByName("id")
	todo, ok := models.TodoStore()[id]
	writer.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if ok == false {
		handleTodoIdNotFound(writer)
		return
	}
	response := models.JsonExtendedResponse{Data: todo}
	err := json.NewEncoder(writer).Encode(response)
	if err != nil {
		panic(err)
	}
}

func handleTodoIdNotFound(writer http.ResponseWriter) {
	// No todo with the id in the url parameters has been found
	writer.WriteHeader(http.StatusNotFound)
	response := models.JsonErrorResponse{Error: models.ApiError{Status: 404, Title: "Record Not Found"}}
	err := json.NewEncoder(writer).Encode(response)
	if err != nil {
		panic(err)
	}
}

// TodoPost Handler for the todos post action
func TodoPost(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	writer.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var todo models.Todo
	err := decodeTodo(request, &todo)

	if err != nil {
		handleTodoNotProperlyTransmitted(writer)
		return
	}

	todoAdded := models.AddTodo(todo)

	response := models.JsonExtendedResponse{Data: todoAdded}
	writer.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(writer).Encode(response)
	if err != nil {
		panic(err)
	}

	err = models.UpdateDataInFile()

	if err != nil {
		panic(err)
	}
}

func handleTodoNotProperlyTransmitted(writer http.ResponseWriter) {
	// todo was not properly transmitted
	writer.WriteHeader(http.StatusBadRequest)
	response := models.JsonErrorResponse{Error: models.ApiError{Status: 400, Title: "Invalid Body"}}
	err := json.NewEncoder(writer).Encode(response)
	if err != nil {
		panic(err)
	}
}

// decodeTodo does decoding of the json request body into a Todo
func decodeTodo(request *http.Request, todo *models.Todo) error {
	if request.Body == nil {
		return errors.New("invalid body")
	}
	err := json.NewDecoder(request.Body).Decode(todo)
	if err != nil {
		return err
	}
	return nil
}

// TodoPut Handler for a todo put by id action
func TodoPut(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	// Get todo id from url parameters
	id := params.ByName("id")
	_, ok := models.TodoStore()[id]
	writer.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if ok == false {
		handleTodoIdNotFound(writer)
		return
	}

	var todoReceived models.Todo
	err := decodeTodo(request, &todoReceived)
	if err != nil {
		handleTodoNotProperlyTransmitted(writer)
		return
	}

	todoUpdated, ok := models.UpdateTodo(id, todoReceived)

	if ok == false {
		handleTodoNotProperlyTransmittedGeneral(writer, "Update data model failed")
		return
	}

	response := models.JsonExtendedResponse{Data: todoUpdated}
	writer.WriteHeader(http.StatusOK)
	err = json.NewEncoder(writer).Encode(response)
	if err != nil {
		panic(err)
	}

	err = models.UpdateDataInFile()
	if err != nil {
		panic(err)
	}
}

func handleTodoNotProperlyTransmittedGeneral(writer http.ResponseWriter, title string) {
	// todo was not properly transmitted
	writer.WriteHeader(http.StatusBadRequest)
	response := models.JsonErrorResponse{Error: models.ApiError{Status: 400, Title: title}}
	err := json.NewEncoder(writer).Encode(response)
	if err != nil {
		panic(err)
	}
}

// TodoDelete Handler for a todo delete by id action
func TodoDelete(writer http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	// Get todo id from url parameters
	id := params.ByName("id")
	_, ok := models.TodoStore()[id]
	writer.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if ok == false {
		handleTodoIdNotFound(writer)
		return
	}

	models.RemoveTodo(id)

	writer.WriteHeader(http.StatusOK)

	err := models.UpdateDataInFile()
	if err != nil {
		panic(err)
	}
}

// DeleteAllTodos Handler for deleting all todo's
func DeleteAllTodos(writer http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	models.DeleteAllTodos()
	err := models.UpdateDataInFile()

	if err != nil {
		panic(err)
	}

	writer.WriteHeader(http.StatusOK)
}
