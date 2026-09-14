set dotenv-load := true

help:
    @just --list

migration name: (check_tool "goose" "go install github.com/pressly/goose/v3/cmd/goose@latest")
    @goose -dir=database/migrations create {{ name }} sql -s

seeder name: (check_tool "goose" "go install github.com/pressly/goose/v3/cmd/goose@latest")
    @goose -dir=database/seeders create {{ name }} sql

mu: migrate-up

migrate-up: (check_tool "goose" "go install github.com/pressly/goose/v3/cmd/goose@latest")
    @echo "Running up migrations..."
    @goose -dir=database/migrations postgres ${DATABASE_URL} up

md: migrate-down

migrate-down: (check_tool "goose" "go install github.com/pressly/goose/v3/cmd/goose@latest")
    @echo "Rolling back migrations..."
    @goose -dir=database/migrations postgres ${DATABASE_URL} down

mf: migrate-fresh

migrate-fresh: (check_tool "goose" "go install github.com/pressly/goose/v3/cmd/goose@latest")
    @echo "Droping..."
    @goose -dir=database/migrations postgres ${DATABASE_URL} reset
    @just migrate-up

mfs: migrate-fresh-seed

migrate-fresh-seed: (check_tool "goose" "go install github.com/pressly/goose/v3/cmd/goose@latest")
    @just migrate-fresh
    @echo "Seeding..."
    @goose -dir=database/seeders -no-versioning postgres ${DATABASE_URL} up

v: vendor

vendor:
    @echo "Tidying and verifying dependencies..."
    @go mod tidy
    @go mod verify
    @echo "Vendoring dependencies..."
    @go mod vendor

a: audit

audit: vendor (check_tool "staticcheck" "go install honnef.co/go/tools/cmd/staticcheck@latest")
    @echo "Formatting code..."
    @go fmt ./...
    @echo "Vetting code..."
    @go vet ./...
    @echo "Running staticcheck..."
    @staticcheck ./...
    @echo "Running test..."
    @go test -race -vet=off ./...

check_tool tool install:
    #!/usr/bin/env bash
    set -e
    if ! command -v {{ tool }} > /dev/null 2>&1; then
        read -p "{{ tool }} is not installed. Download it now? [y/N] " confirm
        if [[ "$confirm" =~ ^[Yy]$ ]]; then
            {{ install }}
        else
            echo "Skipping {{ tool }} install, aborting."
            exit 1
        fi
    fi

sql:
    @sqlc generate

nd name: (new-domain name)

new-domain name:
    #!/usr/bin/env bash
    set -euo pipefail
    if [[ ! "{{ name }}" =~ ^[a-z][a-z0-9]*$ ]]; then
      echo "domain name must be lowercase alphanumeric, e.g. roles" >&2
      exit 1
    fi
    dir="internal/domain/{{ name }}"
    mkdir -p "$dir" "$dir/test"
    for f in types.go service.go handler.go; do
      path="$dir/$f"
      if [ -f "$path" ]; then
        echo "skip (exists): $path"
        continue
      fi
      printf 'package {{ name }}\n' > "$path"
      echo "created: $path"
    done
