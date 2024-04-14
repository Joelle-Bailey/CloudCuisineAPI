package main

import (
	"encoding/json"
	"html/template"
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
	//id := r.URL.Path[len("/recipe-details/"):]

	// Make a GET request to the test recipe endpoint
	resp, err := http.Get("http://localhost:8080/test-recipe")
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

	// Decode the JSON response into a Recipe struct
	var recipe Recipe
	err = json.NewDecoder(resp.Body).Decode(&recipe)
	if err != nil {
		http.Error(w, "Failed to decode recipe details", http.StatusInternalServerError)
		return
	}

	// Render the recipe details page using a template
	tmpl := template.Must(template.ParseFiles("recipe-details.html"))
	err = tmpl.Execute(w, recipe)
	if err != nil {
		http.Error(w, "Failed to render recipe details page", http.StatusInternalServerError)
		return
	}
}
