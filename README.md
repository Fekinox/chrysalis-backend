# Chrysalis

A re-implementation of Chrysalis.

Chrysalis is a tool that can be used by freelance workers to provide status
updates on their tasks.

## About

Users can create accounts by providing their username and password. They can create forms using a Google Forms-like interface to build it out of various form components, and send those forms to other users. Users can fill out those forms, and the creator of the form can organize all the tasks they have.

## Usage

Copy the default configuration located at `default.env` to `.env`, and configure it to your liking.

Start up the backend infrastructure with `docker compose -f infra.yaml up`, then
start up the main server with `go run main.go`

## Technology Used

| Technology | Usage                                                                |
|------------|----------------------------------------------------------------------|
| Gin        | HTTP framework and muxer for Go.                                     |
| Nginx      | Reverse proxy                                                        |
| PostgreSQL | Relational database for persistent data storage.                     |
| KeyDB      | In-memory database for session storage.                              |
| sqlc       | Code generator that converts SQL queries into type-safe Go code.     |
| pgx        | PostgreSQL driver libraries for Go.                                  |
| HTMX       | Framework for being able to make AJAX calls directly from HTML.      |
| Alpine.js  | Lightweight framework for client-side scripting                      |
| MVP.css    | Basic bootstrap styling, to be replaced with Tailwind in the future. |
