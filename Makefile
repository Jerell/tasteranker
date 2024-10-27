.PHONY: migrate-up migrate-down migrate-create

DB_URL=postgres://postgres:jerell@localhost:5432/tasteranker-dev?sslmode=disable

migrate-up:
	migrate -database "${DB_URL}" -path internal/db/migrations up

migrate-down:
	migrate -database "${DB_URL}" -path internal/db/migrations down

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir internal/db/migrations -seq $$name
