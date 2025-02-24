package handlers

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"vpn-bot/internal/db"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// getAdminID получает ID администратора из переменной окружения.
func getAdminID() int64 {
	adminIDStr := os.Getenv("ADMIN_TELEGRAM_ID") // 🔴 ! Убедитесь, что ADMIN_TELEGRAM_ID заполнена корректно!
	adminID, err := strconv.ParseInt(adminIDStr, 10, 64)
	if err != nil {
		log.Fatalf("🔴 Ошибка преобразования ADMIN_TELEGRAM_ID: %v", err)
	}
	return adminID
}

// ListServersHandler обрабатывает команду /listservers для администратора.
func ListServersHandler(bot *tgbotapi.BotAPI, chatID int64) {
	if chatID != getAdminID() {
		bot.Send(tgbotapi.NewMessage(chatID, "⛔ Доступ запрещён"))
		return
	}

	// Получаем список серверов с подсчётом свободных ключей
	var result []struct {
		Name      string
		IP        string
		FreeKeys  int
		TotalKeys int
		IsActive  bool
	}

	err := db.DB.Raw(`
		SELECT s.name, s.ip,
			COUNT(CASE WHEN k.is_used = false THEN 1 END) AS free_keys,
			COUNT(k.id) AS total_keys,
			s.is_active
		FROM servers s
		LEFT JOIN vless_keys k ON s.id = k.server_id
		GROUP BY s.id
		ORDER BY s.name
	`).Scan(&result).Error
	if err != nil {
		log.Printf("🔴 Ошибка запроса серверов: %v", err)
		bot.Send(tgbotapi.NewMessage(chatID, "Ошибка получения списка серверов"))
		return
	}

	message := "📡 Список серверов:\n"
	for _, server := range result {
		status := "🟢 Активен"
		if !server.IsActive {
			status = "🔴 Неактивен"
		}
		message += fmt.Sprintf(
			"\n🌍 *%s* (%s)\n🔑 Свободных ключей: %d / %d\n%s\n",
			server.Name, server.IP, server.FreeKeys, server.TotalKeys, status,
		)
	}

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

// BroadcastHandler обрабатывает команду /broadcast <сообщение> для рассылки всем пользователям.
func BroadcastHandler(bot *tgbotapi.BotAPI, chatID int64, broadcastText string) {
	if chatID != getAdminID() {
		bot.Send(tgbotapi.NewMessage(chatID, "⛔ Доступ запрещён"))
		return
	}

	// Получаем список всех пользователей
	var users []db.User
	err := db.DB.Select("telegram_id").Find(&users).Error
	if err != nil {
		log.Printf("🔴 Ошибка получения пользователей: %v", err)
		bot.Send(tgbotapi.NewMessage(chatID, "Ошибка получения списка пользователей"))
		return
	}

	sent, failed := 0, 0
	for _, user := range users {
		msg := tgbotapi.NewMessage(user.TelegramID, broadcastText)
		_, err := bot.Send(msg)
		if err != nil {
			failed++
		} else {
			sent++
		}
	}

	response := fmt.Sprintf("📢 Рассылка завершена:\n✅ Отправлено: %d\n❌ Ошибок: %d", sent, failed)
	bot.Send(tgbotapi.NewMessage(chatID, response))
}

// HandleAdminCommand определяет, какую админ-команду выполнить, исходя из входящего сообщения.
func HandleAdminCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	text := update.Message.Text

	if text == "/listservers" {
		ListServersHandler(bot, chatID)
	} else if strings.HasPrefix(text, "/broadcast") {
		// Формат команды: /broadcast <сообщение>
		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
			bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Использование: /broadcast <сообщение>"))
			return
		}
		BroadcastHandler(bot, chatID, parts[1])
	} else {
		bot.Send(tgbotapi.NewMessage(chatID, "Неизвестная админ-команда"))
	}
}
