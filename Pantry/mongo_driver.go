// Example use of Go mongo-driver
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongodbEndpoint = "mongodb://172.17.0.2:27017" // Find this from the Mongo container
)

type Post struct {
	ID        primitive.ObjectID `bson:"_id"`
	Title     string             `bson:"title"`
	Body      string             `bson:"body"`
	Tags      []string           `bson:"tags"`
	Comments  uint64             `bson:"comments"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

func main() {
	// create a mongo client
	client, err := mongo.NewClient(
		options.Client().ApplyURI(mongodbEndpoint),
	)
	checkError(err)

	// Connect to mongo
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)

	// Disconnect
	defer client.Disconnect(ctx)

	// select collection from database
	col := client.Database("blog").Collection("posts")

	// Insert one
	res, err := col.InsertOne(ctx, &Post{
		ID:        primitive.NewObjectID(),
		Title:     "post",
		Tags:      []string{"mongodb"},
		Body:      `MongoDB is a NoSQL database`,
		CreatedAt: time.Now(),
	})
	fmt.Printf("inserted id: %s\n", res.InsertedID.(primitive.ObjectID).Hex())

	// filter posts tagged as mongodb
	filter := bson.M{"tags": bson.M{"$elemMatch": bson.M{"$eq": "mongodb"}}}

	// find one document
	var p Post
	if col.FindOne(ctx, filter).Decode(&p); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("post: %+v\n", p)

}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
