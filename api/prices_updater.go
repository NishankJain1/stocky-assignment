package api

import (
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
	"stocky-assignment/db"
)

// mockPrice returns a random price between 900 and 2500
func mockPrice() float64 {
	return 900 + rand.Float64()*1600
}

// StartPriceUpdater runs a background goroutine that updates prices hourly
func StartPriceUpdater() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			updateAllPrices()
			<-ticker.C
		}
	}()
}

// updateAllPrices updates stock prices for all known symbols in DB
func updateAllPrices() {
	rows, err := db.DB.Query("SELECT DISTINCT stock_symbol FROM rewards;")
	if err != nil {
		logrus.Errorf("❌ Error fetching stock symbols: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var symbol string
		rows.Scan(&symbol)

		price := mockPrice()
		_, err := db.DB.Exec(`
			INSERT INTO stock_prices (stock_symbol, price, updated_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (stock_symbol)
			DO UPDATE SET price = EXCLUDED.price, updated_at = NOW();`,
			symbol, price,
		)
		if err != nil {
			logrus.Errorf("❌ Error updating price for %s: %v", symbol, err)
			continue
		}
		logrus.Infof("💰 Updated %s price to %.2f", symbol, price)
	}
}
