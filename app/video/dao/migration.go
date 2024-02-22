package dao

import (
	"Tiktok/model"
	"log"
)

func migration() {
	err := db.Set(`gorm:table_options`, "charset=utf8mb4").
		AutoMigrate(&model.User{}, &model.Follow{}, &model.Video{}, &model.Favorite{}, &model.Comment{})
	if err != nil {
		log.Fatal(err)
	}
}
