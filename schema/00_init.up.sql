CREATE TABLE IF NOT EXISTS payments(
    id VARCHAR(255) PRIMARY KEY NOT NULL,
    version INT NOT NULL DEFAULT 0,
    organisation VARCHAR(255) NOT NULL,
    deleted INT DEFAULT 0,
    attributes TEXT NOT NULL
)
