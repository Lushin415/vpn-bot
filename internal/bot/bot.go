package bot

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// StartBot инициализирует Telegram-бота и начинает обработку обновлений.
func StartBot() {
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("🔴 Ошибка: переменная окружения BOT_TOKEN не задана!")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("🔴 Ошибка создания бота: %v", err)
	}

	bot.Debug = false
	log.Printf("✅ Бот запущен: %s", bot.Self.UserName)

	// Конфигурация получения обновлений
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates, err := bot.GetUpdatesChan(updateConfig)
	if err != nil {
		log.Fatalf("🔴 Ошибка получения обновлений: %v", err)
	}

	// Обрабатываем каждое обновление через соответствующие обработчики
	for update := range updates {
		go HandleUpdate(bot, update)
	}
}
