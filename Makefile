run:
	go run ./cmd/api
psql:
	psql ${MDB_DSN}
up:
	@echo "Running up migration"
	migratea -path ./migrations -database ${MDB_DSN} up