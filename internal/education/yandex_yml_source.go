package education

import (
	"context"
	"database/sql"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"time"

	"ragbot/internal/util"
)

const source = "yandex.yml"

type YandexYMLSource struct {
	URL      string
	Interval time.Duration
}

func (y *YandexYMLSource) Start(ctx context.Context, db *sql.DB) {
	go y.run(ctx, db)
}

func (y *YandexYMLSource) run(ctx context.Context, db *sql.DB) {
	defer util.Recover("YandexYMLSource.run")
	y.process(db)
	ticker := time.NewTicker(y.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			y.process(db)
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

func (y *YandexYMLSource) process(db *sql.DB) {
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
		var id int
		var createdAt time.Time

		err := db.QueryRowContext(context.Background(),
			"SELECT id, created_at FROM chunks WHERE source=$1 AND ext_id=$2",
			source, extID).Scan(&id, &createdAt)
		if err == sql.ErrNoRows {
			_, err = db.ExecContext(context.Background(),
				"INSERT INTO chunks(content, source, ext_id, created_at) VALUES($1,$2,$3,$4)",
				content, source, extID, pubDate)
			if err != nil {
				log.Printf("yandex.yml insert error: %v", err)
			} else {
				log.Printf("Chunk added from yandex.yml: %s", extID)
			}
			continue
		}
		if err != nil {
			log.Printf("yandex.yml select error: %v", err)
			continue
		}
		if pubDate.After(createdAt) {
			_, err = db.ExecContext(context.Background(),
				"UPDATE chunks SET content=$1, created_at=$2, embedding=NULL, processed_at=NULL WHERE id=$3",
				content, pubDate, id)
			if err != nil {
				log.Printf("yandex.yml update error: %v", err)
			} else {
				log.Printf("Chunk updated from yandex.yml: %s", extID)
			}
		}
	}
}
