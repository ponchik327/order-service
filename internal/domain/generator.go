package domain

import (
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/google/uuid"
)

const randChars = "abcdefghijklmnopqrstuvwxyz0123456789"

// appendRandomString дописывает в b префикс и length случайных символов из randChars.
// math/rand/v2 использует per-P состояние и не блокируется на глобальном мьютексе.
func appendRandomString(b []byte, prefix string, length int) []byte {
	b = append(b, prefix...)
	for i := 0; i < length; i++ {
		b = append(b, randChars[rand.IntN(len(randChars))])
	}
	return b
}

func randomString(prefix string, length int) string {
	b := make([]byte, 0, len(prefix)+length)
	b = appendRandomString(b, prefix, length)
	return string(b)
}

// appendZeroPadded дописывает десятичное n в b с ведущими нулями до width символов.
func appendZeroPadded(b []byte, n, width int) []byte {
	var tmp [20]byte
	digits := strconv.AppendInt(tmp[:0], int64(n), 10)
	for i := len(digits); i < width; i++ {
		b = append(b, '0')
	}
	return append(b, digits...)
}

func randomPhone() string {
	b := make([]byte, 0, 11)
	b = append(b, "+972"...)
	b = appendZeroPadded(b, rand.IntN(10000000), 7)
	return string(b)
}

func randomZip() string {
	b := make([]byte, 0, 6)
	b = appendZeroPadded(b, rand.IntN(1000000), 6)
	return string(b)
}

func randomEmail() string {
	const suffix = "@example.com"
	b := make([]byte, 0, 4+6+len(suffix))
	b = appendRandomString(b, "user", 6)
	b = append(b, suffix...)
	return string(b)
}

func GenerateRandomOrder() Order {
	uid := uuid.New().String()
	trackNumber := randomString("WBIL", 10)
	price := rand.IntN(1000) + 100
	sale := rand.IntN(50)
	totalPrice := price * (100 - sale) / 100

	return Order{
		OrderUID:    uid,
		TrackNumber: trackNumber,
		Entry:       "WBIL",
		Delivery: Delivery{
			Name:    randomString("User", 6),
			Phone:   randomPhone(),
			Zip:     randomZip(),
			City:    randomString("City", 5),
			Address: randomString("Street", 8),
			Region:  randomString("Region", 4),
			Email:   randomEmail(),
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
			ChrtID:      rand.IntN(10000000),
			TrackNumber: trackNumber,
			Price:       price,
			Rid:         randomString("rid", 12),
			Name:        randomString("Product", 6),
			Sale:        sale,
			Size:        "0",
			TotalPrice:  totalPrice,
			NmID:        rand.IntN(10000000),
			Brand:       randomString("Brand", 5),
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
