-- Migration 003: Add password_hash to accounts table for hosting tenant login & system auth
ALTER TABLE accounts ADD COLUMN password_hash TEXT;
