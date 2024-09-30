package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/hashicorp/vault/api"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/rand"
	"gopkg.in/gomail.v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	port = os.Getenv("PORT")

	githubToken = os.Getenv("GITHUB_TOKEN")

	vaultAddr   = os.Getenv("VAULT_ADDR")
	vaultToken  = os.Getenv("VAULT_TOKEN")
	vaultKV     = os.Getenv("VAULT_KV")
	vaultCtx    = context.Background()
	vaultClient *api.Client

	smtpHost = os.Getenv("SMTP_HOST")
	smtpPort = os.Getenv("SMTP_PORT")
	smtpUser = os.Getenv("SMTP_USER")
	smtpPass = os.Getenv("SMTP_PASS")

	jwtSecret      = []byte(os.Getenv("JWT_SECRET"))
	jwtExpiryHours = os.Getenv("JWT_EXPIRY_HOURS")

	dbHost         = os.Getenv("DB_HOST")
	dbPort         = os.Getenv("DB_PORT")
	dbUser         = os.Getenv("DB_USER")
	dbRootPassword = os.Getenv("DB_ROOT_PASSWORD")
	dbPassword     = os.Getenv("DB_PASSWORD")
	dbName         = os.Getenv("DB_NAME")
	db             *gorm.DB
)

type User struct {
	gorm.Model
	Username string `gorm:"unique" json:"username"`
	Email    string `json:"email"`
	Password string `json:"-"`
	IsAdmin  bool   `json:"is_admin"`
}

const (
	ApplicationStatusPending  string = "Pending"
	ApplicationStatusApproved string = "Approved"
)

type Application struct {
	gorm.Model
	Name        string `json:"name"`
	GitURL      string `json:"git_url"`
	Branch      string `json:"branch"`
	Port        int    `json:"port"`
	Description string `json:"description"`
	Status      string `json:"status"`
	OwnerID     uint   `json:"owner_id"`
}

type SuccessResponseDTO struct {
	Message string `json:"message"`
}

type ErrorResponseDTO struct {
	Error string `json:"error"`
}

func init() {
	// Connect to database
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(&User{}, &Application{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Connect to Vault
	config := api.DefaultConfig()
	config.Address = vaultAddr

	vaultClient, err = api.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create Vault client: %v", err)
	}

	if vaultClient == nil {
		log.Fatalf("Vault client is nil, check initialization.")
	}
	vaultClient.SetToken(vaultToken)
}

func main() {
	router := gin.Default()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	auth := router.Group("/auth")
	{
		auth.POST("/login", login)
		auth.POST("/register", register)
	}

	users := router.Group("/users")
	users.Use(authMiddleware())
	{
		users.GET("", getUser)
		users.PATCH("", updateUser)
	}

	applications := router.Group("/applications")
	applications.Use(authMiddleware())
	{
		applications.POST("", submitApplication)
		applications.GET("", getApplications)
		applications.GET("/:appId", getApplication)

		environments := applications.Group("/:appId/environments")
		{
			environments.GET("", getEnvironments)
			environments.POST("", updateEnvironment)
		}
	}

	admin := router.Group("/admin")
	admin.Use(authMiddleware(), adminMiddleware())
	{
		users := admin.Group("/users")
		{
			users.GET("", getUsersByAdmin)
			users.GET("/:userId", getUserByAdmin)

			applications := users.Group("/:userId/applications")
			{
				applications.GET("", getApplicationsByAdmin)
			}
		}

		applications := admin.Group("/applications")
		{
			applications.POST("/:appId/approve", approveApplicationByAdmin)
			applications.GET("/:appId", getApplicationByAdmin)
		}
	}

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization token format"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userID := uint(claims["user_id"].(float64))
			isAdmin := claims["is_admin"].(bool)
			c.Set("user_id", userID)
			c.Set("is_admin", isAdmin)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Next()
	}
}

func adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, _ := c.Get("is_admin")
		if !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin permission required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

type LoginRequestDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponseDTO struct {
	Token string `json:"token"`
	SuccessResponseDTO
	ErrorResponseDTO
}

func login(c *gin.Context) {
	var loginRequest LoginRequestDTO
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, LoginResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Invalid request",
			},
		})
		log.Printf("Failed to bind JSON: %v", err)
		return
	}

	var user User
	if err := db.Where("username = ?", loginRequest.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, LoginResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Invalid credentials",
			},
		})
		log.Printf("Failed to find user: %v", err)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, LoginResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Invalid credentials",
			},
		})
		log.Printf("Failed to compare password: %v", err)
		return
	}

	expiryHours, _ := strconv.Atoi(jwtExpiryHours)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"is_admin": user.IsAdmin,
		"exp":      time.Now().Add(time.Hour * time.Duration(expiryHours)).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, LoginResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to generate token",
			},
		})
		log.Printf("Failed to generate token: %v", err)
		return
	}

	c.JSON(http.StatusOK, LoginResponseDTO{
		Token: tokenString,
		SuccessResponseDTO: SuccessResponseDTO{
			Message: "Login successful",
		},
	})
}

type RegisterRequestDTO struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterResponseDTO struct {
	SuccessResponseDTO
	ErrorResponseDTO
}

func register(c *gin.Context) {
	var registerRequest RegisterRequestDTO
	if err := c.ShouldBindJSON(&registerRequest); err != nil {
		c.JSON(http.StatusBadRequest, RegisterResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Invalid request",
			},
		})
		log.Printf("Failed to bind JSON: %v", err)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(registerRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, RegisterResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to hash password",
			},
		})
		log.Printf("Failed to hash password: %v", err)
		return
	}

	user := User{
		Username: registerRequest.Username,
		Email:    registerRequest.Email,
		Password: string(hash),
		IsAdmin:  false,
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, RegisterResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to register user",
			},
		})
		log.Printf("Failed to create user: %v", err)
		return
	}

	c.JSON(http.StatusCreated, RegisterResponseDTO{
		SuccessResponseDTO: SuccessResponseDTO{
			Message: "User registered successfully",
		},
	})
}

type GetUserResponseDTO struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"is_admin"`
	SuccessResponseDTO
	ErrorResponseDTO
}

func getUser(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var user User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, GetUserResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "User not found",
			},
		})
		log.Printf("Failed to find user: %v", err)
		return
	}

	c.JSON(http.StatusOK, GetUserResponseDTO{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		IsAdmin:  user.IsAdmin,
		SuccessResponseDTO: SuccessResponseDTO{
			Message: "User retrieved successfully",
		},
	})
}

type UpdateUserRequestDTO struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

type UpdateUserResponseDTO struct {
	SuccessResponseDTO
	ErrorResponseDTO
}

func updateUser(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var updateRequest UpdateUserRequestDTO
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, UpdateUserResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Invalid request",
			},
		})
		log.Printf("Failed to bind JSON: %v", err)
		return
	}

	var user User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, UpdateUserResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "User not found",
			},
		})
		log.Printf("Failed to find user: %v", err)
		return
	}

	if updateRequest.Email != "" {
		user.Email = updateRequest.Email
	}
	if updateRequest.Username != "" {
		user.Username = updateRequest.Username
	}

	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, UpdateUserResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to update user",
			},
		})
		log.Printf("Failed to save user: %v", err)
		return
	}

	c.JSON(http.StatusOK, UpdateUserResponseDTO{
		SuccessResponseDTO: SuccessResponseDTO{
			Message: "User updated successfully",
		},
	})
}

type GetApplicationsResponseDTO struct {
	Applications []struct {
		ID     uint   `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"applications"`
	SuccessResponseDTO
	ErrorResponseDTO
}

func getApplications(c *gin.Context) {
	userId, _ := c.Get("user_id")

	var applications []Application
	if err := db.Where("owner_id = ?", userId).Find(&applications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GetApplicationsResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to retrieve applications",
			},
		})
		log.Printf("Failed to find applications: %v", err)
		return
	}

	c.JSON(http.StatusOK, GetApplicationsResponseDTO{
		Applications: func() []struct {
			ID     uint   `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} {
			var result []struct {
				ID     uint   `json:"id"`
				Name   string `json:"name"`
				Status string `json:"status"`
			}
			for _, application := range applications {
				result = append(result, struct {
					ID     uint   `json:"id"`
					Name   string `json:"name"`
					Status string `json:"status"`
				}{
					ID:     application.ID,
					Name:   application.Name,
					Status: application.Status,
				})
			}
			return result
		}(),
		SuccessResponseDTO: SuccessResponseDTO{
			Message: "Applications retrieved successfully",
		},
	})
}

type SubmintApplicationRequestDTO struct {
	Name        string `json:"name" binding:"required"`
	GitURL      string `json:"git_url" binding:"required"`
	Branch      string `json:"branch" binding:"required"`
	Port        int    `json:"port" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type SubmitApplicationResponseDTO struct {
	SuccessResponseDTO
	ErrorResponseDTO
}

func submitApplication(c *gin.Context) {
	userId, _ := c.Get("user_id")

	var request SubmintApplicationRequestDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, SubmitApplicationResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Invalid request",
			},
		})
		log.Printf("Failed to bind JSON: %v", err)
		return
	}

	pattern := `^(?!.*(--|#|;|\bSELECT\b|\bINSERT\b|\bUPDATE\b|\bDELETE\b|\bDROP\b|\bEXEC\b|\bUNION\b|\bOR\b|\bAND\b))[a-z0-9-]+$`

	if matched, _ := regexp.MatchString(pattern, request.Name); !matched {
		c.JSON(http.StatusBadRequest, SubmitApplicationResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Invalid application name",
			},
		})
		log.Printf("Invalid application name: %s", request.Name)
		return
	}

	var existingApplication Application
	if err := db.Where("name = ?", request.Name).First(&existingApplication).Error; err == nil {
		c.JSON(http.StatusConflict, SubmitApplicationResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Application name already exists",
			},
		})
		log.Printf("Application name already exists: %s", request.Name)
		return
	}

	if request.Port < 1 || request.Port > 65535 {
		c.JSON(http.StatusBadRequest, SubmitApplicationResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Invalid port number",
			},
		})
		log.Printf("Invalid port number: %d", request.Port)
		return
	}

	application := Application{
		Name:        request.Name,
		GitURL:      request.GitURL,
		Branch:      request.Branch,
		Port:        request.Port,
		Description: request.Description,
		Status:      ApplicationStatusPending,
		OwnerID:     userId.(uint),
	}

	if err := db.Create(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, SubmitApplicationResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to submit application",
			},
		})
		log.Printf("Failed to create application: %v", err)
		return
	}

	c.JSON(http.StatusCreated, SubmitApplicationResponseDTO{
		SuccessResponseDTO: SuccessResponseDTO{
			Message: "Application submitted successfully",
		},
	})
}

type GetApplicationResponseDTO struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	GitURL      string `json:"git_url"`
	Branch      string `json:"branch"`
	Port        int    `json:"port"`
	Description string `json:"description"`
	OwnerID     uint   `json:"owner_id"`
	Status      string `json:"status"`
	SuccessResponseDTO
	ErrorResponseDTO
}

func getApplication(c *gin.Context) {
	userID, _ := c.Get("user_id")
	appID := c.Param("appId")

	var application Application
	if err := db.First(&application, appID).Error; err != nil {
		c.JSON(http.StatusNotFound, GetApplicationResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Application not found",
			},
		})
		log.Printf("Failed to find application: %v", err)
		return
	}

	if application.OwnerID != userID {
		c.JSON(http.StatusForbidden, GetApplicationResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Permission denied",
			},
		})
		log.Printf("Permission denied for application: %d", application.ID)
		return
	}

	c.JSON(http.StatusOK, GetApplicationResponseDTO{
		ID:          application.ID,
		Name:        application.Name,
		GitURL:      application.GitURL,
		Branch:      application.Branch,
		Port:        application.Port,
		Description: application.Description,
		OwnerID:     application.OwnerID,
		Status:      application.Status,
		SuccessResponseDTO: SuccessResponseDTO{
			Message: "Application retrieved successfully",
		},
	})
}

type GetEnvironmentsResponseDTO struct {
	Environments []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"environments"`
	SuccessResponseDTO
	ErrorResponseDTO
}

func getEnvironments(c *gin.Context) {
	userId, _ := c.Get("user_id")
	appID := c.Param("appId")

	var application Application
	if err := db.First(&application, appID).Error; err != nil {
		c.JSON(http.StatusNotFound, GetEnvironmentsResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Application not found",
			},
		})
		log.Printf("Failed to find application: %v", err)
		return
	}

	if application.OwnerID != userId {
		c.JSON(http.StatusForbidden, GetEnvironmentsResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Permission denied",
			},
		})
		log.Printf("Permission denied for application: %d", application.ID)
		return
	}

	if application.Status != ApplicationStatusApproved {
		c.JSON(http.StatusForbidden, GetEnvironmentsResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Application not approved",
			},
		})
		log.Printf("Application not approved: %d", application.ID)
		return
	}

	secret, err := vaultClient.KVv1(vaultKV).Get(vaultCtx, application.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GetEnvironmentsResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to read from Vault",
			},
		})
		log.Printf("Failed to read from Vault: %v", err)
		return
	}

	c.JSON(http.StatusOK, GetEnvironmentsResponseDTO{
		Environments: func() []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} {
			var result []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			}
			for key, value := range secret.Data {
				result = append(result, struct {
					Key   string `json:"key"`
					Value string `json:"value"`
				}{
					Key:   key,
					Value: value.(string),
				})
			}
			return result
		}(),
		SuccessResponseDTO: SuccessResponseDTO{
			Message: "Environments retrieved successfully",
		},
	})
}

type UpdateEnvironmentRequestDTO struct {
	Environments []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"environments"`
}

type UpdateEnvironmentResponseDTO struct {
	SuccessResponseDTO
	ErrorResponseDTO
}

func updateEnvironment(c *gin.Context) {
	userId, _ := c.Get("user_id")
	appID := c.Param("appId")

	var request UpdateEnvironmentRequestDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, UpdateEnvironmentResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Invalid request",
			},
		})
		log.Printf("Failed to bind JSON: %v", err)
		return
	}

	var application Application
	if err := db.First(&application, appID).Error; err != nil {
		c.JSON(http.StatusNotFound, UpdateEnvironmentResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Application not found",
			},
		})
		log.Printf("Failed to find application: %v", err)
		return
	}

	if application.OwnerID != userId {
		c.JSON(http.StatusForbidden, UpdateEnvironmentResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Permission denied",
			},
		})
		log.Printf("Permission denied for application: %d", application.ID)
		return
	}

	if application.Status != ApplicationStatusApproved {
		c.JSON(http.StatusForbidden, GetEnvironmentsResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Application not approved",
			},
		})
		log.Printf("Application not approved: %d", application.ID)
		return
	}

	data := make(map[string]interface{})
	for _, env := range request.Environments {
		data[env.Key] = env.Value
	}

	err := vaultClient.KVv1(vaultKV).Put(vaultCtx, application.Name, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UpdateEnvironmentResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to write to Vault",
			},
		})
		log.Printf("Failed to write to Vault: %v", err)
		return
	}

	c.JSON(http.StatusOK, UpdateEnvironmentResponseDTO{
		SuccessResponseDTO: SuccessResponseDTO{
			Message: "Environment updated successfully",
		},
	})
}

type GetUsersByAdminResponseDTO struct {
	Users []struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	} `json:"users"`
	SuccessResponseDTO
	ErrorResponseDTO
}

func getUsersByAdmin(c *gin.Context) {
	var users []User
	if err := db.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GetUsersByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to retrieve users",
			},
		})
		log.Printf("Failed to find users: %v", err)
		return
	}

	c.JSON(http.StatusOK, GetUsersByAdminResponseDTO{
		Users: func() []struct {
			ID       uint   `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
		} {
			var result []struct {
				ID       uint   `json:"id"`
				Username string `json:"username"`
				Email    string `json:"email"`
			}
			for _, user := range users {
				result = append(result, struct {
					ID       uint   `json:"id"`
					Username string `json:"username"`
					Email    string `json:"email"`
				}{
					ID:       user.ID,
					Username: user.Username,
					Email:    user.Email,
				})
			}
			return result
		}(),
		SuccessResponseDTO: SuccessResponseDTO{
			Message: "Users retrieved successfully",
		},
	})
}

type GetUserByAdminResponseDTO struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"is_admin"`
	SuccessResponseDTO
	ErrorResponseDTO
}

func getUserByAdmin(c *gin.Context) {
	userID := c.Param("userId")

	var user User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, GetUserByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "User not found",
			},
		})
		log.Printf("Failed to find user: %v", err)
		return
	}

	c.JSON(http.StatusOK, GetUserByAdminResponseDTO{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		IsAdmin:  user.IsAdmin,
		SuccessResponseDTO: SuccessResponseDTO{
			Message: "User retrieved successfully",
		},
	})
}

type GetApplicationsByAdminResponseDTO struct {
	Applications []struct {
		ID     uint   `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"applications"`
	SuccessResponseDTO
	ErrorResponseDTO
}

func getApplicationsByAdmin(c *gin.Context) {
	var applications []Application
	if err := db.Find(&applications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GetApplicationsByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to retrieve applications",
			},
		})
		log.Printf("Failed to find applications: %v", err)
		return
	}

	c.JSON(http.StatusOK, GetApplicationsByAdminResponseDTO{
		Applications: func() []struct {
			ID     uint   `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} {
			var result []struct {
				ID     uint   `json:"id"`
				Name   string `json:"name"`
				Status string `json:"status"`
			}
			for _, application := range applications {
				result = append(result, struct {
					ID     uint   `json:"id"`
					Name   string `json:"name"`
					Status string `json:"status"`
				}{
					ID:     application.ID,
					Name:   application.Name,
					Status: application.Status,
				})
			}
			return result
		}(),
		SuccessResponseDTO: SuccessResponseDTO{
			Message: "Applications retrieved successfully",
		},
	})
}

type ApproveApplicationByAdminResponseDTO struct {
	SuccessResponseDTO
	ErrorResponseDTO
}

func approveApplicationByAdmin(c *gin.Context) {
	appID := c.Param("appId")

	var application Application
	if err := db.First(&application, appID).Error; err != nil {
		c.JSON(http.StatusNotFound, ApproveApplicationByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Application not found",
			},
		})
		log.Printf("Failed to find application: %v", err)
		return
	}

	if application.Status != ApplicationStatusPending {
		c.JSON(http.StatusForbidden, ApproveApplicationByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Application already approved",
			},
		})
		log.Printf("Application already approved: %d", application.ID)
		return
	}

	if err := vaultClient.KVv1(vaultKV).Put(vaultCtx, application.Name, map[string]interface{}{"": ""}); err != nil {
		c.JSON(http.StatusInternalServerError, ApproveApplicationByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to write to Vault",
			},
		})
		log.Printf("Failed to write to Vault: %v", err)
		return
	}

	reqBody := fmt.Sprintf(`{
	"event_type": "write-values",
	"client_payload": {
		"appName": "%s",
		"git": "%s",
		"branch": "%s",
		"port": "%d"
	}
}`, application.Name, application.GitURL, application.Branch, application.Port)

	req, err := http.NewRequest("POST", "https://api.github.com/repos/injunweb/gitops-repo/dispatches",
		bytes.NewBuffer([]byte(reqBody)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ApproveApplicationByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to create GitHub request",
			},
		})
		log.Printf("Failed to create GitHub request: %v", err)
		return
	}

	req.Header.Set("Authorization", "token "+githubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ApproveApplicationByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to dispatch GitHub event",
			},
		})
		log.Printf("Failed to dispatch GitHub event: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, ApproveApplicationByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: fmt.Sprintf("GitHub dispatch failed with status: %s", resp.Status),
			},
		})
		log.Printf("GitHub dispatch failed with status: %s", resp.Status)
		return
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
	password := make([]byte, 12)
	for i := range password {
		password[i] = charset[seededRand.Intn(len(charset))]
	}

	rootDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		"root", dbRootPassword, dbHost, dbPort)

	rootDb, err := gorm.Open(mysql.Open(rootDsn), &gorm.Config{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ApproveApplicationByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to connect to root database",
			},
		})
		log.Printf("Failed to connect to root database: %v", err)
		return
	}

	query := fmt.Sprintf(`
		CREATE DATABASE IF NOT EXISTS %s;
		CREATE USER IF NOT EXISTS %s@'%%' IDENTIFIED BY '%s';
		GRANT ALL PRIVILEGES ON %s.* TO %s@'%%';
		FLUSH PRIVILEGES;
	`, application.Name, application.Name, password, application.Name, application.Name)

	querys := strings.Split(query, ";")
	for _, query := range querys {
		if err := rootDb.Exec(query).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ApproveApplicationByAdminResponseDTO{
				ErrorResponseDTO: ErrorResponseDTO{
					Error: "Failed to create database or user",
				},
			})
			log.Printf("Failed to create database or user: %v", err)
			return
		}
	}

	if err := rootDb.Exec(query).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ApproveApplicationByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to create database or user",
			},
		})
		log.Printf("Failed to create database or user: %v", err)
		return
	}

	var owner User
	if err := db.First(&owner, application.OwnerID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ApproveApplicationByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to find user email",
			},
		})
		log.Printf("Failed to find user email: %v", err)
		return
	}

	msg := fmt.Sprintf(
		"To: %s\r\n"+
			"Subject: Application Approved\r\n\r\n"+
			"Database Type: mysql\r\n"+
			"Your application has been approved.\r\n\r\n"+
			"Database Host: %s\r\n"+
			"Database Port: %s\r\n"+
			"Database Name: %s\r\n"+
			"Database User: %s\r\n"+
			"Database Password: %s\r\n",
		owner.Email, dbHost, dbPort, application.Name, application.Name, password,
	)

	m := gomail.NewMessage()
	m.SetHeader("From", smtpUser)
	m.SetHeader("To", owner.Email)
	m.SetHeader("Subject", "Application Approved")
	m.SetBody("text/plain", msg)

	port, err := strconv.Atoi(smtpPort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ApproveApplicationByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to send email",
			},
		})
		log.Printf("Failed to convert SMTP port: %v", err)
		return
	}

	d := gomail.NewDialer(smtpHost, port, smtpUser, smtpPass)
	if err := d.DialAndSend(m); err != nil {
		c.JSON(http.StatusInternalServerError, ApproveApplicationByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to send email",
			},
		})
		log.Printf("Failed to send email: %v", err)
		return
	}

	if err := db.Save(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ApproveApplicationByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to create database or user",
			},
		})
		log.Printf("Failed to create database or user: %v", err)
		return
	}

	c.JSON(http.StatusOK, ApproveApplicationByAdminResponseDTO{
		SuccessResponseDTO: SuccessResponseDTO{
			Message: "Application approved successfully",
		},
	})
}

type GetApplicationByAdminResponseDTO struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	GitURL      string `json:"git_url"`
	Branch      string `json:"branch"`
	Port        int    `json:"port"`
	Description string `json:"description"`
	OwnerID     uint   `json:"owner_id"`
	Status      string `json:"status"`
	SuccessResponseDTO
	ErrorResponseDTO
}

func getApplicationByAdmin(c *gin.Context) {
	appID := c.Param("appId")

	var application Application
	if err := db.First(&application, appID).Error; err != nil {
		c.JSON(http.StatusNotFound, GetApplicationByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Application not found",
			},
		})
		log.Printf("Failed to find application: %v", err)
		return
	}

	c.JSON(http.StatusOK, GetApplicationByAdminResponseDTO{
		ID:          application.ID,
		Name:        application.Name,
		GitURL:      application.GitURL,
		Branch:      application.Branch,
		Port:        application.Port,
		Description: application.Description,
		OwnerID:     application.OwnerID,
		Status:      application.Status,
		SuccessResponseDTO: SuccessResponseDTO{
			Message: "Application retrieved successfully",
		},
	})
}
