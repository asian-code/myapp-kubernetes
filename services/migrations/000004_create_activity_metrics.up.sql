-- Create activity_metrics table
CREATE TABLE IF NOT EXISTS activity_metrics (
    id SERIAL PRIMARY KEY,
    oura_id VARCHAR(255) UNIQUE NOT NULL,
    day DATE NOT NULL,
    score INTEGER,
    active_calories INTEGER,
    steps INTEGER,
    medium_activity_minutes INTEGER,
    high_activity_minutes INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_activity_day ON activity_metrics(day);
CREATE INDEX IF NOT EXISTS idx_activity_oura_id ON activity_metrics(oura_id);
