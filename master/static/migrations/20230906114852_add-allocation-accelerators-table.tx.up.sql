CREATE TABLE public.allocation_accelerators (
    container_id text NOT NULL PRIMARY KEY,
    task_id text NOT NULL REFERENCES public.tasks(task_id),
    allocation_id text NOT NULL REFERENCES public.allocations(allocation_id),
    node_name text NOT NULL,
    accelerator_type text NOT NULL,
    accelerators JSONB NOT NULL
);
