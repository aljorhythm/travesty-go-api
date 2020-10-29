package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// Book struct
type Book struct {
	ID     string  `json:"id"`
	Isbn   string  `json:"isbn"`
	Title  string  `json:"title"`
	Author *Author `json:"author"`
}

// Author struct
type Author struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

// IDCounter struct
type IDCounter struct {
	counter int
}

var idCounter = IDCounter{0}

func (idCounter *IDCounter) next() int {
	fmt.Printf("counter++")
	idCounter.counter++
	return idCounter.counter
}

// Init books var as a slice Book struct
var books []Book

func update(target *Book, source Book) {
	target.Isbn = source.Isbn
	target.Title = source.Title
	target.Author = source.Author
}

func findBookByID(id string) *Book {
	fmt.Printf("find book by id %v\n", id)
	for index, item := range books {
		if item.ID == id {
			return &books[index]
		}
	}
	return &Book{}
}

func getBooks(w http.ResponseWriter, r *http.Request) {
	fmt.Println("getBooks()")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func getBook(w http.ResponseWriter, r *http.Request) {
	fmt.Println("getBook()")
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	item := findBookByID(params["id"])
	if item != nil {
		json.NewEncoder(w).Encode(item)
	} else {
		json.NewEncoder(w).Encode(&Book{})
	}
}

func createBook(w http.ResponseWriter, r *http.Request) {
	fmt.Println("createBook()")
	w.Header().Set("Content-Type", "application/json")
	var book Book
	body, _ := ioutil.ReadAll(r.Body)
	_ = json.NewDecoder(bytes.NewReader(body)).Decode(&book)
	book.ID = strconv.Itoa(idCounter.next())
	books = append(books, book)
	json.NewEncoder(w).Encode(book)
}

func updateBook(w http.ResponseWriter, r *http.Request) {
	fmt.Println("updateBook()")
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	item := findBookByID(params["id"])

	var source Book
	body, _ := ioutil.ReadAll(r.Body)
	_ = json.NewDecoder(bytes.NewReader(body)).Decode(&source)
	update(item, source)
	fmt.Println(item)
	fmt.Println(books)

	json.NewEncoder(w).Encode(item)
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	fmt.Println("deleteBook()")
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	search, success := params["id"]
	if !success {
		fmt.Printf("Error")
		return
	}
	for index, item := range books {
		if item.ID == search {
			fmt.Println(books[:index])
			books = append(books[:index], books[index+1:]...)
			return
		}
	}
	fmt.Println("not found")
}

func printRes(resp *http.Response, err error) {
	fmt.Println(resp.Request.Method)
	fmt.Println(resp.Request.URL)
	fmt.Println(resp.Header)
	if err != nil {
		fmt.Println(err)
	} else {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		if body != nil {
			fmt.Println(string(body))
		}
	}
	fmt.Println()
}

func setUpServer() *http.Server {
	books = append(books, Book{strconv.Itoa(idCounter.next()), "448743", "Book One", &Author{Firstname: "John", Lastname: "Doe"}})

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "hello world")
	}).Methods("GET")
	r.HandleFunc("/api/books", getBooks).Methods("GET")
	r.HandleFunc("/api/books/{id}", getBook).Methods("GET")
	r.HandleFunc("/api/books", createBook).Methods("POST")
	r.HandleFunc("/api/books/{id}", updateBook).Methods("PATCH")
	r.HandleFunc("/api/books/{id}", deleteBook).Methods("DELETE")

	m := http.NewServeMux()
	m.Handle("/", r)
	s := http.Server{Addr: ":8080", Handler: m}

	go func() {
		s.ListenAndServe()
	}()
	return &s
}

func requests() {
	client := &http.Client{}

	fmt.Println("Client: Hello world")
	printRes(http.Get("http://localhost:8080/"))

	fmt.Println("Client: See Book List")
	printRes(http.Get("http://localhost:8080/api/books"))

	fmt.Println("Client: Create Book")
	newBook := map[string]interface{}{
		"isbn":   "BOOK2",
		"title":  "How to code in golang",
		"author": map[string]interface{}{"firstname": "Johnny", "lastname": "Duck"}}
	jsonValue, _ := json.Marshal(newBook)
	printRes(http.Post("http://localhost:8080/api/books", "application/json", bytes.NewBuffer(jsonValue)))

	fmt.Println("Client: See Book List")
	printRes(http.Get("http://localhost:8080/api/books"))

	fmt.Println("Client: Delete Book ID: 1")
	request, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://localhost:8080/api/books/%s", "1"), bytes.NewBuffer(nil))
	printRes(client.Do(request))

	fmt.Println("Client: See Book List")
	printRes(http.Get("http://localhost:8080/api/books"))

	fmt.Println("Client: Update Book")
	newBook = map[string]interface{}{
		"id":     "2",
		"isbn":   "BOOK2",
		"title":  "How to code in golang (Version 2)",
		"author": map[string]interface{}{"firstname": "Johnny", "lastname": "Duck"}}
	jsonValue, _ = json.Marshal(newBook)
	request, _ = http.NewRequest(http.MethodPatch, fmt.Sprintf("http://localhost:8080/api/books/%s", "2"), bytes.NewBuffer(jsonValue))
	printRes(client.Do(request))

	fmt.Println("Client: See Book List")
	printRes(http.Get("http://localhost:8080/api/books"))
}

func main() {
	fmt.Println("Initializing")

	//server
	s := setUpServer()

	time.Sleep(1 * time.Second)

	// client
	requests()
	time.Sleep(1 * time.Second)

	log.Fatal(s.Shutdown(context.Background()))
}
