APP_NAME=auth-api
CONFIG=./config/local.yaml
MAIN=./cmd/auth-api/main.go
MIGRATE_MAIN=./cmd/migrator/main.go

.PHONY: run build clean migrate

run:
	CONFIG_PATH=$(CONFIG) go run $(MAIN)

build:
	go build -o $(APP_NAME) $(MAIN)

clean:
	rm -f $(APP_NAME)

migrate:
	CONFIG_PATH=$(CONFIG) go run $(MIGRATE_MAIN)
