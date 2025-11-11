---
name: moai-lang-sql
version: 4.0.0
created: 2025-10-22
updated: 2025-10-22
status: active
description: Advanced SQL and database development with PostgreSQL 16, MySQL 8.0, SQL Server 2022, Oracle 23c, and modern query patterns. Enterprise database design, optimization, and analytics with Context7 MCP integration.
keywords: ['sql', 'postgresql', 'mysql', 'sql-server', 'oracle', 'database', 'query-optimization', 'data-analytics', 'context7']
allowed-tools:
  - Read
  - Bash
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# Lang SQL Skill - Enterprise v4.0.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-lang-sql |
| **Version** | 4.0.0 (2025-10-22) |
| **Allowed tools** | Read (read_file), Bash (terminal), Context7 MCP |
| **Auto-load** | On demand when keywords detected |
| **Tier** | Language Enterprise |
| **Context7 Integration** | ✅ PostgreSQL/MySQL/SQLServer/Oracle |

---

## What It Does

Advanced SQL and database development featuring PostgreSQL 16, MySQL 8.0, SQL Server 2022, Oracle 23c, and modern query patterns for enterprise database design, optimization, analytics, and data warehousing. Context7 MCP integration provides real-time access to official database documentation.

**Key capabilities**:
- ✅ PostgreSQL 16 with advanced JSON and parallel queries
- ✅ MySQL 8.0 with window functions and CTEs
- ✅ SQL Server 2022 with T-SQL enhancements
- ✅ Oracle 23c with JSON and graph features
- ✅ Advanced query optimization and indexing
- ✅ Database performance tuning and monitoring
- ✅ Enterprise data modeling and design patterns
- ✅ Data warehousing and analytics with window functions
- ✅ Database security and audit patterns
- ✅ NoSQL hybrid features and document storage

---

## When to Use

**Automatic triggers**:
- SQL query development and optimization
- Database schema design discussions
- Performance tuning and indexing
- Data analytics and reporting
- Database administration tasks
- ETL and data pipeline development

**Manual invocation**:
- Design database architecture
- Optimize complex queries
- Implement data security measures
- Review SQL code quality
- Troubleshoot performance issues
- Design analytics solutions

---

## Technology Stack (2025-10-22)

| Component | Version | Purpose | Status |
|-----------|---------|---------|--------|
| **PostgreSQL** | 16.4 | Advanced RDBMS | ✅ Current |
| **MySQL** | 8.0.35 | Popular RDBMS | ✅ Current |
| **SQL Server** | 2022 | Microsoft RDBMS | ✅ Current |
| **Oracle** | 23c | Enterprise RDBMS | ✅ Current |
| **Redis** | 7.2.0 | In-memory cache | ✅ Current |

---

## Code Examples (30+ Enterprise Patterns)

### 1. Advanced PostgreSQL 16 Patterns

```sql
-- Advanced JSON manipulation with PostgreSQL 16
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    profile JSONB NOT NULL DEFAULT '{}',
    preferences JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- JSON schema validation
ALTER TABLE users 
ADD CONSTRAINT valid_email CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$');

-- JSONB indexes for performance
CREATE INDEX idx_users_profile_gin ON users USING GIN (profile);
CREATE INDEX idx_users_preferences_gin ON users USING GIN (preferences);

-- Advanced JSON queries with JSONPath
SELECT 
    id,
    email,
    profile->>'name' AS name,
    profile->>'age' AS age,
    profile->'address'->>'city' AS city,
    preferences->'notifications'->'email' AS email_notifications
FROM users
WHERE JSON_EXISTS(profile, '$.skills ? (@.type == "programming")')
    AND profile->>'age'::integer > 25
ORDER BY (profile->>'last_login')::timestamp DESC;

-- Parallel query optimization
SET max_parallel_workers_per_gather = 4;
SET parallel_tuple_cost = 0.1;
SET parallel_setup_cost = 1000.0;

-- Complex analytics with window functions
WITH user_activity AS (
    SELECT 
        u.id,
        u.email,
        COUNT(a.id) AS total_activities,
        COUNT(DISTINCT DATE(a.created_at)) AS active_days,
        MAX(a.created_at) AS last_activity,
        AVG(EXTRACT(EPOCH FROM (a.ended_at - a.started_at))) AS avg_session_duration
    FROM users u
    LEFT JOIN user_activities a ON u.id = a.user_id
    WHERE a.created_at >= NOW() - INTERVAL '90 days'
    GROUP BY u.id, u.email
),
user_segments AS (
    SELECT 
        *,
        NTILE(4) OVER (ORDER BY total_activities) AS activity_quartile,
        NTILE(4) OVER (ORDER BY active_days) AS engagement_quartile,
        CASE 
            WHEN total_activities = 0 THEN 'inactive'
            WHEN active_days >= 20 AND avg_session_duration > 300 THEN 'highly_engaged'
            WHEN avg_session_duration > 180 THEN 'moderately_engaged'
            ELSE 'low_engaged'
        END AS engagement_segment
    FROM user_activity
)
SELECT 
    engagement_segment,
    COUNT(*) AS user_count,
    AVG(total_activities) AS avg_activities,
    AVG(active_days) AS avg_active_days,
    AVG(avg_session_duration) AS avg_session_duration
FROM user_segments
GROUP BY engagement_segment
ORDER BY user_count DESC;

-- Advanced CTE with recursive queries
WITH RECURSIVE org_hierarchy AS (
    SELECT 
        id,
        name,
        parent_id,
        1 AS level,
        ARRAY[name] AS path,
        0 AS manager_count
    FROM organizations 
    WHERE parent_id IS NULL
    
    UNION ALL
    
    SELECT 
        o.id,
        o.name,
        o.parent_id,
        oh.level + 1,
        oh.path || o.name,
        oh.manager_count + 1
    FROM organizations o
    INNER JOIN org_hierarchy oh ON o.parent_id = oh.id
),
organization_metrics AS (
    SELECT 
        *,
        COUNT(*) OVER (PARTITION BY parent_id) AS direct_reports,
        SUM(CASE WHEN level = 3 THEN 1 ELSE 0 END) OVER (PARTITION BY parent_id) AS team_leads
    FROM org_hierarchy
)
SELECT 
    level,
    COUNT(*) AS total_organizations,
    AVG(direct_reports) AS avg_direct_reports,
    AVG(team_leads) AS avg_team_leads,
    AVG(manager_count) AS avg_management_depth
FROM organization_metrics
GROUP BY level
ORDER BY level;

-- Materialized views for performance
CREATE MATERIALIZED VIEW user_analytics_mv AS
SELECT 
    DATE_TRUNC('day', created_at) AS date,
    COUNT(*) AS new_users,
    COUNT(DISTINCT email) AS unique_emails,
    AVG(
        (profile->>'age')::integer
    ) FILTER (WHERE profile->>'age' IS NOT NULL) AS avg_age
FROM users
GROUP BY DATE_TRUNC('day', created_at)
ORDER BY date DESC;

CREATE INDEX idx_user_analytics_mv_date ON user_analytics_mv (date);

-- Refresh materialized view
REFRESH MATERIALIZED VIEW CONCURRENTLY user_analytics_mv;
```

### 2. Advanced MySQL 8.0 Patterns

```sql
-- Advanced window functions and analytics
CREATE TABLE sales_transactions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    transaction_date DATETIME NOT NULL,
    region VARCHAR(50) NOT NULL,
    channel ENUM('online', 'retail', 'mobile') NOT NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_product_id (product_id),
    INDEX idx_transaction_date (transaction_date),
    INDEX idx_region (region)
);

-- Complex analytics with window functions
WITH monthly_revenue AS (
    SELECT 
        DATE_FORMAT(transaction_date, '%Y-%m') AS month,
        region,
        channel,
        SUM(amount) AS total_revenue,
        COUNT(*) AS transaction_count,
        AVG(amount) AS avg_transaction_value
    FROM sales_transactions
    WHERE transaction_date >= DATE_SUB(NOW(), INTERVAL 2 YEAR)
    GROUP BY DATE_FORMAT(transaction_date, '%Y-%m'), region, channel
),
revenue_growth AS (
    SELECT 
        month,
        region,
        channel,
        total_revenue,
        transaction_count,
        LAG(total_revenue, 1) OVER (PARTITION BY region, channel ORDER BY month) AS prev_month_revenue,
        LAG(transaction_count, 1) OVER (PARTITION BY region, channel ORDER BY month) AS prev_month_transactions,
        (total_revenue - LAG(total_revenue, 1) OVER (PARTITION BY region, channel ORDER BY month)) / 
        LAG(total_revenue, 1) OVER (PARTITION BY region, channel ORDER BY month) * 100 AS revenue_growth_rate
    FROM monthly_revenue
),
rankings AS (
    SELECT 
        *,
        DENSE_RANK() OVER (PARTITION BY month ORDER BY total_revenue DESC) AS region_rank,
        DENSE_RANK() OVER (PARTITION BY month, region ORDER BY total_revenue DESC) AS channel_rank,
        PERCENT_RANK() OVER (PARTITION BY month ORDER BY total_revenue) AS revenue_percentile
    FROM revenue_growth
)
SELECT 
    month,
    region,
    channel,
    total_revenue,
    transaction_count,
    revenue_growth_rate,
    region_rank,
    channel_rank,
    revenue_percentile
FROM rankings
WHERE month >= DATE_FORMAT(DATE_SUB(NOW(), INTERVAL 6 MONTH), '%Y-%m')
ORDER BY month DESC, region_rank, channel_rank;

-- Common Table Expressions (CTE) with recursive queries
WITH RECURSIVE category_hierarchy AS (
    SELECT 
        id,
        name,
        parent_id,
        1 AS level,
        CAST(name AS CHAR(1000)) AS path
    FROM product_categories
    WHERE parent_id IS NULL
    
    UNION ALL
    
    SELECT 
        c.id,
        c.name,
        c.parent_id,
        ch.level + 1,
        CONCAT(ch.path, ' > ', c.name) AS path
    FROM product_categories c
    INNER JOIN category_hierarchy ch ON c.parent_id = ch.id
),
category_metrics AS (
    SELECT 
        ch.*,
        COUNT(p.id) AS product_count,
        COALESCE(SUM(p.price), 0) AS total_inventory_value
    FROM category_hierarchy ch
    LEFT JOIN products p ON ch.id = p.category_id
    GROUP BY ch.id, ch.name, ch.parent_id, ch.level, ch.path
)
SELECT 
    level,
    COUNT(*) AS category_count,
    AVG(product_count) AS avg_products_per_category,
    AVG(total_inventory_value) AS avg_inventory_value
FROM category_metrics
GROUP BY level
ORDER BY level;

-- JSON functions for flexible data storage
CREATE TABLE products (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    sku VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    attributes JSON NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Generated columns for query optimization
    generated_price DECIMAL(10,2) 
        GENERATED ALWAYS AS (JSON_EXTRACT(attributes, '$.price')) VIRTUAL,
    generated_category VARCHAR(100)
        GENERATED ALWAYS AS (JSON_UNQUOTE(JSON_EXTRACT(attributes, '$.category'))) VIRTUAL
);

-- JSON queries with virtual columns
SELECT 
    p.id,
    p.sku,
    p.name,
    p.generated_price AS price,
    p.generated_category AS category,
    JSON_EXTRACT(p.attributes, '$.specifications') AS specifications,
    JSON_CONTAINS_PATH(p.attributes, 'one', '$.tags[*]') AS has_tags,
    JSON_LENGTH(p.attributes, '$.tags') AS tag_count
FROM products p
WHERE p.generated_price BETWEEN 50.00 AND 200.00
    AND JSON_CONTAINS(p.attributes, JSON_OBJECT('brand', 'Apple'))
    AND p.generated_category IN ('Electronics', 'Computers')
ORDER BY p.generated_price DESC
LIMIT 10;

-- Advanced indexing for performance
CREATE INDEX idx_products_generated_price ON products (generated_price);
CREATE INDEX idx_products_generated_category ON products (generated_category);
CREATE INDEX idx_products_attributes ON products ((CAST(attributes AS CHAR(255) ARRAY)));

-- Optimized query for product search
WITH filtered_products AS (
    SELECT id, name, generated_price, attributes
    FROM products
    WHERE generated_price BETWEEN :min_price AND :max_price
        AND generated_category = :category
        AND MATCH(name) AGAINST(:search_query IN NATURAL LANGUAGE MODE)
),
product_scores AS (
    SELECT 
        fp.*,
        CASE 
            WHEN JSON_CONTAINS_PATH(fp.attributes, 'one', '$.rating') THEN
                JSON_UNQUOTE(JSON_EXTRACT(fp.attributes, '$.rating'))
            ELSE 0
        END AS rating,
        CASE 
            WHEN JSON_CONTAINS_PATH(fp.attributes, 'one', '$.stock') THEN
                JSON_UNQUOTE(JSON_EXTRACT(fp.attributes, '$.stock'))
            ELSE 0
        END AS stock,
        (generated_price * 0.7 + 
         CASE WHEN JSON_CONTAINS_PATH(fp.attributes, 'one', '$.rating') THEN
             JSON_UNQUOTE(JSON_EXTRACT(fp.attributes, '$.rating')) * 100
         ELSE 0 END * 0.3) AS relevance_score
    FROM filtered_products fp
)
SELECT 
    id,
    name,
    generated_price,
    rating,
    stock,
    relevance_score
FROM product_scores
WHERE stock > 0
ORDER BY relevance_score DESC, rating DESC
LIMIT 20;
```

### 3. Advanced SQL Server 2022 Patterns

```sql
-- Advanced window functions and analytics in SQL Server 2022
CREATE TABLE FinancialTransactions (
    TransactionID BIGINT IDENTITY(1,1) PRIMARY KEY,
    AccountID VARCHAR(20) NOT NULL,
    TransactionDate DATETIME2 NOT NULL,
    Amount DECIMAL(18,2) NOT NULL,
    TransactionType VARCHAR(50) NOT NULL,
    CurrencyCode CHAR(3) NOT NULL,
    BranchID INT NOT NULL,
    Category VARCHAR(100) NOT NULL
);

-- Create indexes for optimization
CREATE INDEX IX_FinancialTransactions_AccountID ON FinancialTransactions(AccountID);
CREATE INDEX IX_FinancialTransactions_TransactionDate ON FinancialTransactions(TransactionDate);
CREATE INDEX IX_FinancialTransactions_BranchID ON FinancialTransactions(BranchID);

-- Advanced analytics with window functions and percentiles
WITH AccountMetrics AS (
    SELECT 
        AccountID,
        TransactionDate,
        Amount,
        TransactionType,
        CurrencyCode,
        
        -- Moving averages
        AVG(Amount) OVER (
            PARTITION BY AccountID 
            ORDER BY TransactionDate 
            ROWS BETWEEN 29 PRECEDING AND CURRENT ROW
        ) AS Rolling30DayAvg,
        
        -- Cumulative sums
        SUM(CASE WHEN Amount > 0 THEN Amount ELSE 0 END) OVER (
            PARTITION BY AccountID 
            ORDER BY TransactionDate 
            ROWS UNBOUNDED PRECEDING
        ) AS CumulativeDeposits,
        
        SUM(CASE WHEN Amount < 0 THEN ABS(Amount) ELSE 0 END) OVER (
            PARTITION BY AccountID 
            ORDER BY TransactionDate 
            ROWS UNBOUNDED PRECEDING
        ) AS CumulativeWithdrawals,
        
        -- Running totals
        SUM(Amount) OVER (
            PARTITION BY AccountID 
            ORDER BY TransactionDate 
            ROWS UNBOUNDED PRECEDING
        ) AS RunningBalance,
        
        -- Rank functions
        ROW_NUMBER() OVER (
            PARTITION BY AccountID 
            ORDER BY ABS(Amount) DESC
        ) AS TransactionRank,
        
        PERCENT_RANK() OVER (
            PARTITION BY AccountID 
            ORDER BY ABS(Amount)
        ) AS AmountPercentile,
        
        -- Distribution functions
        PERCENTILE_CONT(0.50) WITHIN GROUP (ORDER BY Amount) OVER (
            PARTITION BY AccountID
        ) AS MedianTransactionAmount
        
    FROM FinancialTransactions
    WHERE TransactionDate >= DATEADD(MONTH, -12, GETDATE())
),
AccountSegments AS (
    SELECT 
        AccountID,
        COUNT(*) AS TransactionCount,
        AVG(Amount) AS AvgTransactionAmount,
        STDDEV(Amount) AS TransactionAmountStdDev,
        MAX(Rolling30DayAvg) AS PeakMonthlyActivity,
        MIN(RunningBalance) AS LowestBalance,
        MAX(RunningBalance) AS HighestBalance,
        
        CASE 
            WHEN COUNT(*) > 100 AND AVG(Amount) > 1000 THEN 'HighVolume'
            WHEN COUNT(*) > 50 AND AVG(Amount) > 500 THEN 'MediumVolume'
            WHEN COUNT(*) > 10 THEN 'LowVolume'
            ELSE 'Inactive'
        END AS VolumeSegment,
        
        CASE 
            WHEN MIN(RunningBalance) < -10000 THEN 'HighRisk'
            WHEN MIN(RunningBalance) < 0 THEN 'MediumRisk'
            ELSE 'LowRisk'
        END AS RiskSegment
        
    FROM AccountMetrics
    GROUP BY AccountID
)
SELECT 
    VolumeSegment,
    RiskSegment,
    COUNT(*) AS AccountCount,
    AVG(TransactionCount) AS AvgTransactionCount,
    AVG(AvgTransactionAmount) AS AvgAmount,
    AVG(TransactionAmountStdDev) AS AvgVariability,
    AVG(PeakMonthlyActivity) AS AvgPeakActivity
FROM AccountSegments
GROUP BY VolumeSegment, RiskSegment
ORDER BY AccountCount DESC;

-- STRING_AGG and JSON aggregation
WITH BranchPerformance AS (
    SELECT 
        BranchID,
        TransactionType,
        CurrencyCode,
        COUNT(*) AS TransactionCount,
        SUM(Amount) AS TotalAmount,
        AVG(Amount) AS AvgAmount,
        STRING_AGG(
            CAST(AccountID AS VARCHAR), ','
        ) WITHIN GROUP (ORDER BY Amount DESC) AS TopAccounts,
        
        -- JSON aggregation for detailed breakdown
        (
            SELECT JSON_QUERY(
                (SELECT 
                    TOP 5 
                    AccountID, 
                    Amount,
                    TransactionDate,
                    ROW_NUMBER() OVER (ORDER BY Amount DESC) as RN
                FROM FinancialTransactions ft2
                WHERE ft2.BranchID = ft.BranchID
                    AND ft2.TransactionType = ft.TransactionType
                FOR JSON PATH
            )
            , '$'
            )
        ) AS TopTransactionsJSON
        
    FROM FinancialTransactions ft
    WHERE TransactionDate >= DATEADD(MONTH, -6, GETDATE())
    GROUP BY BranchID, TransactionType, CurrencyCode
)
SELECT 
    b.BranchID,
    b.BranchName,
    bp.TransactionType,
    bp.CurrencyCode,
    bp.TransactionCount,
    bp.TotalAmount,
    bp.AvgAmount,
    bp.TopAccounts,
    JSON_VALUE(bp.TopTransactionsJSON, '$[0].Amount') AS TopTransactionAmount,
    LEN(bp.TopAccounts) - LEN(REPLACE(bp.TopAccounts, ',', '')) + 1 AS AccountCountInTopList
FROM BranchPerformance bp
JOIN Branches b ON bp.BranchID = b.BranchID
ORDER BY bp.TotalAmount DESC;

-- Advanced temporal tables with system-versioning
CREATE TABLE CustomerHistory (
    CustomerID INT IDENTITY(1,1) PRIMARY KEY,
    CustomerName VARCHAR(100) NOT NULL,
    Email VARCHAR(255) UNIQUE NOT NULL,
    PhoneNumber VARCHAR(20),
    Address NVARCHAR(500),
    CustomerStatus VARCHAR(20) NOT NULL,
    ValidFrom DATETIME2 GENERATED ALWAYS AS ROW START NOT NULL,
    ValidTo DATETIME2 GENERATED ALWAYS AS ROW END NOT NULL,
    PERIOD FOR SYSTEM_TIME (ValidFrom, ValidTo)
) WITH (SYSTEM_VERSIONING = ON (HISTORY_TABLE = dbo.CustomerHistory_History));

-- Query historical data
SELECT 
    CustomerID,
    CustomerName,
    Email,
    CustomerStatus,
    ValidFrom,
    ValidTo,
    DATEDIFF(DAY, ValidFrom, 
        COALESCE(LEAD(ValidFrom) OVER (PARTITION BY CustomerID ORDER BY ValidFrom), GETDATE())
    ) AS DurationInStatus
FROM CustomerHistory
FOR SYSTEM_TIME ALL
WHERE CustomerID = 12345
ORDER BY ValidFrom DESC;

-- Temporal comparison between current and historical
WITH CustomerStatusChanges AS (
    SELECT 
        ch.CustomerID,
        ch.CustomerName,
        ch.CustomerStatus AS OldStatus,
        c.CustomerStatus AS NewStatus,
        ch.ValidFrom AS StatusChangeDate,
        LAG(ch.CustomerStatus) OVER (PARTITION BY ch.CustomerID ORDER BY ch.ValidFrom) AS PreviousStatus
    FROM CustomerHistory ch
    LEFT JOIN Customers c ON ch.CustomerID = c.CustomerID
    WHERE ch.ValidFrom >= DATEADD(YEAR, -1, GETDATE())
)
SELECT 
    CustomerID,
    CustomerName,
    PreviousStatus,
    OldStatus,
    NewStatus,
    StatusChangeDate,
    DATEDIFF(DAY, StatusChangeDate, GETDATE()) AS DaysInCurrentStatus
FROM CustomerStatusChanges
WHERE PreviousStatus IS NOT NULL
    AND PreviousStatus <> NewStatus
ORDER BY StatusChangeDate DESC;
