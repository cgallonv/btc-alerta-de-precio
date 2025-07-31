-- Resumen general de tiempos de ganancia
SELECT 
    COUNT(*) as total_opportunities,
    SUM(CASE WHEN time_to_2_percent_gains IS NOT NULL THEN 1 ELSE 0 END) as found_2_percent,
    SUM(CASE WHEN time_to_3_percent_gains IS NOT NULL THEN 1 ELSE 0 END) as found_3_percent,
    SUM(CASE WHEN time_to_4_percent_gains IS NOT NULL THEN 1 ELSE 0 END) as found_4_percent,
    ROUND(AVG(CASE WHEN time_to_2_percent_gains IS NOT NULL THEN time_to_2_percent_gains END), 2) as avg_hours_to_2_percent,
    ROUND(AVG(CASE WHEN time_to_3_percent_gains IS NOT NULL THEN time_to_3_percent_gains END), 2) as avg_hours_to_3_percent,
    ROUND(AVG(CASE WHEN time_to_4_percent_gains IS NOT NULL THEN time_to_4_percent_gains END), 2) as avg_hours_to_4_percent,
    ROUND(MIN(CASE WHEN time_to_2_percent_gains IS NOT NULL THEN time_to_2_percent_gains END), 2) as min_hours_to_2_percent,
    ROUND(MIN(CASE WHEN time_to_3_percent_gains IS NOT NULL THEN time_to_3_percent_gains END), 2) as min_hours_to_3_percent,
    ROUND(MIN(CASE WHEN time_to_4_percent_gains IS NOT NULL THEN time_to_4_percent_gains END), 2) as min_hours_to_4_percent
FROM ticker_data 
WHERE max_discount IS NOT NULL;