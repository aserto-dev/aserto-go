package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

const port = ":8000"

func main() {
	s := newServer()
	r := mux.NewRouter()

	r.Use(basicAuth, jsonContentType)

	r.HandleFunc("/boards", s.CreateBoard).Methods(http.MethodPost)
	r.HandleFunc("/boards", s.ListBoards).Methods(http.MethodGet)
	r.HandleFunc("/boards/{id}", s.PostMessage).Methods(http.MethodPost)
	r.HandleFunc("/boards/{boardID}/messages/{messageID}", s.DeleteMessage).Methods(http.MethodDelete)

	// Start server
	fmt.Printf("Starting server on 'localhost%s'\n", port)
	log.Fatal(http.ListenAndServe(port, r))
}

type server struct {
	boards map[BoardID]MessageBoard
	nextID uint64
}

func newServer() *server {
	return &server{boards: map[BoardID]MessageBoard{}}
}

type (
	BoardID   uint64
	MessageID uint64
)

type MessageBoard struct {
	ID    BoardID `json:"id"`
	Name  string  `json:"name"`
	Owner string  `json:"owner"`

	Messages map[MessageID]BoardMessage `json:"-"`
}

type BoardMessage struct {
	ID           MessageID `json:"id"`
	Sender       string    `json:"sender"`
	BoardID      BoardID   `json:"board_id"`
	CreationTime time.Time `json:"creation_time"`
	Msg          string    `json:"message"`
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

	board := MessageBoard{
		ID:       BoardID(s.newID()),
		Name:     name,
		Owner:    user.(string),
		Messages: map[MessageID]BoardMessage{},
	}

	s.boards[board.ID] = board

	json.NewEncoder(w).Encode(&board)
}

func (s *server) ListBoards(w http.ResponseWriter, r *http.Request) {
	boards := []MessageBoard{}
	for _, b := range s.boards {
		boards = append(boards, b)
	}

	json.NewEncoder(w).Encode(boards)
}

func (s *server) newID() uint64 {
	s.nextID++
	return s.nextID
}

func (s *server) PostMessage(w http.ResponseWriter, r *http.Request) {
	var msg BoardMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, fmt.Errorf("%w: failed to decode posted message").Error(), http.StatusBadRequest)
		return
	}

	boardID, err := boardIDFromString(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	board, ok := s.boards[BoardID(boardID)]
	if !ok {
		http.Error(w, fmt.Sprintf("board with id '%d' doesn't exist", boardID), http.StatusBadRequest)
		return
	}

	msg.ID = MessageID(s.newID())
	msg.Sender = r.Context().Value("user").(string)
	msg.BoardID = board.ID
	msg.CreationTime = time.Now()

	board.Messages[msg.ID] = msg

	json.NewEncoder(w).Encode(msg)
}

func (s *server) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boardID, err := boardIDFromString(vars["boardID"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	messageID, err := messageIDFromString(vars["MessageID"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	board, ok := s.boards[boardID]
	if !ok {
		http.Error(w, "message board not found", http.StatusNotFound)
		return
	}

	delete(board.Messages, messageID)
}

func basicAuth(next http.Handler) http.Handler {
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

func boardIDFromString(id string) (BoardID, error) {
	boardID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid board id '%s': %w", id, err)
	}
	return BoardID(boardID), nil
}

func messageIDFromString(id string) (MessageID, error) {
	messageID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid message id '%s': %w", id, err)
	}
	return MessageID(messageID), nil
}

func checkPassword(user, password string) bool {
	return user == password
}

func jsonContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
