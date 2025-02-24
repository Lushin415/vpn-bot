#!/bin/bash

# Убедитесь, что этот скрипт имеет права на выполнение: chmod +x cron_backup.sh

# Формируем дату для имени файла
DATE=$(date +"%Y-%m-%d_%H-%M-%S")

# 🔴 ! Задайте путь к каталогу для резервных копий
BACKUP_DIR="/path/to/backups"

# 🔴 ! Задайте имя базы данных, пользователя и пароль
DB_NAME="your_db_name"
DB_USER="your_db_user"
DB_PASSWORD="your_db_password"
DB_HOST="localhost"  # При необходимости измените хост

# Экспортируем переменную окружения для pg_dump
export PGPASSWORD="$DB_PASSWORD"

# Создаем каталог для бэкапов, если он не существует
mkdir -p "$BACKUP_DIR"

# Создаем резервную копию базы данных
pg_dump -h "$DB_HOST" -U "$DB_USER" "$DB_NAME" > "$BACKUP_DIR/db_backup_$DATE.sql"

# Опционально: удаляем резервные копии старше 7 дней
find "$BACKUP_DIR" -type f -name "*.sql" -mtime +7 -exec rm {} \;

echo "✅ Резервная копия создана: $BACKUP_DIR/db_backup_$DATE.sql"
