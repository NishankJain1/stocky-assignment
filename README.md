📘 Stocky Assignment – Golang Backend
🧩 Overview

Stocky is a backend system where users earn fractional shares of Indian stocks (like RELIANCE, TCS, INFY) as rewards.
It records these rewards, updates mock INR valuations hourly, and handles corporate actions like stock splits, mergers, and delisting.

⚙️ Setup Instructions

NOTE- go and docker must be installed Postgres is run on docker and go is run locally

1️⃣ Clone the Repository
git clone https://github.com/<your-username>/stocky-assignment.git
cd stocky-assignment

2️⃣ Start PostgreSQL (via Docker)
docker compose up -d


This runs PostgreSQL in a container and prepares it for use.
The .env file is already included — no manual configuration needed.

3️⃣ Create Tables from Schema

After the database is running, execute the schema file:

docker exec -i assignment-db psql -U stocky_user -d assignment < db/schema.sql


You’ll see CREATE TABLE confirmations once all tables are created successfully.

4️⃣ Run the Application

In another terminal:

go run main.go


The server will start at http://localhost:8080

🧠 API Overview
Endpoint	Method	Description
/reward	POST	Record user stock reward
/today-stocks/:userId	GET	Get today’s rewards
/historical-inr/:userId	GET	INR value for past rewards
/stats/:userId	GET	Daily total shares & portfolio INR
/portfolio/:userId	GET	Current holdings & total INR
/stock-adjustment	POST	Add/Update stock split or delisting
/stock-adjustments	GET	Fetch all stock adjustments

📄 All detailed request/response payloads and test cases are available in the Stocky-assignment-postman-api-collection file in this repo.

⚙️ How It Works

Rewards:
Each /reward call credits the user with fractional stock shares.
Duplicate entries are prevented at the database level.

Stock Price Updater:
A background process runs hourly to refresh stock prices using mock random data.

Portfolio & Valuation:
All shares are multiplied by current stock prices, adjusted for stock splits or mergers.
Delisted stocks are automatically excluded from calculations.

Stock Adjustments:
/stock-adjustment handles:

Splits or mergers (multiplier)

Delisting or reinstating (delisted: true/false)
These updates immediately reflect in all portfolio and stats APIs.

Precision:

Shares stored as NUMERIC(18,6)

INR values stored as NUMERIC(18,4)

✅ Edge Cases Handled

Duplicate reward prevention

Stock splits, mergers, and delisting

INR rounding precision

Mock fallback for price updates

Automatic recalculation on adjustment

📦 Deliverables

Golang backend using Gin + Logrus

PostgreSQL schema (db/schema.sql)

.env file included

docker-compose.yml (PostgreSQL only)

Stocky-assignment-postman-api-collection for testing
