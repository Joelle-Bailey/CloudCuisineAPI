package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/rs/cors"
)

// Recipe represents the JSON data structure
type Recipe struct {
	ID                 int      `json:"id"`
	Title              string   `json:"title"`
	Ingredients        []string `json:"ingredients"`
	Instructions       string   `json:"instructions"`
	PhotoURL           string   `json:"image"`
	MealType           []string `json:"dishTypes"`
	DietaryRestriction []string `json:"dietary_restriction"`
}

/* var recipes = map[string]Recipe{
	"1": Recipe{
		ID:                 "1",
		Title:              "Pizza - Test Recipe",
		Ingredients:        []string{"Pizza Dough", "Tomato Sauce", "Mozzarella Cheese", "Pepperoni"},
		Instructions:       "1. Preheat oven to 475°F (245°C).\n2. Roll out the dough on a lightly floured surface.\n3. Spread tomato sauce over the dough.\n4. Sprinkle mozzarella cheese over the sauce.\n5. Add desired toppings like pepperoni.\n6. Bake in preheated oven for 10-15 minutes or until crust is golden brown.",
		PhotoURL:           "https://media.istockphoto.com/id/521403691/photo/hot-homemade-pepperoni-pizza.jpg?s=612x612&w=0&k=20&c=PaISuuHcJWTEVoDKNnxaHy7L2BTUkyYZ06hYgzXmTbo=",
		MealType:           "Dinner",
		DietaryRestriction: []string{"None"},
	},
	"2": Recipe{
		ID:                 "2",
		Title:              "Blueberry Muffins - Test Recipe",
		Ingredients:        []string{"2 cups all-purpose flour", "1/2 cup granulated sugar", "1 tablespoon baking powder", "1/2 teaspoon salt", "1/2 cup unsalted butter, melted", "2 large eggs", "1 cup milk", "1 1/2 cups fresh blueberries"},
		Instructions:       "1. Preheat oven to 375°F (190°C). Grease muffin cups or line with muffin liners.\n2. In a large bowl, combine flour, sugar, baking powder, and salt.\n3. In another bowl, mix together melted butter, eggs, and milk.\n4. Pour the wet ingredients into the dry ingredients and stir until just combined.\n5. Gently fold in the blueberries.\n6. Spoon batter into prepared muffin cups.\n7. Bake in preheated oven for 20 to 25 minutes or until a toothpick inserted into the center comes out clean.\n8. Allow muffins to cool in the pan for 5 minutes before transferring to a wire rack to cool completely.",
		PhotoURL:           "https://www.culinaryhill.com/wp-content/uploads/2022/08/Blueberry-Muffins-Culinary-Hill-1200x800-1.jpg",
		MealType:           "Breakfast",
		DietaryRestriction: []string{"Vegetarian"},
	},
	"3": Recipe{
		ID:                 "3",
		Title:              "Test Recipe 3",
		Ingredients:        []string{"Ingredient A", "Ingredient B", "Ingredient C"},
		Instructions:       "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed non risus. Suspendisse lectus tortor, dignissim sit amet, adipiscing nec, ultricies sed, dolor. Cras elementum ultrices diam. Maecenas ligula massa, varius a, semper congue, euismod non, mi. Proin porttitor, orci nec nonummy molestie, enim est eleifend mi, non fermentum diam nisl sit amet erat. Duis semper. Duis arcu massa, scelerisque vitae, consequat in, pretium a, enim. Pellentesque congue. Ut in risus volutpat libero pharetra tempor. Cras vestibulum bibendum augue. Praesent egestas leo in pede. Praesent blandit odio eu enim. Pellentesque sed dui ut augue blandit sodales. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Aliquam nibh. Mauris ac mauris sed pede pellentesque faucibus. Ut accumsan, velit sit amet aliquam dapibus, libero leo dictum quam, sed tincidunt augue enim eget libero. Suspendisse vitae tortor. Nullam eleifend quam a libero. Integer vitae arcu at urna vehicula consequat. Morbi ipsum ipsum, porta nec, tempor id, vehicula vitae, purus.",
		PhotoURL:           "https://example.com/test-recipe-1.jpg",
		MealType:           "",
		DietaryRestriction: []string{},
	},
} */

func main() {
	mux := http.NewServeMux()

	// Define a handler function for the test recipe endpoint
	http.HandleFunc("/recipe", recipeHandler)
	mux.Handle("/recipe", http.HandlerFunc(recipeHandler))

	// Create a CORS handler
	c := cors.AllowAll()

	// Wrap your ServeMux with the CORS handler
	handler := c.Handler(mux)

	// Start the server
	log.Fatal(http.ListenAndServe("localhost:8081", handler))
}

func recipeHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")             // Allow requests from any origin
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS") // Allow GET and OPTIONS methods
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type") // Include Content-Type header

	// Check if the request method is OPTIONS (preflight request)
	if r.Method == "OPTIONS" {
		return
	}

	// Handle GET request to /recipe
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Fetch recipes from localhost:8003/list
	resp, err := http.Get("http://localhost:8003/list")
	if err != nil {
		http.Error(w, "Failed to fetch recipe data", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Parse the JSON response into a slice of Recipe structs
	var recipes []Recipe

	// Create a new JSON decoder
	decoder := json.NewDecoder(resp.Body)

	// Decode the JSON array into the recipes slice
	if err := decoder.Decode(&recipes); err != nil {
		// If an error occurs, log it
		fmt.Println("Error decoding JSON:", err)
	}

	// Store recipes in a map
	recipeMap := make(map[int]Recipe)
	for _, recipe := range recipes {
		recipeMap[recipe.ID] = recipe
	}

	// Parse the meal types, dietary restrictions, and ingredients from the query parameters
	mealTypes := r.URL.Query()["meal_type"] // Updated to retrieve multiple meal types
	dietaryRestrictions := r.URL.Query()["dietary_restriction"]
	ingredients := r.URL.Query().Get("ingredients")

	// Initialize a slice to store matching recipes
	var matchingRecipes []Recipe

	// Iterate over the recipes map
	for _, recipe := range recipeMap {
		// Check if the recipe matches any of the specified meal types or if no meal types are specified
		if len(mealTypes) == 0 || containsMealType(recipe, mealTypes) {
			// Check if any of the recipe's dietary restrictions match any of the specified dietary restrictions or if no restrictions are specified
			dietaryRestrictions := strings.Join(dietaryRestrictions, " ")
			if len(dietaryRestrictions) == 0 || recipeHasDietaryRestriction(recipe, dietaryRestrictions) {
				// Check if the recipe contains the specified ingredients or if no ingredients are specified
				if containsIngredients(recipe, ingredients) {
					// Add the recipe to the matching recipes slice
					matchingRecipes = append(matchingRecipes, recipe)
				}
			}
		}
	}

	// Check if any matching recipes were found
	if len(matchingRecipes) == 0 {
		http.Error(w, "No recipes found matching the search criteria", http.StatusNotFound)
		return
	}

	// Marshal the matching recipes slice into JSON format
	recipesJSON, err := json.Marshal(matchingRecipes)
	if err != nil {
		http.Error(w, "Failed to marshal recipes JSON", http.StatusInternalServerError)
		return
	}

	// Set the content type header
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response
	w.Write(recipesJSON)
}

func containsMealType(recipe Recipe, mealTypes []string) bool {
	if len(mealTypes) == 0 {
		return true // If no meal types are specified, all recipes are considered to contain the meal types
	}

	recipeMealTypes := strings.Join(recipe.MealType, " ")
	// Convert the recipe's meal types to lowercase for case-insensitive comparison
	recipeMealTypesLower := strings.ToLower(recipeMealTypes)

	// Iterate over the specified meal types
	for _, mealType := range mealTypes {
		// Check if the current meal type is found in the recipe's meal types
		if strings.Contains(recipeMealTypesLower, strings.ToLower(mealType)) {
			return true
		}
	}
	return false
}
func recipeHasDietaryRestriction(recipe Recipe, restriction string) bool {
	// If the restriction is "None" or blank, consider it as no restriction
	if strings.TrimSpace(strings.ToLower(restriction)) == "none" || restriction == "" {
		return true
	}

	// Iterate over the recipe's dietary restrictions
	for _, r := range recipe.DietaryRestriction {
		// Check if the current dietary restriction matches the specified restriction (case-insensitive)
		if strings.EqualFold(strings.TrimSpace(r), strings.TrimSpace(restriction)) {
			return true
		}
	}
	return false
}

func containsIngredients(recipe Recipe, ingredients string) bool {
	if ingredients == "" {
		return true // If no ingredients are specified, all recipes are considered to contain the ingredients
	}

	// Split the required ingredients string into individual words
	requiredIngredients := strings.Fields(ingredients)

	// Convert both the recipe's ingredients and the required ingredients to lowercase for case-insensitive comparison
	recipeIngredientsLower := strings.ToLower(strings.Join(recipe.Ingredients, " "))

	// Check if all the required ingredients are found in the text of the recipe's ingredients
	for _, ingredient := range requiredIngredients {
		if !strings.Contains(recipeIngredientsLower, strings.ToLower(ingredient)) {
			return false
		}
	}
	return true
}
