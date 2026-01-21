-- Rollback departments table creation

-- Remove department permissions
DELETE FROM permissions WHERE resource = 'departments';

-- Drop table (cascades to all dependent objects)
DROP TABLE IF EXISTS departments CASCADE;
