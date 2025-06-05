package main

import (
	"fmt"
	"html/template"
	"os"

	"github.com/daveroberts0321/rpgbackend/controllers"
	"github.com/daveroberts0321/rpgbackend/handlers"
	"github.com/daveroberts0321/rpgbackend/initializers"
	"github.com/daveroberts0321/rpgbackend/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func init() {
	fmt.Println("initializing rpgbackend server...")
	initializers.LoadEnvVars() // Load the environment variables
	initializers.Connect()     // Connect to the database
	initializers.SyncDB()      // Sync the database
}

func main() {
	// Initialize standard Go html template engine
	// Setup the HTML template engine
	safeHTMLFunc := func(s string) template.HTML {
		return template.HTML(s)
	}
	engine := html.New("./templates", ".html") //default template engine and cache
	engine.AddFunc("safeHTML", safeHTMLFunc)

	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "layouts/main",
	})

	// Static files (CSS, JS, images, etc.)
	app.Static("/static", "./static")

	// Routes
	app.Get("/", controllers.LandingPage)
	app.Get("/home", controllers.HomePage)
	app.Get("/blog", controllers.GetBlogList)
	app.Get("/register", controllers.Register)
	app.Post("/register", handlers.RegisterUser)
	app.Get("/howitstarted", controllers.HowitStarted)
	app.Get("/login", controllers.Login)
	app.Post("/login", handlers.LoginUser)
	app.Get("/logout", controllers.Logout)
	//app.Get("/usercheck", handlers.UserCheck)
	app.Get("/dashboard", middleware.RequireAuth(), controllers.UserDashboard)
	app.Get("/startquest", middleware.RequireAuth(), controllers.StartQuest)
	app.Post("/startquest", middleware.RequireAuth(), handlers.UserStartQuest)
	app.Get("/getquest/:id", middleware.RequireAuth(), controllers.GetQuest)
	app.Post("/updatequest/:id", middleware.RequireAuth(), handlers.UpdateQuestProgress)
	app.Get("/deletequest/:id", middleware.RequireAuth(), controllers.UserDeleteQuest)             // render the deletequest page
	app.Post("/deletequest/:id", middleware.RequireAuth(), handlers.DeleteQuest)                   // delete the quest from the database
	app.Get("/updatequest/:id", middleware.RequireAuth(), controllers.UpdateQuest)                 // render the updatequest page
	app.Post("/updatequestvariables/:id", middleware.RequireAuth(), handlers.UpdateQuestVariables) // update the quest in the database
	app.Get("/blogentry", middleware.RequireAuth(), controllers.BlogEntry)
	app.Post("/savepost", middleware.RequireAuth(), handlers.BlogPostEntry)

	app.Get("/quest/:id/history", middleware.RequireAuth(), controllers.QuestHistory)

	//listen
	fmt.Println("listening on Port...")
	error := app.Listen(":" + os.Getenv("PORT"))
	if error != nil {
		fmt.Println("Error starting server:", error)
		os.Exit(1)
	}
}
