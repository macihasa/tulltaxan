-- Additional Codes
CREATE TABLE IF NOT EXISTS additional_code (
	sid INT PRIMARY KEY,
	additional_code_id VARCHAR(255),
	additional_code_type VARCHAR(255),
	change_type VARCHAR(255),
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	is_active BOOLEAN DEFAULT FALSE,
	national INT
);
CREATE INDEX if NOT EXISTS idx_additional_code_id ON additional_code (additional_code_id);
CREATE TABLE IF NOT EXISTS additional_code_description_period (
	sid INT PRIMARY KEY,
	parent_sid INT,
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	national INT,
	FOREIGN key (parent_sid) REFERENCES additional_code (sid) ON DELETE cascade
);
CREATE TABLE IF NOT EXISTS additional_code_description (
	parent_sid INT,
	description TEXT,
	language_id VARCHAR(255),
	national INT,
	PRIMARY KEY (parent_sid, language_id),
	FOREIGN key (parent_sid) REFERENCES additional_code_description_period (sid) ON DELETE cascade
);
CREATE TABLE IF NOT EXISTS additional_code_footnote_association (
	parent_sid INT,
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	footnote_id INT,
	footnote_type VARCHAR(255),
	national INT,
	PRIMARY KEY (parent_sid, footnote_id, footnote_type),
	FOREIGN key (parent_sid) REFERENCES additional_code (sid) ON DELETE cascade
);
-- Certificates
CREATE TABLE IF NOT EXISTS certificate (
	certificate_code VARCHAR(255) NOT NULL,
	certificate_type VARCHAR(255) NOT NULL,
	change_type VARCHAR(255),
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	national INT,
	is_active BOOLEAN DEFAULT FALSE,
	PRIMARY KEY (certificate_code, certificate_type)
);
CREATE INDEX if NOT EXISTS idx_certificate_type_date_end ON certificate (certificate_type, date_end);
CREATE TABLE IF NOT EXISTS certificate_description_period (
	sid INT PRIMARY KEY,
	parent_certificate_code VARCHAR(255),
	parent_certificate_type VARCHAR(255),
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	national INT,
	FOREIGN key (parent_certificate_code, parent_certificate_type) REFERENCES certificate (certificate_code, certificate_type) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS certificate_description (
	parent_sid INT,
	description TEXT,
	language_id VARCHAR(255),
	national INT,
	PRIMARY KEY (parent_sid, language_id),
	FOREIGN key (parent_sid) REFERENCES certificate_description_period (sid) ON DELETE CASCADE
);
-- Goods Nomenclature
CREATE TABLE IF NOT EXISTS goods_nomenclature (
	sid INT PRIMARY KEY,
	change_type VARCHAR(255),
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	goods_nomenclature_code VARCHAR(255),
	national INT,
	product_line_suffix INT,
	statistical_indicator INT
);
CREATE TABLE IF NOT EXISTS goods_nomenclature_indent (
	sid INT PRIMARY KEY,
	parent_sid INT,
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	national INT,
	quantity_indents INT,
	FOREIGN key (parent_sid) REFERENCES goods_nomenclature (sid) ON DELETE cascade
);
CREATE TABLE IF NOT EXISTS goods_nomenclature_description_period (
	sid INT PRIMARY KEY,
	parent_sid INT,
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	national INT,
	FOREIGN key (parent_sid) REFERENCES goods_nomenclature (sid) ON DELETE cascade
);
CREATE TABLE IF NOT EXISTS goods_nomenclature_description (
	parent_sid INT,
	description TEXT,
	language_id VARCHAR(255),
	national INT,
	PRIMARY KEY (parent_sid, language_id),
	FOREIGN key (parent_sid) REFERENCES goods_nomenclature_description_period (sid) ON DELETE cascade
);
CREATE TABLE IF NOT EXISTS goods_nomenclature_footnote_association (
	parent_sid INT,
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	footnote_id INT,
	footnote_type VARCHAR(255),
	national INT,
	PRIMARY KEY (parent_sid, footnote_id, footnote_type),
	FOREIGN key (parent_sid) REFERENCES goods_nomenclature (sid) ON DELETE cascade
);
CREATE TABLE IF NOT EXISTS goods_nomenclature_group_membership (
	parent_sid INT,
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	goods_nomenclature_group_id VARCHAR(255),
	goods_nomenclature_group_type VARCHAR(255),
	national INT,
	PRIMARY KEY (
		parent_sid,
		goods_nomenclature_group_id,
		goods_nomenclature_group_type
	),
	FOREIGN key (parent_sid) REFERENCES goods_nomenclature (sid) ON DELETE cascade
);
-- Geographical Areas
CREATE TABLE IF NOT EXISTS geographical_area (
	sid INT PRIMARY KEY,
	change_type VARCHAR(255),
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	geographical_area_code INT,
	geographical_area_id VARCHAR(255),
	national INT,
	sid_parent_group INT
);
CREATE INDEX if NOT EXISTS idx_geographical_area_id ON geographical_area (geographical_area_id);
CREATE TABLE IF NOT EXISTS geographical_area_membership (
	parent_sid INT,
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	national INT,
	sid_geographical_area_group INT,
	PRIMARY KEY (
		sid_geographical_area_group,
		parent_sid,
		date_start
	),
	FOREIGN key (parent_sid) REFERENCES geographical_area (sid) ON DELETE cascade
);
CREATE INDEX if NOT EXISTS idx_geographical_area_membership_parent_sid ON geographical_area_membership (parent_sid);
CREATE INDEX if NOT EXISTS idx_geographical_area_membership_group_sid ON geographical_area_membership (sid_geographical_area_group, parent_sid);
CREATE TABLE IF NOT EXISTS geographical_area_description_period (
	sid INT PRIMARY KEY,
	parent_sid INT,
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	national INT,
	FOREIGN key (parent_sid) REFERENCES geographical_area (sid) ON DELETE cascade
);
CREATE TABLE IF NOT EXISTS geographical_area_description (
	parent_sid INT,
	description TEXT,
	language_id VARCHAR(255),
	national INT,
	PRIMARY KEY (parent_sid, language_id),
	FOREIGN key (parent_sid) REFERENCES geographical_area_description_period (sid) ON DELETE cascade
);
-- Measure
CREATE TABLE IF NOT EXISTS measure (
	sid INT PRIMARY KEY,
	sid_additional_code INT,
	sid_export_refund_nomenclature INT,
	sid_geographical_area INT,
	sid_goods_nomenclature INT,
	additional_code_id VARCHAR(255),
	additional_code_type VARCHAR(255),
	change_type VARCHAR(255),
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	expression TEXT,
	geographical_area_id VARCHAR(255),
	goods_nomenclature_code VARCHAR(255),
	justification_regulation_id VARCHAR(255),
	justification_regulation_role_type INT,
	measure_type VARCHAR(255),
	national INT,
	quota_order_number INT,
	reduction_indicator INT,
	regulation_id VARCHAR(255),
	regulation_role_type INT,
	stopped_flag INT
);
CREATE INDEX if NOT EXISTS idx_measure_goods_nomenclature_code ON measure (goods_nomenclature_code);
CREATE INDEX if NOT EXISTS idx_measure_additional_code_type ON measure (additional_code_type);
CREATE INDEX if NOT EXISTS idx_measure_geographical_area_id ON measure (geographical_area_id);
CREATE INDEX if NOT EXISTS idx_measure_additional_code_id ON measure (additional_code_id);
CREATE INDEX if NOT EXISTS idx_measure_regulation_id ON measure (regulation_id);
CREATE INDEX if NOT EXISTS idx_measure_type_sid ON measure (measure_type, sid);
CREATE INDEX if NOT EXISTS idx_measure_date_start ON measure (date_start);
CREATE INDEX if NOT EXISTS idx_measure_date_end ON measure (date_end);
CREATE TABLE IF NOT EXISTS measure_condition (
	sid INT PRIMARY KEY,
	parent_sid INT,
	action_code VARCHAR(255),
	certificate_code VARCHAR(255),
	certificate_type VARCHAR(255),
	condition_code_id VARCHAR(255),
	duty_amount FLOAT,
	expression TEXT,
	measurement_unit_code VARCHAR(255),
	measurement_unit_qualifier_code VARCHAR(255),
	monetary_unit_code VARCHAR(255),
	national INT,
	sequence_number INT,
	FOREIGN key (parent_sid) REFERENCES measure (sid) ON DELETE cascade
);
CREATE INDEX if NOT EXISTS idx_measure_condition_parent_sid ON measure_condition (parent_sid);
CREATE INDEX if NOT EXISTS idx_measure_condition_certificate ON measure_condition (certificate_type, certificate_code);
CREATE TABLE IF NOT EXISTS measure_condition_component (
	parent_sid INT,
	duty_amount FLOAT,
	duty_expression_id INT,
	measurement_unit_code VARCHAR(255),
	measurement_unit_qualifier_code VARCHAR(255),
	monetary_unit_code VARCHAR(255),
	national INT,
	PRIMARY KEY (parent_sid, duty_expression_id),
	FOREIGN key (parent_sid) REFERENCES measure_condition (sid) ON DELETE cascade
);
CREATE TABLE IF NOT EXISTS measure_footnote_association (
	parent_sid INT,
	footnote_id VARCHAR(255),
	footnote_type VARCHAR(255),
	national INT,
	PRIMARY KEY (parent_sid, footnote_id, footnote_type),
	FOREIGN key (parent_sid) REFERENCES measure (sid) ON DELETE cascade
);
CREATE TABLE IF NOT EXISTS measure_component (
	parent_sid INT,
	duty_amount FLOAT,
	duty_expression_id INT,
	measurement_unit_code VARCHAR(255),
	measurement_unit_qualifier_code VARCHAR(255),
	monetary_unit_code VARCHAR(255),
	national INT,
	PRIMARY KEY (parent_sid, duty_expression_id),
	FOREIGN key (parent_sid) REFERENCES measure (sid) ON DELETE cascade
);
CREATE TABLE IF NOT EXISTS measure_excluded_geographical_area (
	parent_sid INT,
	geographical_area_id VARCHAR(255),
	national INT,
	sid_geographical_area INT,
	PRIMARY KEY (
		parent_sid,
		geographical_area_id,
		sid_geographical_area
	),
	FOREIGN key (parent_sid) REFERENCES measure (sid) ON DELETE cascade
);
CREATE TABLE IF NOT EXISTS measure_partial_temporary_stop (
	parent_sid INT PRIMARY KEY,
	regulation_id VARCHAR(255),
	regulation_role_type INT,
	national INT,
	FOREIGN key (parent_sid) REFERENCES measure (sid) ON DELETE cascade
);
-- Measure Types
CREATE TABLE IF NOT EXISTS measure_type (
	measure_type VARCHAR(255) PRIMARY KEY,
	change_type VARCHAR(255),
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	explosion_level INT,
	measure_component_applicable_code INT,
	measure_type_series_id VARCHAR(255),
	national INT,
	order_number_capture_code INT,
	origin_destination_code INT,
	priority_code INT,
	trade_movement_code INT
);
CREATE TABLE IF NOT EXISTS measure_type_description (
	parent_measure_type VARCHAR(255),
	description TEXT,
	language_id VARCHAR(255),
	national INT,
	PRIMARY KEY (parent_measure_type, language_id),
	FOREIGN key (parent_measure_type) REFERENCES measure_type (measure_type)
);
-- Goods Nomenclature Groups
CREATE TABLE IF NOT EXISTS goods_nomenclature_group (
	goods_nomenclature_group_id VARCHAR(255) NOT NULL,
	goods_nomenclature_group_type VARCHAR(255) NOT NULL,
	change_type VARCHAR(255),
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	national INT,
	nomenclature_group_facility_code INT,
	PRIMARY KEY (
		goods_nomenclature_group_id,
		goods_nomenclature_group_type
	)
);
CREATE TABLE IF NOT EXISTS goods_nomenclature_group_description (
	parent_goods_nomenclature_group_id VARCHAR(255),
	parent_goods_nomenclature_group_type VARCHAR(255),
	description TEXT,
	language_id VARCHAR(10),
	national INT,
	PRIMARY KEY (
		parent_goods_nomenclature_group_id,
		parent_goods_nomenclature_group_type,
		language_id
	),
	FOREIGN key (
		parent_goods_nomenclature_group_id,
		parent_goods_nomenclature_group_type
	) REFERENCES goods_nomenclature_group (
		goods_nomenclature_group_id,
		goods_nomenclature_group_type
	)
);
-- Monetary Exchange Periods
CREATE TABLE IF NOT EXISTS monetary_exchange_period (
	sid INT PRIMARY KEY,
	change_type VARCHAR(255),
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	monetary_unit_code VARCHAR(255),
	national INT,
	is_quoted BOOLEAN
);
CREATE INDEX if NOT EXISTS idx_mep_date_end ON monetary_exchange_period (date_end);
CREATE INDEX if NOT EXISTS idx_mep_monetary_unit_code ON monetary_exchange_period (monetary_unit_code);
CREATE INDEX if NOT EXISTS idx_monetary_exchange_period_date_end_code ON monetary_exchange_period (date_end, monetary_unit_code);
CREATE TABLE IF NOT EXISTS monetary_exchange_rate (
	parent_sid INT,
	calculation_unit INT,
	monetary_conversion_rate FLOAT,
	monetary_unit_code VARCHAR(255),
	national INT,
	PRIMARY KEY (parent_sid, monetary_unit_code),
	FOREIGN key (parent_sid) REFERENCES monetary_exchange_period (sid) ON DELETE cascade
);
CREATE INDEX if NOT EXISTS idx_mer_monetary_unit_code ON monetary_exchange_rate (monetary_unit_code);
CREATE INDEX if NOT EXISTS idx_monetary_exchange_rate_code ON monetary_exchange_rate (monetary_unit_code);
-- Declarable Goods Nomenclature
CREATE TABLE IF NOT EXISTS declarable_goods_nomenclature (
	change_type VARCHAR(255),
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	goods_nomenclature_code VARCHAR(255) PRIMARY KEY,
	type VARCHAR(255)
);
CREATE INDEX if NOT EXISTS idx_declarable_goods_nomenclature_date_end ON declarable_goods_nomenclature (date_end);
CREATE INDEX if NOT EXISTS idx_declarable_goods_nomenclature_code ON declarable_goods_nomenclature (goods_nomenclature_code);
CREATE INDEX if NOT EXISTS idx_declarable_goods_nomenclature_type ON declarable_goods_nomenclature (type);
CREATE INDEX if NOT EXISTS idx_declarable_goods_nomenclature_code_type ON declarable_goods_nomenclature (goods_nomenclature_code, type);
-- Footnotes
CREATE TABLE IF NOT EXISTS footnote (
	footnote_id VARCHAR(255) NOT NULL,
	footnote_type VARCHAR(255) NOT NULL,
	change_type VARCHAR(255),
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	national INT,
	PRIMARY KEY (footnote_id, footnote_type)
);
CREATE TABLE IF NOT EXISTS footnote_description_period (
	sid INT PRIMARY KEY,
	parent_footnote_id VARCHAR(255),
	parent_footnote_type VARCHAR(255),
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	national INT,
	FOREIGN key (parent_footnote_id, parent_footnote_type) REFERENCES footnote (footnote_id, footnote_type)
);
CREATE TABLE IF NOT EXISTS footnote_description (
	parent_sid INT,
	description TEXT,
	language_id VARCHAR(10),
	national INT,
	PRIMARY KEY (parent_sid, language_id),
	FOREIGN key (parent_sid) REFERENCES footnote_description_period (sid) ON DELETE cascade
);
-- Export Refund Nomenclature
CREATE TABLE IF NOT EXISTS export_refund_nomenclature (
	sid INT PRIMARY KEY,
	additional_code_type VARCHAR(255),
	change_type VARCHAR(255),
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	export_refund_code INT,
	goods_nomenclature_code VARCHAR(255),
	national INT,
	product_line_suffix INT,
	sid_goods_nomenclature INT
);
CREATE TABLE IF NOT EXISTS export_refund_nomenclature_indent (
	sid INT PRIMARY KEY,
	parent_sid INT,
	date_start TIMESTAMP,
	national INT,
	quantity_indents INT,
	FOREIGN key (parent_sid) REFERENCES export_refund_nomenclature (sid) ON DELETE cascade
);
CREATE TABLE IF NOT EXISTS export_refund_nomenclature_description_period (
	sid INT PRIMARY KEY,
	parent_sid INT,
	date_start TIMESTAMP,
	national INT,
	FOREIGN key (parent_sid) REFERENCES export_refund_nomenclature (sid) ON DELETE cascade
);
CREATE TABLE IF NOT EXISTS export_refund_nomenclature_description (
	parent_sid INT,
	description TEXT,
	language_id VARCHAR(10),
	national INT,
	PRIMARY KEY (parent_sid, language_id),
	FOREIGN key (parent_sid) REFERENCES export_refund_nomenclature_description_period (sid) ON DELETE cascade
);
CREATE TABLE IF NOT EXISTS export_refund_nomenclature_footnote_association (
	parent_sid INT,
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	footnote_id INT,
	footnote_type VARCHAR(255),
	national INT,
	PRIMARY KEY (parent_sid, footnote_id, footnote_type),
	FOREIGN key (parent_sid) REFERENCES export_refund_nomenclature (sid) ON DELETE cascade
);
-- Measurement
CREATE TABLE IF NOT EXISTS measurement (
	change_type VARCHAR(255),
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	measurement_unit_code VARCHAR(255),
	measurement_unit_qualifier_code VARCHAR(255),
	national INT,
	PRIMARY KEY (
		measurement_unit_code,
		measurement_unit_qualifier_code
	)
);
-- Measure Action
CREATE TABLE IF NOT EXISTS measure_action (
	action_code VARCHAR(255) PRIMARY KEY,
	change_type VARCHAR(255),
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	national INT
);
CREATE TABLE IF NOT EXISTS measure_action_description (
	parent_action_code VARCHAR(255),
	description TEXT,
	language_id VARCHAR(255),
	national INT,
	PRIMARY KEY (parent_action_code, language_id),
	FOREIGN key (parent_action_code) REFERENCES measure_action (action_code)
);
-- Measure Condition Code
CREATE TABLE IF NOT EXISTS measure_condition_code (
	condition_code VARCHAR(255) PRIMARY KEY,
	change_type VARCHAR(255),
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	national INT,
	type INT
);
CREATE TABLE IF NOT EXISTS measure_condition_code_description (
	parent_condition_code VARCHAR(255),
	description TEXT,
	language_id VARCHAR(255),
	national INT,
	PRIMARY KEY (parent_condition_code, language_id),
	FOREIGN key (parent_condition_code) REFERENCES measure_condition_code (condition_code)
);
-- Code Type
CREATE TABLE IF NOT EXISTS code_type (
	id VARCHAR(255) PRIMARY KEY,
	code_type_id VARCHAR(255),
	change_type VARCHAR(255),
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	export_import_type VARCHAR(255),
	measure_type_series_id VARCHAR(255),
	national INT
);
CREATE TABLE IF NOT EXISTS code_type_description (
	parent_id VARCHAR(255),
	description TEXT,
	language_id VARCHAR(255),
	national INT,
	PRIMARY KEY (parent_id, language_id),
	FOREIGN key (parent_id) REFERENCES code_type (id)
);
-- Duty Expression
CREATE TABLE IF NOT EXISTS duty_expression (
	duty_expression_id VARCHAR(255) PRIMARY KEY,
	change_type VARCHAR(255),
	date_end TIMESTAMP,
	date_start TIMESTAMP,
	duty_amount_applicability_code INT,
	measurement_unit_applicability_code INT,
	monetary_unit_applicability_code INT,
	national INT
);
CREATE TABLE IF NOT EXISTS duty_expression_description (
	parent_duty_expression_id VARCHAR(255),
	description TEXT,
	language_id VARCHAR(255),
	national INT,
	PRIMARY KEY (parent_duty_expression_id, language_id),
	FOREIGN key (parent_duty_expression_id) REFERENCES duty_expression (duty_expression_id)
);
-- Lookup Table
CREATE TABLE IF NOT EXISTS lookup_table (
	sid INT PRIMARY KEY,
	change_type VARCHAR(255),
	date_start TIMESTAMP,
	interpolate BOOLEAN,
	max_interval INT,
	min_interval INT,
	table_id VARCHAR(255)
);
CREATE TABLE IF NOT EXISTS lookup_table_item (
	parent_sid INT,
	threshold FLOAT,
	value FLOAT,
	PRIMARY KEY (parent_sid, threshold, value),
	FOREIGN key (parent_sid) REFERENCES lookup_table (sid) ON DELETE cascade
);
CREATE TABLE IF NOT EXISTS lookup_table_description (
	parent_sid INT,
	description TEXT,
	language_id VARCHAR(255),
	PRIMARY KEY (parent_sid, language_id),
	FOREIGN key (parent_sid) REFERENCES lookup_table (sid) ON DELETE cascade
);
-- Measurement Unit
CREATE TABLE IF NOT EXISTS measurement_unit (
	measurement_unit_code CHAR(3) PRIMARY KEY,
	date_start DATE NOT NULL,
	date_end DATE,
	national SMALLINT CHECK (national IN (0, 1)),
	national_abbreviation VARCHAR(200),
	change_type CHAR(1) NOT NULL
);
CREATE TABLE IF NOT EXISTS measurement_unit_description (
	parent_unit_code CHAR(3) NOT NULL REFERENCES measurement_unit(measurement_unit_code) ON DELETE CASCADE,
	description TEXT NOT NULL,
	language_id CHAR(2) NOT NULL,
	national SMALLINT CHECK (national IN (0, 1)),
	CONSTRAINT unique_description_per_language UNIQUE (parent_unit_code, language_id)
);
CREATE INDEX IF NOT EXISTS idx_measurement_unit_description_language ON measurement_unit_description (language_id);
-- Measurement Unit Qualifier
CREATE TABLE IF NOT EXISTS measurement_unit_qualifier (
	measurement_unit_qualifier_code VARCHAR(255) PRIMARY KEY,
	change_type VARCHAR(255),
	date_start TIMESTAMP,
	national INT
);
CREATE TABLE IF NOT EXISTS measurement_unit_qualifier_description (
	parent_measurement_unit_qualifier_code VARCHAR(255),
	description TEXT,
	language_id VARCHAR(255),
	national INT,
	PRIMARY KEY (
		parent_measurement_unit_qualifier_code,
		language_id
	),
	FOREIGN key (parent_measurement_unit_qualifier_code) REFERENCES measurement_unit_qualifier (measurement_unit_qualifier_code)
);
-- Meursing Additional Code
CREATE TABLE IF NOT EXISTS meursing_additional_code (
	meursing_table_plan_id SERIAL PRIMARY KEY,
	additional_code_id TEXT NOT NULL,
	date_start DATE NOT NULL,
	date_end DATE,
	national INTEGER NOT NULL,
	change_type TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS meursing_table_cell_component (
	id SERIAL PRIMARY KEY,
	parent_table_plan_id INTEGER NOT NULL REFERENCES meursing_additional_code (meursing_table_plan_id) ON DELETE CASCADE,
	heading_number INTEGER NOT NULL,
	row_column_code INTEGER NOT NULL,
	subheading_sequence_number INTEGER NOT NULL,
	date_start DATE NOT NULL,
	national INTEGER NOT NULL,
	UNIQUE (
		parent_table_plan_id,
		heading_number,
		row_column_code,
		subheading_sequence_number
	)
);
CREATE INDEX IF NOT EXISTS idx_meursing_additional_code_id ON meursing_additional_code (meursing_table_plan_id);
CREATE INDEX IF NOT EXISTS idx_meursing_table_cell_component_parent_id ON meursing_table_cell_component (parent_table_plan_id);
-- Meursing heading
CREATE TABLE IF NOT EXISTS meursing_heading (
	heading_number INT NOT NULL,
	meursing_table_plan_id INT NOT NULL,
	row_column_code INT NOT NULL,
	date_start DATE NOT NULL,
	national INT NOT NULL,
	change_type TEXT NOT NULL,
	PRIMARY KEY (
		heading_number,
		meursing_table_plan_id,
		row_column_code
	)
);
CREATE TABLE IF NOT EXISTS meursing_heading_footnote_association (
	id SERIAL PRIMARY KEY,
	parent_heading_id INT NOT NULL,
	footnote_id INT NOT NULL,
	footnote_type TEXT NOT NULL,
	date_start DATE,
	national INT,
	UNIQUE (parent_heading_id, footnote_id, footnote_type)
);
CREATE TABLE IF NOT EXISTS meursing_heading_text (
	id SERIAL PRIMARY KEY,
	parent_heading_id INT NOT NULL,
	description TEXT,
	language_id TEXT NOT NULL,
	national INT NOT NULL,
	UNIQUE (parent_heading_id, language_id)
);
CREATE INDEX IF NOT EXISTS idx_meursing_heading_table_plan ON meursing_heading (meursing_table_plan_id);
CREATE INDEX IF NOT EXISTS idx_meursing_heading_footnote_parent ON meursing_heading_footnote_association (parent_heading_id);
CREATE INDEX IF NOT EXISTS idx_meursing_heading_text_parent ON meursing_heading_text (parent_heading_id);
-- Meursing subheading
CREATE TABLE IF NOT EXISTS meursing_subheading (
	id SERIAL PRIMARY KEY,
	heading_number INT NOT NULL,
	meursing_table_plan_id INT NOT NULL,
	row_column_code INT NOT NULL,
	subheading_sequence_number INT NOT NULL,
	date_start DATE,
	description TEXT,
	national INT,
	change_type TEXT,
	UNIQUE (
		heading_number,
		meursing_table_plan_id,
		row_column_code,
		subheading_sequence_number
	)
);
CREATE INDEX IF NOT EXISTS idx_meursing_subheading_heading_plan ON meursing_subheading (heading_number, meursing_table_plan_id);
CREATE INDEX IF NOT EXISTS idx_meursing_subheading_row_column ON meursing_subheading (row_column_code);
-- Meursing table plan
CREATE TABLE IF NOT EXISTS meursing_table_plan (
	meursing_table_plan_id INT PRIMARY KEY,
	date_start DATE,
	national INT,
	change_type TEXT
);
CREATE INDEX IF NOT EXISTS idx_meursing_table_plan_id ON meursing_table_plan (meursing_table_plan_id);
-- Preference code
CREATE TABLE IF NOT EXISTS preference_code (
	pref_code INT PRIMARY KEY,
	date_start DATE,
	change_type TEXT
);
CREATE TABLE IF NOT EXISTS preference_code_description (
	id SERIAL PRIMARY KEY,
	parent_pref_code INT REFERENCES preference_code (pref_code) ON DELETE CASCADE,
	description TEXT,
	language_id TEXT NOT NULL,
	UNIQUE (parent_pref_code, language_id)
);
CREATE INDEX IF NOT EXISTS idx_preference_code ON preference_code (pref_code);
CREATE INDEX IF NOT EXISTS idx_preference_code_description_parent ON preference_code_description (parent_pref_code);
-- Quotas
CREATE TABLE IF NOT EXISTS quota_definition (
	sid INT PRIMARY KEY,
	sid_quota_order_number INT,
	quota_critical_state_code TEXT,
	quota_critical_threshold INT,
	quota_maximum_precision INT,
	quota_order_number INT,
	change_type TEXT,
	description TEXT,
	initial_volume FLOAT,
	measurement_unit_code TEXT,
	measurement_unit_qualifier_code TEXT,
	monetary_unit_code TEXT,
	volume FLOAT,
	date_start DATE,
	date_end DATE,
	national INT
);
CREATE INDEX IF NOT EXISTS idx_quota_definition_sid_quota_order_number ON quota_definition (sid_quota_order_number);
CREATE TABLE IF NOT EXISTS quota_suspension_period (
	sid INT PRIMARY KEY,
	parent_sid INT NOT NULL,
	description TEXT,
	date_start DATE,
	date_end DATE,
	national INT,
	FOREIGN KEY (parent_sid) REFERENCES quota_definition (sid) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_quota_suspension_period_sid ON quota_suspension_period (sid);
CREATE TABLE IF NOT EXISTS quota_blocking_period (
	sid INT PRIMARY KEY,
	parent_sid INT NOT NULL,
	blocking_period_type INT,
	description TEXT,
	date_start DATE,
	date_end DATE,
	national INT,
	FOREIGN KEY (parent_sid) REFERENCES quota_definition (sid) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_quota_blocking_period_sid ON quota_blocking_period (sid);
CREATE TABLE IF NOT EXISTS quota_association (
	sid_sub_quota INT PRIMARY KEY,
	parent_sid INT NOT NULL,
	relation_type TEXT,
	coefficient FLOAT,
	national INT,
	FOREIGN KEY (parent_sid) REFERENCES quota_definition (sid) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_quota_association_sid_sub_quota ON quota_association (sid_sub_quota);
CREATE INDEX IF NOT EXISTS idx_quota_association_relation_type ON quota_association (relation_type);
-- Base Regulation
CREATE TABLE IF NOT EXISTS base_regulation (
	regulation_id VARCHAR(255) PRIMARY KEY,
	antidumping_regulation_id VARCHAR(255),
	antidumping_regulation_role_type INT,
	change_type VARCHAR(255),
	community_code INT,
	date_end TIMESTAMP,
	date_published TIMESTAMP,
	date_start TIMESTAMP,
	description TEXT,
	effective_end_date TIMESTAMP,
	journal_page INT,
	national INT,
	official_journal_id VARCHAR(255),
	regulation_approved_flag INT,
	regulation_group VARCHAR(255),
	regulation_role_type INT,
	replacement_indicator INT,
	stopped_flag INT,
	url TEXT
);
CREATE INDEX if NOT EXISTS idx_base_regulation_date_end ON base_regulation (date_end);
-- Modification Regulation
CREATE TABLE IF NOT EXISTS modification_regulation (
	modification_regulation_id VARCHAR(255) PRIMARY KEY,
	base_regulation_id VARCHAR(255),
	base_regulation_role_type INT,
	change_type VARCHAR(255),
	date_end TIMESTAMP,
	date_published TIMESTAMP,
	date_start TIMESTAMP,
	description TEXT,
	effective_end_date TIMESTAMP,
	journal_page INT,
	modification_regulation_role_type INT,
	national INT,
	official_journal_id VARCHAR(255),
	regulation_approved_flag INT,
	replacement_indicator INT,
	stopped_flag INT
);
CREATE INDEX if NOT EXISTS idx_modification_regulation_date_end ON modification_regulation (date_end);
-- Full Temporary Stop Regulation
CREATE TABLE IF NOT EXISTS full_temporary_stop_regulation (
	fts_regulation_id VARCHAR(255) PRIMARY KEY,
	change_type VARCHAR(255),
	date_end TIMESTAMP,
	date_published TIMESTAMP,
	date_start TIMESTAMP,
	description TEXT,
	effective_end_date TIMESTAMP,
	fts_regulation_role_type INT,
	journal_page INT,
	national INT,
	official_journal_id VARCHAR(255),
	regulation_approved_flag INT,
	replacement_indicator INT,
	stopped_flag INT
);
-- Full Temporary Stop Regulation Action
CREATE TABLE IF NOT EXISTS full_temporary_stop_regulation_action (
	fts_regulation_id VARCHAR(255),
	stopped_regulation_id VARCHAR(255),
	stopped_regulation_role_type INT,
	national INT,
	PRIMARY KEY (
		fts_regulation_id,
		stopped_regulation_id,
		stopped_regulation_role_type
	),
	FOREIGN key (fts_regulation_id) REFERENCES full_temporary_stop_regulation (fts_regulation_id)
);
-- Inserted files to validate if a file should be processed or not. 
CREATE TABLE IF NOT EXISTS inserted_files (
	file_name VARCHAR(255) PRIMARY KEY,
	date_inserted date,
	time_taken TIME,
	file_size FLOAT
);