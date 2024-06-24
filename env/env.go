package env

import (
	"fmt"
	"os"
)

var DB_USER string
var DB_PASSWORD string
var DB_HOST string
var DB_PORT string
var DB_NAME string
var JWT_SECRET_KEY string
var GITHUB_DISPATCH_TOKEN string
var GITHUB_OWNER string

func init() {
	DB_USER = os.Getenv("DB_USER")
	if DB_USER == "" {
		fmt.Println("Warning: DB_USER environment variable is not set")
	}

	DB_PASSWORD = os.Getenv("DB_PASSWORD")
	if DB_PASSWORD == "" {
		fmt.Println("Warning: DB_PASSWORD environment variable is not set")
	}

	DB_HOST = os.Getenv("DB_HOST")
	if DB_HOST == "" {
		fmt.Println("Warning: DB_HOST environment variable is not set")
	}

	DB_PORT = os.Getenv("DB_PORT")
	if DB_PORT == "" {
		fmt.Println("Warning: DB_PORT environment variable is not set")
	}

	DB_NAME = os.Getenv("DB_NAME")
	if DB_NAME == "" {
		fmt.Println("Warning: DB_NAME environment variable is not set")
	}

	JWT_SECRET_KEY = os.Getenv("JWT_SECRET_KEY")
	if JWT_SECRET_KEY == "" {
		fmt.Println("Warning: JWT_SECRET_KEY environment variable is not set")
	}

	GITHUB_DISPATCH_TOKEN = os.Getenv("GITHUB_DISPATCH_TOKEN")
	if GITHUB_DISPATCH_TOKEN == "" {
		fmt.Println("Warning: GITHUB_TOKEN environment variable is not set")
	}

	GITHUB_OWNER = os.Getenv("GITHUB_OWNER")
	if GITHUB_OWNER == "" {
		fmt.Println("Warning: GITHUB_OWNER environment variable is not set")
	}
}
