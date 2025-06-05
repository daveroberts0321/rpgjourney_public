package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"strconv"
	"time"

	"github.com/daveroberts0321/rpgbackend/initializers"
	"github.com/daveroberts0321/rpgbackend/middleware"
	"github.com/daveroberts0321/rpgbackend/models"
	"github.com/gofiber/fiber/v2"
)

// Struct Literal
type QuestLevel struct {
	Strength uint
	Health   uint
	Wealth   uint
	Wisdom   uint
	Skills   uint
	strint   uint
	heaint   uint
	weaint   uint
	wisint   uint
	skiint   uint
}

type Quote struct {
	Day    int    `json:"day"`
	Quote  string `json:"quote"`
	Author string `json:"author"`
}

type QuotesWrapper struct {
	Quotes []Quote `json:"quotes"`
}

// home page
func LandingPage(c *fiber.Ctx) error {
	// check if user is logged in
	// if not, redirect to login page
	if c.Locals("userID") == nil {
		// check if jwt token is present and exp is valid
		jwt := c.Cookies("jwt")
		// decode the jwt token
		claims, err := middleware.DecodeJWT(jwt)
		if err != nil {
			return c.Redirect("/login")
		}
		// check if claim exp is valid
		if !claims.VerifyExpiresAt(0, true) {
			// redirect to login page
			return c.Redirect("/login")
		}
		return c.Redirect("/dashboard")
	}
	return c.Render("index", fiber.Map{})
}

func HomePage(c *fiber.Ctx) error {
	// render the home page
	return c.Render("index", fiber.Map{
		"Title": "Home",
	})
}

func Register(c *fiber.Ctx) error {
	// render the register page
	return c.Render("register", fiber.Map{
		"Title": "Register New User",
	})
}

func HowitStarted(c *fiber.Ctx) error {
	// render the how it started page
	return c.Render("howitstarted", fiber.Map{
		"Title": "How It Started",
	})
}

func Login(c *fiber.Ctx) error {
	// render the login page
	return c.Render("login", fiber.Map{
		"Title": "Login",
	})
}

func Logout(c *fiber.Ctx) error {
	// clear the JWT token
	c.ClearCookie("jwt")
	// redirect to login page
	return c.Redirect("/login")
}

func UserDashboard(c *fiber.Ctx) error {
	//fmt.Println("User Dashboard")
	// get the user from the JWT token
	userID := uint(c.Locals("userID").(float64))

	// get quest data from the database
	var questlog []models.QuestLog
	result := initializers.DB.Where("user_id = ?", userID).Find(&questlog)
	if result.Error != nil {
		fmt.Println(result.Error)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Fetching Quest Data",
			"Desc":    "Error fetching quest data. Please try again.",
		})
	}
	//get active quest data from the database
	var quest []models.Quest
	result = initializers.DB.Where("user_id = ? AND completed = ?", userID, false).Find(&quest)
	if result.Error != nil {
		fmt.Println(result.Error)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Fetching Active Quest Data",
			"Desc":    "Error fetching active quest data. Please try again.",
		})
	}
	//fmt.Println("Quest Data: ", quest)

	var profile models.UserProfile
	result = initializers.DB.Where("user_id = ?", userID).First(&profile)
	if result.Error != nil {
		fmt.Println(result.Error)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Fetching User Profile",
			"Desc":    "Error fetching user profile. Please try again.",
		})
	}

	// todays quote
	today := time.Now().YearDay()

	// Read the quotes.json file
	data, err := ioutil.ReadFile("quotes.json")
	if err != nil {
		fmt.Println("Error reading quotes.json:", err)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Fetching Quote",
			"Desc":    "Error fetching today's quote. Please try again.",
		})
	}

	// Parse the JSON data
	var quotesWrapper QuotesWrapper
	if err := json.Unmarshal(data, &quotesWrapper); err != nil {
		fmt.Println("Error parsing quotes.json:", err)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Parsing Quote Data",
			"Desc":    "Error parsing today's quote data. Please try again.",
		})
	}

	// Find today's quote
	var todaysQuote Quote
	for _, quote := range quotesWrapper.Quotes {
		if quote.Day == today {
			todaysQuote = quote
			break
		}
	}

	// render the user dashboard
	return c.Render("dashboard", fiber.Map{
		"Title":    "User Dashboard",
		"questlog": questlog,
		"quest":    quest,
		"profile":  profile,
		"Quote":    todaysQuote.Quote,
		"Author":   todaysQuote.Author,
	})
}

func GetQuest(c *fiber.Ctx) error {
	// get the user from the JWT token
	userID := uint(c.Locals("userID").(float64))
	if userID == 0 {
		return c.Redirect("/login")
	}
	//fmt.Println("User ID: ", userID)
	// get the quest ID from the URL
	questID := c.Params("id")
	if questID == "" {
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Fetching Quest",
			"Desc":    "Error fetching quest. Please try again.",
		})
	}
	//fmt.Println("Quest ID: ", questID)
	// get quest data from the database
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

	// compute the percentage of change from starting level as a percentage of the current level
	var compchange float32
	if quest.Ascending {
		// For ascending quests, the change is measured relative to the current level
		compchange = ((quest.Current - quest.Starting) / quest.Current) * 100
	} else {
		// For descending quests, the change is measured relative to the current level
		compchange = ((quest.Starting - quest.Current) / quest.Current) * 100
	}
	// round the percentage to 2 decimal places
	compchange = float32(math.Round(float64(compchange*100)) / 100)
	//fmt.Println("Comp Change: ", compchange)

	// check if user is the owner of the quest
	if quest.UserID != userID {
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Fetching Quest",
			"Desc":    "Error fetching quest. Please try again.",
		})
	}
	// render the quest page
	return c.Render("questpage", fiber.Map{
		"Title":      "Quest",
		"quest":      quest,
		"compchange": compchange,
	})
}

func QuestHistory(c *fiber.Ctx) error {
	questID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid quest ID")
	}

	var history []models.History
	if err := initializers.DB.Where("quest_id = ?", questID).Order("created_at").Find(&history).Error; err != nil {
		return c.Status(500).SendString(err.Error())
	}
	// print json history
	//fmt.Println(history)
	// return json response

	return c.JSON(history)
}

func StartQuest(c *fiber.Ctx) error {
	// get the user from the JWT token
	userID := uint(c.Locals("userID").(float64))
	if userID == 0 {
		return c.Redirect("/login")
	}
	// redirect to the user start quest page
	return c.Render("startquest", fiber.Map{
		"Title": "Start Quest",
	})
}

func UserDeleteQuest(c *fiber.Ctx) error {
	// get the user from the JWT token
	userID := uint(c.Locals("userID").(float64))
	if userID == 0 {
		return c.Redirect("/login")
	}

	// get the quest ID from the URL and convert to uint
	questID := c.Params("id")
	if questID == "" {
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Fetching Quest",
			"Desc":    "Error fetching quest. Please try again.",
		})
	}

	// get quest data from the database
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
	// check if user is the owner of the quest
	if quest.UserID != userID {
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Fetching Quest",
			"Desc":    "Error fetching quest. Please try again.",
		})
	}
	// render the delete quest page
	return c.Render("deletequest", fiber.Map{
		"Title": "Delete Quest",
		"quest": quest,
	})
}

func UpdateQuest(c *fiber.Ctx) error {
	// get the user from the JWT token
	userID := uint(c.Locals("userID").(float64))
	if userID == 0 {
		return c.Redirect("/login")
	}
	// get the quest ID from the URL and convert to uint
	questID := c.Params("id")
	if questID == "" {
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Fetching Quest",
			"Desc":    "Error fetching quest. Please try again.",
		})
	}
	// get quest data from the database
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
	// check if user is the owner of the quest
	if quest.UserID != userID {
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Fetching Quest",
			"Desc":    "Error fetching quest. Please try again.",
		})
	}
	// render the update quest page
	return c.Render("updatequest", fiber.Map{
		"Title": "Update Quest",
		"quest": quest,
	})
}

func BlogEntry(c *fiber.Ctx) error {
	// render the blog entry page
	return c.Render("blogentry", fiber.Map{
		"Title": "Blog Entry",
	})
}

// get last 15 blog posts
func GetBlogList(c *fiber.Ctx) error {
	// get the last 15 blog posts from the database
	var blogposts []models.BlogPost
	result := initializers.DB.Order("created_at desc").Limit(15).Find(&blogposts)
	if result.Error != nil {
		fmt.Println(result.Error)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Error Fetching Blog Posts",
			"Desc":    "Error fetching blog posts. Please try again.",
		})
	}
	// render the blog page
	return c.Render("bloglist", fiber.Map{
		"Title":     "Blog",
		"blogposts": blogposts,
	})
}
