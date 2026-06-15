#!/bin/bash
# ============================================================
# QuantLab — Database Initialization Script
# ============================================================
# CREATE DATABASE cannot run inside a transaction block,
# so we use a shell script instead of a plain SQL file.
# ============================================================

set -e

# List of databases to create
DATABASES=(
  quantlab_user
  quantlab_strategy
  quantlab_portfolio
  quantlab_billing
  quantlab_community
  quantlab_ranking
  quantlab_notification
  quantlab_ai
  quantlab_backtest
  quantlab_formula
  quantlab_market_data
)

for db in "${DATABASES[@]}"; do
  echo "Creating database: $db"
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE "$db"
      WITH ENCODING 'UTF8'
      LC_COLLATE = 'en_US.utf8'
      LC_CTYPE = 'en_US.utf8'
      TEMPLATE template0;
EOSQL
done

echo "All databases created successfully."
