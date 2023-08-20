package api

import (
	"log"
	"time"

	"github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"

	"rinha-backend-go/persistence"

	"github.com/gofiber/fiber/v2"
)

type Server struct {
	Port     string `json:"port"`
	fiberApp *fiber.App
	store    persistence.Store

	cache *redis.Client
}

func (s *Server) Stop() error {
	return s.fiberApp.Shutdown()
}

func (s *Server) Start() error {

	s.fiberApp = fiber.New(
		fiber.Config{
			JSONEncoder:  json.Marshal,
			JSONDecoder:  json.Unmarshal,
			ReadTimeout:  30 * time.Millisecond,
			WriteTimeout: 15 * time.Millisecond,
			IdleTimeout:  10 * time.Millisecond,
		},
	)

	handler := PeopleHandler{store: s.store, cache: s.cache}

	s.fiberApp.Get("contagem-pessoas", handler.GetPeopleCount)
	s.fiberApp.Post("/pessoas", handler.AddPerson)
	s.fiberApp.Get("/pessoas", handler.GetPeople)
	s.fiberApp.Get("/pessoas/:id", handler.GetPerson)

	log.Println("Server listening on port", s.Port)

	return s.fiberApp.Listen(s.Port)
}

func New(store persistence.Store, port string, redisAddress string) *Server {
	return &Server{
		Port:  ":" + port,
		store: store,
		cache: redis.NewClient(&redis.Options{
			Addr: redisAddress,
		}),
	}
}
