package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongodbEndpoint = "mongodb://172.17.0.2:27017" // Find this from the Mongo container
)

type dollars float32

func (d dollars) String() string { return fmt.Sprintf("$%.2f", d) }

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type Ingredient struct {
	Name   string  `bson:"name" json:"name"`
	Amount float64 `bson:"amount" json:"amount"`
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
			return fmt.Errorf("operation failed")
		}
	}

	client, err := mongo.NewClient(
		options.Client().ApplyURI(mongodbEndpoint),
	)
	checkError(err)

	// Connect to MongoDB
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Second)
	err = client.Connect(ctx)

	err = retry(ctx, 4, time.Second, operation)
	if err != nil {
		fmt.Printf("Operation failed after retries: %v\n", err)
	} else {
		fmt.Println("Operation succeeded")
	}

	// Select collection from database
	col := client.Database("inventory").Collection("items")

	return &database{
		data:    col,
		connect: ctx,
		client:  client,
	}
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request)

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f(w, r)
}

func (db database) list(w http.ResponseWriter, r *http.Request) {
	filter := bson.M{"item": bson.M{"$exists": true}}

	// Find all documents
	ctx, cancel := context.WithTimeout(db.connect, 10*time.Second)
	defer cancel()

	cursor, err := db.data.Find(ctx, filter)
	if err != nil {
		http.Error(w, "Error querying database", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var ingredients []Ingredient
	if err := cursor.All(ctx, &ingredients); err != nil {
		http.Error(w, "Error retrieving data", http.StatusInternalServerError)
		return
	}

	for _, ing := range ingredients {
		fmt.Fprintf(w, "%s: %.2f\n", ing.Name, ing.Amount)
	}
}

func (db database) create(w http.ResponseWriter, r *http.Request) {

	name := r.URL.Query().Get("name")
	amountStr := r.URL.Query().Get("amount")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "could not convert amount: %q\n", amountStr)
		return
	}

	filter := bson.M{"name": name}
	found := db.data.FindOne(context.Background(), filter)

	if found.Err() == nil {
		if found.Err() != mongo.ErrNoDocuments {
			// Ingredient already exists
			w.WriteHeader(http.StatusConflict)
			fmt.Fprintf(w, "ingredient already in inventory: %q\n", name)
			return
		}
		// Other error occurred
		// Respond with 500 Internal Server Error
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error checking ingredient: %v\n", found.Err())
		return
	}

	res, err := db.data.InsertOne(db.connect, &Ingredient{
		Name:   name,
		Amount: amount,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error creating ingredient: %s\n", err.Error())
		return
	}

	fmt.Printf("inserted id: %s\n", res.InsertedID.(primitive.ObjectID).Hex())

	fmt.Fprintf(w, "Created ingredient: %s with amount %.2f\n", name, amount)

}

func (db database) update(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	amountStr := r.URL.Query().Get("amount")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "could not convert amount: %q\n", amountStr)
		return
	}

	// Create a filter to find the ingredient document
	filter := bson.M{"name": name}

	// Create an update to set the amount field
	update := bson.M{"$set": bson.M{"amount": amount}}

	// Perform the update operation
	_, err = db.data.UpdateOne(context.Background(), filter, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error updating ingredient: %v\n", err)
		return
	}

	// Send a success response
	fmt.Fprintf(w, "amount of ingredient %q updated to %.2f\n", name, amount)
}

func (db database) remove(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")

	// Create a filter to find the ingredient document
	filter := bson.M{"name": name}

	// Perform the delete operation
	result, err := db.data.DeleteOne(context.Background(), filter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error deleting ingredient: %v\n", err)
		return
	}

	// Check if the ingredient was found and deleted
	if result.DeletedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "no such ingredient: %q\n", name)
		return
	}

	// Send a success response
	fmt.Fprintf(w, "Deleted ingredient: %s\n", name)
}

func main() {

	db := newDatabase()

	mux := http.NewServeMux()
	mux.Handle("/list", http.HandlerFunc(db.list))
	mux.Handle("/create", http.HandlerFunc(db.create))
	mux.Handle("/update", http.HandlerFunc(db.update))
	mux.Handle("/remove", http.HandlerFunc(db.remove))
	log.Fatal(http.ListenAndServe("localhost:8003", mux))

	//defer db.client.Disconnect(db.connect
}
