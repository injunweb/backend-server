package database

import (
	"fmt"
	"log"
	"time"

	"github.com/injunweb/backend-server/internal/config"
	"github.com/injunweb/backend-server/internal/models"
	"golang.org/x/exp/rand"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.AppConfig.DBUser,
		config.AppConfig.DBPassword,
		config.AppConfig.DBHost,
		config.AppConfig.DBPort,
		config.AppConfig.DBName)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	err = DB.AutoMigrate(&models.User{}, &models.Application{}, &models.ExtraHostnames{}, &models.Notification{}, &models.Subscription{})
	if err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	log.Println("Database connection established and migrations completed")
	return nil
}

func CreateDatabaseAndUser(appName string) (string, error) {
	password := generateRandomPassword()

	rootDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		"root", config.AppConfig.DBRootPassword, config.AppConfig.DBHost, config.AppConfig.DBPort)

	rootDb, err := gorm.Open(mysql.Open(rootDsn), &gorm.Config{})
	if err != nil {
		return "", fmt.Errorf("failed to connect to root database: %v", err)
	}

	queries := []string{
		fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`;", appName),
		fmt.Sprintf("CREATE USER IF NOT EXISTS '%s'@'%%' IDENTIFIED BY '%s';", appName, password),
		fmt.Sprintf("GRANT ALL PRIVILEGES ON `%s`.* TO '%s'@'%%';", appName, appName),
		"FLUSH PRIVILEGES;",
	}

	for _, query := range queries {
		if err := rootDb.Exec(query).Error; err != nil {
			return "", fmt.Errorf("failed to execute query: %v", err)
		}
	}

	return password, nil
}

func generateRandomPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
	password := make([]byte, 12)
	for i := range password {
		password[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(password)
}

func DeleteDatabaseAndUser(appName string) error {
	rootDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		"root", config.AppConfig.DBRootPassword, config.AppConfig.DBHost, config.AppConfig.DBPort)

	rootDb, err := gorm.Open(mysql.Open(rootDsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to root database: %v", err)
	}

	queries := []string{
		fmt.Sprintf("DROP DATABASE IF EXISTS `%s`;", appName),
		fmt.Sprintf("DROP USER IF EXISTS '%s'@'%%';", appName),
		"FLUSH PRIVILEGES;",
	}

	for _, query := range queries {
		if err := rootDb.Exec(query).Error; err != nil {
			return fmt.Errorf("failed to execute query: %v", err)
		}
	}

	return nil
}
