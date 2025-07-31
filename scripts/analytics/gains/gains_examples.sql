-- Mostrar 10 ejemplos con todos los detalles
SELECT 
    id,
    datetime(close_time) as time,
    ROUND(last_price, 2) as price,
    ROUND(max_discount, 2) as discount_percent,
    ROUND(time_to_2_percent_gains, 2) as hours_to_2percent,
    ROUND(time_to_3_percent_gains, 2) as hours_to_3percent,
    ROUND(time_to_4_percent_gains, 2) as hours_to_4percent
FROM ticker_data 
WHERE max_discount IS NOT NULL 
    AND (time_to_2_percent_gains IS NOT NULL 
    OR time_to_3_percent_gains IS NOT NULL 
    OR time_to_4_percent_gains IS NOT NULL)
ORDER BY close_time DESC
LIMIT 10;