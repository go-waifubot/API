package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Karitham/httperr"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-waifubot/api/db"
	"github.com/rs/zerolog/log"
)

func main() {
	p := os.Getenv("API_PORT")
	apiPort, err := strconv.Atoi(p)
	if err != nil || apiPort == 0 {
		apiPort = 3333
	}

	conf := db.Config{
		User:     os.Getenv("DB_USER"),
		Database: os.Getenv("DB_NAME"),
		Password: os.Getenv("DB_PASS"),
		Host:     os.Getenv("DB_HOST"),
	}

	log.Debug().Interface("config", conf).Msg("Running with config")

	d, err := db.Init(conf)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not connect to database")
	}
	api := &APIContext{
		db: d,
	}

	r := chi.NewRouter()

	// Timeout
	r.Use(middleware.Timeout(5 * time.Second))

	// Logger
	r.Use(loggerMiddleware(&log.Logger))

	// Set application/json as content type
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*"},
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Implement GET /user/123
	r.Route("/user", func(r chi.Router) {
		r.Route("/{userID}", func(r chi.Router) {
			r.Get("/", api.getUser)
		})
	})

	log.Info().Int("API_PORT", apiPort).Msg("API started")

	if err := http.ListenAndServe(":"+strconv.Itoa(apiPort), r); err != nil {
		log.Fatal().Err(err).Int("Port", apiPort).Msg("Listen and serve crash")
	}
}

type APIContext struct {
	db db.Querier
}

// getUser is the request handler
func (a *APIContext) getUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	id, err := strconv.Atoi(userID)
	if err != nil || id == 0 {
		fmt.Fprintf(w, "invalid ID provided: %d", id)
		return
	}

	user, err := a.db.Profile(r.Context(), int64(id))
	if err != nil {
		httperr.JSON(w, r, &httperr.DefaultError{Message: "user not found", ErrorCode: "GU0001", StatusCode: 404})
		log.Err(err).Msg("fetching user ID")
		return
	}

	if err = json.NewEncoder(w).Encode(user); err != nil {
		log.Err(err).Msg("encoding request")
	}
}
