# Apache Spark 3.5

Comprehensive coverage of Apache Spark 3.5 including DataFrame API, SQL, and Structured Streaming.

## Context7 Library Mappings

```
/apache/spark - Spark 3.5 DataFrame and SQL
/delta-io/delta - Delta Lake 3.0
/apache/kafka - Kafka Clients 3.7
/apache/iceberg - Apache Iceberg
```

---

## SparkSession Setup

### Basic Configuration

```scala
import org.apache.spark.sql.{DataFrame, SparkSession}
import org.apache.spark.sql.functions.*

val spark = SparkSession.builder()
  .appName("Data Analysis")
  .master("local[*]")  // For local development
  .config("spark.sql.adaptive.enabled", "true")
  .config("spark.sql.shuffle.partitions", "200")
  .config("spark.serializer", "org.apache.spark.serializer.KryoSerializer")
  .getOrCreate()

import spark.implicits.*

// Logging level
spark.sparkContext.setLogLevel("WARN")
```

### Production Configuration

```scala
val spark = SparkSession.builder()
  .appName("Production Job")
  .config("spark.sql.adaptive.enabled", "true")
  .config("spark.sql.adaptive.coalescePartitions.enabled", "true")
  .config("spark.sql.adaptive.skewJoin.enabled", "true")
  .config("spark.sql.autoBroadcastJoinThreshold", "10485760")  // 10MB
  .config("spark.sql.shuffle.partitions", "auto")
  .config("spark.dynamicAllocation.enabled", "true")
  .config("spark.dynamicAllocation.minExecutors", "2")
  .config("spark.dynamicAllocation.maxExecutors", "100")
  .enableHiveSupport()
  .getOrCreate()
```

---

## DataFrame Operations

### Reading Data

```scala
// Parquet (columnar, compressed)
val parquetDf = spark.read.parquet("s3://bucket/data.parquet")

// CSV with schema inference
val csvDf = spark.read
  .option("header", "true")
  .option("inferSchema", "true")
  .csv("data/*.csv")

// JSON
val jsonDf = spark.read
  .option("multiLine", "true")
  .json("data.json")

// With explicit schema
import org.apache.spark.sql.types.*

val schema = StructType(Seq(
  StructField("id", LongType, nullable = false),
  StructField("name", StringType, nullable = true),
  StructField("email", StringType, nullable = true),
  StructField("created_at", TimestampType, nullable = true)
))

val dfWithSchema = spark.read.schema(schema).parquet("users.parquet")
```

### Basic Transformations

```scala
// Select and rename
val selected = df
  .select("id", "name", "email")
  .withColumnRenamed("name", "user_name")

// Filter
val active = df.filter(col("status") === "active")
val recent = df.where($"created_at" > lit("2024-01-01"))

// Add columns
val enriched = df
  .withColumn("full_name", concat($"first_name", lit(" "), $"last_name"))
  .withColumn("processed_at", current_timestamp())
  .withColumn("year", year($"created_at"))

// Drop columns
val cleaned = df.drop("temp_col", "debug_info")

// Distinct
val unique = df.dropDuplicates("email")
val distinctAll = df.distinct()
```

### Aggregations

```scala
// Basic aggregations
val stats = orders
  .groupBy("user_id")
  .agg(
    sum("amount").as("total_spent"),
    count("*").as("order_count"),
    avg("amount").as("avg_order_value"),
    min("created_at").as("first_order"),
    max("created_at").as("last_order"),
    collect_list("product_id").as("products")
  )

// Window functions
import org.apache.spark.sql.expressions.Window

val windowSpec = Window
  .partitionBy("user_id")
  .orderBy($"created_at".desc)

val ranked = orders
  .withColumn("row_num", row_number().over(windowSpec))
  .withColumn("rank", rank().over(windowSpec))
  .withColumn("running_total", sum("amount").over(
    Window.partitionBy("user_id").orderBy("created_at")
      .rowsBetween(Window.unboundedPreceding, Window.currentRow)
  ))
```

### Joins

```scala
// Inner join
val joined = orders
  .join(users, Seq("user_id"), "inner")

// Left join with alias
val ordersAlias = orders.as("o")
val usersAlias = users.as("u")

val leftJoined = ordersAlias
  .join(usersAlias, $"o.user_id" === $"u.id", "left")
  .select(
    $"o.id".as("order_id"),
    $"u.name".as("user_name"),
    $"o.amount"
  )

// Broadcast join (for small tables)
val broadcastJoined = orders
  .join(broadcast(countries), Seq("country_code"))

// Cross join (use carefully)
val cross = df1.crossJoin(df2)
```

---

## Spark SQL

### SQL Queries

```scala
// Register temp view
df.createOrReplaceTempView("users")

// Run SQL
val result = spark.sql("""
  SELECT 
    u.id,
    u.name,
    COUNT(o.id) as order_count,
    SUM(o.amount) as total_spent
  FROM users u
  LEFT JOIN orders o ON u.id = o.user_id
  WHERE u.status = 'active'
  GROUP BY u.id, u.name
  HAVING SUM(o.amount) > 1000
  ORDER BY total_spent DESC
""")

// Complex SQL with CTEs
val cteResult = spark.sql("""
  WITH monthly_orders AS (
    SELECT 
      user_id,
      DATE_TRUNC('month', created_at) as month,
      SUM(amount) as monthly_total
    FROM orders
    GROUP BY user_id, DATE_TRUNC('month', created_at)
  ),
  user_trends AS (
    SELECT 
      user_id,
      month,
      monthly_total,
      LAG(monthly_total) OVER (PARTITION BY user_id ORDER BY month) as prev_month
    FROM monthly_orders
  )
  SELECT 
    user_id,
    month,
    monthly_total,
    (monthly_total - prev_month) / prev_month * 100 as growth_pct
  FROM user_trends
  WHERE prev_month IS NOT NULL
""")
```

### User Defined Functions

```scala
// Simple UDF
val toUpperCase = udf((s: String) => Option(s).map(_.toUpperCase).orNull)

val result = df.withColumn("upper_name", toUpperCase($"name"))

// Type-safe UDF
import org.apache.spark.sql.expressions.UserDefinedFunction

val parseJson: UserDefinedFunction = udf { json: String =>
  import io.circe.parser.*
  parse(json).toOption.flatMap(_.as[Map[String, String]].toOption)
}

// Register for SQL
spark.udf.register("parse_json", parseJson)
```

---

## Structured Streaming

### Basic Streaming

```scala
// Read from Kafka
val kafkaStream = spark.readStream
  .format("kafka")
  .option("kafka.bootstrap.servers", "localhost:9092")
  .option("subscribe", "events")
  .option("startingOffsets", "earliest")
  .load()
  .selectExpr("CAST(value AS STRING) as json")

// Parse JSON events
import org.apache.spark.sql.types.*

val eventSchema = StructType(Seq(
  StructField("event_id", StringType),
  StructField("event_type", StringType),
  StructField("user_id", LongType),
  StructField("timestamp", TimestampType),
  StructField("data", MapType(StringType, StringType))
))

val events = kafkaStream
  .select(from_json($"json", eventSchema).as("event"))
  .select("event.*")
```

### Windowed Aggregations

```scala
val aggregated = events
  .withWatermark("timestamp", "10 minutes")
  .groupBy(
    window($"timestamp", "1 hour", "15 minutes"),
    $"event_type"
  )
  .agg(
    count("*").as("event_count"),
    approx_count_distinct("user_id").as("unique_users")
  )

// Write stream
val query = aggregated.writeStream
  .format("delta")
  .outputMode("append")
  .option("checkpointLocation", "/checkpoints/events")
  .partitionBy("date")
  .start("/output/events")

query.awaitTermination()
```

### Stream-Stream Joins

```scala
val orders = spark.readStream
  .format("kafka")
  .option("subscribe", "orders")
  .load()
  .selectExpr("CAST(value AS STRING)")
  .select(from_json($"value", orderSchema).as("order"))
  .select("order.*")
  .withWatermark("order_time", "1 hour")

val payments = spark.readStream
  .format("kafka")
  .option("subscribe", "payments")
  .load()
  .selectExpr("CAST(value AS STRING)")
  .select(from_json($"value", paymentSchema).as("payment"))
  .select("payment.*")
  .withWatermark("payment_time", "1 hour")

val joined = orders.join(
  payments,
  expr("""
    order_id = payment_order_id AND
    payment_time >= order_time AND
    payment_time <= order_time + interval 1 hour
  """),
  "leftOuter"
)
```

---

## Delta Lake Integration

### Delta Operations

```scala
// Write as Delta
df.write
  .format("delta")
  .mode("overwrite")
  .partitionBy("year", "month")
  .save("/data/users")

// Read Delta
val deltaDf = spark.read.format("delta").load("/data/users")

// Time travel
val historyDf = spark.read
  .format("delta")
  .option("versionAsOf", 5)
  .load("/data/users")

val timestampDf = spark.read
  .format("delta")
  .option("timestampAsOf", "2024-01-01")
  .load("/data/users")
```

### MERGE (Upsert)

```scala
import io.delta.tables.*

val deltaTable = DeltaTable.forPath(spark, "/data/users")

deltaTable.as("target")
  .merge(
    updates.as("source"),
    "target.id = source.id"
  )
  .whenMatched.updateAll()
  .whenNotMatched.insertAll()
  .execute()
```

---

## Performance Optimization

### Caching and Persistence

```scala
import org.apache.spark.storage.StorageLevel

// Cache in memory
df.cache()

// With storage level
df.persist(StorageLevel.MEMORY_AND_DISK_SER)

// Unpersist when done
df.unpersist()
```

### Partitioning

```scala
// Repartition for parallelism
val repartitioned = df.repartition(100)

// Coalesce to reduce partitions
val coalesced = df.coalesce(10)

// Repartition by column
val byUser = df.repartition($"user_id")

// Custom partitioner
val rangePartitioned = df.repartitionByRange(100, $"created_at")
```

### Optimization Tips

```scala
// Filter early
val optimized = df
  .filter($"status" === "active")  // Filter first
  .join(largeTable, ...)           // Then join

// Project early
val projected = df
  .select("id", "name", "amount")  // Select needed columns
  .groupBy("name")
  .sum("amount")

// Avoid shuffles
val skipShuffle = df
  .repartition($"user_id")  // Partition once
  .cache()                   // Cache
// Multiple joins on user_id won't shuffle

// Broadcast small tables
val broadcasted = largeTable.join(
  broadcast(smallLookup),
  Seq("key")
)
```

---

## Writing Data

### Output Formats

```scala
// Parquet (recommended)
df.write
  .mode("overwrite")
  .partitionBy("year", "month")
  .parquet("/output/data")

// Delta Lake
df.write
  .format("delta")
  .mode("append")
  .save("/output/delta")

// CSV
df.write
  .option("header", "true")
  .csv("/output/csv")

// Single file
df.coalesce(1).write.json("/output/single.json")
```

---

## Best Practices

Data Reading:
- Always use explicit schemas for production
- Prefer Parquet/Delta over CSV/JSON
- Use pushdown predicates for data sources

Transformations:
- Filter and project early
- Avoid wide transformations when possible
- Use broadcast joins for small tables

Memory:
- Cache only when reusing DataFrames
- Unpersist when done
- Monitor memory with Spark UI

Streaming:
- Set appropriate watermarks
- Use trigger intervals for throughput
- Monitor lag in checkpoints

---

Last Updated: 2026-01-06
Version: 2.0.0
