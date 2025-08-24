package domain

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

func generateRandomString(prefix string, length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return prefix + string(result)
}

func GenerateRandomOrder() Order {
	uid := uuid.New().String()
	trackNumber := generateRandomString("WBIL", 10)
	price := rand.Intn(1000) + 100
	sale := rand.Intn(50)
	totalPrice := price * (100 - sale) / 100

	return Order{
		OrderUID:    uid,
		TrackNumber: trackNumber,
		Entry:       "WBIL",
		Delivery: Delivery{
			Name:    generateRandomString("User", 6),
			Phone:   fmt.Sprintf("+972%07d", rand.Intn(10000000)),
			Zip:     fmt.Sprintf("%06d", rand.Intn(1000000)),
			City:    generateRandomString("City", 5),
			Address: generateRandomString("Street", 8),
			Region:  generateRandomString("Region", 4),
			Email:   generateRandomString("user", 6) + "@example.com",
		},
		Payment: Payment{
			Transaction:  uid,
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       price + 1500,
			PaymentDt:    time.Now().Unix(),
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   totalPrice,
			CustomFee:    0,
		},
		Items: []Item{{
			ChrtID:      rand.Intn(10000000),
			TrackNumber: trackNumber,
			Price:       price,
			Rid:         generateRandomString("rid", 12),
			Name:        generateRandomString("Product", 6),
			Sale:        sale,
			Size:        "0",
			TotalPrice:  totalPrice,
			NmID:        rand.Intn(10000000),
			Brand:       generateRandomString("Brand", 5),
			Status:      202,
		}},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       time.Now(),
		OofShard:          "1",
	}
}
