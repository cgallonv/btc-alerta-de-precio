-- Análisis por rangos de descuento
SELECT 
    CASE 
        WHEN max_discount <= -3 THEN 'Descuento > 3%'
        WHEN max_discount <= -2 THEN 'Descuento 2-3%'
        ELSE 'Descuento 1-2%'
    END as discount_range,
    COUNT(*) as opportunities,
    ROUND(AVG(CASE WHEN time_to_2_percent_gains IS NOT NULL THEN time_to_2_percent_gains END), 2) as avg_hours_to_2percent,
    ROUND(AVG(CASE WHEN time_to_3_percent_gains IS NOT NULL THEN time_to_3_percent_gains END), 2) as avg_hours_to_3percent,
    ROUND(AVG(CASE WHEN time_to_4_percent_gains IS NOT NULL THEN time_to_4_percent_gains END), 2) as avg_hours_to_4percent,
    SUM(CASE WHEN time_to_2_percent_gains IS NOT NULL THEN 1 ELSE 0 END) as reached_2percent,
    SUM(CASE WHEN time_to_3_percent_gains IS NOT NULL THEN 1 ELSE 0 END) as reached_3percent,
    SUM(CASE WHEN time_to_4_percent_gains IS NOT NULL THEN 1 ELSE 0 END) as reached_4percent
FROM ticker_data 
WHERE max_discount IS NOT NULL
GROUP BY 
    CASE 
        WHEN max_discount <= -3 THEN 'Descuento > 3%'
        WHEN max_discount <= -2 THEN 'Descuento 2-3%'
        ELSE 'Descuento 1-2%'
    END
ORDER BY max_discount;