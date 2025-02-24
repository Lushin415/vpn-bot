package bot

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"gorm.io/gorm"
	"vpn-bot/internal/db"
	"vpn-bot/internal/services"
)

// YooKassaWebhook представляет структуру уведомления от Юкассы.
type YooKassaWebhook struct {
	Object struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	} `json:"object"`
}

// StartWebhook запускает HTTP-сервер для обработки веб-хуков от Юкассы.
// 🔴 ! Убедитесь, что порт 8080 не занят другим сервисом.
func StartWebhook() {
	http.HandleFunc("/yookassa-webhook", handleYooKassaWebhook)
	log.Println("✅ Веб-хук Юкассы запущен на порту :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("🔴 Ошибка запуска веб-сервера: %v", err)
	}
}

// handleYooKassaWebhook обрабатывает POST-запросы от Юкассы.
func handleYooKassaWebhook(w http.ResponseWriter, r *http.Request) {
	var webhook YooKassaWebhook
	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		log.Printf("🔴 Ошибка декодирования веб-хука: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	paymentID := webhook.Object.ID
	status := webhook.Object.Status
	log.Printf("Получен веб-хук: PaymentID=%s, статус=%s", paymentID, status)

	// Находим платеж в БД по YooKassaID.
	var payment db.Payment
	if err := db.DB.Where("yoo_kassa_id = ?", paymentID).First(&payment).Error; err != nil {
		log.Printf("🔴 Платеж с ID %s не найден: %v", paymentID, err)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Обновляем статус платежа в БД.
	if err := db.DB.Model(&payment).Update("status", status).Error; err != nil {
		log.Printf("🔴 Ошибка обновления статуса платежа: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Если платеж успешен, активируем VLESS-ключ; иначе снимаем резервирование.
	if status == "succeeded" {
		activateVLESSKey(payment)
	} else {
		releaseReservedKey(payment.UserID)
	}

	w.WriteHeader(http.StatusOK)
}

// activateVLESSKey активирует VLESS-ключ после успешного платежа.
func activateVLESSKey(payment db.Payment) {
	var key db.VLESSKey
	// Ищем зарезервированный ключ, закрепленный за пользователем.
	err := db.DB.Where("user_id = ? AND is_used = false", payment.UserID).First(&key).Error
	if err == gorm.ErrRecordNotFound {
		log.Printf("🔴 Резервированный ключ для пользователя %d не найден", payment.UserID)
		return
	} else if err != nil {
		log.Printf("🔴 Ошибка при поиске ключа: %v", err)
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

	// Отправляем пользователю уведомление о успешной активации.
	services.SendMessage(int64(payment.UserID), "✅ Оплата прошла успешно! Ваш VLESS-ключ активирован.")
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

	services.SendMessage(int64(userID), "❌ Оплата не прошла или была отменена. Резервирование ключа снято.")
}
