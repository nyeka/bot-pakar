package config

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() *gorm.DB {

	DB_USER := os.Getenv("DB_USER")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	DB_HOST := os.Getenv("DB_HOST")
	DB_PORT := os.Getenv("DB_PORT")
	DB_NAME := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=require", DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME)

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage

	}), &gorm.Config{})

	if err != nil {
		fmt.Println(DB_HOST)
		panic(err.Error())
	}
	sql, err := db.DB()
	if err != nil {
		panic(err)
	}

	sql.SetMaxOpenConns(16)
	sql.SetMaxIdleConns(8)
	sql.SetConnMaxLifetime(30 * time.Minute)

	// var models = []interface{}{&model.Notification{}, &model.History{}, &model.Reports{}, &model.Words{}, &model.Users{}, &model.Blog{}, &model.Profiles{}, &model.Review{}, &model.Verification{}}
	// db.AutoMigrate(models...)

	return db
}
