-- Add task kind and root-cause fields for BE-8 task typing.

ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS kind TEXT;

ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS root_cause_analysis TEXT;

UPDATE tasks
SET kind = 'task'
WHERE kind IS NULL;

UPDATE tasks
SET root_cause_analysis = ''
WHERE root_cause_analysis IS NULL;

ALTER TABLE tasks
ALTER COLUMN kind SET DEFAULT 'task';

ALTER TABLE tasks
ALTER COLUMN root_cause_analysis SET DEFAULT '';

ALTER TABLE tasks
ALTER COLUMN kind SET NOT NULL;

ALTER TABLE tasks
ALTER COLUMN root_cause_analysis SET NOT NULL;

ALTER TABLE tasks
DROP CONSTRAINT IF EXISTS tasks_kind_check;

ALTER TABLE tasks
ADD CONSTRAINT tasks_kind_check CHECK (kind IN ('task', 'bug', 'feature', 'chore', 'spike'));

ALTER TABLE tasks
DROP CONSTRAINT IF EXISTS tasks_root_cause_analysis_check;

ALTER TABLE tasks
ADD CONSTRAINT tasks_root_cause_analysis_check CHECK (kind <> 'bug' OR root_cause_analysis <> '');

INSERT INTO schema_migrations (version)
VALUES ('0002_task_kind_root_cause')
ON CONFLICT (version) DO NOTHING;
