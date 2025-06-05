CREATE TYPE Frequency AS ENUM ('daily', 'hourly');

CREATE TABLE IF NOT EXISTS Subscriptions (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    frequency Frequency NOT NULL,
    city VARCHAR(255) NOT NULL,
    activated BOOLEAN NOT NULL,
    token UUID NOT NULL
)