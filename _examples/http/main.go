package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type (
	BoardID   uint64
	MessageID uint64
)

type MessageBoard struct {
	ID       BoardID        `json:"id"`
	Name     string         `json:"name"`
	Owner    string         `json:"owner"`
	Messages []BoardMessage `json:"-"`
}

type BoardMessage struct {
	ID           MessageID `json:"id"`
	Sender       string    `json:"sender"`
	BoardID      uint64    `json:"board_id"`
	CreationTime time.Time `json:"creation_time"`
	Msg          string    `json:"message"`
}

type server struct {
	boards map[BoardID]MessageBoard
	nextID uint64
}

func newServer() *server {
	return &server{boards: map[BoardID]MessageBoard{}}
}

func (s *server) CreateBoard(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		fmt.Println("Missing parameter")
		http.Error(w, "missing query parameter 'name'", http.StatusBadRequest)
		return
	}

	user := r.Context().Value("user")
	if user == nil {
		http.Error(w, "no user in context", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	board := MessageBoard{
		ID:    BoardID(s.newID()),
		Name:  name,
		Owner: user.(string),
	}

	s.boards[board.ID] = board

	json.NewEncoder(w).Encode(&board)
}

func (s *server) newID() uint64 {
	s.nextID++
	return s.nextID
}

func BasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok || !checkPassword(user, password) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "user", user))

		next.ServeHTTP(w, r)
	})
}

func checkPassword(user, password string) bool {
	return user == password
}

const port = ":8000"

func main() {
	s := newServer()
	r := mux.NewRouter()

	r.Use(BasicAuth)

	r.HandleFunc("/boards", s.CreateBoard).Methods("POST")

	fmt.Printf("Starting server on 'localhost%s'\n", port)

	// Start server
	log.Fatal(http.ListenAndServe(port, r))
}
