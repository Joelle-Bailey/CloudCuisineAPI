package main

import (
	"context"
	"fmt"
	"log"
	"html/template"
	"net/http"
	"encoding/json"
	"time"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongodbEndpoint = "mongodb://172.17.0.2:27017"
)

type dollars float32

func (d dollars) String() string { return fmt.Sprintf("$%.2f", d) }

type Field struct {
	Ingredient string    `bson:"ingredient"`
	Username   string    `bson:"username"`
	CreatedAt  time.Time `bson:"created_at"`
	UpdatedAt  time.Time `bson:"updated_at"`
}

type User struct {
	Username string `bson:"username"`
	Password string `bson:"password"`
}

type database struct {
	ingredientCollection *mongo.Collection
	userCollection       *mongo.Collection
	
}
var store = sessions.NewCookieStore([]byte("secret"))
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongodbEndpoint))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	ingredientCollection := client.Database("pantry").Collection("ingredients")
	userCollection := client.Database("pantry").Collection("users")
	db := database{ingredientCollection: ingredientCollection, userCollection: userCollection}

	router := http.NewServeMux()

	router.HandleFunc("/list", db.list)
	router.HandleFunc("/create", db.create)
	router.HandleFunc("/read", db.read)
	router.HandleFunc("/update", db.update)
	router.HandleFunc("/delete", db.delete)
	router.HandleFunc("/register", db.registerPage)
	router.HandleFunc("/registerSubmit", db.registerSubmit)
	router.HandleFunc("/signIn", db.signInPage)
	router.HandleFunc("/signInSubmit", db.signInSubmit)
	router.HandleFunc("/index", db.index)
	router.HandleFunc("/pantry", db.pantry)
	router.HandleFunc("/logout", db.logout)

	// Serve static files from the "static" directory
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Server started on port 9000")
	log.Fatal(http.ListenAndServe(":9000", router))
}

func (db *database) list(w http.ResponseWriter, r *http.Request) {
    // Get the username from the session
    session, err := store.Get(r, "session-name")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    username, ok := session.Values["username"].(string)
    if !ok {
        http.Error(w, "username not found in session", http.StatusUnauthorized)
        return
    }

    ctx := r.Context()
    cursor, err := db.ingredientCollection.Find(ctx, bson.M{"username": username})
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer cursor.Close(ctx)

    var ingredients []string
    for cursor.Next(ctx) {
        var ingredient Field
        if err := cursor.Decode(&ingredient); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        ingredients = append(ingredients, ingredient.Ingredient)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(ingredients)
}
func (db *database) index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}
func (db *database) registerPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/register.html")
}
func (db *database) signInPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/signIn.html")
}


func (db *database) registerSubmit(w http.ResponseWriter, r *http.Request) {
	// Parse username and password from form data
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Check if the username already exists
	exists, err := db.userExists(r.Context(), username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "username already exists", http.StatusBadRequest)
		return
	}

	// Insert the new user document into the collection without hashing the password
	user := User{Username: username, Password: password}
	_, err = db.userCollection.InsertOne(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	

	// Redirect to the sign-in page after a short delay
	http.Redirect(w, r, "/signIn", http.StatusSeeOther)
}

func (db *database) userExists(ctx context.Context, username string) (bool, error) {
	var user User
	err := db.userCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err == nil {
		// User exists
		return true, nil
	} else if err == mongo.ErrNoDocuments {
		// User does not exist
		return false, nil
	} else {
		// Error occurred
		return false, err
	}
}
func (db *database) signInSubmit(w http.ResponseWriter, r *http.Request) {
    // Parse username and password from form data
    username := r.FormValue("username")
    password := r.FormValue("password")

    // Check if the username exists
    ctx := r.Context()
    var existingUser User
    err := db.userCollection.FindOne(ctx, bson.M{"username": username}).Decode(&existingUser)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            // User does not exist
            http.Error(w, "username or password is incorrect", http.StatusUnauthorized)
            return
        }
        // Other error occurred
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Compare the password from the database with the provided password
    if existingUser.Password != password {
        // Incorrect password
        http.Error(w, "username or password is incorrect", http.StatusUnauthorized)
        return
    }

    // Set a session cookie or JWT token to indicate successful login
    session, err := store.Get(r, "session-name")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    session.Values["username"] = username
    session.Save(r, w)

    // Redirect to the home page
    http.Redirect(w, r, "/pantry", http.StatusFound)
}


func (db *database) pantry(w http.ResponseWriter, r *http.Request) {
    // Get the username from the session
    session, err := store.Get(r, "session-name")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    username, ok := session.Values["username"].(string)
    if !ok {
        username = "Guest" // Default to "Guest" if not found
    }

    // Update the pantry.html file to include the username
    data := struct {
        Username string
    }{
        Username: username,
    }

    // Render the pantry.html file with the username
    tmpl, err := template.ParseFiles("static/pantry.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if err := tmpl.Execute(w, data); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}

func (db *database) create(w http.ResponseWriter, req *http.Request) {
    // Get the username from the session
    session, err := store.Get(req, "session-name")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    username, ok := session.Values["username"].(string)
    if !ok {
        http.Error(w, "username not found in session", http.StatusUnauthorized)
        return
    }

    var body struct {
        Ingredient string `json:"ingredient"`
    }
    if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    ingredient := body.Ingredient

    // Create a new ingredient document with the username
    newIngredient := Field{
        Ingredient: ingredient,
        CreatedAt:  time.Now(),
        UpdatedAt:  time.Now(),
        Username:   username, // Add the username to the ingredient
    }

    // Insert the new ingredient document into the collection
    ctx := req.Context()
    _, err = db.ingredientCollection.InsertOne(ctx, newIngredient)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Return the created ingredient as JSON
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(struct {
        Ingredient string `json:"ingredient"`
    }{Ingredient: ingredient})
}


  
	




func (db *database) read(w http.ResponseWriter, req *http.Request) {
	db.list(w, req)
}

func (db *database) update(w http.ResponseWriter, req *http.Request) {
	ingredient := req.URL.Query().Get("ingredient")

	// Check if ingredient exists
	var existingIngredient Field
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := db.ingredientCollection.FindOne(ctx, bson.M{"ingredient": ingredient}).Decode(&existingIngredient)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusBadRequest) // 400
			fmt.Fprintf(w, "ingredient does not exist: %s\n", ingredient)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update ingredient
	// Here you would update the ingredient's information other than price
	// For example, you could update the category or other details
	_, err = db.ingredientCollection.UpdateOne(ctx,
		bson.M{"ingredient": ingredient},
		bson.M{"$set": bson.M{"updated_at": time.Now()}})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "update ingredient: %s\n", ingredient)
}

func (db *database) delete(w http.ResponseWriter, req *http.Request) {
    var data map[string]string
    if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    ingredient := data["ingredient"]

    // Delete ingredient from the database
    _, err := db.ingredientCollection.DeleteOne(req.Context(), bson.M{"ingredient": ingredient})
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Ingredient deleted successfully"))
}

func (db *database) logout(w http.ResponseWriter, r *http.Request) {
    session, err := store.Get(r, "session-name")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Clear the username from the session
    delete(session.Values, "username")
    session.Save(r, w)

    // Redirect to the sign-in page after logging out
    http.Redirect(w, r, "/index", http.StatusSeeOther)
}