# gin-freemarket

## How to Start the Application

1. Set up required environment variables (e.g., create a `.env` file)

2. Run migrations (if needed)
   ```sh
   go run migrations/migration.go
   ```

3. Start the application
   ```sh
   go run main.go
   ```

4. (If using Docker)
   ```sh
   docker-compose up -d
   ```

---

## How to Run Migrations

1. Create migration files
   ```sh
   go run migrations/migration.go
   ```

## How to Run Load Tests

1. Execute load tests
   ```sh
   cd loadtest
   go run loadtest.go
   ```
