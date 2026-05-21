package generator

import (
	"benchmark/generator/internal/model"
	"benchmark/generator/internal/writer"
	"fmt"
	"math/rand"
	"time"
)

func (g *Generator) GenerateOrderItems() error {
	fmt.Println("▶ Generating ORDER_ITEMS...")

	start := time.Now()

	itemsFile, err := writer.NewCSVWriter(g.writer.OutputDir, "order_items.csv")
	if err != nil {
		return err
	}
	defer itemsFile.Close()

	batch := make([][]string, 0, g.cfg.BatchSize)

	g.orderItems = make([]model.OrderItem, 0, g.cfg.OrderItems)

	g.orderItems = make([]model.OrderItem, 0, g.cfg.OrderItems)

	itemID := int64(1)

	// reset order totals (we recompute correctly here)
	orderTotals := make(map[int64]float64, len(g.orders))

	for _, order := range g.orders {

		itemCount := sampleItemsPerOrder()

		for i := 0; i < itemCount; i++ {

			product := g.products[zipfProduct(len(g.products))]

			price := product.Price
			qty := rand.Intn(5) + 1

			total := float64(qty) * price

			oi := model.OrderItem{
				ID:        itemID,
				OrderID:   order.ID,
				ProductID: product.ID,
				Quantity:  qty,
				Price:     price,
				CreatedAt: order.CreatedAt,
			}

			g.orderItems = append(g.orderItems, oi)

			orderTotals[order.ID] += total

			batch = append(batch, []string{
				fmt.Sprint(oi.ID),
				fmt.Sprint(oi.OrderID),
				fmt.Sprint(oi.ProductID),
				fmt.Sprint(oi.Quantity),
				fmt.Sprintf("%.2f", oi.Price),
				oi.CreatedAt.Format(time.RFC3339),
			})

			itemID++

			if len(batch) >= g.cfg.BatchSize {
				if err := itemsFile.WriteBatch(batch); err != nil {
					return err
				}
				batch = batch[:0]
			}
		}
	}

	// final flush
	if len(batch) > 0 {
		if err := itemsFile.WriteBatch(batch); err != nil {
			return err
		}
	}

	itemsFile.Flush()

	// FIX ORDER TOTALS CONSISTENCY
	g.fixOrderTotals(orderTotals)

	fmt.Printf("✅ ORDER_ITEMS done in %s (%d rows)\n",
		time.Since(start), len(g.orderItems))

	return nil
}

func (g *Generator) fixOrderTotals(totals map[int64]float64) {
	for i := range g.orders {
		g.orders[i].TotalAmount = totals[g.orders[i].ID]
	}
}

func sampleItemsPerOrder() int {
	r := rand.Float64()

	switch {
	case r < 0.35:
		return 1
	case r < 0.65:
		return 2
	case r < 0.85:
		return 3
	default:
		return rand.Intn(8) + 4
	}
}

func zipfProduct(n int) int {
	r := rand.Float64()

	// bias toward popular products
	index := int(float64(n) * (1.0 - r*r*r))

	if index < 0 {
		index = 0
	}
	if index >= n {
		index = n - 1
	}

	return index
}
