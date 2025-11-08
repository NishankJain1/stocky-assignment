package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"stocky-assignment/db"
)

// ------------------- Structs -------------------
type RewardRequest struct {
	UserID      string  `json:"user_id" binding:"required"`
	StockSymbol string  `json:"stock_symbol" binding:"required"`
	Shares      float64 `json:"shares" binding:"required"`
}

type RewardResponse struct {
	Message  string    `json:"message"`
	RewardID int       `json:"reward_id"`
	Time     time.Time `json:"reward_time"`
}

type AdjustmentRequest struct {
	StockSymbol   string  `json:"stock_symbol" binding:"required"`
	Multiplier    float64 `json:"multiplier" binding:"required"`
	EffectiveDate string  `json:"effective_date" binding:"required"` // YYYY-MM-DD
	Delisted      bool    `json:"delisted"`                          // optional
}

// ------------------- Add Reward -------------------
func AddReward(c *gin.Context) {
	var req RewardRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.Warn("Invalid request payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	query := `
		INSERT INTO rewards (user_id, stock_symbol, shares, reward_time)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, reward_time;
	`

	var rewardID int
	var rewardTime time.Time

	err := db.DB.QueryRow(query, req.UserID, req.StockSymbol, req.Shares).Scan(&rewardID, &rewardTime)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			logrus.Warnf("Duplicate reward for user %s, stock %s", req.UserID, req.StockSymbol)
			c.JSON(http.StatusConflict, gin.H{"error": "Duplicate reward event — this reward already exists"})
			return
		}
		logrus.Errorf("DB insert error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	logrus.Infof("✅ Reward added: %+v", req)
	c.JSON(http.StatusOK, RewardResponse{
		Message:  "Reward recorded successfully",
		RewardID: rewardID,
		Time:     rewardTime,
	})
}

// ------------------- Stats (Today's Total) -------------------
func GetStats(c *gin.Context) {
	userId := c.Param("userId")

	query := `
		SELECT 
			r.stock_symbol,
			SUM(r.shares * COALESCE(sa.multiplier, 1.0)) AS total_shares
		FROM rewards r
		LEFT JOIN stock_adjustments sa 
			ON r.stock_symbol = sa.stock_symbol
			AND sa.effective_date <= CURRENT_DATE
		WHERE r.user_id = $1
		  AND DATE(r.reward_time) = CURRENT_DATE
		  AND (sa.delisted IS NULL OR sa.delisted = FALSE)
		GROUP BY r.stock_symbol;
	`

	rows, err := db.DB.Query(query, userId)
	if err != nil {
		logrus.Errorf("DB query error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	type StockStat struct {
		StockSymbol string  `json:"stock_symbol"`
		TotalShares float64 `json:"total_shares"`
	}

	var stats []StockStat
	totalValue := 0.0

	for rows.Next() {
		var s StockStat
		rows.Scan(&s.StockSymbol, &s.TotalShares)
		s.TotalShares = round(s.TotalShares, 6)
		price := 1000 + float64(len(s.StockSymbol))*100
		totalValue += price * s.TotalShares
		stats = append(stats, s)
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":             userId,
		"today_rewards":       stats,
		"portfolio_inr_value": round(totalValue, 2),
	})
}

// ------------------- Portfolio -------------------
func GetPortfolio(c *gin.Context) {
	userId := c.Param("userId")

	query := `
		SELECT 
			r.stock_symbol,
			SUM(r.shares * COALESCE(sa.multiplier, 1.0)) AS total_shares,
			COALESCE(sp.price, 1000) AS current_price
		FROM rewards r
		LEFT JOIN stock_prices sp 
			ON r.stock_symbol = sp.stock_symbol
		LEFT JOIN stock_adjustments sa 
			ON r.stock_symbol = sa.stock_symbol
			AND sa.effective_date <= CURRENT_DATE
		WHERE r.user_id = $1
		  AND (sa.delisted IS NULL OR sa.delisted = FALSE)
		GROUP BY r.stock_symbol, sp.price;
	`

	rows, err := db.DB.Query(query, userId)
	if err != nil {
		logrus.Errorf("DB query error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	type Holding struct {
		StockSymbol  string  `json:"stock_symbol"`
		TotalShares  float64 `json:"total_shares"`
		CurrentPrice float64 `json:"current_price"`
		TotalValue   float64 `json:"total_value_inr"`
	}

	var portfolio []Holding
	totalValue := 0.0

	for rows.Next() {
		var h Holding
		rows.Scan(&h.StockSymbol, &h.TotalShares, &h.CurrentPrice)
		h.TotalShares = round(h.TotalShares, 6)
		h.TotalValue = round(h.TotalShares*h.CurrentPrice, 2)
		totalValue += h.TotalValue
		portfolio = append(portfolio, h)
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":             userId,
		"portfolio":           portfolio,
		"portfolio_total_inr": round(totalValue, 2),
	})
}

// ------------------- Historical INR -------------------
func GetHistoricalINR(c *gin.Context) {
	userId := c.Param("userId")

	query := `
		SELECT 
			DATE(r.reward_time) AS date,
			SUM(r.shares * COALESCE(sa.multiplier, 1.0) * COALESCE(sp.price, 1000)) AS total_inr
		FROM rewards r
		LEFT JOIN stock_prices sp 
			ON r.stock_symbol = sp.stock_symbol
		LEFT JOIN stock_adjustments sa 
			ON r.stock_symbol = sa.stock_symbol
			AND sa.effective_date <= CURRENT_DATE
		WHERE r.user_id = $1
		  AND (sa.delisted IS NULL OR sa.delisted = FALSE)
		  AND DATE(r.reward_time) < CURRENT_DATE
		GROUP BY DATE(r.reward_time)
		ORDER BY DATE(r.reward_time);
	`

	rows, err := db.DB.Query(query, userId)
	if err != nil {
		logrus.Errorf("DB query error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	type Historical struct {
		Date     string  `json:"date"`
		TotalINR float64 `json:"total_inr"`
	}

	var data []Historical
	for rows.Next() {
		var h Historical
		rows.Scan(&h.Date, &h.TotalINR)
		h.TotalINR = round(h.TotalINR, 2)
		data = append(data, h)
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":        userId,
		"historical_inr": data,
	})
}

// ------------------- Today’s Stocks -------------------
func GetTodayStocks(c *gin.Context) {
	userId := c.Param("userId")

	query := `
		SELECT r.id, r.stock_symbol, 
		       r.shares * COALESCE(sa.multiplier, 1.0) AS adjusted_shares,
		       r.reward_time
		FROM rewards r
		LEFT JOIN stock_adjustments sa 
			ON r.stock_symbol = sa.stock_symbol
			AND sa.effective_date <= CURRENT_DATE
		WHERE r.user_id = $1
		  AND DATE(r.reward_time) = CURRENT_DATE
		  AND (sa.delisted IS NULL OR sa.delisted = FALSE)
		ORDER BY r.reward_time ASC;
	`

	rows, err := db.DB.Query(query, userId)
	if err != nil {
		logrus.Errorf("DB query error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	type Reward struct {
		RewardID    int       `json:"reward_id"`
		StockSymbol string    `json:"stock_symbol"`
		Shares      float64   `json:"shares"`
		RewardTime  time.Time `json:"reward_time"`
	}

	var rewards []Reward
	for rows.Next() {
		var r Reward
		rows.Scan(&r.RewardID, &r.StockSymbol, &r.Shares, &r.RewardTime)
		r.Shares = round(r.Shares, 6)
		rewards = append(rewards, r)
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":       userId,
		"rewards_today": rewards,
	})
}

// ------------------- Stock Adjustment -------------------
func AddOrUpdateStockAdjustment(c *gin.Context) {
	var req AdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.Warn("Invalid stock adjustment payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	query := `
		INSERT INTO stock_adjustments (stock_symbol, multiplier, effective_date, delisted)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (stock_symbol)
		DO UPDATE SET 
			multiplier = EXCLUDED.multiplier,
			effective_date = EXCLUDED.effective_date,
			delisted = EXCLUDED.delisted;
	`

	_, err := db.DB.Exec(query, req.StockSymbol, req.Multiplier, req.EffectiveDate, req.Delisted)
	if err != nil {
		logrus.Errorf("Failed to add/update stock adjustment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	action := "updated"
	if req.Delisted {
		action = "delisted"
	}
	logrus.Infof("✅ %s stock adjustment for %s (multiplier %.2f, effective %s)",
		action, req.StockSymbol, req.Multiplier, req.EffectiveDate)

	c.JSON(http.StatusOK, gin.H{
		"message": "Stock adjustment recorded successfully",
		"stock":   req.StockSymbol,
	})
}

// ------------------- List Adjustments -------------------
func GetAllStockAdjustments(c *gin.Context) {
	rows, err := db.DB.Query(`
		SELECT stock_symbol, multiplier, effective_date, delisted 
		FROM stock_adjustments 
		ORDER BY stock_symbol;
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	type Adjustment struct {
		StockSymbol   string  `json:"stock_symbol"`
		Multiplier    float64 `json:"multiplier"`
		EffectiveDate string  `json:"effective_date"`
		Delisted      bool    `json:"delisted"`
	}

	var adjustments []Adjustment
	for rows.Next() {
		var adj Adjustment
		rows.Scan(&adj.StockSymbol, &adj.Multiplier, &adj.EffectiveDate, &adj.Delisted)
		adjustments = append(adjustments, adj)
	}

	c.JSON(http.StatusOK, gin.H{"adjustments": adjustments})
}
