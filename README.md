# Glonk

## About
Almost an ORM. Uses reflection to generate sql queries & statements on arbitrary types.

The `glonk` annotation on types indicates the sql column names. Additionally data types may be private, specified with an `owner_id` sql column, or public with an `author_id` sql column.

Glonk requires an `id` and either `owner_id` or `author_id` column on each type.

Types with `owner_id` specified will only be accessible to an authenticated user with that id.

Uses google oauth2 for authentication. Sets `session_id` cookie after authenticating with expiration of 20 mins.

## Getting Started
Authenticate with google, then GET, POST, PUT, or DELETE data.

Janky frontent will not report any errors to the user - console might be helpful.

`owner_id` or `author_id` must be specified on each POST and PUT & must match user id.

## Misc.
data lives at `/data/{data_type}/{id}?{queries}`

queries and data types (poorly documented) live at `/schema`

PUT requests are sparse updates

there is no rate limiting, but I can and will wipe the db for any reason or on a whim

See also the in progress [rewrite in rust](https://github.com/reshane/authrs)

[namesake hint](https://en.wikipedia.org/wiki/Flanimals)
