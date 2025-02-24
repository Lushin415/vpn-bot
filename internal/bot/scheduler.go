package bot

import (
	"log"

	"github.com/robfig/cron/v3"
	"vpn-bot/internal/services"
)

// InitCronJobs инициализирует и запускает фоновые задачи с использованием cron.
func InitCronJobs() {
	c := cron.New()

	// 1. Проверка зависших платежей каждые 2 минуты.
	_, err := c.AddFunc("*/2 * * * *", func() {
		log.Println("🔍 Проверка зависших платежей...")
		services.CheckPendingPayments()
	})
	if err != nil {
		log.Printf("🔴 Ошибка добавления задачи проверки платежей: %v", err)
	}

	// 2. Ежедневная отправка уведомлений о подписке (например, в 10:00 утра).
	_, err = c.AddFunc("0 10 * * *", func() {
		log.Println("📢 Отправка уведомлений о подписке...")
		services.SendSubscriptionReminders()
	})
	if err != nil {
		log.Printf("🔴 Ошибка добавления задачи отправки уведомлений: %v", err)
	}

	// 3. (Опционально) Ежедневный мониторинг серверов, например, в 03:00 утра.
	// 🔴 ! Если реализована функция MonitorServers, раскомментируйте и настройте задачу.
	/*
		_, err = c.AddFunc("0 3 * * *", func() {
			log.Println("📡 Запуск мониторинга серверов...")
			services.MonitorServers()
		})
		if err != nil {
			log.Printf("🔴 Ошибка добавления задачи мониторинга серверов: %v", err)
		}
	*/

	c.Start()
	log.Println("✅ Cron задачи успешно запущены!")
}
