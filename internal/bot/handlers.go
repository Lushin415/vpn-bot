package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"vpn-bot/internal/db"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// HandleUpdate обрабатывает входящие обновления (сообщения и callback'и)
func HandleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	// Обработка текстовых сообщений
	if update.Message != nil {
		switch update.Message.Text {
		case "/start":
			sendStartMenu(bot, update.Message.Chat.ID)
		case "/support":
			sendSupportInfo(bot, update.Message.Chat.ID)
		case "/buy":
			// Отправляем выбор сервера для покупки подписки
			sendServerSelection(bot, update.Message.Chat.ID)
		default:
			sendUnknownCommand(bot, update.Message.Chat.ID)
		}
	}

	// Обработка callback-запросов (inline-кнопки)
	if update.CallbackQuery != nil {
		handleCallback(bot, update.CallbackQuery)
	}
}

// sendStartMenu отправляет главное меню пользователю
func sendStartMenu(bot *tgbotapi.BotAPI, chatID int64) {
	text := "Привет! Выберите действие:"
	// Клавиатура с основными командами
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🚀 Купить подписку"),
			tgbotapi.NewKeyboardButton("🔄 Продлить подписку"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📊 Мои подписки"),
			tgbotapi.NewKeyboardButton("📨 Поддержка"),
		),
	)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	if _, err := bot.Send(msg); err != nil {
		log.Printf("🔴 Ошибка отправки стартового меню: %v", err)
	}
}

// sendSupportInfo отправляет информацию о поддержке
func sendSupportInfo(bot *tgbotapi.BotAPI, chatID int64) {
	// 🔴 ! Проверьте, что @YourSupportChat заменён на актуальный контакт поддержки!
	text := "Чтобы связаться с поддержкой, напишите: @YourSupportChat"
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("🔴 Ошибка отправки поддержки: %v", err)
	}
}

// sendUnknownCommand отправляет сообщение о неизвестной команде
func sendUnknownCommand(bot *tgbotapi.BotAPI, chatID int64) {
	text := "Неизвестная команда. Используйте /start для начала работы."
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("🔴 Ошибка отправки сообщения: %v", err)
	}
}

// sendServerSelection отправляет пользователю список доступных серверов (локаций)
func sendServerSelection(bot *tgbotapi.BotAPI, chatID int64) {
	var servers []db.Server
	if err := db.DB.Where("is_active = ?", true).Find(&servers).Error; err != nil {
		log.Printf("🔴 Ошибка получения серверов: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка при получении серверов.")
		bot.Send(msg)
		return
	}

	if len(servers) == 0 {
		msg := tgbotapi.NewMessage(chatID, "На данный момент нет доступных серверов 😞")
		bot.Send(msg)
		return
	}

	text := "Выберите локацию:"
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, server := range servers {
		btn := tgbotapi.NewInlineKeyboardButtonData(server.Name, fmt.Sprintf("select_server_%d", server.ID))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	if _, err := bot.Send(msg); err != nil {
		log.Printf("🔴 Ошибка отправки выбора сервера: %v", err)
	}
}

// handleCallback обрабатывает callback-запросы от inline-кнопок
func handleCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	data := callback.Data

	if strings.HasPrefix(data, "select_server_") {
		// Обработка выбора сервера
		serverIDStr := strings.TrimPrefix(data, "select_server_")
		serverID, err := strconv.Atoi(serverIDStr)
		if err != nil {
			log.Printf("🔴 Ошибка преобразования serverID: %v", err)
			return
		}
		sendTariffSelection(bot, callback.Message.Chat.ID, serverID)
	} else if strings.HasPrefix(data, "buy_") {
		// Обработка выбора тарифа, формат: buy_<serverID>_<месяцев>
		parts := strings.Split(data, "_")
		if len(parts) < 3 {
			log.Printf("🔴 Некорректный формат данных для покупки: %s", data)
			return
		}
		serverID, err := strconv.Atoi(parts[1])
		if err != nil {
			log.Printf("🔴 Ошибка преобразования serverID в callback: %v", err)
			return
		}
		months, err := strconv.Atoi(parts[2])
		if err != nil {
			log.Printf("🔴 Ошибка преобразования месяцев в callback: %v", err)
			return
		}
		// Вызываем функцию резервирования ключа и создания платежа
		reserveKeyAndCreatePayment(bot, callback.Message.Chat.ID, serverID, months)
	} else {
		// Неизвестный callback
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "Неизвестное действие.")
		bot.Send(msg)
	}
}

// sendTariffSelection отправляет пользователю выбор тарифов для выбранного сервера
func sendTariffSelection(bot *tgbotapi.BotAPI, chatID int64, serverID int) {
	var server db.Server
	if err := db.DB.First(&server, serverID).Error; err != nil {
		msg := tgbotapi.NewMessage(chatID, "Ошибка: сервер не найден.")
		bot.Send(msg)
		return
	}

	text := fmt.Sprintf("Вы выбрали сервер *%s*.\nВыберите тариф подписки:", server.Name)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("1 месяц - 500₽", fmt.Sprintf("buy_%d_1", serverID)),
			tgbotapi.NewInlineKeyboardButtonData("3 месяца - 1425₽ (-5%%)", fmt.Sprintf("buy_%d_3", serverID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("6 месяцев - 2700₽ (-10%%)", fmt.Sprintf("buy_%d_6", serverID)),
			tgbotapi.NewInlineKeyboardButtonData("12 месяцев - 5100₽ (-15%%)", fmt.Sprintf("buy_%d_12", serverID)),
		),
	)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	if _, err := bot.Send(msg); err != nil {
		log.Printf("🔴 Ошибка отправки выбора тарифа: %v", err)
	}
}
