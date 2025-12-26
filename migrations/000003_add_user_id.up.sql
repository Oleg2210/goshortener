ALTER TABLE urls ADD COLUMN user_id TEXT NOT NULL;
ALTER TABLE urls ADD COLUMN is_deleted BOOLEAN DEFAULT false;

CREATE INDEX idx_urls_user_id ON urls(user_id);