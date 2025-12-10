CREATE TABLE IF NOT EXISTS urls (
    id serial PRIMARY KEY,
    short text NOT NULL UNIQUE,
    original text NOT NULL 
); 

CREATE UNIQUE INDEX idx_urls_short ON urls(short);