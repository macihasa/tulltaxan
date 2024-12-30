SELECT
	*
FROM
	MEASURE_ACTION MA
	LEFT JOIN MEASURE_ACTION_DESCRIPTION MAD ON MAD.PARENT_ACTION_CODE = MA.ACTION_CODE;

SELECT
	*
FROM
	MEASURE_CONDITION MC
	LEFT JOIN MEASURE_ACTION MA ON MA.ACTION_CODE = MC.ACTION_CODE
	LEFT JOIN MEASURE_ACTION_DESCRIPTION MAD ON MAD.PARENT_ACTION_CODE = MA.ACTION_CODE
	where mc.certificate_type = 'Y' and mc.certificate_code = '901'
LIMIT
	1000;