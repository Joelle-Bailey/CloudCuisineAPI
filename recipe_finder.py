import requests

def get_user_preferences():
    ingredients = input("Enter the ingredients you have separated by a comma (e.g., apples, flour, sugar): ")
    dietary_restrictions = input("Enter any dietary restrictions (e.g., vegetarian, vegan, gluten-free): ")
    meal_type = input("What type of meal are you looking for? (e.g., breakfast, lunch, dinner): ")
    return ingredients, dietary_restrictions, meal_type

def fetch_recipes(api_key, ingredients, dietary_restrictions, meal_type):
    # Base URL for the API request
    url = f"https://api.spoonacular.com/recipes/complexSearch?apiKey={api_key}&includeIngredients={ingredients}&type={meal_type}&addRecipeInformation=true&number=5"
    
    # Append dietary restrictions only if they are specified
    if dietary_restrictions.lower() not in ["none", ""]:
        url += f"&diet={dietary_restrictions}"

    try:
        print("Fetching recipes...")  # Debug: Notify user that the API call is being made
        print("URL: ", url)  # Debug: Print the URL to help trace what is being sent to the API
        response = requests.get(url)
        response.raise_for_status()  # This will raise an exception for HTTP error codes
        recipes = response.json().get('results', [])
        
        if not recipes:
            print("No recipes found. Try different ingredients or settings.")
            return

        print("\nRecipes you can make:")
        for recipe in recipes:
            print(f" - {recipe['title']} (ID: {recipe['id']})")
    except requests.exceptions.RequestException as e:
        print(f"An error occurred: {e}")

def main():
    api_key = '0ceb43eec36c4d9fa329e4a6c39c9f2a'  # Replace with your actual Spoonacular API key
    while True:
        ingredients, dietary_restrictions, meal_type = get_user_preferences()
        fetch_recipes(api_key, ingredients, dietary_restrictions, meal_type)
        
        user_choice = input("\nWould you like to search again or modify your choices? (yes/no): ")
        if user_choice.lower() != 'yes':
            print("Thank you for using the recipe finder!")
            break

if __name__ == "__main__":
    main()
