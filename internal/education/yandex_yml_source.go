package education

import (
	"context"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"time"

	"ragbot/internal/repository"
	"ragbot/internal/util"
)

const yandexSource = "yandex.yml"

type YandexYMLSource struct {
	URL      string
	Interval time.Duration
}

func (y *YandexYMLSource) Start(ctx context.Context, repo *repository.Repository) {
	go y.run(ctx, repo)
}

func (y *YandexYMLSource) run(ctx context.Context, repo *repository.Repository) {
	defer util.Recover("YandexYMLSource.run")
	y.process(repo)
	ticker := time.NewTicker(y.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			y.process(repo)
		}
	}
}

type ymlCatalog struct {
	Date string  `xml:"date,attr"`
	Shop ymlShop `xml:"shop"`
}

type ymlShop struct {
	Offers []ymlOffer `xml:"offers>offer"`
}

type ymlOffer struct {
	ID          string `xml:"id,attr"`
	CategoryID  string `xml:"categoryId"`
	Name        string `xml:"name"`
	Description string `xml:"description"`
}

func (y *YandexYMLSource) process(repo *repository.Repository) {
	defer util.Recover("YandexYMLSource.process")
	resp, err := http.Get(y.URL)
	if err != nil {
		log.Printf("yandex.yml fetch error: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("yandex.yml read error: %v", err)
		return
	}

	var catalog ymlCatalog
	if err := xml.Unmarshal(body, &catalog); err != nil {
		log.Printf("yandex.yml parse error: %v", err)
		return
	}

	pubDate, err := time.Parse("2006-01-02T15:04-07:00", catalog.Date)
	if err != nil {
		log.Printf("yandex.yml date parse error: %v", err)
		return
	}

	for _, offer := range catalog.Shop.Offers {
		if offer.CategoryID != "1" && offer.CategoryID != "2" {
			continue
		}

		var content string
		switch offer.CategoryID {
		case "1":
			content = "Класс " + offer.Name + ": " + offer.Description
		case "2":
			content = "Абонемент " + offer.Name + ": " + offer.Description
		}

		extID := offer.CategoryID + ":" + offer.ID
		id, createdAt, oldContent, found, err := repo.GetChunkByExtID(context.Background(), yandexSource, extID)
		if err != nil {
			log.Printf("yandex.yml select error: %v", err)
			continue
		}
		if !found {
			err = repo.InsertChunkWithExtID(context.Background(), content, yandexSource, extID, pubDate)
			if err != nil {
				log.Printf("yandex.yml insert error: %v", err)
			} else {
				log.Printf("Chunk added from yandex.yml: %s", extID)
			}
			continue
		}
		if pubDate.After(createdAt) {
			if oldContent == content {
				err = repo.UpdateChunkCreatedAt(context.Background(), id, pubDate)
				if err != nil {
					log.Printf("yandex.yml date update error: %v", err)
				} else {
					log.Printf("Chunk date updated from yandex.yml: %s", extID)
				}
			} else {
				err = repo.UpdateChunkWithCreatedAt(context.Background(), id, content, pubDate)
				if err != nil {
					log.Printf("yandex.yml update error: %v", err)
				} else {
					log.Printf("Chunk updated from yandex.yml: %s", extID)
				}
			}
		}
	}
}
