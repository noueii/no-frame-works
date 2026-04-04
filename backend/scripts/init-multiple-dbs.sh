#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE USER kratos WITH PASSWORD 'secret';
    CREATE DATABASE kratos OWNER kratos;
EOSQL
