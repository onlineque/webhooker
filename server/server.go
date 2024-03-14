package server

import (
	"context"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
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
	Tokens   *mongo.Collection
}

type Message struct {
	Message   string    `bson:"message"`
	Timestamp time.Time `bson:"timestamp"`
}

type Token struct {
	Token string `bson:"token"`
}

func (srv *Server) Listen(address string, certFile string, keyFile string, dbUri string) error {
	srv.DbObj = &database.DbConnection{}
	srv.DbObj.NewDbConnection(dbUri)
	err := srv.DbObj.Connect()
	if err != nil {
		return err
	}
	srv.Messages = srv.DbObj.GetCollection("messages")
	srv.Tokens = srv.DbObj.GetCollection("tokens")

	http.HandleFunc("/webhook", srv.handleWebhook)
	http.HandleFunc("/wall", srv.handleWall)
	log.Printf("webhook server listening on %s", address)
	return http.ListenAndServeTLS(address, certFile, keyFile, nil)
}

func (srv *Server) handleWall(w http.ResponseWriter, _ *http.Request) {
	cursor, err := srv.Messages.Find(context.TODO(), bson.D{})
	if err != nil {
		msg := fmt.Sprintf("error querying MongoDB database: %s", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	for cursor.Next(context.TODO()) {
		// var elem Message
		//var elem bson.M
		var elem = &Message{}
		err := cursor.Decode(&elem)
		if err != nil {
			msg := "cannot decode query result"
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		_, err = w.Write([]byte(fmt.Sprintf("Time: %s\nMessage: %s\n\n", elem.Timestamp, elem.Message)))
		if err != nil {
			msg := fmt.Sprintf("error publishing the message: %s", err)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
	}

	http.Error(w, "", http.StatusOK)
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
			msg := "Request body contains badly-formed JSON"
			http.Error(w, msg, http.StatusBadRequest)

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)",
				unmarshalTypeError.Field, unmarshalTypeError.Offset)
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

	isTokenValid := srv.checkTokenValidity(wr.Token)
	if !isTokenValid {
		msg := "user token is invalid !"
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

	insertResult, err := srv.Messages.InsertOne(context.TODO(), bson.D{
		bson.E{Key: "timestamp", Value: time.Now()},
		bson.E{Key: "message", Value: wr.Message},
	})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Printf("token: %s, channel: %s, message: %s", wr.Token, wr.Channel, wr.Message)
	log.Printf("inserted a single doc: %v", insertResult.InsertedID)
	http.Error(w, "message created", http.StatusCreated)
}

func (srv *Server) checkTokenValidity(token string) bool {
	shaObj := sha512.New()
	shaObj.Write([]byte(token))
	userTokenHash := fmt.Sprintf("%x", shaObj.Sum(nil))
	// search for the hashed token
	filter := bson.D{{Key: "token", Value: userTokenHash}}
	tokenHash := &Token{}
	err := srv.Tokens.FindOne(context.TODO(), filter).Decode(tokenHash)
	if err != nil {
		log.Printf("error decoding token: %s", err)
		return false
	}
	log.Printf("token found in db")
	return true
}
