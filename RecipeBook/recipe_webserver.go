package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	mongodbEndpoint = "mongodb://172.17.0.2:27017" // Find this from the Mongo container
)

// Recipe struct represents a recipe document
type Recipe struct {
	ID                 int      `bson:"_id" json:"id"`
	Title              string   `bson:"title" json:"title"`
	Ingredients        []string `bson:"ingredients" json:"ingredients"`
	Instructions       string   `bson:"instructions" json:"instructions"`
	PhotoURL           string   `bson:"photoURL" json:"image"`
	MealType           []string `bson:"mealType" json:"dishTypes"`
	DietaryRestriction []string `bson:"dietaryRestriction" json:"dietary_restriction"`
}

// database struct represents the database connection and collection
type database struct {
	data    *mongo.Collection
	connect context.Context
	client  *mongo.Client
}

// retry function retries an operation with exponential backoff
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

// newDatabase function initializes a new database connection
func newDatabase() *database {

	client, err := mongo.NewClient(
		options.Client().ApplyURI(mongodbEndpoint),
	)
	checkError(err)

	// Connect to MongoDB
	ctx, _ := context.WithTimeout(context.Background(), 1000*time.Second)
	err = client.Connect(ctx)
	checkError(err)

	// Ping MongoDB to verify connection
	err = client.Ping(ctx, readpref.Primary())
	checkError(err)

	// Retry database operation
	operation := func() error {
		// Simulate an operation that may fail
		if time.Now().Second()%2 == 0 {
			return nil // Success
		}
		return fmt.Errorf("operation failed")
	}

	err = retry(ctx, 4, time.Second, operation)
	if err != nil {
		fmt.Printf("Operation failed after retries: %v\n", err)
	} else {
		fmt.Println("Operation succeeded")
	}

	// Select collection from database
	col := client.Database("inventory").Collection("recipes")

	return &database{
		data:    col,
		connect: ctx,
		client:  client,
	}
}

// list handler function retrieves and returns all recipes
func (db database) list(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

	if r.Method == "OPTIONS" {
		// Respond to preflight requests
		w.WriteHeader(http.StatusOK)
		return
	}

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

	// Marshal recipes into JSON
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

// create handler function creates a new recipe
func (db database) create(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type") // Include Content-Type header

	if r.Method == "OPTIONS" {
		// Respond to preflight requests
		w.WriteHeader(http.StatusOK)
		return
	}

	var recipe Recipe
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&recipe)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "error decoding request body: %v\n", err)
		return
	}

	_, err = db.data.InsertOne(db.connect, recipe)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error creating recipe: %s\n", err.Error())
		return
	}

	json.NewEncoder(w).Encode(recipe)
}

// remove handler function removes a recipe
func (db database) remove(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")

	if r.Method == "OPTIONS" {
		// Respond to preflight requests
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse recipe ID from the query parameter
	title := r.URL.Query().Get("title")
	if title == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "missing recipe title parameter")
		return
	}

	// Delete the recipe from the database
	filter := bson.M{"title": title}
	res, err := db.data.DeleteOne(db.connect, filter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error removing recipe: %s\n", err.Error())
		return
	}

	// Check if a recipe was deleted
	if res.DeletedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "recipe not found with ID: %s\n", title)
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "recipe removed successfully")
}

func main() {
	// Initialize database connection
	db := newDatabase()

	// Create a new HTTP multiplexer
	mux := http.NewServeMux()

	// Handle list endpoint
	mux.HandleFunc("/list", db.list)

	// Handle create endpoint
	mux.HandleFunc("/create", db.create)

	// Handle remove endpoint
	mux.HandleFunc("/remove", db.remove)

	// Start HTTP server
	log.Fatal(http.ListenAndServe(":8003", mux))

	// Disconnect from MongoDB
	defer db.client.Disconnect(db.connect)
}

// checkError function checks for errors and logs them
func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
