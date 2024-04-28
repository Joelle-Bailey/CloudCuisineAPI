package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	mongodbEndpoint = "mongodb://172.17.0.3:27017" // Find this from the Mongo container
)

type dollars float32

func (d dollars) String() string { return fmt.Sprintf("$%.2f", d) }

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type Recipe struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title              string             `bson:"title" json:"title"`
	Ingredients        []string           `bson:"ingredients" json:"ingredients"`
	Instructions       string             `bson:"instructions" json:"instructions"`
	PhotoURL           string             `bson:"photoURL" json:"image"`
	MealType           string             `bson:"mealType" json:"dishTypes"`
	DietaryRestriction []string           `bson:"dietaryRestriction" json:"dietary_restriction"`
}

type database struct {
	data    *mongo.Collection
	connect context.Context
	client  *mongo.Client
}

func retry(ctx context.Context, maxAttempts int, interval time.Duration, operation func() error) error {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := operation()
		if err == nil {
			// Operation succeeded, no need to retry
			return nil
		}

		// Check if the context is done (cancelled or timed out)
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// If this is the last attempt, return the error without retrying
		if attempt == maxAttempts {
			return err
		}

		// Sleep for the specified interval before retrying
		select {
		case <-time.After(interval):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func newDatabase() *database {

	operation := func() error {
		// Simulate an operation that may fail
		if time.Now().Second()%2 == 0 {
			return nil // Success
		} else {
			return errors.New("operation failed")
		}
	}

	client, err := mongo.NewClient(
		options.Client().ApplyURI(mongodbEndpoint),
	)
	checkError(err)

	// Connect to mongo
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	err = retry(ctx, 4, time.Second, operation)

	if err != nil {
		fmt.Printf("Operation failed after retries: %v\n", err)
	} else {
		fmt.Println("Operation succeeded")
	}

	// select collection from database
	col := client.Database("inventory").Collection("recipes")

	return &database{
		data:    col,
		connect: ctx,
		client:  client,
	}
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request)

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) { // adapter function
	f(w, r)
}

func (db database) list(w http.ResponseWriter, r *http.Request) {
	cursor, err := db.data.Find(db.connect, bson.M{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error fetching recipes: %v\n", err)
		return
	}
	defer cursor.Close(db.connect)

	var recipes []Recipe
	if err := cursor.All(db.connect, &recipes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error decoding recipes: %v\n", err)
		return
	}

	// Marshal the recipes slice into JSON
	responseData, err := json.Marshal(recipes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error marshaling response: %v\n", err)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Write JSON response
	w.WriteHeader(http.StatusOK)
	w.Write(responseData)
}

func (db database) create(w http.ResponseWriter, r *http.Request) {
	var recipe Recipe
	err := json.NewDecoder(r.Body).Decode(&recipe)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "error decoding request body: %v\n", err)
		return
	}

	res, err := db.data.InsertOne(db.connect, recipe)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error creating recipe: %s\n", err.Error())
		return
	}

	fmt.Printf("inserted id: %s\n", res.InsertedID.(primitive.ObjectID).Hex())

	json.NewEncoder(w).Encode(recipe)
}

func main() {
	db := newDatabase()

	mux := http.NewServeMux()
	mux.Handle("/list", http.HandlerFunc(db.list))
	mux.Handle("/create", http.HandlerFunc(db.create))
	log.Fatal(http.ListenAndServe("localhost:8003", mux))

	defer db.client.Disconnect(db.connect)
}
