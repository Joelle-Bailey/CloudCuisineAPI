// Example use of Go mongo-driver
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	//"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongodbEndpoint = "mongodb://172.17.0.2:27017" // Find this from the Mongo container
)

type dollars float32

func (d dollars) String() string { return fmt.Sprintf("$%.2f", d) }

type Field struct {
	Ingredient      string             `bson:"ingredient"`
	Price     dollars            `bson:"price"`
	Category  string             `bson:"category"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

type database struct {
	collection *mongo.Collection
}

func main() {
	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongodbEndpoint))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// Select database and collection
	collection := client.Database("pantry").Collection("ingredients")
	db := database{collection: collection}

	// Initialize router
	router := http.NewServeMux()

	// Map the handlers
	router.HandleFunc("/list", db.list)
	router.HandleFunc("/price", db.price)
	router.HandleFunc("/create", db.create)
	router.HandleFunc("/read", db.read)
	router.HandleFunc("/update", db.update)
	router.HandleFunc("/delete", db.delete)

	// Start server
	log.Println("Server started on port  9000")
	log.Fatal(http.ListenAndServe(":9000", router))
}
func (db *database) list(w http.ResponseWriter, r *http.Request) {
	// Get ingredients from the collection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := db.collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	// Retrieve the ingredients from the cursor
	var ingredients []Field
	for cursor.Next(ctx) {
		var ingredient Field
		if err := cursor.Decode(&ingredient); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ingredients = append(ingredients, ingredient)
	}

	// Write the ingredients
	for _, ingredient := range ingredients {
		fmt.Fprintf(w, "%s: %s\n", ingredient.Ingredient, ingredient.Price)
	}
}

func (db *database) price(w http.ResponseWriter, req *http.Request) {
	// Get the ingredient from the query parameter
	ingredient := req.URL.Query().Get("ingredient")

	// Take ingredient from the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result Field
	err := db.collection.FindOne(ctx, bson.M{"ingredient": ingredient}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusNotFound) // 404
			fmt.Fprintf(w, "no such ingredient: %q\n", ingredient)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write ingredient price
	fmt.Fprintf(w, "%s\n", result.Price)
}

func (db *database) create(w http.ResponseWriter, req *http.Request) {
	ingredient := req.URL.Query().Get("ingredient")
	newPrice := req.URL.Query().Get("price")

	price, err := strconv.ParseFloat(newPrice, 32)
	// Handle Parsing Failure
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		fmt.Fprintf(w, "invalid price: %q\n", newPrice)
		return
	}

	// Check if ingredient already exists
	var existingIngredient Field
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.collection.FindOne(ctx, bson.M{"ingredient": ingredient}).Decode(&existingIngredient)
	if err == nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		fmt.Fprintf(w, "ingredient already exists: %s\n", ingredient)
		return
	} else if err != mongo.ErrNoDocuments {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a new ingredient document
	newIngredient := Field{
		Ingredient:      ingredient,
		Price:     dollars(price),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Insert the new ingredient document into the collection
	_, err = db.collection.InsertOne(ctx, newIngredient)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "create ingredient: %s, price: %.2f\n", ingredient, price)
}

func (db *database) read(w http.ResponseWriter, req *http.Request) {
	db.list(w, req)
}

func (db *database) update(w http.ResponseWriter, req *http.Request) {
	ingredient := req.URL.Query().Get("ingredient")
	newPrice := req.URL.Query().Get("price")

	price, err := strconv.ParseFloat(newPrice, 32)
	// Parsing Failure
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		fmt.Fprintf(w, "invalid price: %q\n", newPrice)
		return
	}

	// Check if ingredient exists
	var existingIngredient Field
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.collection.FindOne(ctx, bson.M{"ingredient": ingredient}).Decode(&existingIngredient)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusBadRequest) // 400
			fmt.Fprintf(w, "ingredient does not exist: %s\n", ingredient)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update ingredient price
	_, err = db.collection.UpdateOne(ctx,
		bson.M{"ingredient": ingredient},
		bson.M{"$set": bson.M{"price": dollars(price), "updated_at": time.Now()}})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "update ingredient: %s, price: %.2f\n", ingredient, price)
}

func (db *database) delete(w http.ResponseWriter, req *http.Request) {
	ingredient := req.URL.Query().Get("ingredient")

	// Check if ingredient exists
	var existingIngredient Field
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := db.collection.FindOne(ctx, bson.M{"ingredient": ingredient}).Decode(&existingIngredient)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusBadRequest) // 400
			fmt.Fprintf(w, "ingredient does not exist: %s\n", ingredient)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete ingredient from
	_, err = db.collection.DeleteOne(ctx, bson.M{"ingredient": ingredient})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "delete ingredient: %s\n", ingredient)
}
