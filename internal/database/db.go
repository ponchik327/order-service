package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// InitDB инициализирует подключение к PostgreSQL.
func InitDB() (*sql.DB, error) {
	// Получаем строку подключения из переменной окружения
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://user:password@localhost:5432/orders_db?sslmode=disable" // нужно вынести в конфиг .env файл
	}

	// Открываем соединение с базой данных
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	log.Println("Successfully connected to PostgreSQL")
	return db, nil
}
