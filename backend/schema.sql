
CREATE TABLE users (
  telegram_id BIGINT PRIMARY KEY,
  name TEXT,
  joined_at TIMESTAMP DEFAULT NOW()
);
CREATE TABLE babies (
  baby_id SERIAL PRIMARY KEY,
  name TEXT,
  birth_date DATE
);
CREATE TABLE baby_access (
  baby_id INT REFERENCES babies(baby_id) ON DELETE CASCADE,
  telegram_id BIGINT REFERENCES users(telegram_id) ON DELETE CASCADE,
  role TEXT CHECK (role IN ('admin', 'parent')),
  PRIMARY KEY (baby_id, telegram_id)
);
CREATE TABLE events (
  event_id SERIAL PRIMARY KEY,
  baby_id INT REFERENCES babies(baby_id) ON DELETE CASCADE,
  type TEXT NOT NULL,
  time_str TEXT,
  timestamp BIGINT NOT NULL
);
CREATE TABLE invites (
  code TEXT PRIMARY KEY,
  baby_id INT REFERENCES babies(baby_id) ON DELETE CASCADE,
  created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_events_baby ON events(baby_id);
CREATE INDEX idx_access_user ON baby_access(telegram_id);
