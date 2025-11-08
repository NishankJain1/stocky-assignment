package api

import (
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
	"stocky-assignment/db"
)

// mockPrice generates a pseudo-random stock price between 900 and 2500.
func mockPrice() float64 {
	return 900 + rand.Float64()*1600
}

// StartPriceUpdater launches a background process that refreshes stock prices every hour.
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

// updateAllPrices retrieves all stock symbols from rewards
// and updates or inserts their latest mock prices into stock_prices.
func updateAllPrices() {
	rows, err := db.DB.Query("SELECT DISTINCT stock_symbol FROM rewards;")
	if err != nil {
		logrus.Errorf("Error retrieving stock symbols: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			logrus.Errorf("Error scanning stock symbol: %v", err)
			continue
		}

		price := mockPrice()
		_, err := db.DB.Exec(`
			INSERT INTO stock_prices (stock_symbol, price, updated_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (stock_symbol)
			DO UPDATE SET 
				price = EXCLUDED.price, 
				updated_at = NOW();`,
			symbol, price,
		)
		if err != nil {
			logrus.Errorf("Error updating price for %s: %v", symbol, err)
			continue
		}
		logrus.Infof("Stock %s price updated to %.2f", symbol, price)
	}
}
