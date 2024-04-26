package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"text/template"
)

// Recipe represents the JSON data structure
type Recipe struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Ingredients        []string `json:"ingredients"`
	Instructions       string   `json:"instructions"`
	PhotoURL           string   `json:"image"`
	MealType           string   `json:"dishTypes"`
	DietaryRestriction []string `json:"dietary_restriction"`
}

func main() {
	// Define a handler function for the homepage
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Define a handler function for serving static files (CSS)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Define a handler function for the recipe details page
	http.HandleFunc("/recipe-details/", detailPageHandler)

	// Define a handler function for the recipe details page
	http.HandleFunc("/api/", externalAPIHandler)

	// Start the web server
	http.ListenAndServe(":8080", nil)
}

// detailPageHandler is responsible for rendering the recipe details page using a template
func detailPageHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the recipe ID from the query parameters
	id := r.URL.Query().Get("id")
	call := r.URL.Query().Get("call")

	if call == "favorites" {
		// Make a GET request to fetch the recipe details based on the ID
		resp, err := http.Get(fmt.Sprintf("http://localhost:8081/details?id=%s", id))
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
	if call == "api" {
		apiKey := os.Getenv("SPOONACULAR_API_KEY")
		if apiKey == "" {
			http.Error(w, "Spoonacular API key not found", http.StatusInternalServerError)
			return
		}

		// Make a GET request to fetch the recipe details based on the ID
		url := fmt.Sprintf("https://api.spoonacular.com/recipes/%s/information?apiKey=%s", id, apiKey)

		// Make the GET request to the external API
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

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Failed to read response body", http.StatusInternalServerError)
			return
		}

		// Decode the JSON response into a Recipe struct
		recipe, err := ParseRecipe(body)
		if err != nil {
			http.Error(w, "Error decoding JSON", http.StatusInternalServerError)
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
}

func ParseRecipe(data []byte) (Recipe, error) {
	var recipeData struct {
		ID                  int      `json:"id"`
		Title               string   `json:"title"`
		PhotoURL            string   `json:"image"`
		DishTypes           []string `json:"dishTypes"`
		Vegetarian          bool     `json:"vegetarian"`
		Vegan               bool     `json:"vegan"`
		GlutenFree          bool     `json:"glutenFree"`
		ExtendedIngredients []struct {
			Name string `json:"name"`
		} `json:"extendedIngredients"`
		Instructions string `json:"instructions"`
	}

	err := json.Unmarshal(data, &recipeData)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return Recipe{}, err
	}

	// Initialize a Recipe struct
	recipe := Recipe{
		ID:                 strconv.Itoa(recipeData.ID),
		Title:              recipeData.Title,
		Ingredients:        make([]string, len(recipeData.ExtendedIngredients)),
		Instructions:       recipeData.Instructions,
		PhotoURL:           recipeData.PhotoURL,
		MealType:           "",
		DietaryRestriction: make([]string, 0),
	}

	// Map extended ingredients names to Ingredients
	for i, extIng := range recipeData.ExtendedIngredients {
		recipe.Ingredients[i] = extIng.Name
	}

	// Add dietary restrictions if true
	if recipeData.Vegetarian {
		recipe.DietaryRestriction = append(recipe.DietaryRestriction, "vegetarian")
	}
	if recipeData.Vegan {
		recipe.DietaryRestriction = append(recipe.DietaryRestriction, "vegan")
	}
	if recipeData.GlutenFree {
		recipe.DietaryRestriction = append(recipe.DietaryRestriction, "gluten-free")
	}

	// Combine dish types into a single string
	if len(recipeData.DishTypes) > 0 {
		recipe.MealType = recipeData.DishTypes[0] // Just take the first one for now
	}

	return recipe, nil
}

func externalAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	mealType := query.Get("type")
	dietaryRestriction := query.Get("diet")
	ingredients := query.Get("includeIngredients")

	apiKey := os.Getenv("SPOONACULAR_API_KEY")
	if apiKey == "" {
		http.Error(w, "Spoonacular API key not found", http.StatusInternalServerError)
		return
	}

	// Construct the URL for the external API request
	url := fmt.Sprintf("https://api.spoonacular.com/recipes/complexSearch?apiKey=%s&type=%s&diet=%s&includeIngredients=%s", apiKey, mealType, dietaryRestriction, ingredients)

	// Make the GET request to the external API
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "Failed to fetch data from external API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response from external API", http.StatusInternalServerError)
		return
	}

	// Set the content type header
	w.Header().Set("Content-Type", "application/json")

	// Write the response body to the client
	_, err = w.Write(body)
	if err != nil {
		http.Error(w, "Failed to write response to client", http.StatusInternalServerError)
		return
	}
}
