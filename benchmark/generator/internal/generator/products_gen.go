package generator

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"benchmark/generator/internal/model"
	"benchmark/generator/internal/writer"
)

func (g *Generator) GenerateProducts() error {
	fmt.Println("▶ Generating PRODUCTS...")

	start := time.Now()

	productsFile, err := writer.NewCSVWriter(g.writer.OutputDir, "products.csv")
	if err != nil {
		return err
	}
	defer productsFile.Close()

	batch := make([][]string, 0, g.cfg.BatchSize)

	g.products = make([]model.Product, 0, g.cfg.Products)

	for i := 0; i < g.cfg.Products; i++ {

		categoryID := int64(rand.Intn(g.cfg.Categories) + 1)

		price := skewedPrice()
		rating := skewedRating()

		p := model.Product{
			ID:         int64(i + 1),
			CategoryID: categoryID,
			SKU:        fmt.Sprintf("sku_%d_%d", i, rand.Intn(100000)),
			Name:       fmt.Sprintf("product_%d", i),
			Price:      price,
			Rating:     rating,
			CreatedAt:  randomTime(g.cfg),
		}

		g.products = append(g.products, p)

		batch = append(batch, []string{
			fmt.Sprint(p.ID),
			fmt.Sprint(p.CategoryID),
			p.SKU,
			p.Name,
			fmt.Sprintf("%.2f", p.Price),
			fmt.Sprintf("%.2f", p.Rating),
			p.CreatedAt.Format(time.RFC3339),
		})

		if len(batch) >= g.cfg.BatchSize {
			if err := productsFile.WriteBatch(batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		if err := productsFile.WriteBatch(batch); err != nil {
			return err
		}
	}

	productsFile.Flush()

	fmt.Printf("✅ PRODUCTS done in %s (%d rows)\n",
		time.Since(start), g.cfg.Products)

	return nil
}

func skewedPrice() float64 {
	base := rand.Float64()

	// Pareto-ish skew: more cheap products, few expensive
	price := math.Pow(base, 3) * 1000

	if price < 1 {
		price = 1
	}

	return price
}

func skewedRating() float64 {
	r := rand.Float64()

	// bias towards 3–5 stars
	rating := 1 + math.Pow(r, 0.5)*4

	if rating > 5 {
		rating = 5
	}

	return rating
}
