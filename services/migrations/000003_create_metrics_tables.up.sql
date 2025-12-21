-- Create sleep_metrics table
CREATE TABLE IF NOT EXISTS sleep_metrics (
    id SERIAL PRIMARY KEY,
    oura_id VARCHAR(255) UNIQUE NOT NULL,
    day DATE NOT NULL,
    score INTEGER,
    duration INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sleep_day ON sleep_metrics(day);
CREATE INDEX IF NOT EXISTS idx_sleep_oura_id ON sleep_metrics(oura_id);
