package main

import (
	"fmt"
	"log"
	"strings"
	"url-shortener/internal/config"
	"url-shortener/internal/database"
	"url-shortener/internal/route/user/sign"
	"url-shortener/internal/server"
)

func main() {
	// TODO: connect to database at startup
	// TODO: make env variable configurable
	// TODO: do TDD
	env := config.ReadEnv()

	/**
	Database configuration
	*/
	dbConfig := database.Config{
		Username: env.DBUser,
		Password: env.DBPass,
		Host:     env.DBHost,
		Port:     env.DBPort,
		DBName:   env.DBName,
		DBParams: env.DBParams,
	}
	db, err := database.NewMySQLDatabase(dbConfig)
	if err != nil {
		log.Fatalf("Unable to set up database | Reason: %v\n", err)
	}

	/**
	jwtKey configuration
	*/
	jwtKey := []byte(env.JwtKey)

	/**
	googleOauth configuration
	*/
	gConf := sign.GoogleOauthConfig{
		ClientId:     env.GoogleOauthClientId,
		ClientSecret: env.GoogleOauthClientSecret,
	}

	r := server.SetupServer(db, jwtKey, env.UseHttps, env.BaseUrl.String(), strings.Split(env.BaseUrl.Host, ":")[0], "./internal/template", gConf) // blocking if starting with success
	log.Printf("Server is listening...")
	port := fmt.Sprintf(":%v", env.Port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Unable to start server: %v\n", err)
	}
}
