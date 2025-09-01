package database

import (
	"database/sql"
	"log"
	"order_service/internal/config"

	_ "github.com/lib/pq"
)

// Подключаем к PostgreSql
func InitDB(cfg *config.Config) (*sql.DB, error) {
	var sslMode string
	if cfg.Database.SslMode {
		sslMode = "enable"
	} else {
		sslMode = "disable"
	}

	// Достаём данные для подключения из конифга
	dsn := "postgres://" + cfg.Database.User +
		":" + cfg.Database.Password +
		"@" + cfg.Database.Adress +
		"/" + cfg.Database.Name +
		"?sslmode=" + sslMode

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
