-- Create readiness_metrics table
CREATE TABLE IF NOT EXISTS readiness_metrics (
    id SERIAL PRIMARY KEY,
    oura_id VARCHAR(255) UNIQUE NOT NULL,
    day DATE NOT NULL,
    score INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_readiness_day ON readiness_metrics(day);
CREATE INDEX IF NOT EXISTS idx_readiness_oura_id ON readiness_metrics(oura_id);
