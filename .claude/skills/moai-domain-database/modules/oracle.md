# Oracle Advanced Patterns

## Overview

Comprehensive Oracle patterns covering advanced SQL, PL/SQL, performance optimization, and enterprise features using python-oracledb (successor to cx_Oracle).

## Quick Implementation

### Async Connection Pool Setup

```python
import asyncio
import oracledb

async def create_oracle_pool():
    """Create async connection pool with statement caching."""
    pool = await oracledb.create_pool_async(
        user="app_user",
        password=userpwd,
        dsn="dbhost.example.com/orclpdb",
        min=2,
        max=10,
        increment=1,
        stmtcachesize=50  # Statement cache for performance
    )
    return pool

async def execute_query(pool, query: str, params: dict = None):
    """Execute query with connection from pool."""
    async with pool.acquire() as connection:
        with connection.cursor() as cursor:
            await cursor.execute(query, params or {})
            async for row in cursor:
                yield row
```

### Hierarchical Queries (CONNECT BY)

```sql
-- Organizational hierarchy with path
SELECT
    employee_id,
    name,
    manager_id,
    LEVEL AS depth,
    SYS_CONNECT_BY_PATH(name, '/') AS path
FROM employees
WHERE department_id = :dept_id
START WITH manager_id IS NULL
CONNECT BY PRIOR employee_id = manager_id
ORDER SIBLINGS BY name;

-- Recursive CTE alternative (Oracle 11g+)
WITH employee_hierarchy AS (
    SELECT employee_id, name, manager_id, 1 AS level
    FROM employees
    WHERE manager_id IS NULL

    UNION ALL

    SELECT e.employee_id, e.name, e.manager_id, h.level + 1
    FROM employees e
    JOIN employee_hierarchy h ON e.manager_id = h.employee_id
)
SELECT * FROM employee_hierarchy;
```

### Batch Operations with PL/SQL

```python
import oracledb

def batch_insert_with_plsql(connection, data: list[tuple]):
    """Efficient batch insert using executemany."""
    with connection.cursor() as cursor:
        # Batch insert with array DML row counts
        cursor.executemany(
            """
            INSERT INTO orders (order_id, customer_id, amount, order_date)
            VALUES (:1, :2, :3, :4)
            """,
            data,
            batcherrors=True,
            arraydmlrowcounts=True
        )

        # Check for batch errors
        errors = cursor.getbatcherrors()
        if errors:
            for error in errors:
                print(f"Row {error.offset}: {error.message}")

        # Get row counts
        row_counts = cursor.getarraydmlrowcounts()
        connection.commit()
        return sum(row_counts)

def call_stored_procedure(connection, proc_name: str, params: list):
    """Call PL/SQL procedure with IN/OUT parameters."""
    with connection.cursor() as cursor:
        # OUT parameter setup
        out_val = cursor.var(oracledb.DB_TYPE_NUMBER)
        cursor.callproc(proc_name, [*params, out_val])
        return out_val.getvalue()
```

### Partitioned Table Management

```sql
-- Range partitioned table
CREATE TABLE sales (
    sales_id NUMBER GENERATED ALWAYS AS IDENTITY,
    sale_date DATE NOT NULL,
    amount NUMBER(12,2),
    region VARCHAR2(50),
    CONSTRAINT pk_sales PRIMARY KEY (sales_id, sale_date)
)
PARTITION BY RANGE (sale_date)
INTERVAL (NUMTOYMINTERVAL(1, 'MONTH'))
(
    PARTITION p_initial VALUES LESS THAN (DATE '2024-01-01')
);

-- Hash partitioned for even distribution
CREATE TABLE customers (
    customer_id NUMBER PRIMARY KEY,
    name VARCHAR2(100),
    email VARCHAR2(255)
)
PARTITION BY HASH (customer_id)
PARTITIONS 8;

-- Composite partitioning (Range-Hash)
CREATE TABLE transactions (
    txn_id NUMBER,
    txn_date DATE,
    account_id NUMBER,
    amount NUMBER(15,2)
)
PARTITION BY RANGE (txn_date)
SUBPARTITION BY HASH (account_id)
SUBPARTITIONS 4
(
    PARTITION p_2024_q1 VALUES LESS THAN (DATE '2024-04-01'),
    PARTITION p_2024_q2 VALUES LESS THAN (DATE '2024-07-01')
);
```

### LOB Handling

```python
import oracledb

def insert_document(connection, doc_id: int, content: bytes):
    """Insert BLOB data efficiently."""
    with connection.cursor() as cursor:
        # Create temporary LOB
        blob_var = connection.createlob(oracledb.DB_TYPE_BLOB)
        blob_var.write(content)

        cursor.execute(
            "INSERT INTO documents (id, content) VALUES (:id, :content)",
            {"id": doc_id, "content": blob_var}
        )
        connection.commit()

def stream_large_lob(connection, doc_id: int, chunk_size: int = 65536):
    """Stream large LOB data in chunks."""
    with connection.cursor() as cursor:
        cursor.execute(
            "SELECT content FROM documents WHERE id = :id",
            {"id": doc_id}
        )
        row = cursor.fetchone()
        if row and row[0]:
            lob = row[0]
            offset = 1
            while True:
                chunk = lob.read(offset, chunk_size)
                if not chunk:
                    break
                yield chunk
                offset += len(chunk)
```

## Key Features

1. **Hierarchical Queries**: CONNECT BY, recursive CTEs, tree traversals
2. **Partitioning**: Range, list, hash, composite partitioning strategies
3. **PL/SQL Integration**: Stored procedures, packages, batch operations
4. **Performance Tuning**: Statement caching, connection pooling, optimizer hints

## Best Practices

- Use connection pooling with appropriate min/max settings
- Enable statement caching (stmtcachesize) for repeated queries
- Use bind variables to prevent SQL injection and improve parse efficiency
- Leverage partitioning for large tables (>10GB)
- Use bitmap indexes for low-cardinality columns in data warehouses
- Collect statistics regularly with DBMS_STATS
- Monitor with AWR (Automatic Workload Repository) reports

## Integration Points

- **Python Drivers**: python-oracledb (recommended), cx_Oracle (legacy)
- **ORMs**: SQLAlchemy with Oracle dialect, Django ORM
- **Connection Pooling**: Built-in pooling, Oracle DRCP
- **Migration Tools**: SQL*Plus, Data Pump, Flyway, Liquibase
- **Monitoring**: Oracle Enterprise Manager, AWR, ASH reports
- **Cloud**: Oracle Autonomous Database, OCI integration

## Advanced Patterns

### Optimizer Hints

```sql
-- Force index usage
SELECT /*+ INDEX(orders idx_order_date) */ *
FROM orders
WHERE order_date BETWEEN :start_date AND :end_date;

-- Parallel query execution
SELECT /*+ PARALLEL(sales, 4) */
    region, SUM(amount) as total
FROM sales
GROUP BY region;

-- Materialized view refresh
BEGIN
    DBMS_MVIEW.REFRESH('mv_sales_summary', 'C');  -- Complete refresh
    DBMS_MVIEW.REFRESH('mv_daily_stats', 'F');    -- Fast refresh
END;
```

### JSON Support (Oracle 12c+)

```sql
-- Store and query JSON data
CREATE TABLE api_logs (
    id NUMBER GENERATED ALWAYS AS IDENTITY,
    request_data JSON,
    response_data JSON,
    created_at TIMESTAMP DEFAULT SYSTIMESTAMP
);

-- JSON path queries
SELECT
    j.request_data.method,
    j.request_data.endpoint,
    j.response_data.status_code
FROM api_logs j
WHERE JSON_EXISTS(j.request_data, '$.headers.Authorization');

-- JSON aggregation
SELECT JSON_ARRAYAGG(
    JSON_OBJECT(
        'id' VALUE customer_id,
        'name' VALUE name,
        'orders' VALUE order_count
    )
) AS customers_json
FROM customer_summary;
```
