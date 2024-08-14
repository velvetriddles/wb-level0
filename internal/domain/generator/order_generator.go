package generator

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/velvetriddles/wb-level0/internal/domain"
)

func GenerateRandomOrder() domain.Order {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	order := domain.Order{
		OrderUID:          faker.UUIDDigit(),
		TrackNumber:       faker.Word(),
		Entry:             faker.Word(),
		Locale:            "en",
		InternalSignature: faker.Word(),
		CustomerID:        faker.Username(),
		DeliveryService:   faker.Word(),
		Shardkey:          strconv.Itoa(r.Intn(10)),
		SmID:              r.Intn(100),
		DateCreated:       time.Now(),
		OofShard:          strconv.Itoa(r.Intn(10)),
		Delivery: domain.Delivery{
			Name:    faker.Name(),
			Phone:   faker.Phonenumber(),
			Zip:     strconv.Itoa(r.Intn(100000) + 10000),
			City:    faker.Word(),
			Address: faker.Word() + ", " + strconv.Itoa(r.Intn(100)+1) + " street",
			Region:  faker.Word(),
			Email:   faker.Email(),
		},
		Payment: domain.Payment{
			Transaction:  faker.UUIDDigit(),
			RequestID:    faker.UUIDDigit(),
			Currency:     faker.Currency(),
			Provider:     faker.Word(),
			Amount:       r.Intn(10000),
			PaymentDt:    r.Intn(1000000),
			Bank:         faker.Word(),
			DeliveryCost: r.Intn(2000),
			GoodsTotal:   r.Intn(1000),
			CustomFee:    r.Intn(100),
		},
	}

	itemsCount := r.Intn(3) + 1
	order.Items = make([]domain.Item, itemsCount)
	for i := 0; i < itemsCount; i++ {
		order.Items[i] = domain.Item{
			ChrtID:      r.Intn(10000000),
			TrackNumber: faker.Word(),
			Price:       r.Intn(500),
			RID:         faker.UUIDDigit(),
			Name:        faker.Word(),
			Sale:        r.Intn(100),
			Size:        strconv.Itoa(r.Intn(5)),
			TotalPrice:  r.Intn(1000),
			NmID:        r.Intn(1000000),
			Brand:       faker.Word(),
			Status:      r.Intn(500),
		}
	}

	return order
}
