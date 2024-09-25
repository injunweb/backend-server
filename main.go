package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/hashicorp/vault/api"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/rand"
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

	smtpHost   = os.Getenv("SMTP_HOST")
	smtpPort   = 587
	smtpUser   = os.Getenv("SMTP_USER")
	smtpPass   = os.Getenv("SMTP_PASS")
	smtpClient *smtp.Client

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
	ApplianceStatusPending  string = "Pending"
	ApplianceStatusApproved string = "Approved"
)

type Appliance struct {
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
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FSeoul",
		dbUser, dbPassword, dbHost, dbPort, dbName)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(&User{}, &Appliance{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Connect to Vault
	vaultClient, err = api.NewClient(&api.Config{
		Address: vaultAddr,
	})
	if err != nil {
		log.Fatalf("Failed to connect to Vault: %v", err)
	}
	vaultClient.SetToken(vaultToken)

	// Connect to SMTP
	smtpAuth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	smtpAddr := fmt.Sprintf("%s:%d", smtpHost, smtpPort)
	smtpClient, err = smtp.Dial(smtpAddr)
	if err != nil {
		log.Fatalf("Failed to connect to SMTP: %v", err)
	}
	if err := smtpClient.Auth(smtpAuth); err != nil {
		log.Fatalf("Failed to authenticate SMTP: %v", err)
	}
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
		users.GET("/", getUser)
		users.PATCH("/", updateUser)
	}

	appliances := router.Group("/appliances")
	appliances.Use(authMiddleware())
	{
		appliances.POST("/", submitAppliance)
		appliances.GET("/", getAppliances)
		appliances.GET("/:appId", getAppliance)

		environments := appliances.Group("/:appId/environments")
		{
			environments.GET("/", getEnvironments)
			environments.POST("/", updateEnvironment)
		}
	}

	admin := router.Group("/admin")
	admin.Use(authMiddleware(), adminMiddleware())
	{
		users := admin.Group("/users")
		{
			users.GET("/", getUsersByAdmin)
			users.GET("/:userId", getUsersByAdmin)

			appliances := users.Group("/:userId/appliances")
			{
				appliances.GET("/", getAppliancesByAdmin)
			}
		}

		appliances := admin.Group("/appliances")
		{
			appliances.POST("/:appId/approve", approveApplianceByAdmin)
			appliances.GET("/:appId", getApplianceByAdmin)
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
				Error: err.Error(),
			},
		})
		return
	}

	var user User
	if err := db.Where("username = ?", loginRequest.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, LoginResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Invalid credentials",
			},
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, LoginResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Invalid credentials",
			},
		})
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
				Error: err.Error(),
			},
		})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(registerRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, RegisterResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to hash password",
			},
		})
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
	Email    string `json:"email" binding:"required"`
	Username string `json:"username" binding:"required"`
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
				Error: err.Error(),
			},
		})
		return
	}

	var user User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, UpdateUserResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "User not found",
			},
		})
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
		return
	}

	c.JSON(http.StatusOK, UpdateUserResponseDTO{
		SuccessResponseDTO: SuccessResponseDTO{
			Message: "User updated successfully",
		},
	})
}

type GetAppliancesResponseDTO struct {
	Appliances []struct {
		ID     uint   `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"appliances"`
	SuccessResponseDTO
	ErrorResponseDTO
}

func getAppliances(c *gin.Context) {
	userId, _ := c.Get("user_id")

	var appliances []Appliance
	if err := db.Where("owner_id = ?", userId).Find(&appliances).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GetAppliancesResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to retrieve appliances",
			},
		})
		return
	}

	c.JSON(http.StatusOK, GetAppliancesResponseDTO{
		Appliances: func() []struct {
			ID     uint   `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} {
			var result []struct {
				ID     uint   `json:"id"`
				Name   string `json:"name"`
				Status string `json:"status"`
			}
			for _, appliance := range appliances {
				result = append(result, struct {
					ID     uint   `json:"id"`
					Name   string `json:"name"`
					Status string `json:"status"`
				}{
					ID:     appliance.ID,
					Name:   appliance.Name,
					Status: appliance.Status,
				})
			}
			return result
		}(),
		SuccessResponseDTO: SuccessResponseDTO{
			Message: "Appliances retrieved successfully",
		},
	})
}

type SubmintApplianceRequestDTO struct {
	Name        string `json:"name" binding:"required"`
	GitURL      string `json:"git_url" binding:"required"`
	Branch      string `json:"branch" binding:"required"`
	Port        int    `json:"port" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type SubmitApplianceResponseDTO struct {
	SuccessResponseDTO
	ErrorResponseDTO
}

func submitAppliance(c *gin.Context) {
	userId, _ := c.Get("user_id")

	var request SubmintApplianceRequestDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, SubmitApplianceResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: err.Error(),
			},
		})
		return
	}

	if matched, _ := regexp.MatchString("^[a-z0-9-]+$", request.Name); !matched {
		c.JSON(http.StatusBadRequest, SubmitApplianceResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Invalid appliance name",
			},
		})
		return
	}

	var existingAppliance Appliance
	if err := db.Where("name = ?", request.Name).First(&existingAppliance).Error; err == nil {
		c.JSON(http.StatusConflict, SubmitApplianceResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Appliance name already exists",
			},
		})
		return
	}

	if request.Port < 1 || request.Port > 65535 {
		c.JSON(http.StatusBadRequest, SubmitApplianceResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Invalid port number",
			},
		})
		return
	}

	appliance := Appliance{
		Name:        request.Name,
		GitURL:      request.GitURL,
		Branch:      request.Branch,
		Port:        request.Port,
		Description: request.Description,
		Status:      ApplianceStatusPending,
		OwnerID:     userId.(uint),
	}

	if err := db.Create(&appliance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, SubmitApplianceResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to submit appliance",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, SubmitApplianceResponseDTO{
		SuccessResponseDTO: SuccessResponseDTO{
			Message: "Appliance submitted successfully",
		},
	})
}

type GetApplianceResponseDTO struct {
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

func getAppliance(c *gin.Context) {
	userID, _ := c.Get("user_id")
	appID := c.Param("appId")

	var appliance Appliance
	if err := db.First(&appliance, appID).Error; err != nil {
		c.JSON(http.StatusNotFound, GetApplianceResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Appliance not found",
			},
		})
		return
	}

	if appliance.OwnerID != userID {
		c.JSON(http.StatusForbidden, GetApplianceResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Permission denied",
			},
		})
		return
	}

	c.JSON(http.StatusOK, GetApplianceResponseDTO{
		ID:          appliance.ID,
		Name:        appliance.Name,
		GitURL:      appliance.GitURL,
		Branch:      appliance.Branch,
		Port:        appliance.Port,
		Description: appliance.Description,
		OwnerID:     appliance.OwnerID,
		Status:      appliance.Status,
		SuccessResponseDTO: SuccessResponseDTO{
			Message: "Appliance retrieved successfully",
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

	var appliance Appliance
	if err := db.First(&appliance, appID).Error; err != nil {
		c.JSON(http.StatusNotFound, GetEnvironmentsResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Appliance not found",
			},
		})
		return
	}

	if appliance.OwnerID != userId {
		c.JSON(http.StatusForbidden, GetEnvironmentsResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Permission denied",
			},
		})
		return
	}

	if appliance.Status != ApplianceStatusApproved {
		c.JSON(http.StatusForbidden, GetEnvironmentsResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Appliance not approved",
			},
		})
		return
	}

	secret, err := vaultClient.KVv1(vaultKV).Get(vaultCtx, appliance.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GetEnvironmentsResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to read from Vault",
			},
		})
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
				Error: err.Error(),
			},
		})
		return
	}

	var appliance Appliance
	if err := db.First(&appliance, appID).Error; err != nil {
		c.JSON(http.StatusNotFound, UpdateEnvironmentResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Appliance not found",
			},
		})
		return
	}

	if appliance.OwnerID != userId {
		c.JSON(http.StatusForbidden, UpdateEnvironmentResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Permission denied",
			},
		})
		return
	}

	if appliance.Status != ApplianceStatusApproved {
		c.JSON(http.StatusForbidden, GetEnvironmentsResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Appliance not approved",
			},
		})
		return
	}

	data := make(map[string]interface{})
	for _, env := range request.Environments {
		data[env.Key] = env.Value
	}

	err := vaultClient.KVv1(vaultKV).Put(vaultCtx, appliance.Name, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UpdateEnvironmentResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to write to Vault",
			},
		})
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

type GetAppliancesByAdminResponseDTO struct {
	Appliances []struct {
		ID     uint   `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"appliances"`
	SuccessResponseDTO
	ErrorResponseDTO
}

func getAppliancesByAdmin(c *gin.Context) {
	var appliances []Appliance
	if err := db.Find(&appliances).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GetAppliancesByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to retrieve appliances",
			},
		})
		return
	}

	c.JSON(http.StatusOK, GetAppliancesByAdminResponseDTO{
		Appliances: func() []struct {
			ID     uint   `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} {
			var result []struct {
				ID     uint   `json:"id"`
				Name   string `json:"name"`
				Status string `json:"status"`
			}
			for _, appliance := range appliances {
				result = append(result, struct {
					ID     uint   `json:"id"`
					Name   string `json:"name"`
					Status string `json:"status"`
				}{
					ID:     appliance.ID,
					Name:   appliance.Name,
					Status: appliance.Status,
				})
			}
			return result
		}(),
		SuccessResponseDTO: SuccessResponseDTO{
			Message: "Appliances retrieved successfully",
		},
	})
}

type ApproveApplianceByAdminRequestDTO struct {
	Password string `json:"password" binding:"required"`
}

type ApproveApplianceByAdminResponseDTO struct {
	SuccessResponseDTO
	ErrorResponseDTO
}

func approveApplianceByAdmin(c *gin.Context) {
	appID := c.Param("appId")

	var appliance Appliance
	if err := db.First(&appliance, appID).Error; err != nil {
		c.JSON(http.StatusNotFound, ApproveApplianceByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Appliance not found",
			},
		})
		return
	}

	if appliance.Status != ApplianceStatusPending {
		c.JSON(http.StatusForbidden, ApproveApplianceByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Appliance already approved",
			},
		})
		return
	}

	if err := vaultClient.KVv1(vaultKV).Put(vaultCtx, appliance.Name, map[string]interface{}{}); err != nil {
		c.JSON(http.StatusInternalServerError, ApproveApplianceByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to write to Vault",
			},
		})
		return
	}

	reqBody := fmt.Sprintf(`{
	"event_type": "write-values",
	"client_payload": {
		"appName": "%s",
		"git": "%s",
		"branch": ""%s",
		"port": "%d"
	}
}`, appliance.Name, appliance.GitURL, appliance.Branch, appliance.Port)

	req, err := http.NewRequest("POST", "https://api.github.com/repos/injunweb/gitops-repo/dispatches",
		bytes.NewBuffer([]byte(reqBody)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ApproveApplianceByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to create GitHub request",
			},
		})
		return
	}

	req.Header.Set("Authorization", "token "+githubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ApproveApplianceByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to dispatch GitHub event",
			},
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, ApproveApplianceByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: fmt.Sprintf("GitHub dispatch failed with status: %s", resp.Status),
			},
		})
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
		log.Fatalf("Failed to connect to root database: %v", err)
	}

	query := `
		CREATE DATABASE IF NOT EXISTS ?;
		CREATE USER IF NOT EXISTS ?@'%' IDENTIFIED BY ?;
		GRANT ALL PRIVILEGES ON ?.* TO ?@'%';
		FLUSH PRIVILEGES;
	`

	if err := rootDb.Exec(query, appliance.Name, appliance.Name, password, appliance.Name, appliance.Name).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ApproveApplianceByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to create database or user",
			},
		})
		return
	}

	var owner User
	if err := db.First(&owner, appliance.OwnerID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ApproveApplianceByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to find user email",
			},
		})
		return
	}

	msg := fmt.Sprintf(
		"To: %s\r\n"+
			"Subject: Appliance Approved\r\n\r\n"+
			"Database Type: mysql\r\n"+
			"Your appliance has been approved.\r\n\r\n"+
			"Database Host: %s\r\n"+
			"Database Port: %s\r\n"+
			"Database Name: %s\r\n"+
			"Database User: %s\r\n"+
			"Database Password: %s\r\n",
		owner.Email, dbHost, dbPort, appliance.Name, appliance.Name, password,
	)

	if err := smtpClient.Mail(smtpUser); err != nil {
		c.JSON(http.StatusInternalServerError, ApproveApplianceByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to set SMTP sender",
			},
		})
		return
	}

	if err := smtpClient.Rcpt(owner.Email); err != nil {
		c.JSON(http.StatusInternalServerError, ApproveApplianceByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to set SMTP recipient",
			},
		})
		return
	}

	w, err := smtpClient.Data()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ApproveApplianceByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to send email data",
			},
		})
		return
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ApproveApplianceByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to write email message",
			},
		})
		return
	}

	err = w.Close()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ApproveApplianceByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to close email writer",
			},
		})
		return
	}
	if err := db.Save(&appliance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ApproveApplianceByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Failed to create database or user",
			},
		})
		return
	}

	c.JSON(http.StatusOK, ApproveApplianceByAdminResponseDTO{
		SuccessResponseDTO: SuccessResponseDTO{
			Message: "Appliance approved successfully",
		},
	})
}

type GetApplianceByAdminResponseDTO struct {
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

func getApplianceByAdmin(c *gin.Context) {
	appID := c.Param("appId")

	var appliance Appliance
	if err := db.First(&appliance, appID).Error; err != nil {
		c.JSON(http.StatusNotFound, GetApplianceByAdminResponseDTO{
			ErrorResponseDTO: ErrorResponseDTO{
				Error: "Appliance not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, GetApplianceByAdminResponseDTO{
		ID:          appliance.ID,
		Name:        appliance.Name,
		GitURL:      appliance.GitURL,
		Branch:      appliance.Branch,
		Port:        appliance.Port,
		Description: appliance.Description,
		OwnerID:     appliance.OwnerID,
		Status:      appliance.Status,
		SuccessResponseDTO: SuccessResponseDTO{
			Message: "Appliance retrieved successfully",
		},
	})
}
