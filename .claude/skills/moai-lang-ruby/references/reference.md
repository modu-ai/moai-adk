# Ruby 3.3+ Complete Reference

## Language Features Reference

### Ruby 3.3 Feature Matrix

| Feature               | Status       | Production Ready | Performance Impact |
| --------------------- | ------------ | ---------------- | ------------------ |
| YJIT (JIT Compiler)   | Stable       | Yes              | 15-20% faster      |
| Pattern Matching      | Stable       | Yes              | Minimal            |
| Data Class            | Stable       | Yes              | Minimal            |
| Endless Method        | Stable       | Yes              | None               |
| Error Highlighting    | Stable       | Yes              | N/A                |
| Hash Shorthand Syntax | Experimental | Limited          | None               |
| Prism Parser          | Experimental | No               | N/A                |

### YJIT Configuration

Enable YJIT:

```ruby
# Option 1: Environment variable
RUBY_YJIT_ENABLE=1 ruby app.rb

# Option 2: Command line flag
ruby --yjit app.rb

# Option 3: Programmatically
RubyVM::YJIT.enable
```

Check YJIT Status:

```ruby
# Check if YJIT is enabled
RubyVM::YJIT.enabled?  # => true/false

# Get YJIT statistics
RubyVM::YJIT.runtime_stats
# => {
#   compiled_iseq_count: 1234,
#   compiled_block_count: 5678,
#   invalidation_count: 10,
#   ...
# }
```

YJIT Performance Best Practices:

- Enable YJIT in production for Rails apps (15-20% performance improvement)
- Works best with long-running processes (background workers, web servers)
- Requires additional memory (approximately 50-100MB)
- Most effective for CPU-bound workloads

### Pattern Matching Complete Guide

Basic Patterns:

```ruby
# Literal pattern
case value
in 42
  "The answer"
in "hello"
  "Greeting"
end

# Variable binding
case [1, 2, 3]
in [a, b, c]
  puts "a=#{a}, b=#{b}, c=#{c}"
end

# Array patterns with splat
case [1, 2, 3, 4, 5]
in [first, *middle, last]
  puts "first=#{first}, middle=#{middle}, last=#{last}"
end
```

Hash Patterns:

```ruby
# Basic hash pattern
case { name: "John", age: 30 }
in { name: n, age: a }
  puts "#{n} is #{a} years old"
end

# Hash pattern with rest
case { name: "John", age: 30, city: "NYC" }
in { name:, **rest }
  puts "Name: #{name}, Rest: #{rest}"
end

# Nested patterns
case { user: { name: "John", profile: { age: 30 } } }
in { user: { name:, profile: { age: } } }
  puts "#{name} is #{age}"
end
```

Guard Clauses:

```ruby
case user
in { role: "admin", verified: true } if user.active?
  "Active admin"
in { role: "admin" }
  "Inactive admin"
in { verified: true }
  "Verified user"
else
  "Regular user"
end
```

### Data Class (Immutable Structs)

Basic Usage:

```ruby
# Define Data class
Point = Data.define(:x, :y)

# Create instances
p1 = Point.new(x: 1, y: 2)
p2 = Point.new(1, 2)  # Positional arguments

# Immutable
p1.x = 5  # NoMethodError: undefined method `x='

# Create modified copy
p3 = p1.with(x: 5)  # Point(x: 5, y: 2)
```

Advanced Data Patterns:

```ruby
# Data with methods
User = Data.define(:name, :email) do
  def greeting
    "Hello, #{name}!"
  end

  def domain
    email.split("@").last
  end
end

# Data with validation
SafeUser = Data.define(:name, :email) do
  def initialize(name:, email:)
    raise ArgumentError, "Invalid email" unless email.include?("@")
    super
  end
end
```

### Endless Method Definition

Single-line Methods:

```ruby
class Calculator
  def add(a, b) = a + b
  def subtract(a, b) = a - b
  def multiply(a, b) = a * b
  def divide(a, b) = a / b
end

class User
  def full_name = "#{first_name} #{last_name}"
  def active? = status == "active"
  def admin? = role == "admin"
end
```

---

## Rails 7.2 Component Reference

### Rails 7.2 New Features

| Feature               | Description                            | Availability |
| --------------------- | -------------------------------------- | ------------ |
| Async Queries         | Non-blocking database queries          | Stable       |
| Dev Containers        | Dockerized development environment     | Stable       |
| Bun Support           | JavaScript runtime alternative to Node | Stable       |
| Better Health Checks  | Built-in health check endpoints        | Stable       |
| Query Log Tags        | Better query tracking in logs          | Stable       |
| Trilogy MySQL Adapter | Modern MySQL adapter                   | Experimental |

### ActiveRecord 7.2 Reference

Query Interface:

```ruby
# Basic queries
User.find(1)
User.find_by(email: "user@example.com")
User.where(active: true)
User.where("age > ?", 18)
User.where.not(role: "banned")

# Ordering
User.order(created_at: :desc)
User.order("created_at DESC, name ASC")
User.reorder(name: :asc)  # Replace existing order

# Limiting
User.limit(10)
User.offset(20)
User.first(5)
User.last(5)

# Joins
User.joins(:posts)
User.left_joins(:posts)
User.includes(:posts)  # Eager loading
User.preload(:posts)   # Separate queries
User.eager_load(:posts)  # LEFT OUTER JOIN

# Aggregations
User.count
User.sum(:age)
User.average(:age)
User.minimum(:age)
User.maximum(:age)
User.group(:role).count
```

Scopes:

```ruby
class Post < ApplicationRecord
  # Basic scope
  scope :published, -> { where(published: true) }
  scope :draft, -> { where(published: false) }

  # Parameterized scope
  scope :by_author, ->(author_id) { where(author_id: author_id) }
  scope :recent, ->(days = 7) { where("created_at > ?", days.days.ago) }

  # Chainable scopes
  scope :popular, -> { where("views > ?", 1000) }
  scope :trending, -> { popular.recent(7) }

  # Default scope (use sparingly)
  default_scope { order(created_at: :desc) }
end

# Usage
Post.published.recent.limit(10)
Post.by_author(user.id).popular
```

Associations:

```ruby
class User < ApplicationRecord
  # One-to-many
  has_many :posts, dependent: :destroy
  has_many :comments, dependent: :destroy

  # Many-to-many
  has_many :memberships
  has_many :teams, through: :memberships

  # One-to-one
  has_one :profile, dependent: :destroy

  # Polymorphic
  has_many :images, as: :imageable

  # Association with options
  has_many :published_posts,
           -> { where(published: true) },
           class_name: "Post"

  has_many :recent_posts,
           -> { order(created_at: :desc).limit(5) },
           class_name: "Post"
end

class Post < ApplicationRecord
  belongs_to :user
  belongs_to :category, optional: true
  has_many :comments, dependent: :destroy
  has_many_attached :images
end
```

Validations:

```ruby
class User < ApplicationRecord
  # Presence
  validates :name, presence: true
  validates :email, presence: { message: "can't be blank" }

  # Uniqueness
  validates :email, uniqueness: true
  validates :username, uniqueness: { case_sensitive: false }

  # Format
  validates :email, format: { with: URI::MailTo::EMAIL_REGEXP }
  validates :phone, format: { with: /\A\d{10}\z/ }

  # Length
  validates :name, length: { minimum: 2, maximum: 100 }
  validates :bio, length: { maximum: 500 }

  # Numericality
  validates :age, numericality: { only_integer: true, greater_than: 0 }
  validates :price, numericality: { greater_than_or_equal_to: 0 }

  # Inclusion/Exclusion
  validates :role, inclusion: { in: %w[user admin moderator] }
  validates :subdomain, exclusion: { in: %w[www admin] }

  # Custom validation
  validate :acceptable_email_domain

  private

  def acceptable_email_domain
    return if email.blank?

    domain = email.split("@").last
    errors.add(:email, "domain not allowed") unless allowed_domains.include?(domain)
  end
end
```

Callbacks:

```ruby
class User < ApplicationRecord
  # Before callbacks
  before_validation :normalize_email
  before_save :encrypt_password
  before_create :generate_auth_token

  # After callbacks
  after_create :send_welcome_email
  after_update :notify_changes
  after_destroy :cleanup_assets

  # Around callbacks
  around_save :log_save_time

  # Conditional callbacks
  after_create :send_admin_notification, if: :admin?
  before_destroy :prevent_deletion, unless: :deletable?

  private

  def normalize_email
    self.email = email.downcase.strip if email.present?
  end

  def log_save_time
    start_time = Time.current
    yield
    Rails.logger.info "Save took #{Time.current - start_time} seconds"
  end
end
```

### Hotwire/Turbo Reference

Turbo Drive:

```ruby
# Disable Turbo for specific links
<%= link_to "Legacy Page", legacy_path, data: { turbo: false } %>

# Turbo method
<%= link_to "Delete", user_path(@user), data: { turbo_method: :delete } %>

# Turbo confirmation
<%= link_to "Delete", user_path(@user),
      data: { turbo_method: :delete, turbo_confirm: "Are you sure?" } %>
```

Turbo Frames:

```erb
<!-- Parent page -->
<%= turbo_frame_tag "messages" do %>
  <%= render @messages %>
<% end %>

<!-- Partial -->
<%= turbo_frame_tag dom_id(message) do %>
  <div class="message">
    <%= message.content %>
    <%= link_to "Edit", edit_message_path(message) %>
  </div>
<% end %>

<!-- Lazy loading -->
<%= turbo_frame_tag "lazy_content", src: lazy_content_path %>
```

Turbo Streams:

```ruby
# Controller
def create
  @post = Post.create(post_params)

  respond_to do |format|
    format.turbo_stream
    format.html { redirect_to @post }
  end
end

# create.turbo_stream.erb
<%= turbo_stream.prepend "posts", @post %>
<%= turbo_stream.update "post_form", partial: "form", locals: { post: Post.new } %>

# Available actions
turbo_stream.append(target, content)
turbo_stream.prepend(target, content)
turbo_stream.replace(target, content)
turbo_stream.update(target, content)
turbo_stream.remove(target)
turbo_stream.before(target, content)
turbo_stream.after(target, content)
```

Stimulus Controllers:

```javascript
// app/javascript/controllers/dropdown_controller.js
import { Controller } from "@hotwired/stimulus";

export default class extends Controller {
  static targets = ["menu"];
  static classes = ["open"];
  static values = {
    openDuration: { type: Number, default: 100 },
  };

  connect() {
    this.element.classList.add("dropdown");
  }

  toggle() {
    this.menuTarget.classList.toggle(this.openClass);
  }

  hide(event) {
    if (!this.element.contains(event.target)) {
      this.menuTarget.classList.remove(this.openClass);
    }
  }
}
```

### Action Cable (WebSockets)

Channel:

```ruby
# app/channels/chat_channel.rb
class ChatChannel < ApplicationCable::Channel
  def subscribed
    stream_from "chat_#{params[:room_id]}"
  end

  def unsubscribed
    stop_all_streams
  end

  def speak(data)
    message = current_user.messages.create!(
      content: data["message"],
      room_id: params[:room_id]
    )

    ActionCable.server.broadcast(
      "chat_#{params[:room_id]}",
      message: render_message(message)
    )
  end

  private

  def render_message(message)
    ApplicationController.render(
      partial: "messages/message",
      locals: { message: message }
    )
  end
end

# app/channels/application_cable/connection.rb
module ApplicationCable
  class Connection < ActionCable::Connection::Base
    identified_by :current_user

    def connect
      self.current_user = find_verified_user
    end

    private

    def find_verified_user
      if verified_user = User.find_by(id: cookies.signed[:user_id])
        verified_user
      else
        reject_unauthorized_connection
      end
    end
  end
end
```

Broadcasting:

```ruby
# From anywhere in your app
ActionCable.server.broadcast("chat_#{room.id}", {
  message: "New message",
  user: user.name
})

# From models
class Message < ApplicationRecord
  after_create_commit -> {
    broadcast_append_to "messages",
      target: "messages_list",
      partial: "messages/message",
      locals: { message: self }
  }
end
```

---

## RSpec Testing Reference

### RSpec Configuration

```ruby
# spec/spec_helper.rb
RSpec.configure do |config|
  config.expect_with :rspec do |expectations|
    expectations.include_chain_clauses_in_custom_matcher_descriptions = true
  end

  config.mock_with :rspec do |mocks|
    mocks.verify_partial_doubles = true
  end

  config.shared_context_metadata_behavior = :apply_to_host_groups
  config.filter_run_when_matching :focus
  config.example_status_persistence_file_path = "spec/examples.txt"
  config.disable_monkey_patching!
  config.warnings = true

  config.order = :random
  Kernel.srand config.seed
end
```

### Matcher Reference

Model Matchers (shoulda-matchers):

```ruby
# Associations
it { is_expected.to have_many(:posts) }
it { is_expected.to have_one(:profile) }
it { is_expected.to belong_to(:user) }
it { is_expected.to have_and_belong_to_many(:tags) }

# Validations
it { is_expected.to validate_presence_of(:name) }
it { is_expected.to validate_uniqueness_of(:email) }
it { is_expected.to validate_length_of(:name).is_at_least(2).is_at_most(100) }
it { is_expected.to validate_numericality_of(:age).only_integer }
it { is_expected.to validate_inclusion_of(:role).in_array(%w[user admin]) }

# Database
it { is_expected.to have_db_column(:name).of_type(:string) }
it { is_expected.to have_db_index(:email) }
```

Controller Matchers:

```ruby
# Responses
expect(response).to have_http_status(:success)
expect(response).to have_http_status(200)
expect(response).to redirect_to(root_path)
expect(response).to render_template(:index)

# Content
expect(response.body).to include("Welcome")
expect(response).to match_response_schema("user")
```

### Factory Bot Patterns

```ruby
# spec/factories/users.rb
FactoryBot.define do
  factory :user do
    sequence(:email) { |n| "user#{n}@example.com" }
    name { Faker::Name.name }
    password { "password123" }

    # Traits
    trait :admin do
      role { :admin }
    end

    trait :with_posts do
      transient do
        posts_count { 3 }
      end

      after(:create) do |user, evaluator|
        create_list(:post, evaluator.posts_count, user: user)
      end
    end

    # Nested factory
    factory :admin_user, traits: [:admin]
  end
end

# Usage
create(:user)
create(:user, :admin)
create(:user, :with_posts, posts_count: 5)
build(:user)
build_stubbed(:user)
attributes_for(:user)
```

---

## RuboCop Configuration

### Basic Configuration

```yaml
# .rubocop.yml
require:
  - rubocop-rails
  - rubocop-rspec

AllCops:
  NewCops: enable
  TargetRubyVersion: 3.3
  Exclude:
    - "db/schema.rb"
    - "db/migrate/**/*"
    - "node_modules/**/*"
    - "vendor/**/*"
    - "bin/**/*"

Style/StringLiterals:
  EnforcedStyle: double_quotes

Style/Documentation:
  Enabled: false

Metrics/MethodLength:
  Max: 20
  Exclude:
    - "spec/**/*"

Metrics/BlockLength:
  Exclude:
    - "spec/**/*"
    - "config/**/*"

Layout/LineLength:
  Max: 120
  Exclude:
    - "config/**/*"

Rails/HasManyOrHasOneDependent:
  Enabled: true

RSpec/ExampleLength:
  Max: 20

RSpec/MultipleExpectations:
  Max: 5
```

---

## Bundler Configuration

### Gemfile Best Practices

```ruby
# Gemfile
source "https://rubygems.org"

ruby "3.3.0"

# Core
gem "rails", "~> 7.2.0"

# Database
gem "pg", "~> 1.5"

# Server
gem "puma", ">= 6.0"

# Performance
gem "bootsnap", require: false

group :development do
  gem "web-console"
  gem "listen"
end

group :development, :test do
  gem "debug", platforms: [:mri, :mingw, :x64_mingw]
  gem "rspec-rails"
  gem "factory_bot_rails"
  gem "faker"
end

group :test do
  gem "capybara"
  gem "selenium-webdriver"
  gem "shoulda-matchers"
  gem "simplecov", require: false
end
```

### Bundle Commands

```bash
# Install dependencies
bundle install

# Update all gems
bundle update

# Update specific gem
bundle update rails

# Check for outdated gems
bundle outdated

# Show gem location
bundle show rails

# Clean unused gems
bundle clean

# Check for security vulnerabilities
bundle audit

# Create binstubs
bundle binstubs rspec-core
```

---

## Performance Optimization

### Database Query Optimization

N+1 Query Prevention:

```ruby
# Bad: N+1 queries
users = User.all
users.each do |user|
  puts user.posts.count  # Triggers query for each user
end

# Good: Use includes
users = User.includes(:posts)
users.each do |user|
  puts user.posts.count  # No additional queries
end

# Good: Use counter cache
class Post < ApplicationRecord
  belongs_to :user, counter_cache: true
end

# Migration
add_column :users, :posts_count, :integer, default: 0
User.find_each { |u| User.reset_counters(u.id, :posts) }
```

Database Indexing:

```ruby
# Migration
class AddIndexToUsers < ActiveRecord::Migration[7.2]
  def change
    add_index :users, :email, unique: true
    add_index :posts, :user_id
    add_index :posts, [:user_id, :created_at]
    add_index :comments, :post_id, where: "deleted_at IS NULL"
  end
end
```

### Caching Strategies

Fragment Caching:

```erb
<!-- app/views/posts/index.html.erb -->
<% cache @posts do %>
  <% @posts.each do |post| %>
    <% cache post do %>
      <%= render post %>
    <% end %>
  <% end %>
<% end %>
```

Russian Doll Caching:

```ruby
class Post < ApplicationRecord
  belongs_to :user, touch: true
  has_many :comments, touch: true
end

# View
<% cache [@post, @post.user, @post.comments.maximum(:updated_at)] do %>
  <%= render @post %>
<% end %>
```

Low-Level Caching:

```ruby
class ProductService
  def expensive_calculation(product_id)
    Rails.cache.fetch("product/#{product_id}/calculation", expires_in: 1.hour) do
      # Expensive operation
      perform_calculation(product_id)
    end
  end
end
```

---

## Security Best Practices

### SQL Injection Prevention

```ruby
# Bad: SQL injection vulnerable
User.where("name = '#{params[:name]}'")

# Good: Parameterized query
User.where("name = ?", params[:name])
User.where(name: params[:name])
```

### Mass Assignment Protection

```ruby
# Controller
def create
  @user = User.new(user_params)
  @user.save
end

private

def user_params
  params.require(:user).permit(:name, :email, :password)
end
```

### CSRF Protection

```ruby
# ApplicationController
class ApplicationController < ActionController::Base
  protect_from_forgery with: :exception
end

# API Controller (use token authentication)
class Api::BaseController < ActionController::API
  skip_before_action :verify_authenticity_token
  before_action :authenticate_token
end
```

### Content Security Policy

```ruby
# config/initializers/content_security_policy.rb
Rails.application.config.content_security_policy do |policy|
  policy.default_src :self, :https
  policy.font_src    :self, :https, :data
  policy.img_src     :self, :https, :data
  policy.object_src  :none
  policy.script_src  :self, :https
  policy.style_src   :self, :https
end
```

---

## Context7 Integration

### Library ID Resolution

```ruby
# Rails framework
library_id = resolve_library_id("rails")
# Returns: /rails/rails

# RSpec testing
library_id = resolve_library_id("rspec")
# Returns: /rspec/rspec

# Sidekiq background jobs
library_id = resolve_library_id("sidekiq")
# Returns: /sidekiq/sidekiq
```

### Available Libraries

| Library     | Context7 ID                   | Key Topics                          |
| ----------- | ----------------------------- | ----------------------------------- |
| Rails       | /rails/rails                  | models, controllers, routing, views |
| RSpec       | /rspec/rspec                  | testing, matchers, mocks            |
| Sidekiq     | /sidekiq/sidekiq              | jobs, queues, scheduling            |
| Devise      | /heartcombo/devise            | authentication, sessions            |
| Turbo       | /hotwired/turbo-rails         | frames, streams, drive              |
| Stimulus    | /hotwired/stimulus            | controllers, targets, actions       |
| Pundit      | /varvet/pundit                | authorization, policies             |
| FactoryBot  | /thoughtbot/factory_bot       | factories, traits, sequences        |
| Shrine      | /shrinerb/shrine              | file uploads, storage               |
| RuboCop     | /rubocop/rubocop              | linting, formatting                 |
| ActiveAdmin | /activeadmin/activeadmin      | admin interface                     |
| Kaminari    | /kaminari/kaminari            | pagination                          |
| Ransack     | /activerecord-hackery/ransack | search, filtering                   |

---

## Deployment Reference

### Production Environment Configuration

```ruby
# config/environments/production.rb
Rails.application.configure do
  config.cache_classes = true
  config.eager_load = true
  config.consider_all_requests_local = false
  config.public_file_server.enabled = ENV["RAILS_SERVE_STATIC_FILES"].present?
  config.assets.compile = false
  config.active_storage.service = :amazon
  config.log_level = :info
  config.log_tags = [:request_id]
  config.action_mailer.perform_caching = false
  config.i18n.fallbacks = true
  config.active_support.report_deprecations = false
  config.active_record.dump_schema_after_migration = false
end
```

### Database Configuration

```yaml
# config/database.yml
production:
  adapter: postgresql
  encoding: unicode
  pool: <%= ENV.fetch("RAILS_MAX_THREADS") { 5 } %>
  url: <%= ENV["DATABASE_URL"] %>
  prepared_statements: true
  advisory_locks: true
  connect_timeout: 5
  checkout_timeout: 5
  variables:
    statement_timeout: 30000
```

### Puma Configuration

```ruby
# config/puma.rb
max_threads_count = ENV.fetch("RAILS_MAX_THREADS") { 5 }
min_threads_count = ENV.fetch("RAILS_MIN_THREADS") { max_threads_count }
threads min_threads_count, max_threads_count

worker_timeout 3600 if ENV.fetch("RAILS_ENV", "development") == "development"

port ENV.fetch("PORT") { 3000 }
environment ENV.fetch("RAILS_ENV") { "development" }
pidfile ENV.fetch("PIDFILE") { "tmp/pids/server.pid" }

workers ENV.fetch("WEB_CONCURRENCY") { 2 }
preload_app!

on_worker_boot do
  ActiveRecord::Base.establish_connection
end

plugin :tmp_restart
```

---

Last Updated: 2026-01-10
Version: 1.0.0
