package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"log"
	"net/http"
	"strings"
	"webhooker/database"
)

type WebhookerRequest struct {
	Token   string `json:"token" validate:"required"`
	Channel string `json:"channel" validate:"required"`
	Message string `json:"message" validate:"required"`
}

type Server struct {
	DbObj    *database.DbConnection
	Messages *mongo.Collection
}

type Message struct {
	msg string
}

func (srv *Server) Listen(address string, certFile string, keyFile string, dbUri string) error {
	srv.DbObj = &database.DbConnection{}
	srv.DbObj.NewDbConnection(dbUri)
	err := srv.DbObj.Connect()
	if err != nil {
		return err
	}
	srv.Messages = srv.DbObj.GetCollection("webhooker", "messages")

	http.HandleFunc("/webhook", srv.handleWebhook)
	log.Printf("webhook server listening on %s", address)
	return http.ListenAndServeTLS(address, certFile, keyFile, nil)
}

func (srv *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
	if mediaType != "application/json" {
		msg := "Content-Type header set to application/json is required"
		http.Error(w, msg, http.StatusUnsupportedMediaType)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var wr WebhookerRequest
	err := dec.Decode(&wr)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formatted JSON (at position %d)", syntaxError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			http.Error(w, msg, http.StatusBadRequest)

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			http.Error(w, msg, http.StatusBadRequest)

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			http.Error(w, msg, http.StatusBadRequest)

		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			http.Error(w, msg, http.StatusRequestEntityTooLarge)
		default:
			log.Print(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	validate := validator.New()
	err = validate.Struct(wr)
	if err != nil {
		log.Print(err)
		msg := fmt.Sprintf("failed to validate struct - %s", err)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	msg := Message{wr.Message}
	insertResult, err := srv.Messages.InsertOne(context.TODO(), msg)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Printf("token: %s, channel: %s, message: %s", wr.Token, wr.Channel, wr.Message)
	log.Printf("inserted a single doc: %v", insertResult.InsertedID)
	http.Error(w, "everything's all right", http.StatusOK)
}
