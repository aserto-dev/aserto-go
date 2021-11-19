package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/aserto-dev/aserto-go/client"
	"github.com/aserto-dev/aserto-go/client/grpc/authorizer"
	asertomw "github.com/aserto-dev/aserto-go/middleware/http"
	"github.com/gorilla/mux"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	port = ":8080"
)

func main() {
	s := newServer()
	r := mux.NewRouter()

	ctx := context.Background()
	authzClient, err := authorizer.New(
		ctx,
		client.WithAddr("localhost:8282"),
		client.WithTenantID("0fb5d7eb-8190-4f9d-ac7f-db0ba8374cb7"),
		client.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("Unable to create authorizer client: %v", err)
	}
	mw := asertomw.New(
		authzClient,
		asertomw.Policy{ID: "messageboards", Decision: "allowed"},
	).WithPolicyFromURL("messageboards").WithResourceMapper(resourceContext)
	mw.Identity.Subject().FromContextValue("user")

	r.Use(basicAuth, jsonContentType, s.boardLoader, mw.Handler)

	r.HandleFunc("/boards", s.CreateBoard).Methods(http.MethodPost)
	r.HandleFunc("/boards", s.ListBoards).Methods(http.MethodGet)
	r.HandleFunc("/boards/{boardID}", s.PostMessage).Methods(http.MethodPost)
	r.HandleFunc("/boards/{boardID}/messages", s.ListMessages).Methods(http.MethodGet)
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

func resourceContext(r *http.Request) *structpb.Struct {
	board := getBoard(r.Context())
	if board == nil {
		return nil
	}

	resource := map[string]interface{}{
		"board": structToMap(board),
	}

	if messageID, err := messageIDFromString(mux.Vars(r)["messageID"]); err != nil {
		if message, ok := board.Messages[messageID]; ok {
			resource["message"] = structToMap(message)
		}
	}

	resourceStruct, err := structpb.NewStruct(resource)
	if err != nil {
		return nil
	}

	return resourceStruct
}

func structToMap(data interface{}) map[string]interface{} {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	mapData := make(map[string]interface{})
	err = json.Unmarshal(dataBytes, &mapData)
	if err != nil {
		return nil
	}
	return mapData
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

	if err := json.NewEncoder(w).Encode(boards); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

	json.NewEncoder(w).Encode(msg)
}

func (s *server) ListMessages(w http.ResponseWriter, r *http.Request) {
	board := getBoard(r.Context())
	if board == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	messages := make([]BoardMessage, 0, len(board.Messages))
	for _, message := range board.Messages {
		messages = append(messages, message)
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].CreationTime.After(messages[j].CreationTime)
	})

	json.NewEncoder(w).Encode(messages)
}

func (s *server) DeleteMessage(w http.ResponseWriter, r *http.Request) {
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

func (s *server) boardLoader(next http.Handler) http.Handler {
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

func getBoard(ctx context.Context) *MessageBoard {
	val := ctx.Value(boardkey{})
	if val != nil {
		return val.(*MessageBoard)
	}

	return nil
}

func (s *server) findBoard(id string) (*MessageBoard, error) {
	boardID, err := boardIDFromString(id)
	if err != nil {
		return nil, err
	}

	board, ok := s.boards[BoardID(boardID)]
	if !ok {
		return nil, nil
	}

	return &board, nil
}

func basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if ok {
			if !checkPassword(user, password) {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), "user", user))
		}

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
