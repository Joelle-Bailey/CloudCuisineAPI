from flask import Flask, render_template, request, jsonify
import requests
import json
import os
app = Flask(__name__)


api_key = '0ceb43eec36c4d9fa329e4a6c39c9f2a'

''
def fetch_recipes(ingredients):
    """Fetch recipes from Spoonacular API based on ingredients."""
    url = f"https://api.spoonacular.com/recipes/findByIngredients?ingredients={ingredients}&apiKey={api_key}&number=10"
    response = requests.get(url)
    if response.status_code == 200:
        return response.json()  # This returns a list of recipe dictionaries
    else:
        print("Failed to fetch recipes, status code:", response.status_code)
        return []


@app.route('/', methods=['GET', 'POST'])
def index():
    """Render the main page where users input their ingredients."""
    return render_template('index.html')

@app.route('/results', methods=['POST'])
def results():
    ingredients = request.form.get('ingredients')
    print("Ingredients received:", ingredients)  # Debug print
    recipes = fetch_recipes(ingredients)
    print("Recipes fetched:", recipes)  # Debug print
    return render_template('results.html', recipes=recipes)

@app.route('/recipebook')
def recipebook():
    return render_template('recipebook.html')


if __name__ == '__main__':
    app.run(debug=True)

