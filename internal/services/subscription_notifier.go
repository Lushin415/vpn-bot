package services

import (
	"fmt"
	"log"
	"time"

	"vpn-bot/internal/db"
)

// SendSubscriptionReminders отправляет напоминания пользователям о скором окончании подписки.
// Предполагается, что срок подписки рассчитывается как AssignedAt + 30 дней.
// 🔴 ! Если у вас другая логика продления подписки, измените формулу расчета.
func SendSubscriptionReminders() {
	var keys []db.VLESSKey
	// Выбираем все активные подписки с установленным AssignedAt.
	if err := db.DB.Where("is_used = ? AND assigned_at IS NOT NULL", true).Find(&keys).Error; err != nil {
		log.Printf("🔴 Ошибка получения активных подписок: %v", err)
		return
	}

	now := time.Now()
	for _, key := range keys {
		if key.AssignedAt == nil {
			continue
		}
		// Рассчитываем дату окончания подписки (предполагаем 30-дневный срок).
		expiration := key.AssignedAt.Add(30 * 24 * time.Hour)
		daysLeft := int(expiration.Sub(now).Hours() / 24)

		// Отправляем уведомление, если осталось ровно 7 или 3 дня.
		if daysLeft == 7 || daysLeft == 3 {
			if key.UserID != nil {
				message := fmt.Sprintf("⏳ Ваша подписка истекает через %d дней. Не забудьте продлить её!", daysLeft)
				// Функция SendMessage реализована ниже как заглушка – замените на реальную отправку сообщений через Telegram Bot API.
				SendMessage(int64(*key.UserID), message)
			}
		}
	}
}

// SendMessage отправляет сообщение пользователю через Telegram.
// 🔴 ! Замените эту заглушку на реальную реализацию отправки сообщений.
/*func SendMessage(chatID int64, text string) {
	log.Printf("Отправка сообщения пользователю %d: %s", chatID, text)
}*/
