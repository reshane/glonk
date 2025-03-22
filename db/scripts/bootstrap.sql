-- @COMMAND
DROP TABLE IF EXISTS notes;
-- @COMMAND
DROP TABLE IF EXISTS users;
-- @COMMAND
-- users table
CREATE TABLE users(
    id SERIAL PRIMARY KEY,
    guid TEXT NOT NULL,
    name TEXT,
    email TEXT,
    picture TEXT
);
-- @COMMAND
CREATE UNIQUE INDEX user_guid_idx on users (guid);
-- @COMMAND
-- notes table
CREATE TABLE notes(
    id SERIAL PRIMARY KEY,
    owner_id INT references users(id),
    contents TEXT
);
--@COMMAND
CREATE INDEX notes_owner_id on notes (owner_id);
