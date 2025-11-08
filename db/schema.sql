CREATE TABLE rewards (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(50) NOT NULL,
    stock_symbol VARCHAR(20) NOT NULL,
    shares NUMERIC(18,6) NOT NULL,
    reward_time TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, stock_symbol, reward_time)
);

CREATE TABLE ledger_entries (
    id SERIAL PRIMARY KEY,
    reward_id INT REFERENCES rewards(id),
    entry_type VARCHAR(20) CHECK (entry_type IN ('stock', 'cash', 'fee')),
    stock_symbol VARCHAR(20),
    units NUMERIC(18,6),
    amount_inr NUMERIC(18,4),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS stock_prices (
    id SERIAL PRIMARY KEY,
    stock_symbol VARCHAR(20) NOT NULL,
    price NUMERIC(18,4) NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(stock_symbol)
);

CREATE TABLE IF NOT EXISTS stock_adjustments (
    stock_symbol VARCHAR(20) PRIMARY KEY,         
    multiplier NUMERIC(10,4) DEFAULT 1.0000,      
    effective_date DATE NOT NULL,                 
    delisted BOOLEAN DEFAULT FALSE,               
    updated_at TIMESTAMP DEFAULT NOW()            
);

