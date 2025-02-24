package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB подключается к PostgreSQL и выполняет миграцию моделей.
// 🔴 Убедитесь, что переменная окружения DATABASE_URL заполнена корректно!
func InitDB() {
	dsn := os.Getenv("DATABASE_URL") // 🔴 ! Проверьте, что DATABASE_URL заполнена!
	if dsn == "" {
		log.Fatal("🔴 Ошибка: переменная окружения DATABASE_URL не задана!")
	}

	dbInstance, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("🔴 Ошибка подключения к БД: %v", err)
	}

	// Автоматическая миграция моделей: User, Server, VLESSKey, Payment
	err = dbInstance.AutoMigrate(&User{}, &Server{}, &VLESSKey{}, &Payment{})
	if err != nil {
		log.Fatalf("🔴 Ошибка миграции: %v", err)
	}

	DB = dbInstance
	fmt.Println("✅ База данных успешно подключена и проинициализирована!")
}
