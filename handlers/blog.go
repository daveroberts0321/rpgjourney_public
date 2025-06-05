package handlers

import (
	"html/template"

	"github.com/daveroberts0321/rpgbackend/initializers"
	"github.com/daveroberts0321/rpgbackend/models"
	"github.com/gofiber/fiber/v2"
)

// BlogPostEntry creates a new blog post entry
func BlogPostEntry(c *fiber.Ctx) error {
	// authentication
	if c.Locals("userID") == nil {
		return c.Render("login", fiber.Map{
			"Title": "Login",
		})
	}
	// Parse the JSON body
	var body struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		UserID  uint   `json:"userID"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	// Get the user ID from the context
	userID := c.Locals("userID").(float64) // Assume it's a float64
	userIDUint := uint(userID)

	// Get the user profile from the database
	var user models.UserProfile
	result := initializers.DB.Where("id = ?", userIDUint).First(&user)
	if result.Error != nil {
		return c.SendString(result.Error.Error())
	}

	// Create a new blog post
	blogPost := models.BlogPost{
		Title:      body.Title,
		Content:    template.HTML(body.Content),
		UserID:     user.ID,
		ProfileImg: user.ProfileImg,
		//BlogImg:    user.BlogImg,
		Username: user.Username,
	}

	// Save the blog post to the database
	result = initializers.DB.Create(&blogPost)
	if result.Error != nil {
		return c.SendString(result.Error.Error())
	}

	// Redirect to the blog entry page or return a success message
	return c.JSON(blogPost)
}
