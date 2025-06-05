package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/daveroberts0321/rpgbackend/initializers"
	"github.com/daveroberts0321/rpgbackend/models"
	"github.com/gofiber/fiber/v2"
)

// Function to parse float with error handling and empty string check
func parseFloatWithErrorHandling(valueStr string, fieldName string) (float32, error) {
	if valueStr == "" {
		return 0, fmt.Errorf("empty value for %s", fieldName)
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		fmt.Printf("Error parsing %s: %v\n", fieldName, err)
		return 0, fmt.Errorf("error parsing %s", fieldName)
	}
	return float32(value), nil
}

// Function to calculate the new level based on completed quests
func calculateLevel(completed uint) uint {
	level := uint(0)
	for {
		var required uint
		switch {
		case level < 10:
			required = (level + 1) * 5
		case level < 20:
			required = 50 + (level-10+1)*10
		case level < 30:
			required = 150 + (level-20+1)*15
		default:
			required = 300 + (level-30+1)*20
		}

		if completed >= required {
			level++
		} else {
			break
		}
	}
	return level
}

// start a new quest
func UserStartQuest(c *fiber.Ctx) error {
	// get userid from the JWT token
	userID := uint(c.Locals("userID").(float64))

	// Get the quest ID from the request body
	category := c.FormValue("category")
	title := c.FormValue("title")
	description := c.FormValue("description")
	unit := c.FormValue("unit")            // unit of measurement
	ascending := c.FormValue("ascending")  // goal up or down
	startingStr := c.FormValue("starting") // current value
	starting, err := strconv.ParseFloat(startingStr, 64)
	if err != nil {
		fmt.Println(err)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Starting Quest",
			"Desc":    "Error starting quest. Please try again.",
		})
	}
	// convert starting to float32
	starting32 := float32(starting)
	current := float32(starting)

	// convert current to float32
	var ascendbool bool
	var target float32 = 0
	if ascending == "ascending" {
		ascendbool = true
		//target value = current value + 1%
		target = current + (current * 0.01)
	} else {
		ascendbool = false
		target = current - (current * 0.01)
	}

	// Create a new quest log
	quest := models.Quest{
		UserID:      userID,
		Category:    category,
		Title:       title,
		Description: description,
		Unit:        unit,
		Goalmet:     0,
		Ascending:   ascendbool,
		Starting:    starting32,
		Target:      target,
		Current:     current,
		Ending:      0,
		Completed:   false,
	}

	// Save the quest log to the database
	result := initializers.DB.Create(&quest)
	if result.Error != nil {
		fmt.Println(result.Error)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Starting Quest",
			"Desc":    "Error starting quest. Please try again.",
		})
	}

	// Redirect to the user dashboard
	return c.Redirect("/dashboard")
}

func UpdateQuestProgress(c *fiber.Ctx) error {
	// Get the user from the JWT token
	userID := uint(c.Locals("userID").(float64))

	// Get the user profile from the database
	var profile models.UserProfile
	result := initializers.DB.Where("id = ?", userID).First(&profile)
	if result.Error != nil {
		fmt.Println(result.Error)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Updating Quest",
			"Desc":    "Error updating quest. Please try again.",
		})
	}

	// Get the quest ID from the URL parameter
	questIDStr := c.Params("id")
	questID, err := strconv.Atoi(questIDStr)
	if err != nil {
		fmt.Println(err)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Updating Quest",
			"Desc":    "Error updating quest. Please try again.",
		})
	}

	currentStr := c.FormValue("progress") // Progress value
	current, err := strconv.ParseFloat(currentStr, 64)
	if err != nil {
		fmt.Println(err)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Updating Quest",
			"Desc":    "Error updating quest. Please try again.",
		})
	}
	// Convert current to float32
	current32 := float32(current)

	// Get the quest from the database
	var quest models.Quest
	result = initializers.DB.Where("id = ? AND user_id = ?", questID, userID).First(&quest)
	if result.Error != nil {
		fmt.Println(result.Error)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Updating Quest",
			"Desc":    "Error updating quest. Please try again.",
		})
	}

	// check if quest.UpdatedAt is within the last 24 hours
	timenow := time.Now()
	if timenow.Sub(quest.UpdatedAt).Hours() < 24 {
		// Quest has been updated in the last 24 hours
		return c.Render("error", fiber.Map{
			"Title":   "RPG Quest/ Error",
			"Message": "Quest may only be updated once every 24 hours",
			"Desc":    "Quest may only be updated once every 24 hours. Please try again later.",
		})
	}

	// Get or initialize the quest log for the user
	var questlog models.QuestLog
	result = initializers.DB.Where("user_id = ?", userID).First(&questlog)
	if result.Error != nil {
		// Create a new quest log
		questlog = models.QuestLog{
			UserID:   userID,
			Strength: 0,
			Health:   0,
			Wealth:   0,
			Wisdom:   0,
			Skills:   0,
			Notes:    "",
		}
		// Save the new quest log
		result = initializers.DB.Create(&questlog)
		if result.Error != nil {
			fmt.Println(result.Error)
			return c.Render("error", fiber.Map{
				"Title":   "RPG/ Error",
				"Message": "Error Updating Quest",
				"Desc":    "Error updating quest. Please try again.",
			})
		}
	}

	// Create a new history entry
	history := models.History{
		QuestID: uint(questID),
		UserID:  userID,
		Amount:  current32,
		Unit:    quest.Unit,
	}

	// Update the quest
	quest.Current = current32
	// Check if the quest is completed
	if quest.Ascending {
		if quest.Current >= quest.Target {
			switch quest.Category { // Update the quest log based on the category
			case "Strength":
				questlog.Strength++
			case "Health":
				questlog.Health++
			case "Wealth":
				questlog.Wealth++
			case "Wisdom":
				questlog.Wisdom++
			case "Skills":
				questlog.Skills++
			}
			// Update quest with new target value if target is hit
			quest.Target = quest.Current + (quest.Current * 0.01)
			quest.Goalmet++
		}
	} else { // Descending goal quest
		if quest.Current <= quest.Target {
			switch quest.Category { // Update the quest log based on the category
			case "Strength":
				questlog.Strength++
			case "Health":
				questlog.Health++
			case "Wealth":
				questlog.Wealth++
			case "Wisdom":
				questlog.Wisdom++
			case "Skills":
				questlog.Skills++
			}
			// Update quest with new target value if target is hit
			quest.Target = quest.Current - (quest.Current * 0.01)
			quest.Goalmet++
		}
	}

	// Save the updated quest
	result = initializers.DB.Save(&quest)
	if result.Error != nil {
		fmt.Println(result.Error)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Updating Quest",
			"Desc":    "Error updating quest. Please try again.",
		})
	}

	// Save the updated quest log
	result = initializers.DB.Save(&questlog)
	if result.Error != nil {
		fmt.Println(result.Error)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Updating Quest",
			"Desc":    "Error updating quest. Please try again.",
		})
	}

	// Save the new history entry
	result = initializers.DB.Create(&history)
	if result.Error != nil {
		fmt.Println(result.Error)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Updating Quest History",
			"Desc":    "Error updating quest. Please try again.",
		})
	}

	// Use quest log to update user level based on quest completion
	// Get level from user model
	level := profile.Level
	// Add up all aspects of the quest log
	completed := questlog.Strength + questlog.Health + questlog.Wealth + questlog.Wisdom + questlog.Skills
	// Determine the new level based on completed quests
	newLevel := calculateLevel(completed)

	// Update user profile if the level has changed
	if newLevel != level {
		profile.Level = newLevel
		result = initializers.DB.Save(&profile)
		if result.Error != nil {
			fmt.Println(result.Error)
			return c.Render("error", fiber.Map{
				"Title":   "RPG/ Error",
				"Message": "Error Updating Level",
				"Desc":    "Error updating level. Please try again.",
			})
		}
	}

	// Redirect to the user dashboard
	return c.Redirect("/dashboard")
}

func UpdateQuestVariables(c *fiber.Ctx) error {
	//fmt.Println("Update Quest Variables")
	// Get the user from the JWT token
	userID := uint(c.Locals("userID").(float64))
	if userID == 0 {
		return c.Redirect("/login")
	}
	// Get the quest ID from the URL and convert to uint
	questID := c.Params("id")
	if questID == "" {
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Fetching Quest",
			"Desc":    "Error fetching quest. Please try again.",
		})
	}
	// Get quest data from the database
	var quest models.Quest
	result := initializers.DB.Where("id = ? AND user_id = ?", questID, userID).First(&quest)
	if result.Error != nil {
		fmt.Println(result.Error)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Fetching Quest",
			"Desc":    "Error fetching quest. Please try again.",
		})
	}

	// Get variables from the quest update form
	category := c.FormValue("category")
	title := c.FormValue("title")
	description := c.FormValue("description")
	unit := c.FormValue("unit")
	ascending := c.FormValue("ascending")

	// Parse starting value
	startingStr := c.FormValue("starting")
	starting32, err := parseFloatWithErrorHandling(startingStr, "starting")
	if err != nil {
		fmt.Println("Error parsing starting value:", err)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Updating Quest Starting Value",
			"Desc":    "Error updating quest. Please try again.",
		})
	}

	// Parse current value
	currentStr := c.FormValue("current")
	current32, err := parseFloatWithErrorHandling(currentStr, "current")
	if err != nil {
		fmt.Println("Error parsing current value:", err)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Updating Quest Current Value",
			"Desc":    "Error updating quest. Please try again.",
		})
	}

	var ascendbool bool
	if ascending == "true" {
		ascendbool = true
	} else {
		ascendbool = false
	}

	// Update the quest if form value is different from the quest
	if quest.Category != category {
		quest.Category = category
	}
	if quest.Title != title {
		quest.Title = title
	}
	if quest.Description != description {
		quest.Description = description
	}
	if quest.Unit != unit {
		quest.Unit = unit
	}
	if quest.Ascending != ascendbool {
		quest.Ascending = ascendbool
	}
	if quest.Starting != starting32 {
		quest.Starting = starting32
	}
	if quest.Current != current32 {
		quest.Current = current32
	}

	// Save the updated quest
	result = initializers.DB.Save(&quest)
	if result.Error != nil {
		fmt.Println(result.Error)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Updating Quest",
			"Desc":    "Error updating quest. Please try again.",
		})
	}
	// Redirect to the quest page
	return c.Render("updatequest", fiber.Map{
		"Title": "Update Quest",
		"quest": quest,
	})
}

// delete a quest
func DeleteQuest(c *fiber.Ctx) error {
	// get the user from the JWT token
	userID := uint(c.Locals("userID").(float64))

	// Get the quest ID from the url parameter
	questIDStr := c.Params("id")
	questID, err := strconv.Atoi(questIDStr)
	if err != nil {
		fmt.Println(err)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Deleting Quest",
			"Desc":    "Error deleting quest. Please try again.",
		})
	}

	// get the quest from the database
	var quest models.Quest
	result := initializers.DB.Where("id = ? AND user_id = ?", questID, userID).First(&quest)
	if result.Error != nil {
		fmt.Println(result.Error)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Deleting Quest",
			"Desc":    "Error deleting quest. Please try again.",
		})
	}

	// delete the quest from the database
	result = initializers.DB.Delete(&quest)
	if result.Error != nil {
		fmt.Println(result.Error)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Deleting Quest",
			"Desc":    "Error deleting quest. Please try again.",
		})
	}

	// delete the history entries for the quest
	result = initializers.DB.Where("quest_id = ?", questID).Delete(&models.History{})
	if result.Error != nil {
		fmt.Println(result.Error)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Deleting Quest",
			"Desc":    "Error deleting quest. Please try again.",
		})
	}

	// Redirect to the user dashboard
	return c.Redirect("/dashboard")
}
