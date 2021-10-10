-- CREATE USER lera WITH ENCRYPTED PASSWORD '123456';
-- GRANT ALL PRIVILEGES ON DATABASE proxy_bd TO lera;

CREATE TABLE IF NOT EXISTS request
(
    id      serial NOT NULL PRIMARY KEY,
    method  text   NOT NULL,
    host    text   NOT NULL,
    url     text   NOT NULL,
    body    text   NOT NULL,
    headers jsonb  NOT NULL
);