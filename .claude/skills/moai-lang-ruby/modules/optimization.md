# Ruby Performance Optimization Guide

## YJIT (Just-In-Time Compiler)

### Enable and Configure YJIT

```ruby
# Method 1: Environment variable
# RUBY_YJIT_ENABLE=1 rails server

# Method 2: Programmatic
RubyVM::YJIT.enable

# Check status
if RubyVM::YJIT.enabled?
  puts "YJIT is running"
  puts RubyVM::YJIT.stats
end

# Configuration options
ENV['RUBY_YJIT_EXEC_MEM_SIZE'] = '128'  # MB for JIT code
ENV['RUBY_YJIT_MIN_CALLS'] = '1000'     # Threshold before JIT
```

### Expected Performance Gains

- **Small applications**: 10-20% improvement
- **Moderate applications**: 15-25% improvement
- **CPU-heavy workloads**: 25-50% improvement
- **I/O-heavy applications**: Minimal (1-5%)

---

## Database Query Optimization

### N+1 Query Prevention

```ruby
# BAD: N+1 Query Problem
users = User.all
users.each do |user|
  puts user.posts.count  # N queries
end

# GOOD: Eager Loading with includes
users = User.includes(:posts)
users.each do |user|
  puts user.posts.size  # Uses loaded data
end

# GOOD: Eager Loading with joins
users = User.joins(:posts)
users.each do |user|
  puts user.posts.count
end

# ADVANCED: Batch loading with left_outer_joins
users = User.left_outer_joins(:posts)
              .select('users.*, COUNT(posts.id) as posts_count')
              .group('users.id')

users.each { |user| puts user.posts_count }
```

### Counter Caching

```ruby
# Model definition
class Post < ApplicationRecord
  belongs_to :user, counter_cache: true
end

# Add migration
class AddPostsCountToUsers < ActiveRecord::Migration[7.0]
  def change
    add_column :users, :posts_count, :integer, default: 0
  end
end

# Usage - no extra query needed
user.posts_count  # Cached in database

# Rebuild counter cache
User.reset_counters(user_id, :posts)
```

### Batch Processing

```ruby
# INEFFICIENT: Load all into memory
User.all.each { |user| user.process_data }

# EFFICIENT: Batch processing
User.find_each(batch_size: 1000) do |user|
  user.process_data
end

# For updates
User.find_in_batches(batch_size: 500) do |batch|
  batch.update_all(last_seen: Time.current)
end
```

### Column Selection

```ruby
# INEFFICIENT: Load all columns
users = User.all

# EFFICIENT: Pluck specific columns
ids_and_emails = User.where(active: true).pluck(:id, :email)
# => [["1", "user@example.com"], ["2", "another@example.com"]]

# EFFICIENT: Select specific columns
users = User.where(active: true).select(:id, :name, :email)
```

---

## Caching Strategies

### Fragment Caching in Views

```erb
<!-- Cache user profile section -->
<% cache(@user, expires_in: 1.hour) do %>
  <div class="profile">
    <%= @user.name %>
    <%= @user.email %>
    <%= render_posts(@user) %>
  </div>
<% end %>

<!-- Cache with dependencies -->
<% cache([@user, @user.posts], expires_in: 1.day) do %>
  <!-- Content automatically invalidates on post changes -->
<% end %>
```

### Low-Level Caching

```ruby
class UserRepository
  def self.find_with_cache(id)
    Rails.cache.fetch("user_#{id}", expires_in: 1.hour) do
      User.find(id)
    end
  end

  def self.recent_users
    Rails.cache.fetch('recent_users', expires_in: 30.minutes) do
      User.recent.limit(10)
    end
  end
end

# Cache invalidation
after_save do
  Rails.cache.delete("user_#{id}")
  Rails.cache.delete('recent_users')
end
```

### Redis Caching Backend

```ruby
# config/environments/production.rb
config.cache_store = :redis_cache_store,
  { url: ENV['REDIS_URL'] }

# Advanced Redis caching
class RedisCacheManager
  def self.set_with_expiry(key, value, ttl = 3600)
    Rails.cache.write(key, value, expires_in: ttl.seconds)
  end

  def self.increment(key, amount = 1)
    Rails.cache.increment(key, amount)
  end

  def self.delete_pattern(pattern)
    Rails.cache.redis.keys(pattern).each do |key|
      Rails.cache.delete(key)
    end
  end
end
```

---

## Memory Optimization

### Object Allocation Reduction

```ruby
# INEFFICIENT: Creates new strings constantly
result = ""
1000.times { result += "x" }

# EFFICIENT: String builder pattern
result = String.new
1000.times { result << "x" }

# EFFICIENT: Use join
result = Array.new(1000, "x").join

# EFFICIENT: Use string interpolation
data = "#{data_source}"
```

### Lazy Evaluation

```ruby
# INEFFICIENT: Evaluates entire collection
users = User.all.map { |u| u.name }.select { |n| n.present? }

# EFFICIENT: Lazy evaluation stops early
users = User.all.lazy
             .map { |u| u.name }
             .select { |n| n.present? }
             .force

# EFFICIENT: Use enumerator chains
results = (1..1000000)
  .lazy
  .select { |n| n.even? }
  .map { |n| n * 2 }
  .take(10)
  .force
```

---

## Concurrency for I/O Bound Operations

### Fiber Pooling

```ruby
require 'async'

class AsyncTaskProcessor
  def self.process_batch(items, concurrency: 10)
    Async do |task|
      queue = items.dup
      results = []

      concurrency.times do
        task.async do
          while item = queue.shift
            result = yield(item)
            results << result
          end
        end
      end

      task.wait
      results
    end
  end
end

# Usage
AsyncTaskProcessor.process_batch(users) do |user|
  fetch_external_data(user)
end
```

### Parallel Execution with Ractors

```ruby
class ParallelProcessor
  def self.map(items, &block)
    ractors = items.map do |item|
      Ractor.new(item, &block)
    end

    ractors.map(&:take)
  end

  def self.reduce(items, initial = nil, &block)
    ractors = items.each_slice(items.size / 4).map do |slice|
      Ractor.new(slice) { |s| s.reduce(&block) }
    end

    ractors.map(&:take).reduce(&block)
  end
end

# Usage
ParallelProcessor.map(expensive_items) { |item| compute(item) }
```

---

## Method Dispatch Optimization

### Avoid Reflection Overhead

```ruby
# SLOW: Method dispatch through public_send
class SlowProxy
  def initialize(target)
    @target = target
  end

  def method_missing(method, *args)
    @target.public_send(method, *args)
  end
end

# FAST: Direct delegation
class FastProxy
  def initialize(target)
    @target = target
  end

  def name
    @target.name
  end

  def email
    @target.email
  end
end

# BEST: Use delegate macro
class BestProxy
  delegate :name, :email, to: :@target

  def initialize(target)
    @target = target
  end
end
```

### Inline Caching for Frequently Called Methods

```ruby
class OptimizedModel
  def expensive_calculation
    @cached_result ||= perform_expensive_work
  end

  def refresh_cache
    @cached_result = nil
  end

  private

  def perform_expensive_work
    sleep 2  # Simulated expensive operation
  end
end
```

---

## Request Optimization in Rails

### Controller-Level Caching

```ruby
class PostsController < ApplicationController
  before_action :set_cache_headers, only: [:show, :index]

  def show
    @post = Post.find(params[:id])
    fresh_when(@post)  # Sets etag/last-modified
  end

  def index
    @posts = Post.includes(:author).limit(20)
    expires_in 30.minutes, public: true
  end

  private

  def set_cache_headers
    response.headers['Cache-Control'] = 'public, max-age=3600'
  end
end
```

### Database Connection Pooling

```ruby
# config/database.yml
production:
  adapter: postgresql
  pool: 25           # Connection pool size
  timeout: 5000      # Connection timeout in ms
  reap_frequency: 10 # Reap idle connections every 10s

# Verify pool configuration
ActiveRecord::Base.connection_pool.size
ActiveRecord::Base.connection_pool.idle_connections.size
```

---

## Profiling and Benchmarking

### Benchmarking Code Segments

```ruby
require 'benchmark'

# Simple timing
time = Benchmark.measure { expensive_operation }
puts time.real

# Comparing multiple approaches
Benchmark.bm do |x|
  x.report("map:") { users.map(&:name) }
  x.report("each:") { users.each { |u| u.name } }
  x.report("pluck:") { users.pluck(:name) }
end
```

### Memory Profiling

```ruby
require 'memory_profiler'

report = MemoryProfiler.report do
  User.includes(:posts).to_a
end

report.pretty_print
```

### Query Logging and Analysis

```ruby
# config/initializers/query_logging.rb
if Rails.env.development?
  ActiveRecord::LogSubscriber.begin_transaction_event_logging = true

  ActiveSupport::Notifications.subscribe 'sql.active_record' do |*args|
    event = ActiveSupport::Notifications::Event.new(*args)
    puts "#{event.duration}ms - #{event.payload[:sql]}"
  end
end
```

---

**Best Practices Summary**:
1. Enable YJIT for automatic 15-30% speedup
2. Use eager loading to prevent N+1 queries
3. Implement caching at all levels (HTTP, fragment, object)
4. Profile before optimizing - measure improvements
5. Use batch processing for large datasets
6. Leverage concurrency for I/O operations
7. Monitor connection pool and database performance
8. Cache computed properties with memoization
