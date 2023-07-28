\set mypid `echo $PPID`
\set logfile add_summary_metrics_:mypid.log
\set logpipe | cat >> add_summary_metrics_:mypid.log
\echo PID :mypid logging to :logfile
\o :logfile
\! sleep 1

BEGIN;
\echo `date +%F_%T` [:mypid] Initial setup
\qecho `date +%F_%T` [:mypid] Initial setup
\! sleep 1

CREATE TEMPORARY TABLE trial_ids (
  id INT
);

SELECT setseed(0.5);
WITH randomized_ids AS (
  SELECT id, NTILE(:number_workers) OVER (ORDER BY RANDOM()) AS bucket_number
  FROM trials
)
INSERT INTO trial_ids
SELECT id
FROM randomized_ids
WHERE bucket_number = :worker_index;

SELECT
    :mypid AS PID,
    :worker_index AS worker_index,
    :number_workers AS number_total_workers,
    COUNT(*) AS number_trials
FROM trial_ids;

\o
SELECT
    :mypid aS PID,
    :worker_index AS worker_index,
    :number_workers AS number_total_workers,
    COUNT(*) AS number_trials
FROM trial_ids;
\o :logpipe

-- displays this worker's list of trials to touch
SELECT id AS trial_id FROM trial_ids ORDER BY id;


-- Validations.
\echo `date +%F_%T` [:mypid] Validations - Create temp tables and indices
\qecho `date +%F_%T` [:mypid] Validations - Create temp tables and indices

CREATE TEMPORARY TABLE val_metric_values (
  id SERIAL,
  trial_id INT,
  name TEXT,
  value TEXT,
  type TEXT,
  end_time timestamptz
);

CREATE TEMPORARY TABLE val_numeric_aggs (
  id SERIAL,
  trial_id INT,
  name TEXT,
  count INT,
  sum FLOAT8,
  min FLOAT8,
  max FLOAT8
);

CREATE TEMPORARY TABLE val_metric_types (
  id SERIAL,
  trial_id INT,
  name TEXT,
  type TEXT
);

CREATE INDEX val_metric_types_idx ON val_metric_types (trial_id, name);

CREATE TEMPORARY TABLE val_metric_latest (
  id SERIAL,
  trial_id INT,
  name TEXT,
  value jsonb
);

CREATE INDEX metric_latest_idx ON val_metric_latest (trial_id, name);

CREATE TEMPORARY TABLE val_summary_metrics (
  id SERIAL,
  trial_id INT,
  summary_metrics JSONB
);

-- Extract training metrics.
\echo `date +%F_%T` [:mypid] Validations - Extract training metrics
\qecho `date +%F_%T` [:mypid] Validations - Extract training metrics

INSERT INTO val_metric_values(trial_id, name, value, type, end_time)
SELECT
    trial_id AS trial_id,
    key AS name,
    CASE value
        WHEN '"NaN"' THEN 'NaN'
        WHEN '"Infinity"' THEN 'Infinity'
        WHEN '"-Infinity"' THEN '-Infinity'
        ELSE value::text
    END AS value,
    CASE
        WHEN jsonb_typeof(value) = 'string' THEN
            CASE
                WHEN value::text = '"Infinity"'::text THEN 'number'
                WHEN value::text = '"-Infinity"'::text THEN 'number'
                WHEN value::text = '"NaN"'::text THEN 'number'
                WHEN value::text ~
                    '^"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})?"$' THEN 'date'
                ELSE 'string'
            END
        ELSE jsonb_typeof(value)::text
    END AS type,
    end_time AS end_time
FROM (
    SELECT
        validations.trial_id,
        (jsonb_each(metrics->'validation_metrics')).key,
        (jsonb_each(metrics->'validation_metrics')).value,
        validations.end_time
    FROM validations
    JOIN trials ON trials.id = validations.trial_id
    WHERE trials.id IN (SELECT id FROM trial_ids)
) AS subquery;

-- Numeric aggregates.
\echo `date +%F_%T` [:mypid] Validations - Numeric aggregates
\qecho `date +%F_%T` [:mypid] Validations - Numeric aggregates

INSERT INTO val_numeric_aggs(trial_id, name, count, sum, min, max)
SELECT
    trial_id AS trial_id,
    name AS name,
    COUNT(*) AS count,
    safe_sum(value::double precision) AS sum,
    MIN(value::double precision) AS min,
    MAX(value::double precision) AS max
FROM val_metric_values
WHERE type = 'number'
GROUP BY trial_id, name;

-- Types.
\echo `date +%F_%T` [:mypid] Validations - Types
\qecho `date +%F_%T` [:mypid] Validations - Types

INSERT INTO val_metric_types(trial_id, name, type)
SELECT
    trial_id AS trial_id,
    name AS name,
    CASE
        WHEN COUNT(DISTINCT type) = 1 THEN MAX(type)
        ELSE 'string'
    END AS type
FROM val_metric_values
GROUP BY trial_id, name;

-- Latest.
\echo `date +%F_%T` [:mypid] Validations - Latest
\qecho `date +%F_%T` [:mypid] Validations - Latest

INSERT INTO val_metric_latest(trial_id, name, value)
SELECT
    s.trial_id AS trial_id,
    unpacked.key as name,
    unpacked.value as value
FROM (
    SELECT s.*,
        ROW_NUMBER() OVER(
            PARTITION BY s.trial_id
            ORDER BY s.end_time DESC
        ) as rank
    FROM validations s
    JOIN trials ON s.trial_id = trials.id
    WHERE trials.id IN (SELECT id FROM trial_ids)
) s, jsonb_each(s.metrics->'validation_metrics') unpacked
WHERE s.rank = 1;

-- Summary metrics.
\echo `date +%F_%T` [:mypid] Validations - Summary metrics (this may take quite a while)
\qecho `date +%F_%T` [:mypid] Validations - Summary metrics (this may take quite a while)

INSERT INTO val_summary_metrics(trial_id, summary_metrics)
SELECT
    trial_id, jsonb_collect(jsonb_build_object(
        name, jsonb_build_object(
        'count', CASE WHEN sub.type = 'number' THEN sub.count ELSE 0 END,
        'sum', CASE WHEN sub.type = 'number' THEN sub.sum ELSE 0 END,
        'min', CASE WHEN sub.type = 'number' THEN
            CASE WHEN sub.max = 'NaN'::double precision
                THEN 'NaN'::double precision ELSE sub.min END
            ELSE 0 END,
        'max', CASE WHEN sub.type = 'number' THEN sub.max ELSE 0 END,
        'last', sub.latest,
        'type', sub.type
    )
)) as summary_metrics
FROM (SELECT
    val_metric_types.trial_id,
    val_metric_types.name,
    count,
    sum,
    min,
    max,
    val_metric_types.type AS type,
    val_metric_latest.value AS latest
FROM val_metric_types
LEFT JOIN val_numeric_aggs ON
     val_numeric_aggs.trial_id = val_metric_types.trial_id AND
     val_numeric_aggs.name = val_metric_types.name
LEFT JOIN val_metric_latest ON
     val_metric_types.trial_id = val_metric_latest.trial_id AND
     val_metric_types.name = val_metric_latest.name) sub
GROUP BY trial_id;

-- Training.
\echo `date +%F_%T` [:mypid] Training - Create temp tables and indices
\qecho `date +%F_%T` [:mypid] Training - Create temp tables and indices

CREATE TEMPORARY TABLE train_metric_values (
  id SERIAL,
  trial_id INT,
  name TEXT,
  value TEXT,
  type TEXT,
  end_time timestamptz
);

CREATE TEMPORARY TABLE train_numeric_aggs (
  id SERIAL,
  trial_id INT,
  name TEXT,
  count INT,
  sum FLOAT8,
  min FLOAT8,
  max FLOAT8
);

CREATE TEMPORARY TABLE train_metric_types (
  id SERIAL,
  trial_id INT,
  name TEXT,
  type TEXT
);

CREATE INDEX train_metric_types_idx ON train_metric_types (trial_id, name);

CREATE TEMPORARY TABLE train_metric_latest (
  id SERIAL,
  trial_id INT,
  name TEXT,
  value jsonb
);

CREATE INDEX train_metric_latest_idx ON train_metric_latest (trial_id, name);

CREATE TEMPORARY TABLE train_summary_metrics (
  id SERIAL,
  trial_id INT,
  summary_metrics JSONB
);

-- Extract training metrics.
\echo `date +%F_%T` [:mypid] Training - Extract training metrics
\qecho `date +%F_%T` [:mypid] Training - Extract training metrics

INSERT INTO train_metric_values(trial_id, name, value, type, end_time)
SELECT
    trial_id AS trial_id,
    key AS name,
    CASE value
        WHEN '"NaN"' THEN 'NaN'
        WHEN '"Infinity"' THEN 'Infinity'
        WHEN '"-Infinity"' THEN '-Infinity'
        ELSE value::text
    END AS value,
    CASE
        WHEN jsonb_typeof(value) = 'string' THEN
            CASE
                WHEN value::text = '"Infinity"'::text THEN 'number'
                WHEN value::text = '"-Infinity"'::text THEN 'number'
                WHEN value::text = '"NaN"'::text THEN 'number'
                WHEN value::text ~
                    '^"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})?"$' THEN 'date'
                ELSE 'string'
            END
        ELSE jsonb_typeof(value)::text
    END AS type,
    end_time AS end_time
FROM (
    SELECT
        steps.trial_id,
        (jsonb_each(metrics->'avg_metrics')).key,
        (jsonb_each(metrics->'avg_metrics')).value,
        steps.end_time
    FROM steps
    JOIN trials ON trials.id = steps.trial_id
    WHERE trials.id IN (SELECT id FROM trial_ids)
) AS subquery;

-- Numeric aggregates.
\echo `date +%F_%T` [:mypid] Training - Numeric aggregates
\qecho `date +%F_%T` [:mypid] Training - Numeric aggregates

INSERT INTO train_numeric_aggs(trial_id, name, count, sum, min, max)
SELECT
    trial_id AS trial_id,
    name AS name,
    COUNT(*) AS count,
    safe_sum(value::double precision) AS sum,
    MIN(value::double precision) AS min,
    MAX(value::double precision) AS max
FROM train_metric_values
WHERE type = 'number'
GROUP BY trial_id, name;

-- Types.
\echo `date +%F_%T` [:mypid] Training - Types
\qecho `date +%F_%T` [:mypid] Training - Types

INSERT INTO train_metric_types(trial_id, name, type)
SELECT
    trial_id AS trial_id,
    name AS name,
    CASE
        WHEN COUNT(DISTINCT type) = 1 THEN MAX(type)
        ELSE 'string'
    END AS type
FROM train_metric_values
GROUP BY trial_id, name;

-- Latest.
\echo `date +%F_%T` [:mypid] Training - Latest
\qecho `date +%F_%T` [:mypid] Training - Latest

INSERT INTO train_metric_latest(trial_id, name, value)
SELECT
    s.trial_id AS trial_id,
    unpacked.key as name,
    unpacked.value as value
FROM (
    SELECT s.*,
        ROW_NUMBER() OVER(
            PARTITION BY s.trial_id
            ORDER BY s.end_time DESC
        ) as rank
    FROM steps s
    JOIN trials ON s.trial_id = trials.id
    WHERE trials.id IN (SELECT id FROM trial_ids)
) s, jsonb_each(s.metrics->'avg_metrics') unpacked
WHERE s.rank = 1;

-- Summary metrics.
\echo `date +%F_%T` [:mypid] Training - Summary metrics (this may take quite a while)
\qecho `date +%F_%T` [:mypid] Training - Summary metrics (this may take quite a while)

INSERT INTO train_summary_metrics(trial_id, summary_metrics)
SELECT
    trial_id, jsonb_collect(jsonb_build_object(
        name, jsonb_build_object(
        'count', CASE WHEN sub.type = 'number' THEN sub.count ELSE 0 END,
        'sum', CASE WHEN sub.type = 'number' THEN sub.sum ELSE 0 END,
        'min', CASE WHEN sub.type = 'number' THEN
            CASE WHEN sub.max = 'NaN'::double precision
                THEN 'NaN'::double precision ELSE sub.min END
            ELSE 0 END,
        'max', CASE WHEN sub.type = 'number' THEN sub.max ELSE 0 END,
        'last', sub.latest,
        'type', sub.type
    )
)) as summary_metrics
FROM (SELECT
    train_metric_types.trial_id,
    train_metric_types.name,
    count,
    sum,
    min,
    max,
    train_metric_types.type AS type,
    train_metric_latest.value AS latest
FROM train_metric_types
LEFT JOIN train_numeric_aggs ON
     train_numeric_aggs.trial_id = train_metric_types.trial_id AND
     train_numeric_aggs.name = train_metric_types.name
LEFT JOIN train_metric_latest ON
     train_metric_types.trial_id = train_metric_latest.trial_id AND
     train_metric_types.name = train_metric_latest.name) sub
GROUP BY trial_id;

UPDATE trials SET
    summary_metrics = (CASE
        WHEN tsm.summary_metrics IS NOT NULL AND vsm.summary_metrics IS NOT NULL THEN
            jsonb_build_object(
                'avg_metrics', tsm.summary_metrics,
                'validation_metrics', vsm.summary_metrics
            )
        WHEN tsm.summary_metrics IS NOT NULL THEN
            jsonb_build_object(
                'avg_metrics', tsm.summary_metrics
            )
        WHEN vsm.summary_metrics IS NOT NULL THEN jsonb_build_object(
                'validation_metrics', vsm.summary_metrics
           )
        ELSE '{}'::jsonb END)
FROM train_summary_metrics tsm
FULL OUTER JOIN val_summary_metrics vsm ON tsm.trial_id = vsm.trial_id
WHERE coalesce(tsm.trial_id, vsm.trial_id) = trials.id;

\echo `date +%F_%T` [:mypid] Done. Clean-up and COMMIT
\qecho `date +%F_%T` [:mypid] Done. Clean-up and COMMIT

DROP TABLE trial_ids;

DROP TABLE train_metric_values;
DROP TABLE train_numeric_aggs;
DROP TABLE train_metric_types;
DROP TABLE train_metric_latest;
DROP TABLE train_summary_metrics;

DROP TABLE val_metric_values;
DROP TABLE val_numeric_aggs;
DROP TABLE val_metric_types;
DROP TABLE val_metric_latest;
DROP TABLE val_summary_metrics;

COMMIT;
