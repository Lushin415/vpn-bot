package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config содержит настройки приложения.
type Config struct {
	BotToken          string
	AdminTelegramID   int64
	YooKassaShopID    string
	YooKassaSecretKey string
	DatabaseURL       string
	Port              string // Порт для веб-сервера (например, для вебхуков)
}

// AppConfig – глобальная переменная для доступа к настройкам.
var AppConfig Config

// LoadConfig загружает конфигурацию из файла .env и переменных окружения.
func LoadConfig() {
	// Попытка загрузить .env файл (если он есть)
	err := godotenv.Load()
	if err != nil {
		log.Println("ℹ️ .env файл не найден, читаем настройки из переменных окружения")
	}

	AppConfig.BotToken = os.Getenv("BOT_TOKEN")
	if AppConfig.BotToken == "" {
		log.Fatal("🔴 Ошибка: BOT_TOKEN не задан!")
	}

	adminIDStr := os.Getenv("ADMIN_TELEGRAM_ID")
	if adminIDStr == "" {
		log.Fatal("🔴 Ошибка: ADMIN_TELEGRAM_ID не задан!")
	}
	adminID, err := strconv.ParseInt(adminIDStr, 10, 64)
	if err != nil {
		log.Fatalf("🔴 Ошибка преобразования ADMIN_TELEGRAM_ID: %v", err)
	}
	AppConfig.AdminTelegramID = adminID

	AppConfig.YooKassaShopID = os.Getenv("YOOKASSA_SHOP_ID")
	if AppConfig.YooKassaShopID == "" {
		log.Println("ℹ️ Предупреждение: YooKassaShopID не задан")
	}

	AppConfig.YooKassaSecretKey = os.Getenv("YOOKASSA_SECRET_KEY")
	if AppConfig.YooKassaSecretKey == "" {
		log.Println("ℹ️ Предупреждение: YooKassaSecretKey не задан")
	}

	AppConfig.DatabaseURL = os.Getenv("DATABASE_URL")
	if AppConfig.DatabaseURL == "" {
		log.Println("ℹ️ Предупреждение: DATABASE_URL не задан. Приложение может не запуститься без БД.")
	}

	// Устанавливаем порт для веб-сервера, если он не задан, используем значение по умолчанию (8080)
	AppConfig.Port = os.Getenv("PORT")
	if AppConfig.Port == "" {
		AppConfig.Port = "8080"
	}
}
