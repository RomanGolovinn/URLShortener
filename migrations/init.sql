CREATE TABLE IF NOT EXISTS links (
    short_code VARCHAR(10) PRIMARY KEY,
    url TEXT NOT NULL,
    number_of_transitions BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_transition TIMESTAMP WITH TIME ZONE
);