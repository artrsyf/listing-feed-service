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

	if *seed != 42 {
		cfg.Seed = *seed
	}

	rand.Seed(cfg.Seed)

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		log.Fatalf("failed to create output dir: %v", err)
	}

	start := time.Now()
	fmt.Printf("Starting dataset generation...\n")
	fmt.Printf("   Config: users=%d, orders=%d, order_items=%d, products=%d, categories=%d\n",
		cfg.Users, cfg.Orders, cfg.OrderItems, cfg.Products, cfg.Categories)
	fmt.Printf("   Output: %s\n", *outDir)
	fmt.Printf("   Mode: %s\n\n", *mode)

	w := writer.NewFileWriter(*outDir)
	defer w.Close()

	gen := generator.NewGenerator(cfg, w)

	switch *mode {
	case "all":
		err = runAll(gen)
	case "postgres":
		err = runPostgres(gen)
	case "elastic":
		err = runElastic(gen)
	default:
		log.Fatalf("unknown mode: %s", *mode)
	}
	if err != nil {
		log.Fatalf("generation failed: %v", err)
	}

	fmt.Printf("\nDone in %s\n", time.Since(start))
}

func runAll(gen *generator.Generator) error {
	fmt.Println("Generating full dataset for PostgreSQL and Elasticsearch")

	if err := runPostgres(gen); err != nil {
		return err
	}
	return gen.GenerateElasticOrders()
}

func runPostgres(gen *generator.Generator) error {
	fmt.Println("Generating normalized PostgreSQL dataset")

	if err := gen.GenerateCategories(); err != nil {
		return err
	}
	if err := gen.GenerateUsers(); err != nil {
		return err
	}
	if err := gen.GenerateProducts(); err != nil {
		return err
	}
	if err := gen.GenerateOrders(); err != nil {
		return err
	}
	return gen.GenerateOrderItems()
}

func runElastic(gen *generator.Generator) error {
	fmt.Println("Generating denormalized Elasticsearch dataset")

	if err := gen.GenerateCategories(); err != nil {
		return err
	}
	if err := gen.GenerateUsers(); err != nil {
		return err
	}
	if err := gen.GenerateProducts(); err != nil {
		return err
	}
	if err := gen.GenerateOrders(); err != nil {
		return err
	}
	return gen.GenerateElasticOrders()
}
