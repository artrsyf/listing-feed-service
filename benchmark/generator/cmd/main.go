package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"benchmark/generator/internal/config"
	"benchmark/generator/internal/generator"
	"benchmark/generator/internal/writer"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	outDir := flag.String("out", "./output", "output directory")
	mode := flag.String("mode", "all", "generation mode: all|postgres|elastic")
	seed := flag.Int64("seed", 42, "random seed for reproducibility")

	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Override seed if provided via flag
	if *seed != 42 {
		cfg.Seed = *seed
	}

	rand.Seed(cfg.Seed)

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		log.Fatalf("failed to create output dir: %v", err)
	}

	start := time.Now()
	fmt.Printf("🚀 Starting dataset generation...\n")
	fmt.Printf("   Config: users=%d, orders=%d, order_items=%d, products=%d, categories=%d\n",
		cfg.Users, cfg.Orders, cfg.OrderItems, cfg.Products, cfg.Categories)
	fmt.Printf("   Output: %s\n", *outDir)
	fmt.Printf("   Mode: %s\n", *mode)
	fmt.Println()

	w := writer.NewFileWriter(*outDir)
	defer w.Close()

	gen := generator.NewGenerator(cfg, w)

	switch *mode {
	case "all":
		runAll(gen)

	case "postgres":
		runPostgres(gen)

	case "elastic":
		runElastic(gen)

	default:
		log.Fatalf("unknown mode: %s", *mode)
	}

	fmt.Printf("\n✅ Done in %s\n", time.Since(start))
}

func runAll(gen *generator.Generator) {
	fmt.Println("▶ Generating FULL dataset (PostgreSQL + Elasticsearch)")

	gen.GenerateCategories()
	gen.GenerateUsers()
	gen.GenerateProducts()
	gen.GenerateOrders()
	gen.GenerateOrderItems()
	gen.GenerateElasticOrders()
}

func runPostgres(gen *generator.Generator) {
	fmt.Println("▶ Generating POSTGRES dataset")

	gen.GenerateCategories()
	gen.GenerateUsers()
	gen.GenerateProducts()
	gen.GenerateOrders()
	gen.GenerateOrderItems()
}

func runElastic(gen *generator.Generator) {
	fmt.Println("▶ Generating ELASTIC dataset (denormalized)")

	// For elastic-only mode, we need to generate all referenced data first
	// but only output denormalized documents
	gen.GenerateCategories()
	gen.GenerateUsers()
	gen.GenerateProducts()
	gen.GenerateOrders()
	gen.GenerateElasticOrders()
}
