-- Enable column headers and pretty printing
.headers on
.mode column

-- Show all tables
.tables

-- Show schema for ticker_data
.schema ticker_data

-- Most recent price data
SELECT 
    timestamp,
    last_price,
    price_change_percent,
    volume,
    quote_volume,
    total_trades
FROM ticker_data 
ORDER BY timestamp DESC 
LIMIT 5;

-- Price statistics for the last hour
SELECT 
    MIN(last_price) as min_price,
    MAX(last_price) as max_price,
    AVG(last_price) as avg_price,
    SUM(volume) as total_volume,
    COUNT(*) as data_points
FROM ticker_data 
WHERE timestamp >= datetime('now', '-1 hour');

-- Price changes over time
SELECT 
    strftime('%Y-%m-%d %H:%M', timestamp) as time_bucket,
    ROUND(AVG(last_price), 2) as avg_price,
    ROUND(AVG(price_change_percent), 4) as avg_change_percent,
    COUNT(*) as samples
FROM ticker_data 
GROUP BY time_bucket
ORDER BY time_bucket DESC
LIMIT 10;

-- Volume analysis
SELECT 
    strftime('%Y-%m-%d %H:%M', timestamp) as time_bucket,
    ROUND(SUM(volume), 4) as btc_volume,
    ROUND(SUM(quote_volume), 2) as usd_volume,
    SUM(total_trades) as trades
FROM ticker_data 
GROUP BY time_bucket
ORDER BY time_bucket DESC
LIMIT 10;

-- Price volatility (high-low spread)
SELECT 
    strftime('%Y-%m-%d %H:%M', timestamp) as time_bucket,
    ROUND(MAX(high_price), 2) as period_high,
    ROUND(MIN(low_price), 2) as period_low,
    ROUND(MAX(high_price) - MIN(low_price), 2) as price_spread,
    ROUND(((MAX(high_price) - MIN(low_price)) / MIN(low_price) * 100), 2) as spread_percent
FROM ticker_data 
GROUP BY time_bucket
ORDER BY time_bucket DESC
LIMIT 10;

-- Data collection health check
SELECT 
    COUNT(*) as total_records,
    COUNT(DISTINCT strftime('%Y-%m-%d %H:%M', timestamp)) as unique_minutes,
    MIN(timestamp) as oldest_record,
    MAX(timestamp) as newest_record,
    ROUND(AVG(total_trades)) as avg_trades_per_record
FROM ticker_data;

-- Find gaps in data collection
WITH RECURSIVE
    minutes(dt) AS (
        SELECT datetime(MIN(timestamp), 'start of minute')
        FROM ticker_data
        UNION ALL
        SELECT datetime(dt, '+1 minute')
        FROM minutes
        WHERE dt < (SELECT datetime(MAX(timestamp), 'start of minute') FROM ticker_data)
    )
SELECT 
    minutes.dt as expected_minute,
    ticker_data.timestamp as actual_record
FROM minutes
LEFT JOIN ticker_data ON datetime(ticker_data.timestamp, 'start of minute') = minutes.dt
WHERE ticker_data.timestamp IS NULL
LIMIT 10;

-- Storage usage analysis
SELECT 
    COUNT(*) as record_count,
    ROUND(AVG(LENGTH(CAST(last_price AS TEXT)) + 
              LENGTH(CAST(volume AS TEXT)) + 
              LENGTH(CAST(quote_volume AS TEXT)) + 
              LENGTH(timestamp) + 
              LENGTH(symbol)), 2) as avg_record_size_bytes
FROM ticker_data;

-- Exit SQLite
.quit