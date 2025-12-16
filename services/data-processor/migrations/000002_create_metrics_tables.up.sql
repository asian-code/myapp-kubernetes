-- Create metrics tables for storing Oura Ring data (single-user, matches InitSchema)
CREATE TABLE IF NOT EXISTS sleep_metrics (
    id SERIAL PRIMARY KEY,
    oura_id VARCHAR(255) UNIQUE,
    day DATE NOT NULL,
    score INTEGER,
    duration INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS activity_metrics (
    id SERIAL PRIMARY KEY,
    oura_id VARCHAR(255) UNIQUE,
    day DATE NOT NULL,
    score INTEGER,
    active_calories INTEGER,
    steps INTEGER,
    medium_activity_minutes INTEGER,
    high_activity_minutes INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS readiness_metrics (
    id SERIAL PRIMARY KEY,
    oura_id VARCHAR(255) UNIQUE,
    day DATE NOT NULL,
    score INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for faster queries
CREATE INDEX IF NOT EXISTS idx_sleep_day ON sleep_metrics(day DESC);
CREATE INDEX IF NOT EXISTS idx_activity_day ON activity_metrics(day DESC);
CREATE INDEX IF NOT EXISTS idx_readiness_day ON readiness_metrics(day DESC);
