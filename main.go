package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection

func main() {
	log := zerolog.New(os.Stderr)

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://0.0.0.0:27017"
	}

	p := os.Getenv("API_PORT")
	apiPort, err := strconv.Atoi(p)
	if err != nil || apiPort == 0 {
		apiPort = 3333
	}

	ctx, fn := context.WithTimeout(nil, time.Second)
	defer fn()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal().Err(err).Msg("could not connect to database")
	}

	// waifu collection
	collection = client.Database("waifubot").Collection("waifus")

	r := chi.NewRouter()

	// Timeout
	r.Use(middleware.Timeout(5 * time.Second))

	// Logger
	r.Use(loggerMiddleware(&log))

	// Cors
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "OPTIONS"},
		MaxAge:         300,
	}))

	// Set application/json as content type
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	// Implement GET /user/123
	r.Route("/user", func(r chi.Router) {
		r.Route("/{userID}", func(r chi.Router) {
			r.Get("/", getUser)
		})
	})

	log.Info().Int("API_PORT", apiPort).Str("MONGO_URI", mongoURI).Msg("API started")

	if err := http.ListenAndServe(":"+strconv.Itoa(apiPort), r); err != nil {
		log.Fatal().Err(err).Int("Port", apiPort).Msg("Listen and serve crash")
	}
}

// getUser is the request handler
func getUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	id, err := strconv.Atoi(userID)
	if err != nil || id == 0 {
		fmt.Fprintf(w, "invalid ID provided: %d", id)
		return
	}

	data, err := getWaifus(r.Context(), id)
	if err != nil {
		fmt.Fprintf(w, "invalid user %d", id)
		log.Err(err).Msg("fetching user ID")
		return
	}

	if err = json.NewEncoder(w).Encode(data); err != nil {
		log.Err(err).Msg("encoding request")
	}
}

// getWaifus queries the database
func getWaifus(ctx context.Context, id int) (*UserDataStruct, error) {
	bytesWaifu, err := collection.FindOne(ctx, bson.M{"_id": id}).DecodeBytes()
	if err != nil {
		return nil, err
	}
	data := new(UserDataStruct)

	err = bson.Unmarshal(bytesWaifu, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// UserDataStruct is a representation of the data inside the database, it's used to retrieve data
type UserDataStruct struct {
	ID            uint         `bson:"_id" json:",omitempty"`
	Quote         string       `bson:"Quote,omitempty" json:",omitempty"`
	Favorite      CharLayout   `bson:"Favourite,omitempty" json:",omitempty"`
	ClaimedWaifus int          `bson:"ClaimedWaifus,omitempty" json:"-"`
	Date          time.Time    `bson:"Date,omitempty" json:",omitempty"`
	Waifus        []CharLayout `bson:"Waifus,omitempty" json:",omitempty"`
}

// CharLayout is how each character is stored
type CharLayout struct {
	ID    uint   `bson:"ID" json:",omitempty"`
	Name  string `bson:"Name" json:",omitempty"`
	Image string `bson:"Image" json:",omitempty"`
}
