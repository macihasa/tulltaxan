
-- HS DESCRIPTION
DROP MATERIALIZED VIEW IF EXISTS mv_hs_desc CASCADE;

CREATE MATERIALIZED VIEW mv_hs_desc AS
SELECT 
    gn.goods_nomenclature_code AS hs_code,
    string_agg(gnd.description, '\n') AS description,
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

CREATE INDEX IF NOT EXISTS idx_mv_hs_desc_hs_code ON mv_hs_desc (hs_code);
CREATE INDEX IF NOT EXISTS idx_mv_hs_desc_hs_code_left_6 ON mv_hs_desc (LEFT(hs_code, 6));
CREATE INDEX IF NOT EXISTS idx_mv_hs_desc_hs_code_left_4 ON mv_hs_desc (LEFT(hs_code, 4));
CREATE INDEX IF NOT EXISTS idx_mv_hs_desc_hs_code_left_2 ON mv_hs_desc (LEFT(hs_code, 2));


-- HS LEVEL DESCRIPTIONS
DROP MATERIALIZED VIEW IF EXISTS mv_hs_level_desc;
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_hs_level_desc AS
SELECT 
    mvd.hs_code AS cn_code,
    'CN' AS level,
    (SELECT string_agg(description, '\n') 
     FROM mv_hs_desc 
     WHERE hs_code = mvd.hs_code) AS cn_descriptions,
    
    hs_undernumber.hs_code AS hs_undernumber,
    'HS Undernumber' AS hs_undernumber_level,
    (SELECT string_agg(description, '\n') 
     FROM mv_hs_desc 
     WHERE hs_code = hs_undernumber.hs_code) AS hs_undernumber_descriptions,
    
    hs.hs_code AS hs_code,
    'HS' AS hs_level,
    (SELECT string_agg(description, '\n') 
     FROM mv_hs_desc 
     WHERE hs_code = hs.hs_code) AS hs_descriptions,
    
    chapter.hs_code AS chapter,
    'Chapter' AS chapter_level,
    (SELECT string_agg(description, '\n') 
     FROM mv_hs_desc 
     WHERE hs_code = chapter.hs_code) AS chapter_descriptions

FROM mv_hs_desc mvd
LEFT JOIN mv_hs_desc hs_undernumber 
    ON LEFT(mvd.hs_code, 6) || '0000' = hs_undernumber.hs_code
LEFT JOIN mv_hs_desc hs 
    ON LEFT(mvd.hs_code, 4) || '000000' = hs.hs_code
LEFT JOIN mv_hs_desc chapter 
    ON LEFT(mvd.hs_code, 2) || '00000000' = chapter.hs_code;



-- HS SEARCH

DROP MATERIALIZED VIEW IF EXISTS mv_goods_nomenclature_search;

CREATE MATERIALIZED VIEW mv_goods_nomenclature_search AS
SELECT 
    cn_code,
    -- JSONB object with one concatenated description and language_id
    jsonb_build_object(
        'description', CONCAT(
            COALESCE('Chapter: ' || chapter_descriptions, ''), '\n',
            COALESCE('HS: ' || hs_descriptions, ''), '\n',
            COALESCE('HS Undernumber: ' || hs_undernumber_descriptions, ''), '\n',
            COALESCE('CN: ' || cn_descriptions, '')
        ),
        'language_id', 'SV'
    ) AS descriptions,

    -- Full-text search vector
    to_tsvector(
        'swedish',
        CONCAT(
            COALESCE('Chapter: ' || chapter_descriptions, ''), '\n',
            COALESCE('HS: ' || hs_descriptions, ''), '\n',
            COALESCE('HS Undernumber: ' || hs_undernumber_descriptions, ''), '\n',
            COALESCE('CN: ' || cn_descriptions, '')
        )
    ) AS search_vector

FROM mv_hs_level_desc;


select * from mv_goods_nomenclature_search