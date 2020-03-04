package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/pgxpool"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Book struct {
	Id          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Author      string `json:"author"`
}

func CreateTableDb(pool *pgxpool.Pool) {
	conn, err := getConnectionForContext(pool)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error database not available")
	}
	defer conn.Release()

	conn.Query(context.Background(), "CREATE TABLE public.books (id serial NOT NULL, title character varying(100), description character varying(1000),author character varying(100), PRIMARY KEY (id));")

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error occured while creating database")
	}
}

func GetBooksDb(pool *pgxpool.Pool) ([]Book, error) {
	var books []Book
	conn, err := getConnectionForContext(pool)
	defer conn.Release()

	rows, err := conn.Query(context.Background(), "SELECT id, title, description, author FROM books")
	if err != nil {
		return books, err
	}

	for rows.Next() {
		var book Book
		err := rows.Scan(&book.Id, &book.Title, &book.Description, &book.Author)
		if err != nil {
			return books, err
		}
		books = append(books, book)
	}

	return books, err
}

func GetBookDb(pool *pgxpool.Pool, bookId int) (Book, error) {
	var book Book
	conn, err := getConnectionForContext(pool)
	defer conn.Release()
	err = conn.QueryRow(context.Background(),
		"SELECT id, title, description, author FROM books WHERE id = $1;", bookId).Scan(&book.Id, &book.Title, &book.Description, &book.Author)
	return book, err
}

func ExistBookDb(pool *pgxpool.Pool, bookId int) (bool, error) {
	var result int
	conn, err := getConnectionForContext(pool)
	defer conn.Release()

	err = conn.QueryRow(context.Background(),
		"SELECT 1 FROM books WHERE id = $1;", bookId).Scan(&result)
	return result == 1, err
}

func AddBookDb(pool *pgxpool.Pool, book Book) (Book, error) {
	var bookId int
	conn, err := getConnectionForContext(pool)
	defer conn.Release()
	err = conn.QueryRow(context.Background(), "INSERT INTO books (title, description, author) VALUES ($1, $2, $3) RETURNING id", book.Title, book.Description, book.Author).Scan(&bookId)
	book.Id = bookId
	return book, err
}

func UpdateBookDb(pool *pgxpool.Pool, bookId int, book Book) (Book, error) {
	conn, err := getConnectionForContext(pool)
	defer conn.Release()
	err = conn.QueryRow(context.Background(), "UPDATE books SET title=$2, description=$3, author=$4 WHERE id=$1", book.Id, book.Title, book.Description, book.Author).Scan()
	return book, err
}

func DeleteBookDb(pool *pgxpool.Pool, bookId int) (Book, error) {
	var book Book
	conn, err := getConnectionForContext(pool)
	defer conn.Release()
	book, err = GetBookDb(pool, bookId)
	if err != nil {
		return book, err
	}
	err = conn.QueryRow(context.Background(), "DELETE FROM books WHERE id=$1", bookId).Scan()
	return book, err
}

func getConnectionForContext(pool *pgxpool.Pool) (*pgxpool.Conn, error) {
	conn, err := pool.Acquire(context.Background())
	return conn, err
}

var books []Book
var pool *pgxpool.Pool

func getBooks(w http.ResponseWriter, r *http.Request) {
	resp := prepareResponseHeaders(w)
	books, err := GetBooksDb(pool)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(resp).Encode(books)
}

func getBook(w http.ResponseWriter, r *http.Request) {
	resp := prepareResponseHeaders(w)
	params := mux.Vars(r)

	bookId, err := strconv.Atoi(params["id"])
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	book, err := GetBookDb(pool, bookId)

	if (book == Book{}) {
		resp.WriteHeader(http.StatusNotFound)
		return
	}
	json.NewEncoder(resp).Encode(book)
}

func addBook(w http.ResponseWriter, r *http.Request) {
	resp := prepareResponseHeaders(w)
	var book Book

	json.NewDecoder(r.Body).Decode(&book)

	if book.Id != 0 {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	createdBook, err := AddBookDb(pool, book)

	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	books = append(books, createdBook)
	json.NewEncoder(resp).Encode(createdBook)
}

func updateBook(w http.ResponseWriter, r *http.Request) {
	resp := prepareResponseHeaders(w)
	params := mux.Vars(r)
	bookId, err := strconv.Atoi(params["id"])
	var updatingBook Book

	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	json.NewDecoder(r.Body).Decode(updatingBook)

	updatedBook, err := UpdateBookDb(pool, bookId, updatingBook)

	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	if (updatedBook == Book{}) {
		resp.WriteHeader(http.StatusNotFound)
		return
	}

	json.NewEncoder(resp).Encode(updatedBook)
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	resp := prepareResponseHeaders(w)
	params := mux.Vars(r)
	bookId, err := strconv.Atoi(params["id"])
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	deletedBook, err := DeleteBookDb(pool, bookId)

	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	if (deletedBook == Book{}) {
		resp.WriteHeader(http.StatusNotFound)
		return
	}

	json.NewEncoder(resp).Encode(deletedBook)
}

func checkBook(w http.ResponseWriter, r *http.Request) {
	resp := prepareResponseHeaders(w)
	params := mux.Vars(r)

	bookId, err := strconv.Atoi(params["id"])

	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := ExistBookDb(pool, bookId)

	if err != nil || !result {
		resp.WriteHeader(http.StatusNotFound)
		return
	}

	resp.WriteHeader(http.StatusOK)
}

func prepareResponseHeaders(w http.ResponseWriter) http.ResponseWriter {
	w.Header().Set("Content-Type", "application/json")
	return w
}

func main() {
	pool, _ = pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	defer pool.Close()

	CreateTableDb(pool)

	r := mux.NewRouter()

	r.HandleFunc("/books", getBooks).Methods("GET")
	r.HandleFunc("/books/{id}", getBook).Methods("GET")
	r.HandleFunc("/books", addBook).Methods("POST")
	r.HandleFunc("/books/{id}", updateBook).Methods("PUT")
	r.HandleFunc("/books/{id}", deleteBook).Methods("DELETE")
	r.HandleFunc("/books/{id}", checkBook).Methods("HEAD")

	log.Fatal(http.ListenAndServe(":8080", r))
}
