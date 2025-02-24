package main

import (
	"fmt"
	"log"

	"vpn-bot/internal/bot"
	"vpn-bot/internal/db"

	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные окружения из файла .env
	err := godotenv.Load()
	if err != nil {
		log.Println("🔴 Не удалось загрузить .env файл, используются системные переменные окружения")
	}

	// Инициализируем подключение к базе данных
	db.InitDB()

	// Запускаем веб-сервер для обработки веб-хуков от Юкассы
	go bot.StartWebhook()

	// Запускаем фоновые задачи (cron jobs)
	go bot.InitCronJobs()

	// Запускаем Telegram-бота
	fmt.Println("✅ Запуск Telegram-бота...")
	bot.StartBot()
}
