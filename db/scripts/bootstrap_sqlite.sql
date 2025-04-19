-- @COMMAND
DROP TABLE IF EXISTS posts;
-- @COMMAND
DROP TABLE IF EXISTS notes;
-- @COMMAND
DROP TABLE IF EXISTS users;
-- @COMMAND
CREATE TABLE users (
    id integer primary key autoincrement,
    guid text not null,
    name text,
    email text,
    picture text);
-- @COMMAND
CREATE TABLE notes (
    id integer primary key autoincrement,
    owner_id integer,
    contents text,
    foreign key(owner_id) references users(id));
-- @COMMAND
CREATE TABLE posts (
    id integer primary key autoincrement,
    author_id integer,
    contents text,
    foreign key(author_id) references users(id));
