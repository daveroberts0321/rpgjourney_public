// handlers/auth.go
package handlers

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/daveroberts0321/rpgbackend/initializers"
	"github.com/daveroberts0321/rpgbackend/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(os.Getenv("SECRET_KEY"))

// auth helper functions
func PasswordHasher(pw string) (string, error) {
	// Get the password from the user and hash it
	password := pw

	// Generate the hash for the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func PasswordCompare(hashedPassword string, password string) bool {
	// Compare the password with the hash
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

// app functions

// database test
func UserCheck(c *fiber.Ctx) error {
	// Check if the user exists
	var user models.User
	initializers.DB.First(&user)
	return c.JSON(user)
}

func RegisterUser(c *fiber.Ctx) error {
	// Get the email and password from the request body
	firstname := c.FormValue("first_name")
	lastname := c.FormValue("last_name")
	email := c.FormValue("email")
	password := c.FormValue("password")
	username := c.FormValue("username")

	if firstname == "" || email == "" || password == "" {
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Please fill in all the fields.",
			"Desc":    "Please fill in all the fields to register a new account.",
		})
	}

	//email to lowercase
	email = strings.ToLower(email)

	// Call NewUserController to create a new user
	hashedPassword, err := PasswordHasher(password)
	if err != nil {
		fmt.Println(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Check if user exists
	var thisuser models.User
	initializers.DB.Where("email = ?", email).First(&thisuser)
	if thisuser.Email == email {
		fmt.Println("User already exists")
		return c.SendString("User already exists")
	}

	// Create a new user
	user := models.User{
		Email:    email,
		Password: hashedPassword,
		Active:   true,
		Role:     "user", // default role
	}

	// Save the user
	result := initializers.DB.Create(&user)
	if result.Error != nil {
		fmt.Println(result.Error)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// create UserProfile for user
	userProfile := models.UserProfile{
		UserID:    user.ID,
		Username:  username,
		FirstName: firstname,
		LastName:  lastname,
		Email:     email,
		Level:     0,
	}

	// Save the user profile
	result = initializers.DB.Create(&userProfile)
	if result.Error != nil {
		fmt.Println(result.Error)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	fmt.Println("User created: ", user.Email)
	// Redirect to login page
	return c.Redirect("/login")

}

func LoginUser(c *fiber.Ctx) error {
	// Get the email and password from the request body
	email := c.FormValue("email")
	password := c.FormValue("password")

	if email == "" || password == "" {
		fmt.Println("Empty email or password")
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Incorrect Credentials",
			"Desc":    "Incorrect Login Credentials. Please try again.",
		})
	}

	email = strings.ToLower(email)

	var user models.User
	userresult := initializers.DB.Where("email = ?", email).First(&user)
	if userresult.Error != nil {
		fmt.Println("User DB Lookup Error: ", userresult.Error)
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Incorrect Credentials",
			"Desc":    "Incorrect Login Credentials. Please try again.",
		})
	}

	if !PasswordCompare(user.Password, password) {
		fmt.Println("Incorrect password")
		return c.Render("error", fiber.Map{
			"Title":   "RPG/ Error",
			"Message": "Incorrect Credentials",
			"Desc":    "Incorrect Login Credentials. Please try again.",
		})
	}

	// Generate JWT token https://github.com/golang-jwt/jwt
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": user.Email,
		"exp":   time.Now().Add(time.Hour * 12).Unix(),
		"ID":    user.ID,
		"role":  user.Role,
	})
	secretKey := []byte(os.Getenv("SECRET_KEY"))
	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(secretKey) // Uses the secret from the .env file
	if err != nil {
		fmt.Println("Generate Token string Login error: ", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Set cookie
	cookie := new(fiber.Cookie)
	if os.Getenv("PRODUCTIONENV") == "production" {
		//fmt.Println("Setting cookie properties for production")
		cookie.HTTPOnly = true
		cookie.Secure = true
	} else {
		//fmt.Println("Setting cookie properties for development")
		cookie.HTTPOnly = false
		cookie.Secure = false
	}

	// Set cookie properties
	cookie.Name = "jwt"
	cookie.Value = tokenString
	cookie.Expires = time.Now().Add(time.Hour * 72) // Expiration time of 72 hours
	c.Cookie(cookie)

	if user.Active {
		//fmt.Println("redirecting to Dashboard")
		return c.Redirect("/dashboard")
	}

	return c.Redirect("/")
}

// Logout Handler
func LogoutHandler(c *fiber.Ctx) error {
	//clear cookie
	cookie := new(fiber.Cookie)
	cookie.Name = "jwt"
	cookie.Value = ""
	cookie.Expires = time.Now().Add(-time.Hour * 12) //expiration time of 12 hours
	cookie.HTTPOnly = false                          //change in production //cookie is not accessible by JavaScript
	cookie.Secure = false                            //change in production //cookie will only be sent with an HTTPS request
	c.Cookie(cookie)

	//redirect to login page
	return c.Redirect("/")
}
