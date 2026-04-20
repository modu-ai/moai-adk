# Ruby Production-Ready Code Examples

## Complete Rails 7.2 Application

### Project Structure

```
rails_app/
├── app/
│   ├── controllers/
│   │   ├── application_controller.rb
│   │   ├── api/
│   │   │   └── v1/
│   │   │       ├── base_controller.rb
│   │   │       └── users_controller.rb
│   │   └── users_controller.rb
│   ├── models/
│   │   ├── application_record.rb
│   │   ├── concerns/
│   │   │   ├── sluggable.rb
│   │   │   └── searchable.rb
│   │   ├── user.rb
│   │   └── post.rb
│   ├── services/
│   │   └── user_registration_service.rb
│   ├── repositories/
│   │   └── user_repository.rb
│   ├── jobs/
│   │   └── send_welcome_email_job.rb
│   ├── mailers/
│   │   └── user_mailer.rb
│   └── views/
│       ├── users/
│       │   ├── index.html.erb
│       │   ├── show.html.erb
│       │   └── _user.html.erb
│       └── layouts/
│           └── application.html.erb
├── config/
│   ├── application.rb
│   ├── database.yml
│   ├── routes.rb
│   └── initializers/
│       └── sidekiq.rb
├── db/
│   └── migrate/
├── spec/
│   ├── spec_helper.rb
│   ├── rails_helper.rb
│   ├── factories/
│   │   └── users.rb
│   ├── models/
│   │   └── user_spec.rb
│   ├── requests/
│   │   └── users_spec.rb
│   └── system/
│       └── users_spec.rb
├── Gemfile
└── Dockerfile
```

### Main Application Configuration

```ruby
# config/application.rb
require_relative "boot"

require "rails/all"

Bundler.require(*Rails.groups)

module RailsApp
  class Application < Rails::Application
    config.load_defaults 7.2

    # API-only mode (optional)
    # config.api_only = true

    # Custom configuration
    config.time_zone = "UTC"
    config.active_record.default_timezone = :utc
    config.i18n.default_locale = :en

    # Background job configuration
    config.active_job.queue_adapter = :sidekiq

    # CORS configuration (for API)
    config.middleware.insert_before 0, Rack::Cors do
      allow do
        origins "localhost:3000", "example.com"
        resource "*",
          headers: :any,
          methods: [:get, :post, :put, :patch, :delete, :options, :head]
      end
    end
  end
end
```

### Gemfile Configuration

```ruby
# Gemfile
source "https://rubygems.org"

ruby "3.3.0"

# Core Rails
gem "rails", "~> 7.2.0"
gem "pg", "~> 1.5"
gem "puma", ">= 6.0"

# Frontend
gem "turbo-rails"
gem "stimulus-rails"
gem "importmap-rails"

# Background Jobs
gem "sidekiq", "~> 7.0"
gem "redis", "~> 5.0"

# Authentication & Authorization
gem "devise", "~> 4.9"
gem "pundit", "~> 2.3"

# API
gem "jbuilder"
gem "active_model_serializers", "~> 0.10.0"

# File Upload
gem "shrine", "~> 3.5"

# Performance
gem "bootsnap", require: false

group :development, :test do
  gem "debug"
  gem "factory_bot_rails"
  gem "faker"
  gem "rspec-rails", "~> 7.0"
  gem "rubocop-rails", require: false
  gem "rubocop-rspec", require: false
end

group :test do
  gem "capybara"
  gem "selenium-webdriver"
  gem "shoulda-matchers", "~> 6.0"
  gem "simplecov", require: false
  gem "database_cleaner-active_record"
end

group :development do
  gem "web-console"
  gem "listen"
end
```

### ActiveRecord Models

```ruby
# app/models/application_record.rb
class ApplicationRecord < ActiveRecord::Base
  primary_abstract_class
end

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
    return if slug.present?
    return unless respond_to?(:title)

    base_slug = title.parameterize
    candidate = base_slug
    counter = 1

    while self.class.exists?(slug: candidate)
      candidate = "#{base_slug}-#{counter}"
      counter += 1
    end

    self.slug = candidate
  end
end

# app/models/user.rb
class User < ApplicationRecord
  # Include Devise modules
  devise :database_authenticatable, :registerable,
         :recoverable, :rememberable, :validatable,
         :confirmable, :lockable, :timeoutable, :trackable

  # Associations
  has_many :posts, dependent: :destroy
  has_one :profile, dependent: :destroy
  has_many :comments, dependent: :destroy

  # Validations
  validates :email, presence: true, uniqueness: true, format: { with: URI::MailTo::EMAIL_REGEXP }
  validates :name, presence: true, length: { minimum: 2, maximum: 100 }
  validates :username, uniqueness: true, allow_nil: true,
                       format: { with: /\A[a-zA-Z0-9_]+\z/, message: "only allows letters, numbers, and underscores" }

  # Scopes
  scope :active, -> { where(active: true) }
  scope :inactive, -> { where(active: false) }
  scope :recent, -> { order(created_at: :desc) }
  scope :by_name, ->(name) { where("name ILIKE ?", "%#{sanitize_sql_like(name)}%") }
  scope :verified, -> { where.not(confirmed_at: nil) }

  # Enums
  enum :role, { user: 0, moderator: 1, admin: 2 }, prefix: true

  # Callbacks
  before_create :set_default_role
  after_create :send_welcome_email

  # Class methods
  def self.search(query)
    by_name(query).or(where("email ILIKE ?", "%#{sanitize_sql_like(query)}%"))
  end

  # Instance methods
  def full_name
    "#{first_name} #{last_name}".strip.presence || name
  end

  def display_name
    username || full_name
  end

  def admin?
    role_admin?
  end

  private

  def set_default_role
    self.role ||= :user
  end

  def send_welcome_email
    UserMailer.welcome(self).deliver_later
  end
end

# app/models/post.rb
class Post < ApplicationRecord
  include Sluggable

  belongs_to :user
  has_many :comments, dependent: :destroy
  has_many_attached :images

  validates :title, presence: true, length: { minimum: 5, maximum: 200 }
  validates :content, presence: true, length: { minimum: 10 }

  scope :published, -> { where(published: true) }
  scope :draft, -> { where(published: false) }
  scope :recent, -> { order(published_at: :desc) }

  # Virtual attributes
  def reading_time
    words_per_minute = 200
    word_count = content.split.size
    (word_count / words_per_minute.to_f).ceil
  end

  # State machine pattern
  def publish!
    update!(published: true, published_at: Time.current)
  end

  def unpublish!
    update!(published: false, published_at: nil)
  end
end
```

### Service Objects

```ruby
# app/services/user_registration_service.rb
class UserRegistrationService
  Result = Data.define(:success, :user, :errors) do
    def success? = success
    def failure? = !success
  end

  def initialize(user_params)
    @user_params = user_params
  end

  def call
    user = User.new(@user_params)

    ActiveRecord::Base.transaction do
      user.save!
      create_default_profile(user)
      send_welcome_email(user)
      track_registration(user)
    end

    Result.new(success: true, user: user, errors: nil)
  rescue ActiveRecord::RecordInvalid => e
    Result.new(success: false, user: nil, errors: e.record.errors)
  rescue StandardError => e
    Rails.logger.error("Registration failed: #{e.message}")
    Result.new(success: false, user: nil, errors: ["Registration failed. Please try again."])
  end

  private

  def create_default_profile(user)
    user.create_profile!(
      bio: "New user",
      avatar_url: default_avatar_url
    )
  end

  def send_welcome_email(user)
    UserMailer.welcome(user).deliver_later
  end

  def track_registration(user)
    # Analytics tracking
    AnalyticsService.track(
      event: "user_registered",
      user_id: user.id,
      properties: {
        email: user.email,
        created_at: user.created_at
      }
    )
  end

  def default_avatar_url
    "https://example.com/default-avatar.png"
  end
end
```

### Repository Pattern

```ruby
# app/repositories/user_repository.rb
class UserRepository
  def find(id)
    User.find_by(id: id)
  end

  def find_by_email(email)
    User.find_by(email: email)
  end

  def all(page: 1, per_page: 20, filters: {})
    scope = User.all
    scope = apply_filters(scope, filters)
    scope.page(page).per(per_page)
  end

  def create(attributes)
    User.create(attributes)
  end

  def update(user, attributes)
    user.update(attributes)
    user
  end

  def delete(user)
    user.destroy
  end

  def search(query, page: 1, per_page: 20)
    User.search(query).page(page).per(per_page)
  end

  private

  def apply_filters(scope, filters)
    scope = scope.active if filters[:active]
    scope = scope.verified if filters[:verified]
    scope = scope.where(role: filters[:role]) if filters[:role]
    scope
  end
end
```

### Rails Controllers

```ruby
# app/controllers/application_controller.rb
class ApplicationController < ActionController::Base
  include Pundit::Authorization

  before_action :configure_permitted_parameters, if: :devise_controller?

  rescue_from Pundit::NotAuthorizedError, with: :user_not_authorized
  rescue_from ActiveRecord::RecordNotFound, with: :record_not_found

  private

  def configure_permitted_parameters
    devise_parameter_sanitizer.permit(:sign_up, keys: [:name, :username])
    devise_parameter_sanitizer.permit(:account_update, keys: [:name, :username])
  end

  def user_not_authorized
    flash[:alert] = "You are not authorized to perform this action."
    redirect_back(fallback_location: root_path)
  end

  def record_not_found
    render file: "#{Rails.root}/public/404.html", status: :not_found, layout: false
  end
end

# app/controllers/users_controller.rb
class UsersController < ApplicationController
  before_action :authenticate_user!
  before_action :set_user, only: %i[show edit update destroy]
  before_action :authorize_user, only: %i[edit update destroy]

  # GET /users
  def index
    @users = User.active.recent.page(params[:page])
  end

  # GET /users/:id
  def show
    @posts = @user.posts.published.recent.limit(10)
  end

  # GET /users/new
  def new
    @user = User.new
  end

  # POST /users
  def create
    result = UserRegistrationService.new(user_params).call

    respond_to do |format|
      if result.success?
        format.html { redirect_to result.user, notice: "User was successfully created." }
        format.turbo_stream { flash.now[:notice] = "User was successfully created." }
        format.json { render json: result.user, status: :created }
      else
        @user = User.new(user_params)
        @user.errors.merge!(result.errors)
        format.html { render :new, status: :unprocessable_entity }
        format.json { render json: @user.errors, status: :unprocessable_entity }
      end
    end
  end

  # PATCH/PUT /users/:id
  def update
    respond_to do |format|
      if @user.update(user_params)
        format.html { redirect_to @user, notice: "User was successfully updated." }
        format.turbo_stream
        format.json { render json: @user }
      else
        format.html { render :edit, status: :unprocessable_entity }
        format.json { render json: @user.errors, status: :unprocessable_entity }
      end
    end
  end

  # DELETE /users/:id
  def destroy
    @user.destroy!

    respond_to do |format|
      format.html { redirect_to users_path, notice: "User was successfully deleted." }
      format.turbo_stream
      format.json { head :no_content }
    end
  end

  private

  def set_user
    @user = User.find(params[:id])
  end

  def authorize_user
    authorize @user
  end

  def user_params
    params.require(:user).permit(:name, :email, :username, :password, :password_confirmation)
  end
end
```

### API Controllers

```ruby
# app/controllers/api/v1/base_controller.rb
module Api
  module V1
    class BaseController < ActionController::API
      include Pundit::Authorization

      before_action :authenticate_user!

      rescue_from Pundit::NotAuthorizedError, with: :user_not_authorized
      rescue_from ActiveRecord::RecordNotFound, with: :record_not_found
      rescue_from ActiveRecord::RecordInvalid, with: :record_invalid

      private

      def authenticate_user!
        token = request.headers["Authorization"]&.split(" ")&.last
        return render_unauthorized unless token

        begin
          decoded = JWT.decode(token, Rails.application.secret_key_base, true, { algorithm: "HS256" })
          @current_user = User.find(decoded[0]["user_id"])
        rescue JWT::DecodeError, ActiveRecord::RecordNotFound
          render_unauthorized
        end
      end

      attr_reader :current_user

      def render_unauthorized
        render json: { error: "Unauthorized" }, status: :unauthorized
      end

      def user_not_authorized
        render json: { error: "Forbidden" }, status: :forbidden
      end

      def record_not_found
        render json: { error: "Not found" }, status: :not_found
      end

      def record_invalid(exception)
        render json: { errors: exception.record.errors.full_messages }, status: :unprocessable_entity
      end
    end
  end
end

# app/controllers/api/v1/users_controller.rb
module Api
  module V1
    class UsersController < BaseController
      skip_before_action :authenticate_user!, only: [:create]

      # GET /api/v1/users
      def index
        @users = User.active.page(params[:page]).per(params[:per_page] || 20)
        render json: @users, each_serializer: UserSerializer
      end

      # GET /api/v1/users/:id
      def show
        @user = User.find(params[:id])
        render json: @user, serializer: UserDetailSerializer
      end

      # POST /api/v1/users
      def create
        result = UserRegistrationService.new(user_params).call

        if result.success?
          render json: result.user, serializer: UserSerializer, status: :created
        else
          render json: { errors: result.errors }, status: :unprocessable_entity
        end
      end

      # PATCH /api/v1/users/:id
      def update
        @user = User.find(params[:id])
        authorize @user

        if @user.update(user_params)
          render json: @user, serializer: UserSerializer
        else
          render json: { errors: @user.errors.full_messages }, status: :unprocessable_entity
        end
      end

      private

      def user_params
        params.require(:user).permit(:name, :email, :username, :password, :password_confirmation)
      end
    end
  end
end
```

### Hotwire/Turbo Patterns

```erb
<!-- app/views/users/index.html.erb -->
<div class="users-container">
  <h1>Users</h1>

  <%= turbo_frame_tag "users" do %>
    <div class="users-list">
      <%= render @users %>
    </div>

    <%= turbo_frame_tag "pagination" do %>
      <%= paginate @users %>
    <% end %>
  <% end %>
</div>

<!-- app/views/users/_user.html.erb -->
<%= turbo_frame_tag dom_id(user) do %>
  <article class="user-card">
    <h2><%= link_to user.name, user %></h2>
    <p class="email"><%= user.email %></p>
    <p class="role"><%= user.role %></p>

    <div class="actions">
      <%= link_to "Edit", edit_user_path(user), data: { turbo_frame: "_top" } %>
      <%= button_to "Delete", user, method: :delete,
                    form: { data: { turbo_confirm: "Are you sure?" } },
                    class: "btn-danger" %>
    </div>
  </article>
<% end %>
```

```erb
<!-- app/views/users/create.turbo_stream.erb -->
<%= turbo_stream.prepend "users", @user %>
<%= turbo_stream.update "user_form", partial: "users/form", locals: { user: User.new } %>
<%= turbo_stream.update "flash", partial: "shared/flash" %>
```

```javascript
// app/javascript/controllers/form_controller.js
import { Controller } from "@hotwired/stimulus";

export default class extends Controller {
  static targets = ["input", "submit", "errors"];

  connect() {
    this.validate();
  }

  validate() {
    const allValid = this.inputTargets.every((input) => {
      return input.value.trim().length > 0;
    });

    this.submitTarget.disabled = !allValid;
  }

  clearErrors() {
    if (this.hasErrorsTarget) {
      this.errorsTarget.innerHTML = "";
    }
  }
}
```

### Background Jobs with Sidekiq

```ruby
# app/jobs/send_welcome_email_job.rb
class SendWelcomeEmailJob < ApplicationJob
  queue_as :default

  retry_on StandardError, wait: :exponentially_longer, attempts: 5
  discard_on ActiveJob::DeserializationError

  def perform(user_id)
    user = User.find(user_id)
    UserMailer.welcome(user).deliver_now
  rescue ActiveRecord::RecordNotFound => e
    Rails.logger.error("User not found: #{user_id}")
    # Don't retry for non-existent users
  end
end

# app/jobs/data_export_job.rb
class DataExportJob < ApplicationJob
  queue_as :low_priority

  def perform(user_id, format: "csv")
    user = User.find(user_id)
    exporter = DataExporter.new(user, format: format)

    file_path = exporter.export
    UserMailer.export_ready(user, file_path).deliver_now
  ensure
    # Cleanup temporary files
    File.delete(file_path) if file_path && File.exist?(file_path)
  end
end

# config/initializers/sidekiq.rb
Sidekiq.configure_server do |config|
  config.redis = { url: ENV.fetch("REDIS_URL", "redis://localhost:6379/0") }
end

Sidekiq.configure_client do |config|
  config.redis = { url: ENV.fetch("REDIS_URL", "redis://localhost:6379/0") }
end
```

---

## Complete RSpec Test Suite

```ruby
# spec/rails_helper.rb
require "spec_helper"
require "rspec/rails"
require "capybara/rspec"
require "shoulda/matchers"

begin
  ActiveRecord::Migration.maintain_test_schema!
rescue ActiveRecord::PendingMigrationError => e
  abort e.to_s.strip
end

RSpec.configure do |config|
  config.fixture_path = "#{Rails.root}/spec/fixtures"
  config.use_transactional_fixtures = true
  config.infer_spec_type_from_file_location!
  config.filter_rails_from_backtrace!

  # FactoryBot
  config.include FactoryBot::Syntax::Methods

  # DatabaseCleaner
  config.before(:suite) do
    DatabaseCleaner.strategy = :transaction
    DatabaseCleaner.clean_with(:truncation)
  end

  config.around(:each) do |example|
    DatabaseCleaner.cleaning do
      example.run
    end
  end
end

Shoulda::Matchers.configure do |config|
  config.integrate do |with|
    with.test_framework :rspec
    with.library :rails
  end
end
```

```ruby
# spec/factories/users.rb
FactoryBot.define do
  factory :user do
    sequence(:email) { |n| "user#{n}@example.com" }
    sequence(:username) { |n| "user#{n}" }
    name { Faker::Name.name }
    password { "password123" }
    confirmed_at { Time.current }

    trait :unconfirmed do
      confirmed_at { nil }
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
  end
end
```

```ruby
# spec/models/user_spec.rb
RSpec.describe User, type: :model do
  describe "associations" do
    it { is_expected.to have_many(:posts).dependent(:destroy) }
    it { is_expected.to have_one(:profile).dependent(:destroy) }
    it { is_expected.to have_many(:comments).dependent(:destroy) }
  end

  describe "validations" do
    subject { build(:user) }

    it { is_expected.to validate_presence_of(:email) }
    it { is_expected.to validate_uniqueness_of(:email).case_insensitive }
    it { is_expected.to validate_presence_of(:name) }
    it { is_expected.to validate_length_of(:name).is_at_least(2).is_at_most(100) }
  end

  describe "scopes" do
    let!(:active_user) { create(:user, active: true) }
    let!(:inactive_user) { create(:user, active: false) }

    describe ".active" do
      it "returns only active users" do
        expect(User.active).to include(active_user)
        expect(User.active).not_to include(inactive_user)
      end
    end

    describe ".recent" do
      it "orders users by created_at desc" do
        expect(User.recent.first).to eq(inactive_user)
      end
    end
  end

  describe "callbacks" do
    describe "after_create" do
      it "sends welcome email" do
        expect {
          create(:user)
        }.to have_enqueued_job(ActionMailer::MailDeliveryJob)
      end
    end
  end

  describe "#full_name" do
    context "when first_name and last_name are present" do
      let(:user) { build(:user, first_name: "John", last_name: "Doe") }

      it "returns the full name" do
        expect(user.full_name).to eq("John Doe")
      end
    end

    context "when first_name and last_name are blank" do
      let(:user) { build(:user, first_name: nil, last_name: nil, name: "John Doe") }

      it "returns the name" do
        expect(user.full_name).to eq("John Doe")
      end
    end
  end

  describe "#admin?" do
    context "when user is admin" do
      let(:user) { build(:user, :admin) }

      it { expect(user.admin?).to be true }
    end

    context "when user is not admin" do
      let(:user) { build(:user) }

      it { expect(user.admin?).to be false }
    end
  end
end
```

```ruby
# spec/requests/users_spec.rb
RSpec.describe "Users", type: :request do
  let(:user) { create(:user) }
  let(:valid_attributes) { attributes_for(:user) }
  let(:invalid_attributes) { { email: "" } }

  describe "GET /users" do
    it "returns success response" do
      get users_path
      expect(response).to have_http_status(:success)
    end

    it "renders index template" do
      get users_path
      expect(response).to render_template(:index)
    end
  end

  describe "GET /users/:id" do
    it "returns success response" do
      get user_path(user)
      expect(response).to have_http_status(:success)
    end

    context "when user does not exist" do
      it "returns 404" do
        expect {
          get user_path(id: "invalid")
        }.to raise_error(ActiveRecord::RecordNotFound)
      end
    end
  end

  describe "POST /users" do
    context "with valid parameters" do
      it "creates a new user" do
        expect {
          post users_path, params: { user: valid_attributes }
        }.to change(User, :count).by(1)
      end

      it "redirects to the created user" do
        post users_path, params: { user: valid_attributes }
        expect(response).to redirect_to(User.last)
      end
    end

    context "with invalid parameters" do
      it "does not create a new user" do
        expect {
          post users_path, params: { user: invalid_attributes }
        }.not_to change(User, :count)
      end

      it "returns unprocessable entity status" do
        post users_path, params: { user: invalid_attributes }
        expect(response).to have_http_status(:unprocessable_entity)
      end
    end
  end
end
```

```ruby
# spec/system/users_spec.rb
RSpec.describe "Users", type: :system do
  before do
    driven_by(:selenium_chrome_headless)
  end

  describe "User registration" do
    it "allows user to sign up" do
      visit new_user_registration_path

      fill_in "Name", with: "John Doe"
      fill_in "Email", with: "john@example.com"
      fill_in "Password", with: "password123"
      fill_in "Password confirmation", with: "password123"

      click_button "Sign up"

      expect(page).to have_content("Welcome! You have signed up successfully.")
    end
  end

  describe "User profile" do
    let(:user) { create(:user, :with_posts) }

    before { sign_in user }

    it "displays user posts" do
      visit user_path(user)

      expect(page).to have_content(user.name)
      user.posts.each do |post|
        expect(page).to have_content(post.title)
      end
    end
  end
end
```

---

## Ruby 3.3 Pattern Matching Examples

```ruby
# Pattern matching with case/in
def process_api_response(response)
  case response
  in { status: 200, data: data }
    puts "Success: #{data}"
  in { status: 201, data: data, location: location }
    puts "Created at: #{location}"
  in { status: (400..499), error: error }
    puts "Client error: #{error}"
  in { status: (500..599), error: error }
    puts "Server error: #{error}"
  else
    puts "Unknown response: #{response}"
  end
end

# Array pattern matching
def process_coordinates(coords)
  case coords
  in [x, y]
    Point2D.new(x, y)
  in [x, y, z]
    Point3D.new(x, y, z)
  in []
    raise "Empty coordinates"
  else
    raise "Invalid coordinates"
  end
end

# Hash pattern matching with guards
def validate_user_data(user_data)
  case user_data
  in { age: age } if age < 0
    [:error, "Invalid age"]
  in { age: age } if age < 18
    [:minor, "User is a minor"]
  in { age: age, verified: true } if age >= 18
    [:adult, "Verified adult user"]
  in { age: age } if age >= 18
    [:adult, "Unverified adult user"]
  else
    [:error, "Invalid user data"]
  end
end
```

---

## Docker Production Dockerfile

```dockerfile
# Dockerfile
FROM ruby:3.3.0-slim AS builder

WORKDIR /app

# Install dependencies
RUN apt-get update -qq && \
    apt-get install -y --no-install-recommends \
      build-essential \
      libpq-dev \
      nodejs \
      npm && \
    rm -rf /var/lib/apt/lists/*

# Install gems
COPY Gemfile Gemfile.lock ./
RUN bundle config set --local deployment 'true' && \
    bundle config set --local without 'development test' && \
    bundle install -j4 --retry 3

# Production stage
FROM ruby:3.3.0-slim AS runtime

WORKDIR /app

# Install runtime dependencies
RUN apt-get update -qq && \
    apt-get install -y --no-install-recommends \
      libpq5 \
      nodejs && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -r rails && useradd -r -g rails rails

# Copy gems from builder
COPY --from=builder /usr/local/bundle /usr/local/bundle

# Copy application
COPY --chown=rails:rails . .

# Precompile assets
RUN RAILS_ENV=production SECRET_KEY_BASE=dummy bundle exec rails assets:precompile

# Switch to non-root user
USER rails

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:3000/health || exit 1

# Start server
CMD ["bundle", "exec", "puma", "-C", "config/puma.rb"]
```

---

Last Updated: 2026-01-10
Version: 1.0.0
