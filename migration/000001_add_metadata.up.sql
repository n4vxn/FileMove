CREATE TABLE IF NOT EXISTS upload_metadata (
    id SERIAL PRIMARY KEY,
    filename TEXT NOT NULL,
    file_size INT,
    checksum TEXT,
    action TEXT NOT NULL,
    uploaded_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS download_metadata (
    id SERIAL PRIMARY KEY,
    filename TEXT NOT NULL,
    file_size INT,
    checksum TEXT,
    action TEXT NOT NULL,
    uploaded_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);