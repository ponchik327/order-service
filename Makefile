run:
	go run cmd/main.go

docker-up:
	docker-compose up -d

DB_URL = "user=user password=password host=localhost port=5432 dbname=orders_db sslmode=disable"

up:
	goose -dir migrations postgres $(DB_URL) up

down:
	goose -dir migrations postgres $(DB_URL) down