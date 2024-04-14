package main

import (
	"encoding/json"
	"net/http"
)

type Recipe struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Ingredients  []string `json:"ingredients"`
	Instructions string   `json:"instructions"`
	PhotoURL     string   `json:"photo_url"`
}

func main() {
	// Define a handler function for the test recipe endpoint
	http.HandleFunc("/test-recipe", testRecipeHandler)

	// Start the web server
	http.ListenAndServe(":8000", nil)
}

func testRecipeHandler(w http.ResponseWriter, r *http.Request) {
	// Create a test recipe
	recipe := Recipe{
		ID:           "1",
		Title:        "Test Recipe",
		Ingredients:  []string{"Ingredient 1", "Ingredient 2", "Ingredient 3"},
		Instructions: "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed aliquam velit eu urna maximus, ac blandit arcu consequat.",
		PhotoURL:     "https://example.com/test-recipe.jpg",
	}

	// Marshal the recipe struct into JSON format
	recipeJSON, err := json.Marshal(recipe)
	if err != nil {
		http.Error(w, "Failed to marshal recipe JSON", http.StatusInternalServerError)
		return
	}

	// Set the content type header
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response
	w.Write(recipeJSON)
}
