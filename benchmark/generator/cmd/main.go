package main

import (
	"flag"
	"fmt"
	"log"
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

	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		log.Fatalf("failed to create output dir: %v", err)
	}

	start := time.Now()
	fmt.Printf("🚀 Starting dataset generation...\n")

	w := writer.NewFileWriter(*outDir)

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

	fmt.Printf("✅ Done in %s\n", time.Since(start))
}

func runAll(gen *generator.Generator) {
	fmt.Println("▶ Generating FULL dataset")

	gen.GenerateCategories()
	gen.GenerateProducts()
	gen.GenerateUsers()
	gen.GenerateOrders()
	gen.GenerateOrderItems()
}

func runPostgres(gen *generator.Generator) {
	fmt.Println("▶ Generating POSTGRES dataset")

	gen.GenerateCategories()
	gen.GenerateProducts()
	gen.GenerateUsers()
	gen.GenerateOrders()
	gen.GenerateOrderItems()
}

func runElastic(gen *generator.Generator) {
	fmt.Println("▶ Generating ELASTIC dataset")

	// Elasticsearch uses denormalized docs
	gen.GenerateElasticOrders()
}
