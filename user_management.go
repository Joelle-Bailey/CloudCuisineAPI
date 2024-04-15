package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type Recipe struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Ingredients        []string `json:"ingredients"`
	Instructions       string   `json:"instructions"`
	PhotoURL           string   `json:"photo_url"`
	MealType           string   `json:"meal_type"`
	DietaryRestriction string   `json:"dietary_restriction"`
}

func main() {
	// Define a handler function for the homepage
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Define a handler function for serving static files (CSS)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Define a handler function for the recipe details page
	http.HandleFunc("/recipe-details/", recipeDetailsHandler)

	// Start the web server
	http.ListenAndServe(":8080", nil)
}

func recipeDetailsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the URL to extract the recipe ID
	// id := r.URL.Query().Get("id")

	// Get selected values from dropdowns
	mealType := strings.TrimSpace(r.URL.Query().Get("meal_type"))
	dietaryRestrictions := r.URL.Query()["dietary_restriction"]
	ingredients := r.URL.Query()["ingredients"]

	// Construct the URL with selected values
	url := "http://localhost:8081/recipe?meal_type=" + mealType
	for _, restriction := range dietaryRestrictions {
		url += "&dietary_restriction=" + restriction
	}
	for _, ingredient := range ingredients {
		url += "&ingredients=" + ingredient
	}

	// Make a GET request to the recipe endpoint with filters and ingredients
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "Failed to fetch recipe details", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check the status code of the response
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to fetch recipe details", http.StatusInternalServerError)
		return
	}

	// Decode the JSON response into a slice of Recipe structs
	var recipes []Recipe
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&recipes)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to decode recipe details: %v", err), http.StatusInternalServerError)
		return
	}

	// Render the recipe details page using a template
	tmpl := template.Must(template.ParseFiles("recipe-details.html"))
	err = tmpl.Execute(w, recipes)
	if err != nil {
		http.Error(w, "Failed to render recipe details page", http.StatusInternalServerError)
		return
	}
}
