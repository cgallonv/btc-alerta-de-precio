# Bitcoin Price Analysis Scripts

This directory contains scripts for analyzing Bitcoin price data, including historical data loading, discount detection, and gain potential analysis.

## Scripts Overview

### 1. Data Collection
- `backfill_historical_data.go`: Loads historical price data from Binance
  ```bash
  go run scripts/analytics/backfill_historical_data.go
  ```
  - Fetches 1-minute candles for the last 60 days
  - Processes data in 24-hour chunks to respect rate limits
  - Stores data in SQLite database

### 2. Discount Analysis
- `discount_opportunities.go`: Identifies price drop opportunities
  ```bash
  go run scripts/analytics/discount_opportunities.go
  ```
  - Analyzes price drops in two windows:
    - 24-hour comparison (day-to-day drops)
    - 5-hour window (short-term drops)
  - Only stores significant drops (> 1%)
  - Uses optimized queries and batch processing

### 3. Gain Potential Analysis
- `potential_gains.go`: Analyzes how long it takes to reach specific gain targets
  ```bash
  go run scripts/analytics/potential_gains.go
  ```
  - Tracks time to reach:
    - 2% gains
    - 3% gains
    - 4% gains
  - Uses a 2-week maximum window
  - Optimized with single-query analysis

## Database Indexes

### Required Indexes
```sql
-- 1. Main index for time-based price queries
CREATE INDEX idx_ticker_time_price 
ON ticker_data (close_time, last_price);

-- 2. Index for discount opportunities
CREATE INDEX idx_ticker_discount 
ON ticker_data (max_discount) 
WHERE max_discount IS NOT NULL;

-- 3. Composite index for gain calculations
CREATE INDEX idx_ticker_gains_search 
ON ticker_data (close_time, last_price, id);

-- 4. Base indexes for common queries
CREATE INDEX idx_ticker_data_timestamp ON ticker_data(timestamp);
CREATE INDEX idx_ticker_data_symbol ON ticker_data(symbol);
CREATE INDEX idx_ticker_data_source ON ticker_data(source);
```

### Index Usage
1. **idx_ticker_time_price**:
   - Used for: Time-based price comparisons
   - Benefits: 
     - 24-hour price comparisons
     - 5-hour window analysis
     - Future price searches

2. **idx_ticker_discount**:
   - Used for: Finding records with discounts
   - Benefits:
     - Filtered index (only non-null discounts)
     - Speeds up opportunity searches
     - Reduces scan size

3. **idx_ticker_gains_search**:
   - Used for: Gain potential calculations
   - Benefits:
     - Optimizes time window searches
     - Includes ID for unique identification
     - Supports ordered results

### Index Verification
```sql
-- Check existing indexes
SELECT 
    type,
    name,
    tbl_name,
    sql
FROM sqlite_master
WHERE type = 'index' 
AND tbl_name = 'ticker_data'
ORDER BY name;

-- Verify index usage
EXPLAIN QUERY PLAN
SELECT * FROM ticker_data 
WHERE max_discount IS NOT NULL 
LIMIT 5;
```

## SQL Analysis Scripts

### Setup Scripts
```sql
-- Reset gain analysis columns
UPDATE ticker_data 
SET time_to_2_percent_gains = NULL,
    time_to_3_percent_gains = NULL,
    time_to_4_percent_gains = NULL;
```

### Analysis Queries
- `gains_analysis.sql`: Overall statistics
- `gains_examples.sql`: Specific examples of opportunities
- `gains_correlation.sql`: Correlation between discounts and gains

## Execution Order

1. **Database Setup**:
   ```bash
   # Create required indexes (if not exists)
   sqlite3 btc_market_data_dev.db < scripts/analytics/create_indexes.sql
   
   # Verify indexes
   sqlite3 btc_market_data_dev.db < scripts/analytics/verify_indexes.sql
   ```

2. **Initial Data Load**:
   ```bash
   go run scripts/analytics/backfill_historical_data.go
   ```

3. **Discount Analysis**:
   ```bash
   go run scripts/analytics/discount_opportunities.go
   ```

4. **Gain Analysis**:
   ```bash
   # First, reset previous analysis if needed
   sqlite3 btc_market_data_dev.db < scripts/analytics/gains_reset.sql
   
   # Run the analysis
   go run scripts/analytics/potential_gains.go
   ```

5. **Review Results**:
   ```bash
   sqlite3 btc_market_data_dev.db < scripts/analytics/gains_analysis.sql
   sqlite3 btc_market_data_dev.db < scripts/analytics/gains_examples.sql
   sqlite3 btc_market_data_dev.db < scripts/analytics/gains_correlation.sql
   ```

## Database Schema

### Table Structure
The scripts require the following columns in the `ticker_data` table:
- `max_discount` (REAL): Stores percentage drops
- `time_to_2_percent_gains` (REAL): Hours to reach 2% gain
- `time_to_3_percent_gains` (REAL): Hours to reach 3% gain
- `time_to_4_percent_gains` (REAL): Hours to reach 4% gain

## Performance Optimizations

1. **Batch Processing**:
   - All scripts use batch processing (1000 records)
   - Prevents memory overload
   - Provides progress feedback

2. **Query Optimizations**:
   - Uses CTEs for complex calculations
   - Single query for multiple thresholds
   - Efficient time window calculations

3. **Index Strategy**:
   - Filtered indexes to reduce index size
   - Composite indexes for complex queries
   - Covers most common query patterns