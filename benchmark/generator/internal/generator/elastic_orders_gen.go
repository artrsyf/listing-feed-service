package generator

import (
	"benchmark/generator/internal/writer"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

func (g *Generator) GenerateElasticOrders() error {
	fmt.Println("▶ Generating ELASTIC ORDERS (denormalized docs)...")

	start := time.Now()

	nd, err := writer.NewNDJSONWriter(g.writer.OutputDir, "orders.ndjson")
	if err != nil {
		return err
	}
	defer nd.Close()

	docCount := 0

	for _, order := range g.orders {

		user := g.users[int(order.UserID)-1]

		items := g.buildElasticItems(order.ID)

		total := computeTotal(items)

		doc := map[string]interface{}{
			"order_id": order.ID,

			"user": map[string]interface{}{
				"id":      user.ID,
				"country": user.Country,
				"status":  user.Status,
			},

			"created_at": order.CreatedAt,

			"items": items,

			"total_amount": total,
		}

		b, err := json.Marshal(doc)
		if err != nil {
			return err
		}

		if err := nd.WriteBulkItem("index", string(b)); err != nil {
			return err
		}

		docCount++

		// Progress indicator
		if docCount%100000 == 0 {
			fmt.Printf("  Generated %d documents...\n", docCount)
		}
	}

	nd.Flush()

	fmt.Printf("✅ ELASTIC ORDERS done in %s (%d docs)\n",
		time.Since(start), docCount)

	return nil
}

func (g *Generator) buildElasticItems(orderID int64) []map[string]interface{} {

	count := sampleItemsPerOrder()

	items := make([]map[string]interface{}, 0, count)

	for i := 0; i < count; i++ {

		product := g.products[randomIndex(len(g.products))]
		categoryID := product.CategoryID

		items = append(items, map[string]interface{}{
			"product_id":  product.ID,
			"category_id": categoryID,
			"price":       product.Price,
			"quantity":    rand.Intn(5) + 1,
		})
	}

	return items
}

func computeTotal(items []map[string]interface{}) float64 {
	total := 0.0

	for _, i := range items {
		price := i["price"].(float64)
		qty := float64(i["quantity"].(int))

		total += price * qty
	}

	return total
}

func randomIndex(n int) int {
	return rand.Intn(n)
}
