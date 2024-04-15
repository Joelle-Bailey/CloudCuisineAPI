package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
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

var recipes = map[string]Recipe{
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
}

func main() {
	mux := http.NewServeMux()
	// Define a handler function for the test recipe endpoint
	http.HandleFunc("/recipe", recipeHandler)
	mux.Handle("/recipe", http.HandlerFunc(recipeHandler))
	mux.Handle("/details", http.HandlerFunc(detailHandler))

	log.Fatal(http.ListenAndServe("localhost:8081", mux))
}

func recipeHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")             // Allow requests from any origin
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS") // Allow GET and OPTIONS methods

	// Check if the request method is OPTIONS (preflight request)
	if r.Method == "OPTIONS" {
		return
	}

	// Handle GET request to /recipe
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Parse the meal type, dietary restriction, and ingredients from the query parameters
	mealType := r.URL.Query().Get("meal_type")
	dietaryRestrictions := r.URL.Query()["dietary_restriction"]
	ingredients := r.URL.Query().Get("ingredients")

	// Initialize a slice to store matching recipes
	var matchingRecipes []Recipe

	// Iterate over the recipes map
	for _, recipe := range recipes {
		// Check if the meal type matches the query or the query is empty
		if mealType == "" || strings.EqualFold(mealType, "none") || strings.EqualFold(recipe.MealType, mealType) {
			// Check if any of the recipe's dietary restrictions match any of the dietary restrictions specified in the query
			if len(dietaryRestrictions) == 0 {
				// If no dietary restrictions are specified in the query, add the recipe to the matching recipes slice
				if containsIngredients(recipe, ingredients) {
					matchingRecipes = append(matchingRecipes, recipe)
				}
			} else {
				// Iterate over the dietary restrictions specified in the query
				for _, restriction := range dietaryRestrictions {
					// Check if the recipe has the current dietary restriction
					if recipeHasDietaryRestriction(recipe, restriction) && containsIngredients(recipe, ingredients) {
						// If the recipe has the dietary restriction and contains the ingredients, add it to the matching recipes slice
						matchingRecipes = append(matchingRecipes, recipe)
						// Break out of the loop since the recipe has already been added
						break
					}
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

func detailHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the recipe ID from the query parameters
	id := r.URL.Query().Get("id")

	// Fetch the recipe with the corresponding ID from your data source
	recipe, found := recipes[id]
	if !found {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}

	// Marshal the recipe into JSON format
	recipeJSON, err := json.Marshal(recipe)
	if err != nil {
		http.Error(w, "Failed to marshal recipe JSON", http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response
	w.Write(recipeJSON)

}
