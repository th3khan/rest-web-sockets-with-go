package server

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/th3khan/rest-web-sockets-with-go/database"
	"github.com/th3khan/rest-web-sockets-with-go/repositories"
	"github.com/th3khan/rest-web-sockets-with-go/websocket"
)

type Config struct {
	Port        string
	JWTSecret   string
	DataBaseUrl string
}

type Server interface {
	Config() *Config
	Hub() *websocket.Hub
}

type Broker struct {
	config *Config
	router *mux.Router
	hub    *websocket.Hub
}

func (b *Broker) Config() *Config {
	return b.config
}

func NewServer(ctx context.Context, config *Config) (*Broker, error) {
	if config.Port == "" {
		return nil, errors.New("Port is required")
	}
	if config.JWTSecret == "" {
		return nil, errors.New("Secret Key is Required")
	}
	if config.DataBaseUrl == "" {
		return nil, errors.New("Database url is required")
	}
	broker := &Broker{
		config: config,
		router: mux.NewRouter(),
		hub:    websocket.NewHub(),
	}
	return broker, nil
}

func (b *Broker) Start(binder func(s Server, r *mux.Router)) {
	b.router = mux.NewRouter()
	binder(b, b.router)

	handler := cors.Default().Handler(b.router)

	repo, err := database.NewMySQLRepository(b.config.DataBaseUrl)
	if err != nil {
		log.Fatal("Error", err)
	}
	go b.hub.Run()
	repositories.SetRepository(repo)

	log.Println("Starting server on Port", b.config.Port)

	if err := http.ListenAndServe(b.config.Port, handler); err != nil {
		log.Fatal("ListenAndServer: ", err)
	}
}

func (b *Broker) Hub() *websocket.Hub {
	return b.hub
}
