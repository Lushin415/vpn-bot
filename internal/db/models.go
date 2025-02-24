package db

import (
	"time"
)

// User представляет пользователя бота.
type User struct {
	ID              int       `gorm:"primaryKey"`
	TelegramID      int64     `gorm:"uniqueIndex"` // Telegram ID пользователя
	CurrentDiscount int       `gorm:"default:0"`   // Текущая скидка в процентах
	CreatedAt       time.Time // Дата создания записи
	UpdatedAt       time.Time // Дата обновления записи
}

// Server представляет сервер (локацию) для VPN-подписок.
type Server struct {
	ID        int     `gorm:"primaryKey"`
	Name      string  `gorm:"unique;not null"` // Название сервера
	IP        string  `gorm:"unique;not null"` // IP-адрес сервера
	Price1    float64 // Цена за 1 месяц
	Price3    float64 // Цена за 3 месяца
	Price6    float64 // Цена за 6 месяцев
	Price12   float64 // Цена за 12 месяцев
	IsActive  bool    `gorm:"default:true"` // Статус активности сервера
	CreatedAt time.Time
	UpdatedAt time.Time
}

// VLESSKey представляет VLESS-ключ, привязанный к серверу.
type VLESSKey struct {
	ID            int        `gorm:"primaryKey"`
	ServerID      int        `gorm:"index;not null"`  // ID сервера
	Key           string     `gorm:"unique;not null"` // VLESS-ключ в формате vless://...
	IsUsed        bool       `gorm:"default:false"`   // Флаг использования ключа
	ReservedUntil *time.Time // Время, до которого ключ зарезервирован
	UserID        *int       `gorm:"index"` // ID пользователя, если ключ закреплен
	AssignedAt    *time.Time // Время закрепления ключа за пользователем
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Payment представляет платеж, произведенный пользователем через Юкассу.
type Payment struct {
	ID         int     `gorm:"primaryKey"`
	UserID     int     `gorm:"index;not null"`       // ID пользователя, совершившего платеж
	YooKassaID string  `gorm:"uniqueIndex;not null"` // Идентификатор платежа в Юкассе
	Amount     float64 // Сумма платежа
	Status     string  `gorm:"default:'pending'"` // Статус платежа (pending, succeeded, failed, и т.д.)
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
