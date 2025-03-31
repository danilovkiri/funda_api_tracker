BIN1 := ./cmd/bot/app
BIN2 := ./cmd/cli/app

# Main files
SRC1 := ./cmd/bot/main.go
SRC2 := ./cmd/cli/main.go

# Default target
.PHONY: all
all: build

# Build both binaries
.PHONY: build
build:
	go build -o $(BIN1) $(SRC1)
	go build -o $(BIN2) $(SRC2)

# Clean built binaries
.PHONY: clean
clean:
	rm -f $(BIN1) $(BIN2)

# Run the bot app
.PHONY: bot
run-bot: $(BIN1)
	./$(BIN1)

# Migrate the DB to latest version
.PHONY: migrate-up
migrate-up: $(BIN2)
	./$(BIN2) storage:migrate --direction=up
