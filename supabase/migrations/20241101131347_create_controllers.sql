CREATE TABLE IF NOT EXISTS controllers(
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT NOT NULL,
    initials CHAR(2) NOT NULL,
    email TEXT UNIQUE NOT NULL,
    role_id INT REFERENCES roles(id) NOT NULL,
    facility_id INT REFERENCES facilities(id) NOT NULL,
    schedule_id INT REFERENCES schedule(id)
);