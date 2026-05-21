package generator

import (
	"benchmark/generator/internal/model"
	"benchmark/generator/internal/writer"
	"fmt"
	"math/rand"
	"time"
)

func (g *Generator) GenerateOrders() error {
	fmt.Println("▶ Generating ORDERS...")

	start := time.Now()

	ordersFile, err := writer.NewCSVWriter(g.writer.OutputDir, "orders.csv")
	if err != nil {
		return err
	}
	defer ordersFile.Close()

	batch := make([][]string, 0, g.cfg.BatchSize)

	g.orders = make([]model.Order, 0, g.cfg.Orders)

	for i := 0; i < g.cfg.Orders; i++ {

		user := g.users[zipfUser(len(g.users))]

		createdAt := randomTime(g.cfg)

		total := 0.0 // will be corrected later during order_items stage

		o := model.Order{
			ID:          int64(i + 1),
			UserID:      user.ID,
			Status:      rand.Intn(5),
			TotalAmount: total,
			CreatedAt:   createdAt,
		}

		g.orders = append(g.orders, o)

		batch = append(batch, []string{
			fmt.Sprint(o.ID),
			fmt.Sprint(o.UserID),
			fmt.Sprint(o.Status),
			fmt.Sprintf("%.2f", o.TotalAmount),
			o.CreatedAt.Format(time.RFC3339),
		})

		if len(batch) >= g.cfg.BatchSize {
			if err := ordersFile.WriteBatch(batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		if err := ordersFile.WriteBatch(batch); err != nil {
			return err
		}
	}

	ordersFile.Flush()

	fmt.Printf("✅ ORDERS done in %s (%d rows)\n",
		time.Since(start), g.cfg.Orders)

	return nil
}

func zipfUser(n int) int {
	// heavy-tail: few users generate most orders
	// simple approximation (we will replace later with real Zipf generator)

	r := rand.Float64()

	// power bias toward low-index users
	index := int(float64(n) * (1.0 - r*r))

	if index < 0 {
		index = 0
	}
	if index >= n {
		index = n - 1
	}

	return index
}
