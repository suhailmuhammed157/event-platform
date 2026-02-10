# Makefile for database migration management

# Database connection string (override with environment variable)
DATABASE_URL ?= postgres://user:pass@localhost:5532/events_db?sslmode=disable
MIGRATIONS_PATH = ./sink-service/migrations

# -----------------------------
# Install migrate CLI tool
# -----------------------------
.PHONY: install-migrate
install-migrate:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# -----------------------------
# Create a new migration file pair
# Usage: make create-migration name=create_users_table
# -----------------------------
.PHONY: create-migration
create-migration:
	@if [ -z "$(name)" ]; then \
		echo "Error: Please provide a migration name, e.g., make create-migration name=create_users_table"; \
		exit 1; \
	fi
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)

# -----------------------------
# Apply all pending migrations
# -----------------------------
.PHONY: migrate-up
migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" up

# -----------------------------
# Apply N migrations
# Usage: make migrate-up-n n=2
# -----------------------------
.PHONY: migrate-up-n
migrate-up-n:
	@if [ -z "$(n)" ]; then \
		echo "Error: Please provide number of migrations, e.g., make migrate-up-n n=2"; \
		exit 1; \
	fi
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" up $(n)

# -----------------------------
# Revert the last migration
# -----------------------------
.PHONY: migrate-down
migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down 1

# -----------------------------
# Revert N migrations
# Usage: make migrate-down-n n=2
# -----------------------------
.PHONY: migrate-down-n
migrate-down-n:
	@if [ -z "$(n)" ]; then \
		echo "Error: Please provide number of migrations, e.g., make migrate-down-n n=2"; \
		exit 1; \
	fi
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down $(n)

# -----------------------------
# Revert all migrations (dangerous!)
# -----------------------------
.PHONY: migrate-reset
migrate-reset:
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down

# -----------------------------
# Show current migration version
# -----------------------------
.PHONY: migrate-version
migrate-version:
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" version

# -----------------------------
# Force set migration version (fix dirty state)
# Usage: make migrate-force version=3
# -----------------------------
.PHONY: migrate-force
migrate-force:
	@if [ -z "$(version)" ]; then \
		echo "Error: Please provide a version, e.g., make migrate-force version=3"; \
		exit 1; \
	fi
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" force $(version)

# -----------------------------
# Go to specific version
# Usage: make migrate-goto version=5
# -----------------------------
.PHONY: migrate-goto
migrate-goto:
	@if [ -z "$(version)" ]; then \
		echo "Error: Please provide a version, e.g., make migrate-goto version=5"; \
		exit 1; \
	fi
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" goto $(version)
