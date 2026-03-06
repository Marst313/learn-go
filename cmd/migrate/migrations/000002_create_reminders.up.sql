DROP TABLE IF EXISTS reminders;

CREATE TABLE reminders(
    id SERIAL PRIMARY KEY,
    user_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    date DATE NOT NULL, 
    time TIME NOT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'personal'
        CHECK (category IN ('personal','work','health','shopping','other')),
    priority VARCHAR(20) NOT NULL DEFAULT 'medium'
        CHECK (priority IN ('low','medium','high')),
    recurring BOOLEAN NOT NULL DEFAULT FALSE,
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE INDEX idx_reminders_user_id ON reminders(user_id);

CREATE INDEX idx_reminders_date ON reminders(date);

CREATE INDEX idx_reminders_completed ON reminders(is_completed);