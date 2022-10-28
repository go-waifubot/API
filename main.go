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
	"github.com/go-chi/stampede"
	"github.com/go-waifubot/api/db"
	"github.com/rs/zerolog/log"
)

func main() {
	p := os.Getenv("PORT")
	apiPort, err := strconv.Atoi(p)
	if err != nil || apiPort == 0 {
		apiPort = 3333
	}

	url := os.Getenv("DB_URL")
	d, err := db.Init(url)
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
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		MaxAge:           300, // Maximum value not ignored by any of major browsers
		AllowCredentials: true,
	}))

	// Implement GET /user/123
	r.Route("/user", func(r chi.Router) {
		r.Use(stampede.Handler(512, 5*time.Second))
		r.Get("/find", api.findUser)
		r.Get("/{userID}", api.getUser)
	})
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "Hello user, you shouldn't be there, direct yourself to https://github.com/go-waifubot/api for docs")
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
		herr := &httperr.DefaultError{
			Message:    "invalid id provided",
			ErrorCode:  "GU0002",
			StatusCode: 400,
		}
		httperr.JSON(w, r, herr)
		log.Debug().Err(herr).Msg("invalid ID")
		return
	}

	user, err := a.db.Profile(r.Context(), int64(id))
	if err != nil || user.ID == 0 {
		herr := &httperr.DefaultError{
			Message:    "user not found",
			ErrorCode:  "GU0001",
			StatusCode: 404,
		}
		httperr.JSON(w, r, herr)
		log.Debug().Err(herr).Msg("fetching user ID")
		return
	}

	if err = json.NewEncoder(w).Encode(user); err != nil {
		log.Err(err).Msg("encoding request")
	}
}

func (a *APIContext) findUser(w http.ResponseWriter, r *http.Request) {
	anilist := r.URL.Query().Get("anilist")
	if anilist == "" {
		herr := &httperr.DefaultError{
			Message:    "anilist query param is required",
			ErrorCode:  "FU0001",
			StatusCode: 400,
		}

		httperr.JSON(w, r, herr)
		log.Debug().Err(herr).Msg("fetching user ID")
		return
	}

	user, err := a.db.UserByAnilistURL(r.Context(), fmt.Sprintf("https://anilist.co/user/%s", anilist))
	if err != nil || user.UserID == 0 {
		herr := &httperr.DefaultError{
			Message:    "user not found",
			ErrorCode:  "FU0002",
			StatusCode: 404,
		}
		httperr.JSON(w, r, herr)
		log.Debug().Err(herr).Msg("fetching user ID")
		return
	}

	type resp struct {
		ID int64 `json:"id,string"`
	}

	if err = json.NewEncoder(w).Encode(resp{
		ID: user.UserID,
	}); err != nil {
		log.Err(err).Msg("encoding request")
	}
}
