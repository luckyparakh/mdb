# Include variables from the .envrc file
include .envrc
# ==================================================================================== #
# HELPER
# ==================================================================================== #

## help: print this help message
.PHONY: help run/api confirm
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

confirm:
	@echo -n "Are you sure[y/N]:" && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run/api: run the cmd/api application
# use @ so that password in MDB_DSN is not printed on command line
run/api:
	@go run ./cmd/api -dsn=${MDB_DSN}

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	@psql ${MDB_DSN}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migration/up: confirm
	@echo "Running up migration"
	@migrate -path ./migrations -database ${MDB_DSN} up

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #
.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor


# ==================================================================================== #
# BUILD
# ==================================================================================== #
## It’s possible, with ldflags option, to reduce the binary size by around 25% by instructing the Go linker to strip the 
## DWARF debugging information and symbol table from the binary. 
current_time = $(shell date --iso-8601=seconds)
version = $(shell git describe --always --dirty --long --tags)
linker_flags = '-s -X main.buildTime=${current_time} -X main.version=${version}'
.PHONY: build/api
build/api:
	$(info "Building ./cmd/api ...")
	go clean -cache
	go build -ldflags=${linker_flags} -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o=./bin/linux_amd64/api ./cmd/api