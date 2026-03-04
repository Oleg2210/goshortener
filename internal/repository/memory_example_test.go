package repository_test

import (
	"context"
	"fmt"
	"log"

	"github.com/Oleg2210/goshortener/internal/entities"
	"github.com/Oleg2210/goshortener/internal/repository"
)

func ExampleMemoryRepository_Save() {
	ctx := context.Background()
	repo := repository.NewMemoryRepository()

	id := "abc123"
	url := "https://example.com"
	userID := "user1"

	_, err := repo.Save(ctx, id, url, userID, false)
	if err != nil {
		log.Fatal(err)
	}

	record, exists := repo.Get(ctx, id)
	if !exists {
		log.Fatal("record not found")
	}

	fmt.Println(record.OriginalURL, record.Short)
	// Output:
	// https://example.com abc123
}

func ExampleMemoryRepository_BatchSave() {
	ctx := context.Background()
	repo := repository.NewMemoryRepository()

	records := []entities.URLRecord{
		{OriginalURL: "https://example.com/1", Short: "id1"},
		{OriginalURL: "https://example.com/2", Short: "id2"},
	}

	userID := "user2"
	err := repo.BatchSave(ctx, records, userID)
	if err != nil {
		log.Fatal(err)
	}

	for _, r := range records {
		rec, exists := repo.Get(ctx, r.Short)
		if !exists {
			log.Fatal("record not found")
		}
		fmt.Println(rec.OriginalURL, rec.Short)
	}
	// Output:
	// https://example.com/1 id1
	// https://example.com/2 id2
}
