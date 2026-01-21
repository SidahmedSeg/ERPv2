-- Add department_id column to users table

ALTER TABLE users ADD COLUMN department_id UUID;

-- Add foreign key constraint to departments
ALTER TABLE users ADD CONSTRAINT fk_users_department
    FOREIGN KEY (tenant_id, department_id)
    REFERENCES departments(tenant_id, id)
    ON DELETE SET NULL;

-- Add index for performance
CREATE INDEX idx_users_department_id ON users(tenant_id, department_id) WHERE department_id IS NOT NULL;

-- Comment
COMMENT ON COLUMN users.department_id IS 'Department membership - nullable FK to departments';
