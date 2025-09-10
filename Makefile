run:
	go run cmd/main.go

docker-up:
	docker-compose up -d

clear-volumes:
	docker volume rm order-service_postgres_data order-service_redis_data order-service_kafka_data	

refresh-app:
	docker-compose build --no-cache app
