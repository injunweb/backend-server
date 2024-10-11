package config

import "os"

type Config struct {
	Port              string
	GithubToken       string
	VaultAddr         string
	VaultToken        string
	VaultKV           string
	InCluster         string
	KubeConfig        string
	HarborURL         string
	HarborUsername    string
	HarborPassword    string
	HarborProjectName string
	SMTPHost          string
	SMTPPort          string
	SMTPSenderEmail   string
	SMTPUser          string
	SMTPPass          string
	JWTSecret         string
	JWTExpiryHours    string
	DBHost            string
	DBPort            string
	DBUser            string
	DBRootPassword    string
	DBPassword        string
	DBName            string
}

var AppConfig Config

func Load() {
	AppConfig = Config{
		Port:              os.Getenv("PORT"),
		GithubToken:       os.Getenv("GITHUB_TOKEN"),
		VaultAddr:         os.Getenv("VAULT_ADDR"),
		VaultToken:        os.Getenv("VAULT_TOKEN"),
		VaultKV:           os.Getenv("VAULT_KV"),
		InCluster:         os.Getenv("IN_CLUSTER"),
		KubeConfig:        os.Getenv("KUBE_CONFIG"),
		HarborURL:         os.Getenv("HARBOR_URL"),
		HarborUsername:    os.Getenv("HARBOR_USERNAME"),
		HarborPassword:    os.Getenv("HARBOR_PASSWORD"),
		HarborProjectName: os.Getenv("HARBOR_PROJECT_NAME"),
		SMTPHost:          os.Getenv("SMTP_HOST"),
		SMTPPort:          os.Getenv("SMTP_PORT"),
		SMTPSenderEmail:   os.Getenv("SMTP_SENDER_EMAIL"),
		SMTPUser:          os.Getenv("SMTP_USER"),
		SMTPPass:          os.Getenv("SMTP_PASS"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		JWTExpiryHours:    os.Getenv("JWT_EXPIRY_HOURS"),
		DBHost:            os.Getenv("DB_HOST"),
		DBPort:            os.Getenv("DB_PORT"),
		DBUser:            os.Getenv("DB_USER"),
		DBRootPassword:    os.Getenv("DB_ROOT_PASSWORD"),
		DBPassword:        os.Getenv("DB_PASSWORD"),
		DBName:            os.Getenv("DB_NAME"),
	}
}
