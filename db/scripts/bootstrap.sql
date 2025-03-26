-- @COMMAND
DROP TABLE IF EXISTS notes;
-- @COMMAND
DROP TABLE IF EXISTS posts;
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
-- @COMMAND
-- posts table
CREATE TABLE posts(
    id SERIAL PRIMARY KEY,
    author_id INT references users(id),
    contents TEXT
);
--@COMMAND
CREATE INDEX posts_author_id on posts (author_id);
