package main

import (
	"errors"
	"go-projects/database"
	_ "go-projects/docs"
	"go-projects/file"
	"strconv"
	"strings"
	"unicode"

	"go-projects/token"
	"path/filepath"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"time"
	"go-projects/date"
	"fmt"
)

type User struct {
	ID           bson.ObjectID `bson:"_id,omitempty"`
	FirstName    string        `bson:"firstName" form:"firstName"`
	LastName     string        `bson:"lastName" form:"lastName"`
	NationalCode string        `bson:"nationalCode" form:"nationalCode"`
	ProfileImage string        `bson:"profileImage" json:"profileImage"`
	Username     string        `bson:"username" form:"username"`
	Password     string        `bson:"password" form:"password"`
	BirthDate time.Time `bson:"birthDate" json:"birthDate"`}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateUserRequest struct {
	FirstName    string `form:"firstName"`
	LastName     string `form:"lastName"`
	NationalCode string `form:"nationalCode"`
	BirthDate    string `form:"birthDate"`
	Username     string `form:"username"`
	Password     string `form:"password"`
}

func isValidNationalCode(code string) bool {
	if len(code) != 10 {
		return false
	}

	for _, char := range code {
		if !unicode.IsDigit(char) {
			return false
		}
	}

	weights := []int{10, 9, 8, 7, 6, 5, 4, 3, 2}
	sum := 0

	for i := 0; i < 9; i++ {
		digit, _ := strconv.Atoi(string(code[i]))
		sum += digit * weights[i]
	}

	remainder := sum % 11

	checkDigit := 0

	if remainder < 2 {
		checkDigit = remainder
	} else {
		checkDigit = 11 - remainder
	}

	lastDigit, _ := strconv.Atoi(string(code[9]))

	return checkDigit == lastDigit
}

// GetUser godoc
// @Summary Get current user
// @Description Get information of the currently authenticated user
// @Success 200 {object} User
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /users [get]
func GetUser(c *gin.Context, users *mongo.Collection) {

	userID, exists := c.Get("userID")

	if !exists {
		c.JSON(401, gin.H{
			"error": "User ID not found",
		})
		return
	}

	objectID, err := bson.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	var user User

	err = database.FindOne(
		c.Request.Context(),
		users,
		bson.M{"_id": objectID},
		&user,
	)

	if err != nil {
		c.JSON(404, gin.H{
			"error": "User not found",
		})
		return
	}

	if user.ProfileImage != "" {
		user.ProfileImage = "http://localhost:8080/uploads/profile/" + user.ProfileImage
	}
	jy, jm, jd, err := date.GregorianToJalali(user.BirthDate)

	if err != nil {
		c.JSON(500, gin.H{
			"error": "Failed to convert birth date",
		})
		return
	}

	birthDate := fmt.Sprintf("%04d/%02d/%02d", jy, jm, jd)

	c.JSON(200, gin.H{
	"id":           user.ID,
	"firstName":    user.FirstName,
	"lastName":     user.LastName,
	"nationalCode": user.NationalCode,
	"profileImage": user.ProfileImage,
	"username":     user.Username,
	"birthDate":    birthDate,
})
}

// CreateUser godoc
// @Summary Create a user
// @Description Create a new user with profile image
// @Accept multipart/form-data
// @Produce json
// @Param firstName formData string true "First name"
// @Param lastName formData string true "Last name"
// @Param nationalCode formData string true "National code"
// @Param birthDate formData string true "Birth date (Jalali)"
// @Param username formData string true "Username"
// @Param password formData string true "Password"
// @Param profileImage formData file false "Profile image"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users [post]
func CreateUser(c *gin.Context, users *mongo.Collection) {
	var req CreateUserRequest

	err := c.ShouldBind(&req)

	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid request body",
		})
		return
	}
	parts := strings.Split(req.BirthDate, "/")

	if len(parts) != 3 {
		c.JSON(400, gin.H{
			"error": "Invalid birth date format",
		})
		return
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid birth year",
		})
		return
	}

	month, err := strconv.Atoi(parts[1])
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid birth month",
		})
		return
	}

	day, err := strconv.Atoi(parts[2])
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid birth day",
		})
		return
	}

	birthDate, err := date.JalaliToGregorian(year, month, day)

	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid birth date",
		})
		return
	}

	user := User{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		NationalCode: req.NationalCode,
		Username:     req.Username,
		Password:     req.Password,
		BirthDate:    birthDate,
	}

	if user.FirstName == "" || user.LastName == "" {
		c.JSON(400, gin.H{
			"error": "First name and last name are required",
		})
		return
	}
	if user.Username == "" || user.Password == "" {
		c.JSON(400, gin.H{
			"error": "Username and password are required",
		})
		return
	}

	if !isValidName(user.FirstName) || !isValidName(user.LastName) {
		c.JSON(400, gin.H{
			"error": "First name and last name must contain only letters",
		})
		return
	}
	if !isValidNationalCode(user.NationalCode) {
		c.JSON(400, gin.H{
			"error": "Invalid national code",
		})
		return
	}

	var existingUser User

	err = database.FindOne(
		c.Request.Context(),
		users,
		bson.M{"nationalCode": user.NationalCode},
		&existingUser,
	)

	if err == nil {
		c.JSON(400, gin.H{
			"error": "National code already exists",
		})
		return
	}

	if !errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(500, gin.H{
			"error": "Database error",
		})
		return
	}

	fileName, err := file.UploadProfileImage(c)

	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	user.ProfileImage = fileName

	insertedID, err := database.InsertOne(
		c.Request.Context(),
		users,
		user,
	)

	if err != nil {
		c.JSON(500, gin.H{
			"error": "Failed to create user",
		})
		return
	}
	c.JSON(201, gin.H{
		"message": "User created successfully",
		"id":      insertedID,
		"user":    user,
	})
}

// UpdateUser godoc
// @Summary Update a user
// @Description Update user information by ID with optional profile image
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param firstName formData string true "First name"
// @Param lastName formData string true "Last name"
// @Param nationalCode formData string true "National code"
// @Param birthDate formData string true "Birth date (Jalali)"
// @Param profileImage formData file false "Profile image"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users [put]
func UpdateUser(c *gin.Context, users *mongo.Collection) {
	userID, exists := c.Get("userID")

	if !exists {
		c.JSON(401, gin.H{
			"error": "User ID not found",
		})
		return
	}

	objectID, err := bson.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	var existingUser User

	err = database.FindOne(
		c.Request.Context(),
		users,
		bson.M{"_id": objectID},
		&existingUser,
	)

	if errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(400, gin.H{
			"error": "User not found",
		})
		return
	}

	if err != nil {
		c.JSON(500, gin.H{
			"error": "Database error",
		})
		return
	}


	var req CreateUserRequest

	err = c.ShouldBind(&req)


	parts := strings.Split(req.BirthDate, "/")

	if len(parts) != 3 {
		c.JSON(400, gin.H{
			"error": "Invalid birth date format",
		})
		return
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid birth year",
		})
		return
	}

	month, err := strconv.Atoi(parts[1])
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid birth month",
		})
		return
	}

	day, err := strconv.Atoi(parts[2])
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid birth day",
		})
		return
	}

	birthDate, err := date.JalaliToGregorian(year, month, day)
	

	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid birth date",
		})
		return
	}	
	err = c.ShouldBind(&req)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	if req.FirstName == "" || req.LastName == "" {
		c.JSON(400, gin.H{
			"error": "First name and last name are required",
		})
		return
	}

	if !isValidName(req.FirstName) || !isValidName(req.LastName) {
		c.JSON(400, gin.H{
			"error": "First name and last name must contain only letters",
		})
		return
	}

	if !isValidNationalCode(req.NationalCode) {
		c.JSON(400, gin.H{
			"error": "Invalid national code",
		})
		return
	}

	var duplicateUser User

	err = database.FindOne(
		c.Request.Context(),
		users,
		bson.M{
			"nationalCode": req.NationalCode,
			"_id": bson.M{
				"$ne": objectID,
			},
		},
		&duplicateUser,
	)

	if err == nil {
		c.JSON(400, gin.H{
			"error": "National code already exists",
		})
		return
	}

	if !errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(500, gin.H{
			"error": "Database error",
		})
		return
	}

	updateFields := bson.M{
		"firstName":    req.FirstName,
		"lastName":     req.LastName,
		"nationalCode": req.NationalCode,
		"birthDate":    birthDate,
	}

	fileName, err := file.UploadProfileImage(c)

	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	if fileName != "" {
		if existingUser.ProfileImage != "" {
			filePath := filepath.Join(file.UploadPath, existingUser.ProfileImage)
			err = file.DeleteFile(filePath)
			if err != nil {
				c.JSON(500, gin.H{
					"error": "Failed to delete old profile image",
				})
				return
			}
		}

		updateFields["profileImage"] = fileName
		existingUser.ProfileImage = fileName
	}

	err = database.UpdateOne(
		c.Request.Context(),
		users,
		bson.M{"_id": objectID},
		bson.M{
			"$set": updateFields,
		},
	)

	if err != nil {
		c.JSON(500, gin.H{
			"error": "Failed to update user",
		})
		return
	}

	existingUser.FirstName = req.FirstName
	existingUser.LastName = req.LastName
	existingUser.NationalCode = req.NationalCode

	if existingUser.ProfileImage != "" {
		existingUser.ProfileImage = "http://localhost:8080/uploads/profile/" + existingUser.ProfileImage
	}

	jy, jm, jd, err := date.GregorianToJalali(existingUser.BirthDate)

	if err != nil {
		c.JSON(500, gin.H{
			"error": "Failed to convert birth date",
		})
		return
	}

	birthDateString := fmt.Sprintf("%04d/%02d/%02d", jy, jm, jd)	

	c.JSON(200, gin.H{
		"message": "User updated successfully",
		"user": gin.H{
			"id":           existingUser.ID,
			"firstName":    existingUser.FirstName,
			"lastName":     existingUser.LastName,
			"nationalCode": existingUser.NationalCode,
			"profileImage": existingUser.ProfileImage,
			"username":     existingUser.Username,
			"birthDate":    birthDateString,
		},
	})
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Delete a user by ID
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users [delete]
func DeleteUser(c *gin.Context, users *mongo.Collection) {

	userID, exists := c.Get("userID")

	if !exists {
		c.JSON(401, gin.H{
			"error": "User ID not found",
		})
		return
	}

	objectID, err := bson.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	filter := bson.M{
		"_id": objectID,
	}
	var user User

	err = database.FindOne(
		c.Request.Context(),
		users,
		bson.M{"_id": objectID},
		&user,
	)

	if errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(404, gin.H{
			"error": "User not found",
		})
		return
	}

	if err != nil {
		c.JSON(500, gin.H{
			"error": "Database error",
		})
		return
	}

	if user.ProfileImage != "" {
		filePath := filepath.Join(file.UploadPath, user.ProfileImage)
		err = file.DeleteFile(filePath)
		if err != nil {
			c.JSON(500, gin.H{
				"error": "Failed to delete profile image",
			})
			return
		}
	}

	var deletedCount int64

	deletedCount, err = database.DeleteOne(
		c.Request.Context(),
		users,
		filter,
	)

	if err != nil {
		c.JSON(500, gin.H{
			"error": "Failed to delete user",
		})
		return
	}

	if deletedCount == 0 {
		c.JSON(404, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "User deleted successfully",
	})
}

func isValidName(name string) bool {
	for _, char := range name {
		if !unicode.IsLetter(char) {
			return false
		}
	}

	return true
}

// Login godoc
// @Summary Login user
// @Description Login with username and password
// @Accept json
// @Produce json
// @Param user body LoginRequest true "Username and password"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /login [post]
func Login(c *gin.Context, users *mongo.Collection) {

	var login LoginRequest

	err := c.ShouldBindJSON(&login)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid request",
		})
		return
	}

	username := login.Username
	password := login.Password

	var user User

	err = database.FindOne(
		c.Request.Context(),
		users,
		bson.M{"username": username},
		&user,
	)

	if err != nil {
		c.JSON(404, gin.H{
			"error": "User not found",
		})
		return
	}

	if password != user.Password {
		c.JSON(401, gin.H{
			"error": "Invalid password",
		})
		return
	}

	tokenString, err := token.GenerateToken(user.ID.Hex())
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Failed to generate token",
		})
		return
	}

	c.JSON(200, gin.H{
		"token": tokenString,
	})
}

func AuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(401, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")

		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, gin.H{
				"error": "Invalid authorization header",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		userID, err := token.GetUserID(tokenString)

		if err != nil {
			c.JSON(401, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		c.Set("userID", userID)

		c.Next()
	}
}

// @title User Management API
// @version 1.0
// @description A RESTful API for managing users.
// @host localhost:8080
// @BasePath
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	client, _, cancel := database.Connect()
	defer cancel()

	users := database.GetCollection(client, "product", "users")

	router := gin.Default()

	router.GET("/users", AuthMiddleware(), func(c *gin.Context) {
		GetUser(c, users)
	})

	router.POST("/users", func(c *gin.Context) {
		CreateUser(c, users)
	})

	router.PUT("/users", AuthMiddleware(), func(c *gin.Context) {
		UpdateUser(c, users)
	})

	router.DELETE("/users", AuthMiddleware(), func(c *gin.Context) {
		DeleteUser(c, users)
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Static("/uploads", "./uploads")
	router.POST("/login", func(c *gin.Context) {
		Login(c, users)
	})
	router.Run(":8080")
}
