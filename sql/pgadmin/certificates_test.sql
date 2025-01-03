SELECT (mc.certificate_type || mc.certificate_code) AS y_code,
    (m.additional_code_type || m.additional_code_id) AS additional_code
FROM measure m
    LEFT JOIN measure_type mt ON m.measure_type = mt.measure_type
    LEFT JOIN measure_type_description mtd ON mt.measure_type = mtd.parent_measure_type
    LEFT JOIN measure_condition mc ON m.sid = mc.parent_sid
    LEFT JOIN measure_action ma ON mc.action_code = ma.action_code
    LEFT JOIN base_regulation br ON m.regulation_id = br.regulation_id
    LEFT JOIN modification_regulation mr ON m.regulation_id = mr.modification_regulation_id
    LEFT JOIN certificate c ON mc.certificate_code = c.certificate_code
    LEFT JOIN additional_code ac ON m.additional_code_id = ac.additional_code_id
WHERE m.goods_nomenclature_code IN (
        '8419508000',
        '8419500000',
        '8419000000',
        '8400000000'
    )
    AND mt.trade_movement_code IN ('1', '2')
    AND (
        m.geographical_area_id = 'RU'
        OR m.geographical_area_id IN (
            SELECT ga2.geographical_area_id
            FROM geographical_area ga1
                JOIN geographical_area_membership gam ON ga1.sid = gam.parent_sid
                JOIN geographical_area ga2 ON gam.sid_geographical_area_group = ga2.sid
            WHERE ga1.geographical_area_id = 'RU'
        )
    )
    AND (
        mc.certificate_type = 'Y'
        OR m.additional_code_type != ''
    )
    AND (
        m.date_end IS NULL
        OR m.date_end > CURRENT_TIMESTAMP
    )
    AND (m.date_start < CURRENT_TIMESTAMP)
    AND (
        c.date_end IS NULL
        OR c.date_end > CURRENT_TIMESTAMP
    )
    AND (c.date_start < CURRENT_TIMESTAMP)
    AND (
        br.date_end IS NULL
        OR br.date_end > CURRENT_TIMESTAMP
    )
    AND (
        mr.date_end IS NULL
        OR mr.date_end > CURRENT_TIMESTAMP
    )
GROUP BY y_code,
    additional_code