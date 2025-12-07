-- Create metrics tables for storing Oura Ring data
CREATE TABLE IF NOT EXISTS sleep_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    score INTEGER,
    duration INTEGER,
    deep_sleep INTEGER,
    rem_sleep INTEGER,
    light_sleep INTEGER,
    awake_time INTEGER,
    bedtime_start TIMESTAMP WITH TIME ZONE,
    bedtime_end TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, date)
);

CREATE TABLE IF NOT EXISTS activity_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    score INTEGER,
    active_calories INTEGER,
    total_calories INTEGER,
    steps INTEGER,
    distance_meters INTEGER,
    medium_activity_minutes INTEGER,
    high_activity_minutes INTEGER,
    inactive_time INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, date)
);

CREATE TABLE IF NOT EXISTS readiness_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    score INTEGER,
    temperature_deviation FLOAT,
    temperature_trend_deviation FLOAT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, date)
);

-- Create indexes for faster queries
CREATE INDEX idx_sleep_metrics_user_date ON sleep_metrics(user_id, date DESC);
CREATE INDEX idx_activity_metrics_user_date ON activity_metrics(user_id, date DESC);
CREATE INDEX idx_readiness_metrics_user_date ON readiness_metrics(user_id, date DESC);
