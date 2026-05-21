package generator

import (
	"benchmark/generator/internal/config"
	"benchmark/generator/internal/model"
	"benchmark/generator/internal/writer"
	"fmt"
	"math/rand"
	"time"
)

type Generator struct {
	cfg    *config.Config
	writer *writer.FileWriter

	// cached datasets (populated sequentially)
	users []model.User
}

func NewGenerator(cfg *config.Config, w *writer.FileWriter) *Generator {
	return &Generator{
		cfg:    cfg,
		writer: w,
	}
}

func (g *Generator) GenerateUsers() error {
	fmt.Println("▶ Generating USERS...")

	start := time.Now()

	usersFile, err := writer.NewCSVWriter(g.writer.OutputDir, "users.csv")
	if err != nil {
		return err
	}
	defer usersFile.Close()

	batch := make([][]string, 0, g.cfg.BatchSize)

	g.users = make([]model.User, 0, g.cfg.Users)

	for i := 0; i < g.cfg.Users; i++ {

		u := model.User{
			ID:        int64(i + 1),
			Email:     fmt.Sprintf("user_%d@mail.com", i),
			Country:   randomCountry(),
			Status:    rand.Intn(3),
			CreatedAt: randomTime(g.cfg),
		}

		g.users = append(g.users, u)

		batch = append(batch, []string{
			fmt.Sprint(u.ID),
			u.Email,
			u.Country,
			fmt.Sprint(u.Status),
			u.CreatedAt.Format(time.RFC3339),
		})

		// flush batch
		if len(batch) >= g.cfg.BatchSize {
			if err := usersFile.WriteBatch(batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	// final flush
	if len(batch) > 0 {
		if err := usersFile.WriteBatch(batch); err != nil {
			return err
		}
	}

	usersFile.Flush()

	fmt.Printf("✅ USERS done in %s (%d rows)\n",
		time.Since(start), g.cfg.Users)

	return nil
}

func randomCountry() string {
	countries := []string{
		"US", "DE", "UK", "FR", "IN", "ES", "IT", "PL", "NL", "BR",
	}

	return countries[rand.Intn(len(countries))]
}

func randomTime(cfg *config.Config) time.Time {
	now := time.Now()
	past := now.AddDate(0, 0, -cfg.TimeRangeDays)

	delta := now.Sub(past).Seconds()
	r := rand.Float64()

	sec := r * delta

	return past.Add(time.Duration(sec) * time.Second)
}
