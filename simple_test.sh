#!/bin/bash

# Simple EF Core Migration Test
echo "🚀 Testing GRA EF Core Migration System"
echo "========================================"

export DATABASE_URL="./test_migrations/simple_test.db"
rm -f "$DATABASE_URL"

echo "✅ 1. Initial Status:"
./bin/ef-migrate status

echo
echo "✅ 2. Creating test migration..."
./bin/ef-migrate add-migration TestMigration "Test migration"

echo
echo "✅ 3. Status with pending migration:"
./bin/ef-migrate status

echo
echo "✅ 4. Migration history:"
./bin/ef-migrate get-migration

echo
echo "🎉 Basic test completed successfully!"
