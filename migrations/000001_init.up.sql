CREATE TABLE links (
    short_code VARCHAR(10) PRIMARY KEY,
    long_url   VARCHAR(1000) NOT NULL UNIQUE
);