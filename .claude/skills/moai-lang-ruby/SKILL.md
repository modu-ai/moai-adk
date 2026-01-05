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
```

Data Class (Immutable Structs):
```ruby
User = Data.define(:name, :email) do
  def greeting
    "Hello, #{name}!"
  end
end

user = User.new(name: "John", email: "john@example.com")
user.name         # => "John"
user.greeting     # => "Hello, John!"
```

Endless Method Definition:
```ruby
class Calculator
  def add(a, b) = a + b
  def multiply(a, b) = a * b
  def positive?(n) = n > 0
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
end
```

Service Objects:
```ruby
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
  static targets = ["input", "submit"]

  connect() {
    this.validate()
  }

  validate() {
    const isValid = this.inputTargets.every(input => input.value.length > 0)
    this.submitTarget.disabled = !isValid
  }
}
```

### RSpec Testing Basics

Factory Bot Patterns:
```ruby
# spec/factories/users.rb
FactoryBot.define do
  factory :user do
    sequence(:email) { |n| "user#{n}@example.com" }
    name { Faker::Name.name }
    password { "password123" }

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
  end
end
```

---

## Advanced Implementation (10+ minutes)

For comprehensive coverage including:
- Production deployment patterns (Docker, Kubernetes)
- Advanced ActiveRecord patterns (polymorphic, STI, query objects)
- Action Cable real-time features
- Performance optimization techniques
- Security best practices
- CI/CD integration patterns
- Complete RSpec testing patterns

See:
- [Advanced Patterns](modules/advanced-patterns.md) - Production patterns and advanced features
- [Testing Patterns](modules/testing-patterns.md) - Complete RSpec testing guide

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
- `moai-workflow-testing` - TDD and testing strategies
- `moai-essentials-debug` - AI-powered debugging
- `moai-foundation-quality` - TRUST 5 quality principles

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
rails assets:precompile
rails assets:clobber
```

RSpec Setup Issues:
```bash
rails generate rspec:install
bundle exec rspec spec/models/user_spec.rb
bundle exec rspec --format documentation
```

Turbo/Stimulus Issues:
```bash
rails javascript:install:esbuild
rails turbo:install
```

---

Last Updated: 2025-12-07
Status: Active (v1.0.0)
