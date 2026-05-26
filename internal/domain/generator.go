package domain

import (
	"math/rand/v2"
	"strconv"
	"time"
	"unsafe"

	"github.com/google/uuid"
)

const randChars = "abcdefghijklmnopqrstuvwxyz0123456789"

// takeRandom апендит в bb префикс и randLen случайных символов из randChars,
// и возвращает строку, ссылающуюся на этот участок backing array через
// unsafe.String. Все строки заказа делят один buf, благодаря чему
// GenerateRandomOrder делает одну аллокацию буфера вместо одиннадцати
// отдельных make'ов. Buf должен иметь достаточный cap, иначе append
// перевыделит массив и старые строки укажут на другой backing array.
func takeRandom(bb *[]byte, prefix string, randLen int) string {
	start := len(*bb)
	*bb = append(*bb, prefix...)
	for i := 0; i < randLen; i++ {
		*bb = append(*bb, randChars[rand.IntN(len(randChars))])
	}
	return unsafe.String(&(*bb)[start], len(*bb)-start)
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

func GenerateRandomOrder() Order {
	// Один backing array на все random-строки заказа.
	// Сумма длин: WBIL+10, User+6, +972+7, 6, City+5, Street+8, Region+4,
	// user+6+@example.com, rid+12, Product+6, Brand+5 = 134 байта. cap=160
	// с запасом — append гарантированно не вызовет growSlice.
	bb := make([]byte, 0, 160)

	trackNumber := takeRandom(&bb, "WBIL", 10)
	name := takeRandom(&bb, "User", 6)

	// Phone: "+972" + 7 zero-padded digits
	phoneStart := len(bb)
	bb = append(bb, "+972"...)
	bb = appendZeroPadded(bb, rand.IntN(10000000), 7)
	phone := unsafe.String(&bb[phoneStart], len(bb)-phoneStart)

	// Zip: 6 zero-padded digits
	zipStart := len(bb)
	bb = appendZeroPadded(bb, rand.IntN(1000000), 6)
	zip := unsafe.String(&bb[zipStart], len(bb)-zipStart)

	city := takeRandom(&bb, "City", 5)
	address := takeRandom(&bb, "Street", 8)
	region := takeRandom(&bb, "Region", 4)

	// Email: "user" + 6 random + "@example.com"
	emailStart := len(bb)
	bb = append(bb, "user"...)
	for i := 0; i < 6; i++ {
		bb = append(bb, randChars[rand.IntN(len(randChars))])
	}
	bb = append(bb, "@example.com"...)
	email := unsafe.String(&bb[emailStart], len(bb)-emailStart)

	rid := takeRandom(&bb, "rid", 12)
	productName := takeRandom(&bb, "Product", 6)
	brand := takeRandom(&bb, "Brand", 5)

	uid := uuid.New().String()
	price := rand.IntN(1000) + 100
	sale := rand.IntN(50)
	totalPrice := price * (100 - sale) / 100

	return Order{
		OrderUID:    uid,
		TrackNumber: trackNumber,
		Entry:       "WBIL",
		Delivery: Delivery{
			Name:    name,
			Phone:   phone,
			Zip:     zip,
			City:    city,
			Address: address,
			Region:  region,
			Email:   email,
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
			Rid:         rid,
			Name:        productName,
			Sale:        sale,
			Size:        "0",
			TotalPrice:  totalPrice,
			NmID:        rand.IntN(10000000),
			Brand:       brand,
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
