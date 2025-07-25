#!/bin/bash

# Database Performance Monitoring Script
# This script helps monitor PostgreSQL performance and identify slow queries

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Database connection details
DB_HOST="localhost"
DB_PORT="5433"
DB_NAME="testdb"
DB_USER="user"
DB_PASSWORD="password"

# Function to run SQL query
run_query() {
    local query="$1"
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "$query"
}

# Function to check if PostgreSQL is running
check_postgres() {
    echo -e "${BLUE}🔍 Checking PostgreSQL connection...${NC}"
    if run_query "SELECT 1;" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ PostgreSQL is running${NC}"
        return 0
    else
        echo -e "${RED}❌ PostgreSQL is not accessible${NC}"
        return 1
    fi
}

# Function to show database statistics
show_db_stats() {
    echo -e "${BLUE}📊 Database Statistics${NC}"
    echo "----------------------------------------"
    
    # Database size
    local db_size=$(run_query "SELECT pg_size_pretty(pg_database_size('$DB_NAME'));")
    echo "Database Size: $db_size"
    
    # Connection count
    local connections=$(run_query "SELECT count(*) FROM pg_stat_activity;")
    echo "Active Connections: $connections"
    
    # Table sizes
    echo -e "\n${YELLOW}Table Sizes:${NC}"
    run_query "
        SELECT 
            schemaname,
            tablename,
            pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
        FROM pg_tables 
        WHERE schemaname = 'public'
        ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
    "
}

# Function to show slow queries
show_slow_queries() {
    echo -e "\n${BLUE}🐌 Slow Queries (>100ms)${NC}"
    echo "----------------------------------------"
    
    local slow_count=$(run_query "SELECT count(*) FROM pg_stat_statements WHERE mean_time > 100;")
    echo "Total slow queries: $slow_count"
    
    if [ "$slow_count" -gt 0 ]; then
        echo -e "\n${YELLOW}Top 5 Slowest Queries:${NC}"
        run_query "
            SELECT 
                substring(query, 1, 100) as query_preview,
                calls,
                round(mean_time::numeric, 2) as avg_time_ms,
                round(total_time::numeric, 2) as total_time_ms,
                rows
            FROM pg_stat_statements 
            WHERE mean_time > 100 
            ORDER BY mean_time DESC 
            LIMIT 5;
        "
    fi
}

# Function to show index usage
show_index_usage() {
    echo -e "\n${BLUE}📈 Index Usage Statistics${NC}"
    echo "----------------------------------------"
    
    run_query "
        SELECT 
            indexname,
            idx_scan as scans,
            idx_tup_read as tuples_read,
            idx_tup_fetch as tuples_fetched
        FROM pg_stat_user_indexes 
        ORDER BY idx_scan DESC 
        LIMIT 10;
    "
}

# Function to show table statistics
show_table_stats() {
    echo -e "\n${BLUE}📋 Table Statistics${NC}"
    echo "----------------------------------------"
    
    run_query "
        SELECT 
            tablename,
            n_live_tup as live_rows,
            n_dead_tup as dead_rows,
            seq_scan,
            idx_scan,
            n_tup_ins as inserts,
            n_tup_upd as updates,
            n_tup_del as deletes
        FROM pg_stat_user_tables 
        ORDER BY n_live_tup DESC;
    "
}

# Function to show PostgreSQL settings
show_pg_settings() {
    echo -e "\n${BLUE}⚙️ PostgreSQL Settings${NC}"
    echo "----------------------------------------"
    
    run_query "
        SELECT 
            name,
            setting,
            unit
        FROM pg_settings 
        WHERE name IN (
            'shared_buffers',
            'effective_cache_size',
            'work_mem',
            'maintenance_work_mem',
            'max_connections',
            'log_min_duration_statement'
        )
        ORDER BY name;
    "
}

# Function to show cache hit ratio
show_cache_stats() {
    echo -e "\n${BLUE}💾 Cache Statistics${NC}"
    echo "----------------------------------------"
    
    run_query "
        SELECT 
            schemaname,
            tablename,
            heap_blks_read,
            heap_blks_hit,
            round(100.0 * heap_blks_hit / (heap_blks_hit + heap_blks_read), 2) as cache_hit_ratio
        FROM pg_statio_user_tables 
        WHERE heap_blks_hit + heap_blks_read > 0
        ORDER BY cache_hit_ratio DESC;
    "
}

# Function to show locks
show_locks() {
    echo -e "\n${BLUE}🔒 Active Locks${NC}"
    echo "----------------------------------------"
    
    local lock_count=$(run_query "SELECT count(*) FROM pg_locks WHERE NOT granted;")
    echo "Blocked queries: $lock_count"
    
    if [ "$lock_count" -gt 0 ]; then
        run_query "
            SELECT 
                l.pid,
                l.mode,
                l.granted,
                a.query
            FROM pg_locks l
            JOIN pg_stat_activity a ON l.pid = a.pid
            WHERE NOT l.granted;
        "
    fi
}

# Function to show vacuum status
show_vacuum_status() {
    echo -e "\n${BLUE}🧹 Vacuum Status${NC}"
    echo "----------------------------------------"
    
    run_query "
        SELECT 
            schemaname,
            tablename,
            last_vacuum,
            last_autovacuum,
            vacuum_count,
            autovacuum_count
        FROM pg_stat_user_tables 
        ORDER BY vacuum_count DESC;
    "
}

# Main function
main() {
    echo -e "${GREEN}🚀 PostgreSQL Performance Monitor${NC}"
    echo "========================================"
    
    # Check if PostgreSQL is running
    if ! check_postgres; then
        echo -e "${RED}Exiting: Cannot connect to PostgreSQL${NC}"
        exit 1
    fi
    
    # Show all statistics
    show_db_stats
    show_pg_settings
    show_slow_queries
    show_index_usage
    show_table_stats
    show_cache_stats
    show_locks
    show_vacuum_status
    
    echo -e "\n${GREEN}✅ Monitoring complete${NC}"
}

# Check if script is run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi 