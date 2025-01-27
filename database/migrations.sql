CREATE TABLE playback_history (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(50) NOT NULL,
    category VARCHAR(50) NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE ad_click_history (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(50) NOT NULL,
    ad_id VARCHAR(50) NOT NULL,
    category VARCHAR(50),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE users (
    user_id VARCHAR(50) PRIMARY KEY,
    preferences TEXT[],
    playback_history JSONB,
    ad_click_history JSONB
);
