// Package models contains the todo backend models
package models

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"
)

type Todo struct {
	// The main identifier for the Todo. This will be unique.
	Id          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Terminated  bool   `json:"terminated"`
}

func (t Todo) Serialize() []string {
	todoSerialized := []string{t.Id, t.Title, t.Description, strconv.FormatBool(t.Terminated)}
	return todoSerialized
}

type JsonExtendedResponse struct {
	// Reserved field to add some meta information to the API response
	Meta interface{} `json:"meta"`
	Data interface{} `json:"data"`
}

type JsonDataResponse struct {
	Data []Todo `json:"data"`
}

type JsonErrorResponse struct {
	Error ApiError `json:"error"`
}

type ApiError struct {
	Status int16  `json:"status"`
	Title  string `json:"title"`
}

const FileName = "data.csv"

// Todo persistence
var filePersistence = false

// EnableFilePersistence enables the file persistence
func EnableFilePersistence() {
	filePersistence = true
}

// DisableFilePersistence disables the file persistence
func DisableFilePersistence() {
	filePersistence = false
}

// A map to store the todos with the ID as the key
// This acts as the storage in lieu of an actual database
var todoStore = make(map[string]Todo)

// TodoStore Getter method
func TodoStore() map[string]Todo {
	// Note that maps and slices are descriptors. If you return a map value, it will refer to the same underlying data structures.
	// Therefore, a clone is created.
	return clone(todoStore)
}

func clone(m map[string]Todo) map[string]Todo {
	m2 := make(map[string]Todo, len(m))

	for k, v := range m {
		m2[k] = v
	}
	return m2
}

// AddTodo adds a todo to the store
func AddTodo(todo Todo) Todo {
	indexAsInt := len(todoStore)
	indexAsString := strconv.Itoa(indexAsInt)

	todo.Id = indexAsString
	todoStore[indexAsString] = todo

	return todo
}

// UpdateTodo allows to set a todo
// If id not equals to todo.Id, then the todo.Id is set based on id.
func UpdateTodo(id string, todo Todo) (Todo, bool) {
	_, ok := todoStore[id]
	if ok == false {
		return Todo{}, false
	}

	if id != todo.Id {
		todo.Id = id
	}

	todoStore[id] = todo

	return todo, true
}

// RemoveTodo removes a todo from the store
func RemoveTodo(id string) bool {
	_, ok := todoStore[id]
	if ok == false {
		return false
	}

	var tempTodoStore = make(map[string]Todo)
	var index int = 0

	for _, currentTodo := range todoStore {
		if id != currentTodo.Id {
			// Add todo's from the original store to the temp store except the one to be deleted
			indexAsString := strconv.Itoa(index)
			currentTodo.Id = indexAsString
			tempTodoStore[indexAsString] = currentTodo
			index += 1
		}
	}

	todoStore = tempTodoStore

	return true
}

// Initialize does the initialization of the repository
func Initialize() {
	if filePersistence {
		todoStore, _ = getDataFromFile()
	}
}

func getDataFromFile() (map[string]Todo, error) {
	// open file
	//
	file, err := os.Open(FileName)
	if err != nil {
		return nil, err
	}

	var readTodos = make(map[string]Todo)

	// read csv values using csv.Reader
	//
	csvReader := csv.NewReader(file)
	rowIndex := 0
	for {
		records, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		rowIndexAsString := strconv.Itoa(rowIndex)

		// Add todo to map
		//
		readTodos[rowIndexAsString] = parseTodoData(records)
		rowIndex = rowIndex + 1
	}

	// remember to close the file at the end
	//
	err = file.Close()

	if err != nil {
		return nil, err
	}

	return readTodos, nil
}

func parseTodoData(rec []string) Todo {
	// Parse todo
	//
	id := rec[0]
	title := rec[1]
	description := rec[2]
	terminated := ToBool(rec[3])

	// Create new todo based on parsed values
	//
	todo := Todo{Id: id, Title: title, Description: description, Terminated: terminated}
	return todo
}

// ToBool converts a string to a boolean value
func ToBool(info string) bool {
	aBool, _ := strconv.ParseBool(info)
	return aBool
}

// UpdateDataInFile updates the data in the file by writing todo store to file.
func UpdateDataInFile() error {
	if filePersistence == false {
		return nil
	}

	file, err := os.OpenFile(FileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	checkError("Cannot open file", err)
	writer := csv.NewWriter(file)

	for _, todo := range todoStore {
		err := writer.Write(todo.Serialize())
		checkError("Cannot write to file", err)
	}

	writer.Flush()
	err = file.Close()

	if err != nil {
		return err
	}

	return nil
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

func DeleteAllTodos() {
	todoStore = make(map[string]Todo)
}
