package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	url2 "net/url"
	"os"
	"reflect"
)

type Env struct {
	DBUser                  string
	DBPass                  string
	DBHost                  string
	DBPort                  string
	DBName                  string
	DBParams                string
	JwtKey                  string
	GoogleOauthClientId     string
	GoogleOauthClientSecret string
	BaseUrl                 *url2.URL
	Port                    string
	UseHttps                bool
}

func ReadEnv() Env {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Unable to read .env | Reason: %v ...skipped\n", err)
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		log.Printf("DB_USER is empty. Default as \"root\"\n")
		dbUser = "root"
	}

	dbPass := os.Getenv("DB_PASSWORD")
	if dbPass == "" {
		log.Printf("DB_PASSWORD is empty\n")
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		log.Printf("DB_HOST is empty. Default as \"localhost\"\n")
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		log.Printf("DB_PORT is empty. Default as \"3306\"\n")
		dbPort = "3306"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		log.Printf("DB_NAME is empty. Default as \"url_shortener\"\n")
		dbName = "url_shortener"
	}

	dbParams := os.Getenv("DB_PARAMS")
	if dbParams == "" {
		log.Printf("DB_PARAMS is empty. Default as \"charset=utf8&parseTime=True&loc=Local\"\n")
		dbParams = "charset=utf8&parseTime=True&loc=Local"
	}

	jwtKey := os.Getenv("JWT_KEY")
	if jwtKey == "" {
		log.Printf("JWT_KEY is empty. Default as \"testKey\"\n")
		jwtKey = "testKey"
	}

	googleClientId := os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	if googleClientId == "" {
		log.Printf("GOOGLE_OAUTH_CLIENT_ID is empty. Default as \"959723324236-0e23oe704fp1rtf3k5qc780mijahd1b3.apps.googleusercontent.com\"\n")
		googleClientId = "959723324236-0e23oe704fp1rtf3k5qc780mijahd1b3.apps.googleusercontent.com"
	}

	googleClientSecret := os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")
	if googleClientSecret == "" {
		log.Printf("GOOGLE_OAUTH_CLIENT_SECRET is empty. Default as \"xG1-yt61nKfvPUAfZumduCNO\"\n")
		googleClientSecret = "xG1-yt61nKfvPUAfZumduCNO"
	}

	port := os.Getenv("API_PORT")
	if port == "" {
		log.Printf("API_PORT is empty. Default as \"8080\"\n")
		port = "8080"
	}

	baseUrl := os.Getenv("BASE_URL")
	if baseUrl == "" {
		log.Printf("BASE_URL is empty. Default as \"http://url-shortener.com:%v\"\n", port)
		baseUrl = fmt.Sprintf("http://url-shortener.com:%v", port)
	}

	u, err := url2.ParseRequestURI(baseUrl)
	if err != nil {
		panic("Invalid baseUrl")
	}
	var useHttps bool
	switch u.Scheme {
	case "http":
		useHttps = false
	case "https":
		useHttps = true
	default:
		panic("Invalid baseUrl")
	}

	env := Env{
		DBUser:                  dbUser,
		DBPass:                  dbPass,
		DBHost:                  dbHost,
		DBPort:                  dbPort,
		DBName:                  dbName,
		DBParams:                dbParams,
		JwtKey:                  jwtKey,
		GoogleOauthClientId:     googleClientId,
		GoogleOauthClientSecret: googleClientSecret,
		BaseUrl:                 u,
		Port:                    port,
		UseHttps:                useHttps,
	}

	fmt.Printf("===========================\n")
	v := reflect.ValueOf(env)
	for i := 0; i < v.NumField(); i++ {
		fmt.Printf("%v: %v\n", v.Type().Field(i).Name, v.Field(i).Interface())
	}
	fmt.Printf("===========================\n")

	return env
}
