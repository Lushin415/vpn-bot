package bot

import (
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"vpn-bot/internal/db"
	"vpn-bot/internal/services"
)

// reserveKeyAndCreatePayment резервирует VLESS-ключ и инициирует создание платежа через Юкассу.
func reserveKeyAndCreatePayment(bot *tgbotapi.BotAPI, chatID int64, serverID, months int) {
	// Поиск свободного ключа, который не используется и не зарезервирован.
	var key db.VLESSKey
	err := db.DB.
		Where("server_id = ? AND is_used = false AND (reserved_until IS NULL OR reserved_until < NOW())", serverID).
		First(&key).Error
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "К сожалению, на данном сервере нет доступных ключей 😞")
		bot.Send(msg)
		return
	}

	// Резервирование ключа на 5 минут
	reservedUntil := time.Now().Add(5 * time.Minute)
	if err := db.DB.Model(&key).Update("reserved_until", reservedUntil).Error; err != nil {
		log.Printf("🔴 Ошибка резервирования ключа: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка при резервировании ключа. Попробуйте позже.")
		bot.Send(msg)
		return
	}

	// Получаем информацию о сервере для расчета цены
	var server db.Server
	if err := db.DB.First(&server, serverID).Error; err != nil {
		msg := tgbotapi.NewMessage(chatID, "Ошибка: сервер не найден.")
		bot.Send(msg)
		return
	}

	// Расчет стоимости подписки
	// 🔴 ! Убедитесь, что в БД для сервера Price1 задан базовый тариф (например, 500₽)
	price := server.Price1 * float64(months)
	if months == 3 {
		price *= 0.95
	} else if months == 6 {
		price *= 0.90
	} else if months == 12 {
		price *= 0.85
	}

	// Создаем платеж через Юкассу
	paymentID, paymentURL, err := services.CreateYooKassaPayment(chatID, price)
	if err != nil {
		log.Printf("🔴 Ошибка создания платежа: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка при создании платежа. Попробуйте позже.")
		bot.Send(msg)
		return
	}

	// Записываем платеж в БД
	payment := db.Payment{
		UserID:     int(chatID), // 🔴 ! Убедитесь, что Telegram ID корректно конвертируется в int
		YooKassaID: paymentID,
		Amount:     price,
		Status:     "pending",
	}
	if err := db.DB.Create(&payment).Error; err != nil {
		log.Printf("🔴 Ошибка записи платежа в БД: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка при записи платежа. Попробуйте позже.")
		bot.Send(msg)
		return
	}

	// Информируем пользователя
	text := fmt.Sprintf("✅ Ваш VLESS-ключ зарезервирован!\n💰 Сумма: %.2f₽\n\nПерейдите по ссылке для оплаты:\n%s", price, paymentURL)
	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}
