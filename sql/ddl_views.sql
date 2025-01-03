-------------- VIEWS ----------------

-- HS DESCRIPTION
DROP MATERIALIZED VIEW IF EXISTS mv_hs_desc CASCADE;

CREATE MATERIALIZED VIEW mv_hs_desc AS
SELECT 
    gn.goods_nomenclature_code AS hs_code,
    string_agg(gnd.description, ' ') AS description,
    to_tsvector('swedish', string_agg(gnd.description, ' ')) AS search_vector,
    gnd.language_id
FROM 
    goods_nomenclature gn
LEFT JOIN goods_nomenclature_description_period gndp 
    ON gn.sid = gndp.parent_sid
LEFT JOIN goods_nomenclature_description gnd 
    ON gndp.sid = gnd.parent_sid
WHERE 
    (gn.date_end IS NULL OR gn.date_end > now()) 
    AND (gndp.date_end IS NULL OR gndp.date_end > now()) 
    AND gnd.language_id = 'SV'
GROUP BY 
    gn.goods_nomenclature_code, gnd.language_id;

CREATE INDEX IF NOT EXISTS idx_mv_hs_desc_search_vector ON mv_hs_desc USING gin (search_vector);

-- HS LEVEL DESCRIPTIONS
DROP MATERIALIZED VIEW IF EXISTS mv_hs_level_desc CASCADE;

CREATE MATERIALIZED VIEW mv_hs_level_desc AS
SELECT 
    mvd.hs_code AS cn_code,
    'CN' AS level,
    mvd.description AS cn_descriptions,

    hs_undernumber.hs_code AS hs_undernumber,
    'HS Undernumber' AS hs_undernumber_level,
    hs_undernumber.description AS hs_undernumber_descriptions,

    hs.hs_code AS hs_code,
    'HS' AS hs_level,
    hs.description AS hs_descriptions,

    chapter.hs_code AS chapter,
    'Chapter' AS chapter_level,
    chapter.description AS chapter_descriptions
FROM mv_hs_desc mvd
INNER JOIN declarable_goods_nomenclature dgn
    ON LEFT(mvd.hs_code, 8) = dgn.goods_nomenclature_code AND RIGHT(mvd.hs_code, 2) = '00'
LEFT JOIN mv_hs_desc hs_undernumber 
    ON LEFT(mvd.hs_code, 6) || '0000' = hs_undernumber.hs_code
LEFT JOIN mv_hs_desc hs 
    ON LEFT(mvd.hs_code, 4) || '000000' = hs.hs_code
LEFT JOIN mv_hs_desc chapter 
    ON LEFT(mvd.hs_code, 2) || '00000000' = chapter.hs_code
;


-- HS SEARCH
DROP MATERIALIZED VIEW IF EXISTS mv_goods_nomenclature_search CASCADE;

CREATE MATERIALIZED VIEW mv_goods_nomenclature_search AS
SELECT 
    cn_code,
    CONCAT(
        'CH: '|| COALESCE(chapter_descriptions, ''), '<br>',
        'HS: '|| COALESCE(hs_descriptions, ''), '<br>',
        'HSU:'|| COALESCE(hs_undernumber_descriptions, ''), '<br>',
        'CN: '|| COALESCE(cn_descriptions, '')
    ) AS descriptions,
    to_tsvector(
        'swedish',
        CONCAT(     
            COALESCE(chapter_descriptions, ''), ' ',
            COALESCE(hs_descriptions, ''), ' ',
            COALESCE(hs_undernumber_descriptions, ''), ' ',
            COALESCE(cn_descriptions, '')
        )
    ) AS search_vector
FROM mv_hs_level_desc;

CREATE INDEX IF NOT EXISTS idx_mv_goods_nomenclature_search_vector ON mv_goods_nomenclature_search USING gin (search_vector);
-- Enable pg_trgm extension
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Create a trigram index for the descriptions column
CREATE INDEX IF NOT EXISTS idx_goods_nomenclature_descriptions_trgm
ON mv_goods_nomenclature_search USING gin (descriptions gin_trgm_ops);
