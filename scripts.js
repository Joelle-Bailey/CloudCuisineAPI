document.addEventListener("DOMContentLoaded", function() {
    const searchButton = document.getElementById("searchButton");
    const ingredientsInput = document.getElementById("ingredients");

    searchButton.addEventListener("click", function() {
        // Reload the page
        window.location.reload();
    });
});
