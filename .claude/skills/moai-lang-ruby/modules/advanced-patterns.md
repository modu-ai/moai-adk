# Ruby Advanced Patterns

## Sidekiq Background Jobs

Job Definition:
```ruby
# app/jobs/process_order_job.rb
class ProcessOrderJob < ApplicationJob
  queue_as :default
  retry_on ActiveRecord::Deadlocked, wait: 5.seconds, attempts: 3
  discard_on ActiveJob::DeserializationError

  def perform(order_id)
    order = Order.find(order_id)

    ActiveRecord::Base.transaction do
      order.process!
      order.update!(processed_at: Time.current)
      OrderMailer.confirmation(order).deliver_later
    end
  end
end

# Sidekiq configuration
# config/initializers/sidekiq.rb
Sidekiq.configure_server do |config|
  config.redis = { url: ENV.fetch("REDIS_URL", "redis://localhost:6379/1") }
end

Sidekiq.configure_client do |config|
  config.redis = { url: ENV.fetch("REDIS_URL", "redis://localhost:6379/1") }
end
```

## ActiveRecord Advanced Patterns

Scopes and Query Objects:
```ruby
class Post < ApplicationRecord
  scope :published, -> { where(published: true) }
  scope :recent, -> { order(created_at: :desc) }
  scope :by_author, ->(author) { where(author: author) }
  scope :search, ->(query) { where("title ILIKE ?", "%#{query}%") }

  # Complex scope with joins
  scope :with_comments, -> {
    joins(:comments).group(:id).having("COUNT(comments.id) > 0")
  }

  # Scope returning specific columns
  scope :titles_only, -> { select(:id, :title) }
end

# Query Object
class PostSearchQuery
  def initialize(relation = Post.all)
    @relation = relation
  end

  def call(params)
    @relation
      .then { |r| filter_by_status(r, params[:status]) }
      .then { |r| filter_by_date(r, params[:start_date], params[:end_date]) }
      .then { |r| search_by_title(r, params[:query]) }
      .then { |r| paginate(r, params[:page], params[:per_page]) }
  end

  private

  def filter_by_status(relation, status)
    return relation if status.blank?
    relation.where(status: status)
  end

  def filter_by_date(relation, start_date, end_date)
    relation = relation.where("created_at >= ?", start_date) if start_date
    relation = relation.where("created_at <= ?", end_date) if end_date
    relation
  end

  def search_by_title(relation, query)
    return relation if query.blank?
    relation.where("title ILIKE ?", "%#{query}%")
  end

  def paginate(relation, page, per_page)
    page ||= 1
    per_page ||= 25
    relation.limit(per_page).offset((page.to_i - 1) * per_page.to_i)
  end
end
```

## Production Deployment Patterns

Docker Configuration:
```dockerfile
# Dockerfile
FROM ruby:3.3-slim

RUN apt-get update -qq && apt-get install -y \
    build-essential \
    libpq-dev \
    nodejs \
    npm \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY Gemfile Gemfile.lock ./
RUN bundle config set --local deployment 'true' && \
    bundle config set --local without 'development test' && \
    bundle install

COPY . .

RUN bundle exec rails assets:precompile

ENV RAILS_ENV=production
ENV RAILS_LOG_TO_STDOUT=true

EXPOSE 3000

CMD ["bundle", "exec", "puma", "-C", "config/puma.rb"]
```

Puma Configuration:
```ruby
# config/puma.rb
workers ENV.fetch("WEB_CONCURRENCY") { 2 }
threads_count = ENV.fetch("RAILS_MAX_THREADS") { 5 }
threads threads_count, threads_count

preload_app!

port ENV.fetch("PORT") { 3000 }
environment ENV.fetch("RAILS_ENV") { "development" }

plugin :tmp_restart

on_worker_boot do
  ActiveRecord::Base.establish_connection if defined?(ActiveRecord)
end
```

## Advanced ActiveRecord Patterns

Polymorphic Associations:
```ruby
class Comment < ApplicationRecord
  belongs_to :commentable, polymorphic: true
  belongs_to :user
end

class Post < ApplicationRecord
  has_many :comments, as: :commentable, dependent: :destroy
end

class Image < ApplicationRecord
  has_many :comments, as: :commentable, dependent: :destroy
end
```

Single Table Inheritance (STI):
```ruby
class Vehicle < ApplicationRecord
  validates :brand, presence: true
end

class Car < Vehicle
  validates :doors, numericality: { greater_than: 0 }
end

class Motorcycle < Vehicle
  validates :engine_cc, numericality: { greater_than: 0 }
end
```

## Action Cable Real-Time Features

Channel Implementation:
```ruby
# app/channels/notification_channel.rb
class NotificationChannel < ApplicationCable::Channel
  def subscribed
    stream_from "notifications:#{current_user.id}"
  end

  def unsubscribed
    stop_all_streams
  end
end

# Broadcasting notifications
class NotificationService
  def self.notify(user, message)
    ActionCable.server.broadcast(
      "notifications:#{user.id}",
      { message: message, timestamp: Time.current.iso8601 }
    )
  end
end
```

## Performance Optimization

N+1 Query Prevention:
```ruby
# Bad: N+1 queries
User.all.each do |user|
  puts user.posts.count  # Executes a query for each user
end

# Good: Eager loading
User.includes(:posts).each do |user|
  puts user.posts.size  # Uses preloaded data
end

# Counter cache for counts
class Post < ApplicationRecord
  belongs_to :user, counter_cache: true
end

# Migration for counter cache
add_column :users, :posts_count, :integer, default: 0
```

Database Indexing:
```ruby
# migration
class AddIndexesToPosts < ActiveRecord::Migration[7.2]
  def change
    add_index :posts, :user_id
    add_index :posts, :status
    add_index :posts, [:status, :published_at]
    add_index :posts, :title, using: :gin, opclass: :gin_trgm_ops
  end
end
```

## Security Best Practices

Strong Parameters:
```ruby
class UsersController < ApplicationController
  private

  def user_params
    params.require(:user).permit(:name, :email, :password, :password_confirmation)
  end
end
```

CSRF Protection:
```ruby
class ApplicationController < ActionController::Base
  protect_from_forgery with: :exception
end
```

SQL Injection Prevention:
```ruby
# Bad: Direct interpolation
User.where("email = '#{params[:email]}'")

# Good: Parameterized query
User.where(email: params[:email])
User.where("email = ?", params[:email])
```

## CI/CD Integration

GitHub Actions Workflow:
```yaml
# .github/workflows/ci.yml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: postgres
        ports:
          - 5432:5432

    steps:
      - uses: actions/checkout@v4
      - uses: ruby/setup-ruby@v1
        with:
          ruby-version: '3.3'
          bundler-cache: true

      - name: Setup database
        env:
          DATABASE_URL: postgres://postgres:postgres@localhost:5432/test
        run: |
          bundle exec rails db:create db:migrate

      - name: Run tests
        env:
          DATABASE_URL: postgres://postgres:postgres@localhost:5432/test
        run: bundle exec rspec
```
