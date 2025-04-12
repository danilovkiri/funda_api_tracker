# Telegram bot for receiving notifications upon new listings addition at Funda.nl

## Description

The bot functionality includes the following:

1. Polling https://www.funda.nl HTTP API using user-defined search-query by user-defined schedule and by manual trigger;
2. Storing and updating retrieved listings in a DB;
3. Sending messages to user containing statistics on each polling run;
4. Sending messages to user containing currently stored or newly added listings;
5. Filtering listings by user-defined regions and cities;
6. Adding listings to 'favorites';
7. Sending messages to user containing 'favorite' listings;
8. DND mode

## Key concepts

### Sessions

The bot operates within a user session, which is created upon `/start` and is removed upon `/stop` command, which also
deletes **all** data, accumulated or generated within that session.

The session stores user-defined parameters:
1. API polling interval — must be set with `set_polling_interval` followed by a valid duration string (e.g. `1000s`,
   `3m` `1h`, `1.5h`, `2h30m15s`, min allowed value is 900s). This parameter defined how often the scheduled API polling
   runs will be executed with data being sent to user. Can be changed at any moment. When changed, the next polling will
   commence if the newly defined interval has passed since the last polling.
2. API polling status — a boolean flag which either enables or disables scheduled API polling runs. API polling can be 
   turned on via `/run` and off via `/pause`. Invoking `/run` after a period of pause exceeding the API polling interval
   duration will trigger an immediate polling run.
3. Regions and cities used for listing filtering at the bot-level. Regions can be set via `/set_regions` followed by
   a comma-separated list of region names (case-insensitive) or via `/add_region` followed by a single region name. The
   same applies to city names via `/set_cities` and `/add_city`. These filters are applied only when sending data to user,
   not when collecting or storing data, thus can be changed any moment. To show the list of regions with corresponding
   top 5 cities run `/show_locations`.
4. DND mode state — a boolean flag which either enables or disables DND mode, which pauses any scheduled API polling
   runs. Can be turned on and off via `/dnd_activate` and `/dnd_deactivate`, respectfully. DND schedule can be set via
   `/dnd_set_schedule` followed by two comma-separated values, defining time in a format of HH:MM **in UTC**
   (e.g. `/dnd_set_schedule 23:00,09:00`).

### Search query

The bot can parse the HTTP API responses only when received from Funda list-type page. You can try the
[example URL] (https://www.funda.nl/en/zoeken/huur/?selected_area=[%22nl%22]&price=%221000-2000%22&object_type=[%22apartment%22,%22house%22]&publication_date=%223%22)
and adapt it as you see fit. You just need to copy the URL from the browser and paste in after the command
`/set_search_query`. The URL can be changed any moment, and any API polling runs executed after search query change will
overwrite existing data.

### Listings

Listings are retrieved each time the scheduled API polling is commenced and when a manual trigger `/update_now` is
invoked. Each iteration removes listings from DB, which are not currently listed, and adds new listings. The user
can retrieve either all listings from DB via `/show_current_listings` or only newly added ones via `/show_new_listings`.

### Favorites

User can add a listing to a list of favorites by clicking the button provided under each listing when invoking
`/tap_current_listings` or `/tap_new_listings`. To access a list of favorites one must use `/show_favorites` command.
The list cannot be edited or deleted except when calling `/stop`. You can add a listing to favorites only when it is
present in the DB, which means that you cannot add a listing to favorites if it was removed from storage.

### Other

Always see `/help` to get a full list of available commands with their description. The commands in `/help` precede over
the description here.

## Prerequisites

1. Your telegram bot token available as an ENV variable `TELEGRAM_BOT_TOKEN`
2. Optional: comma-separated list of authorized users' usernames available as an ENV variable `TELEGRAM_USERS`
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