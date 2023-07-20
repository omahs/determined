ALTER TABLE public.trials
    ADD COLUMN task_id text;

UPDATE trials
SET task_id = trial_id_task_id.task_id
FROM trial_id_task_id
WHERE trials.id = trial_id_task_id.trial_id;

ALTER TABLE public.trials
    ALTER COLUMN task_id SET NOT NULL,
    ADD CONSTRAINT fk_trials_tasks
    FOREIGN KEY (task_id)
    REFERENCES public.tasks(task_id);

DROP VIEW public.proto_checkpoints_view;
DROP VIEW public.checkpoints_view;

CREATE OR REPLACE VIEW public.checkpoints_view AS
    SELECT
        c.id AS id,
        c.uuid AS uuid,
        c.task_id,
        c.allocation_id,
        c.report_time,
        c.state,
        c.resources,
        c.metadata,
        t.id AS trial_id,
        e.id AS experiment_id,
        e.config AS experiment_config,
        t.hparams AS hparams,
        s.metrics AS training_metrics,
        v.metrics->'validation_metrics' AS validation_metrics,
        (v.metrics->'validation_metrics'->>(e.config->'searcher'->>'metric'))::float8 AS searcher_metric,
        CAST(c.metadata->>'steps_completed' AS int) as steps_completed,
        -- Removing checkpoint version since it doesn't make sense anymore.
        c.size
    FROM public.checkpoints_v2 AS c
    LEFT JOIN public.trials AS t on c.task_id = t.task_id
    LEFT JOIN public.experiments AS e on t.experiment_id = e.id
    LEFT JOIN public.raw_validations AS v on CAST(c.metadata->>'steps_completed' AS int) = v.total_batches and t.id = v.trial_id
    LEFT JOIN public.raw_steps AS s on CAST(c.metadata->>'steps_completed' AS int) = s.total_batches and t.id = s.trial_id
    -- avoiding the steps view causes Postgres to not "Materialize" in this join.
    WHERE s.archived IS NULL OR s.archived = false
      AND v.archived IS NULL OR v.archived = false;

CREATE OR REPLACE VIEW public.proto_checkpoints_view AS
    SELECT
        c.uuid::text AS uuid,
        c.task_id,
        c.allocation_id,
        c.report_time as report_time,
        'STATE_' || c.state AS state,
        c.resources,
        c.metadata,
        -- Build a training substruct for protobuf.
        jsonb_build_object(
            'trial_id', c.trial_id,
            'experiment_id', c.experiment_id,
            'experiment_config', c.experiment_config,
            'hparams', c.hparams,
            -- construct training metrics from the untyped jsonb deterministically, since older
            -- versions may have old keys (e.g., num_inputs) and our unmarshaling is strict.
            'training_metrics', jsonb_build_object(
                'avg_metrics', c.training_metrics->'avg_metrics',
                'batch_metrics', c.training_metrics->'batch_metrics'
            ),
            'validation_metrics', json_build_object('avg_metrics', c.validation_metrics),
            'searcher_metric', c.searcher_metric
        ) AS training
    FROM public.checkpoints_view AS c;

DROP TABLE public.trial_id_task_id;

