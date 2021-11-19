package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	port = ":8080"
)

func main() {
	identity := &BasicAuth{}
	mw, err := NewAuthorizaionMiddleware(identity)
	if err != nil {
		log.Fatal(err)
	}

	svc := newMsgBoardsSvc(identity)
	r := mux.NewRouter()

	r.Use(identity.Middleware, jsonContentType, svc.boardLoader, mw.Handler)

	r.HandleFunc("/boards", svc.CreateBoard).Methods(http.MethodPost)
	r.HandleFunc("/boards", svc.ListBoards).Methods(http.MethodGet)
	r.HandleFunc("/boards/{boardID}", svc.PostMessage).Methods(http.MethodPost)
	r.HandleFunc("/boards/{boardID}/messages", svc.ListMessages).Methods(http.MethodGet)
	r.HandleFunc("/boards/{boardID}/messages/{messageID}", svc.DeleteMessage).Methods(http.MethodDelete)

	// Start server
	fmt.Printf("Starting server on 'localhost%s'\n", port)
	log.Fatal(http.ListenAndServe(port, r))
}

func jsonContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
