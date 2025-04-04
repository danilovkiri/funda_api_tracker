# Telegram bot for receiving notifications upon new listings addition at Funda.nl

## Description

The bot polls the Funda API and sends the notification each time the polling was successfully run. The notification
contains the amount of listings (added and removed) as compared to the previous run. The bot can provide information on
these listings when asked to do so.

Each user has a session. Session exists in DB from `/start` till `/stop`, it is deleted with all data related to
session's UserID (session, search parameters and listings).

Each user can define a list of regions and cities for filtering. Regions can be set with `/set_regions` followed by a
comma-separated list of regions (case-insensitive). Cities can be set with `/set_cities` followed by a comma-separated
list of cities (case-insensitive). Both regions and cities can be reset with same commands without specifying a list of
regions/cities. Filtering by regions works in a way that listings pass filters if their corresponding attributes are
specified in regions and/or cities.

`/run` activates scheduled API polling until turned off by `/pause`. The schedule can be defined by `/set_polling_interval`
followed by a value of seconds between consecutive API pollings (integer, min 300, default 3600 if not set), the values
is stored in UpdateIntervalSeconds and can be altered any time. You can optionally trigger a manual listings update via
`/update_now`.

Each session has a list of currently stored listings in DB with their attributes (location, unique URL, prices, etc.),
these listings are updated each time the API polling is successful. At any given time you can invoke `/show_current_listings`
to see all currently stored listings from DB (price and URL), or `/show_new_listings` to see only those listings which
were added in the previous run.

Listings are retrieved using a search query which you must define by running `/set_search_query` followed by a URL (e.g.
this URL `https://www.funda.nl/en/zoeken/huur/?selected_area=[%22nl%22]&price=%221000-2000%22&object_type=[%22apartment%22,%22house%22]&publication_date=%221%22`).
You can change the search query any time by invoking the same command again with a new URL. Note that the next scheduled
or triggered API polling after the URL is changed will compare the newly retrieved listings with the currently stored in
DB (which were probably retrieved using another URL).


## Prerequisites

1. Your telegram bot token available as an ENV variable `TELEGRAM_BOT_TOKEN`
2. Comma-separated list of authorized users' usernames available as an ENV variable `TELEGRAM_USERS`
3. Optional: set logging level with `LOG_LEVEL` (0-3) 
4. Optional: set specific sqlite DB location with `SQLITE_DNS`, do not forget to add `?_loc=auto`

## Building

To build both bot and CLI applications, run:
```bash
make build
```

To prepare DB (required):
```bash
make migrate-up
```
## Running

To run bot:
```bash
make run-bot
```

To explicitly run migrations, execute:
```bash
./cmd/cli/app storage:migrate --direction up
./cmd/cli/app storage:migrate --direction down
```

To show all sessions in terminal, execute:
```bash
./cmd/cli/app manager:showSessions # shows all sessions
./cmd/cli/app manager:showSessions --onlyActive true # shows only active sessions 
```

To send a message to subscriber(s), execute:
```bash
./cmd/cli/app manager:sendMessage --message "Generic message" --userID genericUserName --chatID 0 # sends message to one user
./cmd/cli/app manager:sendMessage --message "Generic message" --sendToAll true # sends message to all user with active sessions
```