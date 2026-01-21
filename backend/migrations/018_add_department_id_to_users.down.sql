-- Rollback department_id column from users table

-- Remove index
DROP INDEX IF EXISTS idx_users_department_id;

-- Remove foreign key constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_department;

-- Remove column
ALTER TABLE users DROP COLUMN IF EXISTS department_id;
