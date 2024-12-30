SELECT 
	mt.measure_type,
	mt.explosion_level,
	mt.measure_type_series_id,
	mtd.description,
	mt.order_number_capture_code,
	mt.measure_component_applicable_code,
	mt.trade_movement_code,
	mt.priority_code
FROM measure_type mt 
	LEFT JOIN measure_type_description mtd 
		ON mtd.parent_measure_type = mt.measure_type 
WHERE 
	mt.trade_movement_code IN ('1', '2')
ORDER BY mt.measure_type DESC