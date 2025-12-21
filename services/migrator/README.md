# Database Migrator

This tool runs database migrations using golang-migrate.

## Usage

### Local Development

```bash
cd services/migrator

# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=myhealth_user
export DB_PASSWORD=your_password
export DB_NAME=myhealth
export DB_SSLMODE=disable
export MIGRATIONS_PATH=file://../migrations

# Run migrations
go run main.go
```

### Docker

```bash
# Build image
docker build -t myhealth/migrator:latest .

# Run migrations
docker run --rm \
  -e DB_HOST=your-db-host \
  -e DB_PORT=5432 \
  -e DB_USER=myhealth_user \
  -e DB_PASSWORD=your_password \
  -e DB_NAME=myhealth \
  -e DB_SSLMODE=require \
  myhealth/migrator:latest
```

### Kubernetes

The migrator runs as an init container before the main application containers start.

See the Helm chart for configuration.

## Environment Variables

- `DB_HOST` - Database host (required)
- `DB_PORT` - Database port (default: 5432)
- `DB_USER` - Database user (required)
- `DB_PASSWORD` - Database password (required)
- `DB_NAME` - Database name (required)
- `DB_SSLMODE` - SSL mode: disable, require, verify-ca, verify-full (default: require)
- `MIGRATIONS_PATH` - Path to migrations directory (default: file://migrations)

## Migration Files

Migrations are stored in `services/migrations/` directory.

File naming format: `{version}_{description}.{up|down}.sql`

Example:
- `000001_create_users.up.sql` - Creates users table
- `000001_create_users.down.sql` - Drops users table

## Creating New Migrations

```bash
# Create a new migration
touch services/migrations/000006_add_user_preferences.up.sql
touch services/migrations/000006_add_user_preferences.down.sql
```

Edit the `.up.sql` file with your schema changes and the `.down.sql` file with the rollback logic.
