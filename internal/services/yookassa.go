package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// YooKassaPaymentRequest структура запроса на создание платежа в Юкассе
type YooKassaPaymentRequest struct {
	Amount   YooKassaAmount   `json:"amount"`
	Capture  bool             `json:"capture"`
	Payment  YooKassaPayment  `json:"payment_method_data"`
	Metadata YooKassaMetadata `json:"metadata"`
}

// YooKassaAmount структура суммы платежа
type YooKassaAmount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

// YooKassaPayment структура способа оплаты
type YooKassaPayment struct {
	Type string `json:"type"`
}

// YooKassaMetadata дополнительные метаданные (например, ID пользователя)
type YooKassaMetadata struct {
	UserID int `json:"user_id"`
}

// YooKassaResponse структура ответа от Юкассы
type YooKassaResponse struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Confirm struct {
		ConfirmationURL string `json:"confirmation_url"`
	} `json:"confirmation"`
}

// CreateYooKassaPayment создаёт платёж через Юкассу
func CreateYooKassaPayment(userID int64, amount float64) (string, string, error) {
	url := "https://api.yookassa.ru/v3/payments"

	// Форматируем сумму в строку с двумя знаками после запятой
	amountStr := fmt.Sprintf("%.2f", amount)

	// Формируем JSON-запрос
	requestBody := YooKassaPaymentRequest{
		Amount: YooKassaAmount{
			Value:    amountStr,
			Currency: "RUB",
		},
		Capture: true,
		Payment: YooKassaPayment{
			Type: "bank_card",
		},
		Metadata: YooKassaMetadata{
			UserID: int(userID),
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", "", fmt.Errorf("ошибка кодирования JSON: %v", err)
	}

	// Подготавливаем HTTP-запрос
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", "", fmt.Errorf("ошибка создания HTTP-запроса: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotence-Key", fmt.Sprintf("%d", time.Now().Unix())) // Уникальный ключ для предотвращения дублирования

	// 🔴 ! Вставьте свои ключи из .env
	req.SetBasicAuth(os.Getenv("YOOKASSA_SHOP_ID"), os.Getenv("YOOKASSA_SECRET_KEY"))

	// Отправляем запрос
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("ошибка отправки запроса: %v", err)
	}
	defer resp.Body.Close()

	// Декодируем ответ
	var yooResp YooKassaResponse
	if err := json.NewDecoder(resp.Body).Decode(&yooResp); err != nil {
		return "", "", fmt.Errorf("ошибка декодирования ответа Юкассы: %v", err)
	}

	if yooResp.ID == "" || yooResp.Confirm.ConfirmationURL == "" {
		return "", "", fmt.Errorf("невалидный ответ Юкассы")
	}

	// Возвращаем ID платежа и URL для оплаты
	return yooResp.ID, yooResp.Confirm.ConfirmationURL, nil
}
