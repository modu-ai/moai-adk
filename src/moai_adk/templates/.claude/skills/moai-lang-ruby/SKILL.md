---
name: moai-lang-ruby
description: Ruby 3.3+ development specialist covering Rails 7.2, ActiveRecord, Hotwire/Turbo, and modern Ruby patterns. Use when developing Ruby APIs, web applications, or Rails projects.
version: 1.0.0
updated: 2025-12-07
status: active
allowed-tools: Read, Grep, Glob, Bash, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
---

## Quick Reference (30 seconds)

Ruby 3.3+ Development Specialist - Rails 7.2, ActiveRecord, Hotwire/Turbo, RSpec, and modern Ruby patterns.

Auto-Triggers: `.rb` files, `Gemfile`, `Rakefile`, `config.ru`, Rails/Ruby discussions

Core Capabilities:
- Ruby 3.3 Features: YJIT production-ready, pattern matching, Data class, endless methods
- Web Framework: Rails 7.2 with Turbo, Stimulus, ActiveRecord
- Frontend: Hotwire (Turbo + Stimulus) for SPA-like experiences
- Testing: RSpec with factories, request specs, system specs
- Background Jobs: Sidekiq with ActiveJob
- Package Management: Bundler with Gemfile
- Code Quality: RuboCop with Rails cops
- Database: ActiveRecord with migrations, associations, scopes

### Quick Patterns

Rails Controller:
```ruby
class UsersController < ApplicationController
  before_action :set_user, only: %i[show edit update destroy]

  def index
    @users = User.all
  end

  def create
    @user = User.new(user_params)

    respond_to do |format|
      if @user.save
        format.html { redirect_to @user, notice: "User was successfully created." }
        format.turbo_stream
      else
        format.html { render :new, status: :unprocessable_entity }
      end
    end
  end

  private

  def set_user
    @user = User.find(params[:id])
  end

  def user_params
    params.require(:user).permit(:name, :email)
  end
end
```

ActiveRecord Model:
```ruby
class User < ApplicationRecord
  has_many :posts, dependent: :destroy
  has_one :profile, dependent: :destroy

  validates :email, presence: true, uniqueness: true, format: { with: URI::MailTo::EMAIL_REGEXP }
  validates :name, presence: true, length: { minimum: 2, maximum: 100 }

  scope :active, -> { where(active: true) }
  scope :recent, -> { order(created_at: :desc) }

  def full_name
    "#{first_name} #{last_name}".strip
  end
end
```

RSpec Test:
```ruby
RSpec.describe User, type: :model do
  describe "validations" do
    it { is_expected.to validate_presence_of(:email) }
    it { is_expected.to validate_uniqueness_of(:email) }
  end

  describe "#full_name" do
    let(:user) { build(:user, first_name: "John", last_name: "Doe") }

    it "returns the full name" do
      expect(user.full_name).to eq("John Doe")
    end
  end
end
```

---

## Implementation Guide (5 minutes)

### Ruby 3.3 New Features

YJIT (Production-Ready):
- Enabled by default in Ruby 3.3
- 15-20% performance improvement for Rails apps
- Enable: `ruby --yjit` or `RUBY_YJIT_ENABLE=1`
- Check status: `RubyVM::YJIT.enabled?`
- Memory optimization with better code caching

Pattern Matching (case/in):
```ruby
def process_response(response)
  case response
  in { status: "ok", data: data }
    puts "Success: #{data}"
  in { status: "error", message: msg }
    puts "Error: #{msg}"
  in { status: status } if %w[pending processing].include?(status)
    puts "In progress..."
  else
    puts "Unknown response"
  end
end

# Array pattern matching
case [1, 2, 3]
in [first, *rest]
  puts "First: #{first}, Rest: #{rest}"
end

# Hash pattern matching with guard
case { name: "John", age: 25 }
in { name:, age: } if age >= 18
  puts "Adult: #{name}"
end
```

Data Class (Immutable Structs):
```ruby
# Define immutable data class
User = Data.define(:name, :email) do
  def greeting
    "Hello, #{name}!"
  end
end

user = User.new(name: "John", email: "john@example.com")
user.name         # => "John"
user.greeting     # => "Hello, John!"
# user.name = "Jane"  # => FrozenError

# With default values
Point = Data.define(:x, :y) do
  def self.origin
    new(x: 0, y: 0)
  end
end
```

Endless Method Definition:
```ruby
class Calculator
  # Single expression methods
  def add(a, b) = a + b
  def multiply(a, b) = a * b

  # With method chaining
  def double(n) = n * 2
  def square(n) = n ** 2

  # Predicate methods
  def positive?(n) = n > 0
  def even?(n) = n.even?
end
```

### Rails 7.2 Patterns

Application Setup:
```ruby
# Gemfile
source "https://rubygems.org"

gem "rails", "~> 7.2.0"
gem "pg", "~> 1.5"
gem "puma", ">= 6.0"
gem "turbo-rails"
gem "stimulus-rails"
gem "jbuilder"
gem "redis", ">= 5.0"
gem "sidekiq", "~> 7.0"

group :development, :test do
  gem "rspec-rails", "~> 7.0"
  gem "factory_bot_rails"
  gem "faker"
  gem "rubocop-rails", require: false
end

group :test do
  gem "capybara"
  gem "shoulda-matchers"
end
```

Model with Concerns:
```ruby
# app/models/concerns/sluggable.rb
module Sluggable
  extend ActiveSupport::Concern

  included do
    before_validation :generate_slug, on: :create
    validates :slug, presence: true, uniqueness: true
  end

  def to_param
    slug
  end

  private

  def generate_slug
    self.slug = title.parameterize if title.present? && slug.blank?
  end
end

# app/models/post.rb
class Post < ApplicationRecord
  include Sluggable

  belongs_to :user
  has_many :comments, dependent: :destroy
  has_many_attached :images

  validates :title, presence: true, length: { minimum: 5 }
  validates :content, presence: true

  scope :published, -> { where(published: true) }
  scope :by_user, ->(user) { where(user: user) }
end
```

Service Objects:
```ruby
# app/services/user_registration_service.rb
class UserRegistrationService
  def initialize(user_params)
    @user_params = user_params
  end

  def call
    user = User.new(@user_params)

    ActiveRecord::Base.transaction do
      user.save!
      create_profile(user)
      send_welcome_email(user)
    end

    Result.new(success: true, user: user)
  rescue ActiveRecord::RecordInvalid => e
    Result.new(success: false, errors: e.record.errors)
  end

  private

  def create_profile(user)
    user.create_profile!(bio: "New user")
  end

  def send_welcome_email(user)
    UserMailer.welcome(user).deliver_later
  end

  Result = Data.define(:success, :user, :errors) do
    def success? = success
    def failure? = !success
  end
end
```

### Hotwire (Turbo + Stimulus)

Turbo Frames:
```erb
<!-- app/views/posts/index.html.erb -->
<%= turbo_frame_tag "posts" do %>
  <% @posts.each do |post| %>
    <%= render post %>
  <% end %>
<% end %>

<!-- app/views/posts/_post.html.erb -->
<%= turbo_frame_tag dom_id(post) do %>
  <article class="post">
    <h2><%= link_to post.title, post %></h2>
    <p><%= truncate(post.content, length: 200) %></p>
    <%= link_to "Edit", edit_post_path(post) %>
  </article>
<% end %>
```

Turbo Streams:
```ruby
# app/controllers/posts_controller.rb
def create
  @post = current_user.posts.build(post_params)

  respond_to do |format|
    if @post.save
      format.turbo_stream
      format.html { redirect_to @post }
    else
      format.html { render :new, status: :unprocessable_entity }
    end
  end
end

# app/views/posts/create.turbo_stream.erb
<%= turbo_stream.prepend "posts", @post %>
<%= turbo_stream.update "new_post", partial: "posts/form", locals: { post: Post.new } %>
```

Stimulus Controller:
```javascript
// app/javascript/controllers/form_controller.js
import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
  static targets = ["input", "submit", "output"]
  static values = { url: String }

  connect() {
    this.validate()
  }

  validate() {
    const isValid = this.inputTargets.every(input => input.value.length > 0)
    this.submitTarget.disabled = !isValid
  }

  async submit(event) {
    event.preventDefault()

    const response = await fetch(this.urlValue, {
      method: "POST",
      body: new FormData(this.element),
      headers: { "Accept": "text/vnd.turbo-stream.html" }
    })

    if (response.ok) {
      this.element.reset()
    }
  }
}
```

### RSpec Testing Patterns

Model Specs:
```ruby
# spec/models/user_spec.rb
RSpec.describe User, type: :model do
  describe "associations" do
    it { is_expected.to have_many(:posts).dependent(:destroy) }
    it { is_expected.to have_one(:profile).dependent(:destroy) }
  end

  describe "validations" do
    subject { build(:user) }

    it { is_expected.to validate_presence_of(:email) }
    it { is_expected.to validate_uniqueness_of(:email).case_insensitive }
    it { is_expected.to validate_length_of(:name).is_at_least(2).is_at_most(100) }
  end

  describe "scopes" do
    describe ".active" do
      let!(:active_user) { create(:user, active: true) }
      let!(:inactive_user) { create(:user, active: false) }

      it "returns only active users" do
        expect(described_class.active).to contain_exactly(active_user)
      end
    end
  end

  describe "#full_name" do
    context "when both names are present" do
      let(:user) { build(:user, first_name: "John", last_name: "Doe") }

      it "returns the full name" do
        expect(user.full_name).to eq("John Doe")
      end
    end

    context "when last name is missing" do
      let(:user) { build(:user, first_name: "John", last_name: nil) }

      it "returns only first name" do
        expect(user.full_name).to eq("John")
      end
    end
  end
end
```

Request Specs:
```ruby
# spec/requests/posts_spec.rb
RSpec.describe "Posts", type: :request do
  let(:user) { create(:user) }

  before { sign_in user }

  describe "GET /posts" do
    let!(:posts) { create_list(:post, 3, user: user) }

    it "returns a successful response" do
      get posts_path
      expect(response).to have_http_status(:ok)
    end

    it "displays all posts" do
      get posts_path
      posts.each do |post|
        expect(response.body).to include(post.title)
      end
    end
  end

  describe "POST /posts" do
    let(:valid_params) { { post: attributes_for(:post) } }
    let(:invalid_params) { { post: { title: "" } } }

    context "with valid parameters" do
      it "creates a new post" do
        expect {
          post posts_path, params: valid_params
        }.to change(Post, :count).by(1)
      end

      it "redirects to the created post" do
        post posts_path, params: valid_params
        expect(response).to redirect_to(Post.last)
      end
    end

    context "with invalid parameters" do
      it "does not create a new post" do
        expect {
          post posts_path, params: invalid_params
        }.not_to change(Post, :count)
      end

      it "returns unprocessable entity status" do
        post posts_path, params: invalid_params
        expect(response).to have_http_status(:unprocessable_entity)
      end
    end
  end
end
```

Factory Bot Patterns:
```ruby
# spec/factories/users.rb
FactoryBot.define do
  factory :user do
    sequence(:email) { |n| "user#{n}@example.com" }
    name { Faker::Name.name }
    password { "password123" }
    active { true }

    trait :inactive do
      active { false }
    end

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

    factory :admin_user, traits: [:admin]
  end
end
```

### Sidekiq Background Jobs

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

### ActiveRecord Advanced Patterns

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

---

## Advanced Implementation (10+ minutes)

For comprehensive coverage including:
- Production deployment patterns (Docker, Kubernetes)
- Advanced ActiveRecord patterns (polymorphic, STI)
- Action Cable real-time features
- Performance optimization techniques
- Security best practices
- CI/CD integration patterns

See:
- [reference.md](reference.md) - Complete reference documentation
- [examples.md](examples.md) - Production-ready code examples

---

## Context7 Library Mappings

```
/rails/rails - Ruby on Rails web framework
/rspec/rspec - RSpec testing framework
/hotwired/turbo-rails - Turbo for Rails
/hotwired/stimulus-rails - Stimulus for Rails
/sidekiq/sidekiq - Background job processing
/rubocop/rubocop - Ruby style guide enforcement
/thoughtbot/factory_bot - Test data factories
```

---

## Works Well With

- `moai-domain-backend` - REST API and web application architecture
- `moai-domain-database` - SQL patterns and ActiveRecord optimization
- `moai-quality-testing` - TDD and testing strategies
- `moai-essentials-debug` - AI-powered debugging
- `moai-foundation-trust` - TRUST 5 quality principles

---

## Troubleshooting

Common Issues:

Ruby Version Check:
```bash
ruby --version  # Should be 3.3+
ruby -e "puts RubyVM::YJIT.enabled?"  # Check YJIT status
```

Rails Version Check:
```bash
rails --version  # Should be 7.2+
bundle exec rails about  # Full environment info
```

Database Connection Issues:
- Check `config/database.yml` configuration
- Ensure PostgreSQL/MySQL service is running
- Run `rails db:create` if database doesn't exist

Asset Pipeline Issues:
```bash
# Precompile assets
rails assets:precompile

# Clear asset cache
rails assets:clobber
```

RSpec Setup Issues:
```bash
# Install RSpec
rails generate rspec:install

# Run specific test
bundle exec rspec spec/models/user_spec.rb

# Run with verbose output
bundle exec rspec --format documentation
```

Turbo/Stimulus Issues:
```bash
# Rebuild JavaScript
rails javascript:install:esbuild

# Clear Turbo cache
rails turbo:install
```

---

Last Updated: 2025-12-07
Status: Active (v1.0.0)
