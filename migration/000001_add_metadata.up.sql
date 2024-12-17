CREATE TABLE IF NOT EXISTS metadata (
    id SERIAL PRIMARY KEY,
    filename TEXT NOT NULL,
    file_size INT,
    checksum TEXT,
    action TEXT NOT NULL,
    uploaded_at TIMESTAMPTZ DEFAULT NOW()
);