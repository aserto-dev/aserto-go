package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type (
	BoardID   uint64
	MessageID uint64
)

type Board struct {
	ID    BoardID `json:"id"`
	Name  string  `json:"name"`
	Owner string  `json:"owner"`

	Messages map[MessageID]Message `json:"-"`
}

type Message struct {
	ID           MessageID `json:"id"`
	Sender       string    `json:"sender"`
	BoardID      BoardID   `json:"board_id"`
	CreationTime time.Time `json:"creation_time"`
	Msg          string    `json:"message"`
}

type MsgBoardsSvc struct {
	boards   map[BoardID]Board
	nextID   uint64
	identity Identity
}

func newMsgBoardsSvc(identity Identity) *MsgBoardsSvc {
	return &MsgBoardsSvc{boards: map[BoardID]Board{}, identity: identity}
}

func (s *MsgBoardsSvc) CreateBoard(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "missing query parameter 'name'", http.StatusBadRequest)
		return
	}

	user := s.identity.User(r)
	if user == "" {
		http.Error(w, "missing user identity", http.StatusInternalServerError)
		return
	}

	board := Board{
		ID:       BoardID(s.newID()),
		Name:     name,
		Owner:    user,
		Messages: map[MessageID]Message{},
	}

	s.boards[board.ID] = board

	if err := json.NewEncoder(w).Encode(&board); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode JSON response: %v", err), http.StatusInternalServerError)
	}
}

func (s *MsgBoardsSvc) ListBoards(w http.ResponseWriter, r *http.Request) {
	boards := []Board{}
	for _, b := range s.boards {
		boards = append(boards, b)
	}

	if err := json.NewEncoder(w).Encode(boards); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *MsgBoardsSvc) newID() uint64 {
	s.nextID++
	return s.nextID
}

func (s *MsgBoardsSvc) PostMessage(w http.ResponseWriter, r *http.Request) {
	var msg Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, fmt.Sprintf("%v: failed to decode posted message", err), http.StatusBadRequest)
		return
	}

	board := getBoard(r.Context())
	if board == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	msg.ID = MessageID(s.newID())
	msg.Sender = r.Context().Value("user").(string)
	msg.BoardID = board.ID
	msg.CreationTime = time.Now()

	board.Messages[msg.ID] = msg

	if err := json.NewEncoder(w).Encode(msg); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode JSON response: %v", err), http.StatusInternalServerError)
	}
}

func (s *MsgBoardsSvc) ListMessages(w http.ResponseWriter, r *http.Request) {
	board := getBoard(r.Context())
	if board == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	messages := make([]Message, 0, len(board.Messages))
	for _, message := range board.Messages {
		messages = append(messages, message)
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].CreationTime.After(messages[j].CreationTime)
	})

	if err := json.NewEncoder(w).Encode(messages); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode JSON response: %v", err), http.StatusInternalServerError)
	}
}

func (s *MsgBoardsSvc) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	boardID, err := boardIDFromString(vars["boardID"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	messageID, err := messageIDFromString(vars["messageID"])
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

type boardkey struct{}

func (s *MsgBoardsSvc) boardLoader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if boardID, ok := vars["boardID"]; ok {
			board, err := s.findBoard(boardID)
			if err == nil {
				r = r.WithContext(
					context.WithValue(r.Context(), boardkey{}, board),
				)
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (s *MsgBoardsSvc) findBoard(id string) (*Board, error) {
	boardID, err := boardIDFromString(id)
	if err != nil {
		return nil, err
	}

	board, ok := s.boards[boardID]
	if !ok {
		return nil, nil
	}

	return &board, nil
}

func boardIDFromString(id string) (BoardID, error) {
	boardID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid board id '%s': %w", id, err)
	}

	return BoardID(boardID), nil
}

func getBoard(ctx context.Context) *Board {
	val := ctx.Value(boardkey{})
	if val != nil {
		return val.(*Board)
	}

	return nil
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
