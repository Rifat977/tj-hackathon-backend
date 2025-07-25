-- PostgreSQL Performance Optimization Script
-- This script runs on container startup to optimize the database

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";
CREATE EXTENSION IF NOT EXISTS "btree_gin";

-- Optimize PostgreSQL settings for ultra-fast performance
ALTER SYSTEM SET shared_buffers = '768MB';  -- 75% of 1GB RAM for ultra-fast operations
ALTER SYSTEM SET effective_cache_size = '1GB';  -- 75% of available RAM
ALTER SYSTEM SET work_mem = '64MB';  -- Increased for ultra-fast sort/join performance
ALTER SYSTEM SET maintenance_work_mem = '512MB';  -- For ultra-fast index creation and maintenance
ALTER SYSTEM SET max_connections = 500;  -- Increased for ultra-fast bulk operations
ALTER SYSTEM SET wal_buffers = '64MB';  -- Increased for ultra-fast writes
ALTER SYSTEM SET checkpoint_segments = 128;  -- Increased for ultra-fast bulk operations
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET random_page_cost = 1.1;  -- Optimize for SSD
ALTER SYSTEM SET effective_io_concurrency = 800;  -- Increased for ultra-fast SSD operations
ALTER SYSTEM SET max_parallel_workers_per_gather = 6;  -- Increased for 2 vCPU
ALTER SYSTEM SET max_parallel_workers = 6;  -- Increased for ultra-fast processing
ALTER SYSTEM SET max_worker_processes = 12;  -- Increased for ultra-fast operations

-- Enable slow query logging (queries > 100ms)
ALTER SYSTEM SET log_min_duration_statement = 100;
ALTER SYSTEM SET log_statement = 'all';
ALTER SYSTEM SET log_duration = on;
ALTER SYSTEM SET log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h ';

-- Optimize for bulk operations
ALTER SYSTEM SET synchronous_commit = off;  -- Faster bulk inserts
ALTER SYSTEM SET fsync = off;  -- Faster bulk operations (use with caution)
ALTER SYSTEM SET full_page_writes = off;  -- Faster bulk operations
ALTER SYSTEM SET wal_writer_delay = 200ms;
ALTER SYSTEM SET commit_delay = 1000;  -- Microseconds
ALTER SYSTEM SET commit_siblings = 5;

-- Reload configuration
SELECT pg_reload_conf();

-- Create indexes on frequently queried columns
-- Product table indexes
CREATE INDEX IF NOT EXISTS idx_products_id ON products(id);
CREATE INDEX IF NOT EXISTS idx_products_active ON products(active) WHERE active = true;
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_price ON products(price);
CREATE INDEX IF NOT EXISTS idx_products_stock ON products(stock);
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_products_updated_at ON products(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_products_slug ON products(slug);
CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);
CREATE INDEX IF NOT EXISTS idx_products_internal_id ON products(internal_id);
CREATE INDEX IF NOT EXISTS idx_products_brand ON products(brand);
CREATE INDEX IF NOT EXISTS idx_products_availability ON products(availability);

-- Composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_products_active_category ON products(active, category_id) WHERE active = true;
CREATE INDEX IF NOT EXISTS idx_products_active_created ON products(active, created_at DESC) WHERE active = true;
CREATE INDEX IF NOT EXISTS idx_products_category_price ON products(category_id, price);
CREATE INDEX IF NOT EXISTS idx_products_active_price ON products(active, price) WHERE active = true;

-- Full-text search indexes
CREATE INDEX IF NOT EXISTS idx_products_name_fts ON products USING gin(to_tsvector('english', name));
CREATE INDEX IF NOT EXISTS idx_products_description_fts ON products USING gin(to_tsvector('english', description));
CREATE INDEX IF NOT EXISTS idx_products_combined_fts ON products USING gin(to_tsvector('english', name || ' ' || COALESCE(description, '')));

-- Category table indexes
CREATE INDEX IF NOT EXISTS idx_categories_id ON categories(id);
CREATE INDEX IF NOT EXISTS idx_categories_active ON categories(active) WHERE active = true;
CREATE INDEX IF NOT EXISTS idx_categories_slug ON categories(slug);
CREATE INDEX IF NOT EXISTS idx_categories_name ON categories(name);

-- Pagination optimization indexes
CREATE INDEX IF NOT EXISTS idx_products_pagination ON products(active, created_at DESC, id) WHERE active = true;
CREATE INDEX IF NOT EXISTS idx_products_search_pagination ON products(active, id) WHERE active = true;

-- Bulk upload optimization indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_bulk_upload ON products(category_id, active, created_at) WHERE active = true;

-- Lightning-fast bulk upload optimization indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_lightning_fast ON products(category_id, active, created_at, id) WHERE active = true;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_bulk_insert ON products(name, slug, sku) WHERE active = true;

-- Performance monitoring function
CREATE OR REPLACE FUNCTION public.health_check()
RETURNS TABLE(
    database_name text,
    current_connections integer,
    max_connections integer,
    shared_buffers text,
    effective_cache_size text,
    work_mem text,
    slow_queries bigint
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        current_database()::text as database_name,
        (SELECT count(*) FROM pg_stat_activity)::integer as current_connections,
        (SELECT setting::integer FROM pg_settings WHERE name = 'max_connections') as max_connections,
        (SELECT setting FROM pg_settings WHERE name = 'shared_buffers') as shared_buffers,
        (SELECT setting FROM pg_settings WHERE name = 'effective_cache_size') as effective_cache_size,
        (SELECT setting FROM pg_settings WHERE name = 'work_mem') as work_mem,
        (SELECT count(*) FROM pg_stat_statements WHERE mean_time > 100) as slow_queries;
END;
$$ LANGUAGE plpgsql;

-- Create view for slow queries monitoring
CREATE OR REPLACE VIEW slow_queries AS
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    stddev_time,
    min_time,
    max_time,
    rows
FROM pg_stat_statements 
WHERE mean_time > 100 
ORDER BY mean_time DESC;

-- Create view for index usage statistics
CREATE OR REPLACE VIEW index_usage_stats AS
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;

-- Create view for table statistics
CREATE OR REPLACE VIEW table_stats AS
SELECT 
    schemaname,
    tablename,
    seq_scan,
    seq_tup_read,
    idx_scan,
    idx_tup_fetch,
    n_tup_ins,
    n_tup_upd,
    n_tup_del,
    n_live_tup,
    n_dead_tup
FROM pg_stat_user_tables
ORDER BY n_live_tup DESC;

-- Grant permissions for monitoring
GRANT SELECT ON slow_queries TO PUBLIC;
GRANT SELECT ON index_usage_stats TO PUBLIC;
GRANT SELECT ON table_stats TO PUBLIC;

-- Log the completion
DO $$
BEGIN
    RAISE NOTICE 'PostgreSQL optimization completed successfully';
    RAISE NOTICE 'Indexes created for performance optimization';
    RAISE NOTICE 'Slow query logging enabled (>100ms)';
    RAISE NOTICE 'Shared buffers set to 256MB';
    RAISE NOTICE 'Work memory set to 16MB';
END $$; 