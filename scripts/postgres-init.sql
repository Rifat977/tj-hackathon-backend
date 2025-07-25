-- PostgreSQL initialization script for production resilience
-- This script runs when the PostgreSQL container starts for the first time

-- Create extension for better performance monitoring
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Configure shared_preload_libraries for better monitoring
-- Note: This needs to be set in postgresql.conf, but we'll prepare the extension
SELECT 'pg_stat_statements extension installed' as status;

-- Set optimal settings for the application
-- Connection and authentication settings
ALTER SYSTEM SET max_connections = 200;
ALTER SYSTEM SET shared_buffers = '128MB';
ALTER SYSTEM SET effective_cache_size = '256MB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.7;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;
ALTER SYSTEM SET random_page_cost = 1.1;
ALTER SYSTEM SET effective_io_concurrency = 200;

-- Logging configuration for better debugging
ALTER SYSTEM SET log_destination = 'stderr';
ALTER SYSTEM SET logging_collector = on;
ALTER SYSTEM SET log_directory = 'pg_log';
ALTER SYSTEM SET log_filename = 'postgresql-%Y-%m-%d_%H%M%S.log';
ALTER SYSTEM SET log_min_messages = warning;
ALTER SYSTEM SET log_min_error_statement = error;
ALTER SYSTEM SET log_min_duration_statement = 1000; -- Log slow queries (>1s)

-- Connection settings for better resilience
ALTER SYSTEM SET tcp_keepalives_idle = 600;
ALTER SYSTEM SET tcp_keepalives_interval = 30;
ALTER SYSTEM SET tcp_keepalives_count = 3;

-- Reload configuration
SELECT pg_reload_conf();

-- Create a simple health check function
CREATE OR REPLACE FUNCTION public.health_check()
RETURNS TABLE(status text, uptime interval, connections integer, database_size text)
LANGUAGE sql
AS $$
    SELECT 
        'healthy' as status,
        now() - pg_postmaster_start_time() as uptime,
        numbackends as connections,
        pg_size_pretty(pg_database_size(current_database())) as database_size
    FROM pg_stat_database 
    WHERE datname = current_database();
$$;

-- Grant necessary permissions
GRANT EXECUTE ON FUNCTION public.health_check() TO public;

SELECT 'PostgreSQL initialization completed successfully' as status; 