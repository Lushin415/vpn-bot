package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"vpn-bot/internal/db"
)

// YooKassaStatusResponse представляет ответ API Юкассы при запросе статуса платежа.
type YooKassaStatusResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// getYooKassaPaymentStatus запрашивает статус платежа у Юкассы по его ID.
func getYooKassaPaymentStatus(paymentID string) (string, error) {
	url := fmt.Sprintf("https://api.yookassa.ru/v3/payments/%s", paymentID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// 🔴 ! Убедитесь, что переменные YOOKASSA_SHOP_ID и YOOKASSA_SECRET_KEY заданы в .env.
	req.SetBasicAuth(os.Getenv("YOOKASSA_SHOP_ID"), os.Getenv("YOOKASSA_SECRET_KEY"))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка отправки запроса: %v", err)
	}
	defer resp.Body.Close()

	var statusResp YooKassaStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return "", fmt.Errorf("ошибка декодирования ответа: %v", err)
	}

	return statusResp.Status, nil
}

// CheckPendingPayments ищет платежи со статусом "pending", которые ожидаются более 3 минут,
// запрашивает их актуальный статус у Юкассы и обновляет БД.
// При успешном платеже активирует VLESS-ключ, при неуспешном снимает резервирование.
func CheckPendingPayments() {
	var payments []db.Payment
	threshold := time.Now().Add(-3 * time.Minute)

	if err := db.DB.Where("status = ? AND created_at < ?", "pending", threshold).Find(&payments).Error; err != nil {
		log.Printf("🔴 Ошибка выборки зависших платежей: %v", err)
		return
	}

	for _, payment := range payments {
		status, err := getYooKassaPaymentStatus(payment.YooKassaID)
		if err != nil {
			log.Printf("🔴 Ошибка проверки статуса платежа %s: %v", payment.YooKassaID, err)
			continue
		}

		// Обновляем статус платежа в БД
		if err := db.DB.Model(&payment).Update("status", status).Error; err != nil {
			log.Printf("🔴 Ошибка обновления статуса платежа %s: %v", payment.YooKassaID, err)
			continue
		}

		// Если платеж успешен – активируем ключ, иначе снимаем резервирование
		if status == "succeeded" {
			activateVLESSKey(payment)
		} else {
			releaseReservedKey(payment.UserID)
		}
	}
}

// activateVLESSKey активирует VLESS-ключ для платежа, если оплата прошла успешно.
// 🔴 Данный код дублирует логику из webhook'а – убедитесь, что он синхронизирован!
func activateVLESSKey(payment db.Payment) {
	var key db.VLESSKey
	err := db.DB.Where("user_id = ? AND is_used = false", payment.UserID).First(&key).Error
	if err != nil {
		log.Printf("🔴 Резервированный ключ для пользователя %d не найден: %v", payment.UserID, err)
		return
	}

	now := time.Now()
	if err := db.DB.Model(&key).Updates(db.VLESSKey{
		IsUsed:     true,
		AssignedAt: &now,
	}).Error; err != nil {
		log.Printf("🔴 Ошибка активации ключа: %v", err)
		return
	}

	// Отправляем уведомление пользователю (реализуйте SendMessage по своему усмотрению)
	SendMessage(int64(payment.UserID), "✅ Оплата прошла успешно! Ваш VLESS-ключ активирован.")
}

// releaseReservedKey снимает резервирование ключа, если оплата не прошла.
func releaseReservedKey(userID int) {
	if err := db.DB.Model(&db.VLESSKey{}).
		Where("user_id = ? AND is_used = false", userID).
		Updates(map[string]interface{}{
			"reserved_until": nil,
			"user_id":        nil,
		}).Error; err != nil {
		log.Printf("🔴 Ошибка снятия резервирования ключа для пользователя %d: %v", userID, err)
		return
	}

	// Отправляем уведомление пользователю (реализуйте SendMessage по своему усмотрению)
	SendMessage(int64(userID), "❌ Оплата не прошла или была отменена. Резервирование ключа снято.")
}

// SendMessage отправляет сообщение пользователю через Telegram.
// 🔴 ! Необходимо реализовать реальную отправку сообщений через Telegram Bot API.
func SendMessage(chatID int64, text string) {
	// Для демонстрации просто логируем сообщение.
	log.Printf("Отправка сообщения пользователю %d: %s", chatID, text)
}
