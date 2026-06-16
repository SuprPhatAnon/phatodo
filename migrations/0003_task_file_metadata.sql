-- Add planned and changed file metadata for BE-11.

ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS planned_files JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS changed_files JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE tasks
DROP CONSTRAINT IF EXISTS tasks_planned_files_array_check;

ALTER TABLE tasks
ADD CONSTRAINT tasks_planned_files_array_check CHECK (jsonb_typeof(planned_files) = 'array');

ALTER TABLE tasks
DROP CONSTRAINT IF EXISTS tasks_changed_files_array_check;

ALTER TABLE tasks
ADD CONSTRAINT tasks_changed_files_array_check CHECK (jsonb_typeof(changed_files) = 'array');

INSERT INTO schema_migrations (version)
VALUES ('0003_task_file_metadata')
ON CONFLICT (version) DO NOTHING;
