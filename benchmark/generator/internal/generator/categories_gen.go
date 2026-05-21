package generator

import (
	"benchmark/generator/internal/model"
	"benchmark/generator/internal/writer"
	"fmt"
	"math/rand"
	"time"
)

func (g *Generator) GenerateCategories() error {
	fmt.Println("▶ Generating CATEGORIES...")

	start := time.Now()

	categoriesFile, err := writer.NewCSVWriter(g.writer.OutputDir, "categories.csv")
	if err != nil {
		return err
	}
	defer categoriesFile.Close()

	batch := make([][]string, 0, g.cfg.BatchSize)

	g.categories = make([]model.Category, 0, g.cfg.Categories)

	// root categories (level 0)
	rootCount := g.cfg.Categories / 10
	if rootCount < 1 {
		rootCount = 1
	}

	for i := 0; i < int(rootCount); i++ {
		cat := model.Category{
			ID:        int64(i + 1),
			ParentID:  nil,
			Name:      fmt.Sprintf("category_root_%d", i),
			CreatedAt: time.Now(),
		}

		g.categories = append(g.categories, cat)

		batch = append(batch, []string{
			fmt.Sprint(cat.ID),
			"\\N", // NULL for parent_id
			cat.Name,
			cat.CreatedAt.Format(time.RFC3339),
		})

		if len(batch) >= g.cfg.BatchSize {
			if err := categoriesFile.WriteBatch(batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	// child categories (level 1+)
	for i := int(rootCount); i < g.cfg.Categories; i++ {
		// assign random parent from existing categories
		parentID := int64(rand.Intn(i) + 1)

		cat := model.Category{
			ID:        int64(i + 1),
			ParentID:  &parentID,
			Name:      fmt.Sprintf("category_%d", i),
			CreatedAt: time.Now(),
		}

		g.categories = append(g.categories, cat)

		batch = append(batch, []string{
			fmt.Sprint(cat.ID),
			fmt.Sprint(*cat.ParentID),
			cat.Name,
			cat.CreatedAt.Format(time.RFC3339),
		})

		if len(batch) >= g.cfg.BatchSize {
			if err := categoriesFile.WriteBatch(batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	// final flush
	if len(batch) > 0 {
		if err := categoriesFile.WriteBatch(batch); err != nil {
			return err
		}
	}

	categoriesFile.Flush()

	fmt.Printf("✅ CATEGORIES done in %s (%d rows)\n",
		time.Since(start), g.cfg.Categories)

	return nil
}
