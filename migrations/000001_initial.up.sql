CREATE TABLE IF NOT EXISTS urls (
    id serial PRIMARY KEY,
    short text NOT NULL UNIQUE,
    original text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
); 

CREATE UNIQUE INDEX idx_urls_short ON urls(short);