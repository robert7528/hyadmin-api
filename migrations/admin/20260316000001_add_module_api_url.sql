-- Atlas migration: add api_url column to hyadmin_modules
-- Generated: 2026-03-16
-- Purpose: Store backend API base URL for each module (relative or absolute).

ALTER TABLE hyadmin_modules ADD COLUMN IF NOT EXISTS api_url TEXT;
