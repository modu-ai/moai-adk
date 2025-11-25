# Ruby 3.4 - Practical Working Examples

_Last updated: 2025-11-22 | Enterprise Production Ready_

---

## Example 1: Rails API with async/await

```ruby
# config/routes.rb
Rails.application.routes.draw do
  namespace :api do
    resources :users, only: [:create, :show, :index]
    resources :posts, only: [:index, :create]
  end
end

# app/controllers/api/users_controller.rb
module Api
  class UsersController < ApplicationController
    def index
      users = User.includes(:posts)
                   .where(active: true)
                   .order(created_at: :desc)
                   .limit(20)

      render json: {
        data: users.map { |u| user_serializer(u) },
        meta: { total: User.count, page: 1 }
      }
    end

    def create
      user = User.new(user_params)

      if user.save
        render json: { data: user_serializer(user) }, status: :created
      else
        render json: { errors: user.errors }, status: :unprocessable_entity
      end
    end

    private

    def user_serializer(user)
      {
        id: user.id,
        name: user.full_name,
        email: user.email,
        posts_count: user.posts.count
      }
    end

    def user_params
      params.require(:user).permit(:first_name, :last_name, :email, :password)
    end
  end
end
```

---

## Example 2: RSpec Model Testing Suite

```ruby
# spec/models/user_spec.rb
RSpec.describe User, type: :model do
  # Factories
  let(:user) { build(:user) }
  let!(:admin) { create(:user, :admin) }

  describe 'associations' do
    it { is_expected.to have_many(:posts).dependent(:destroy) }
    it { is_expected.to have_one(:profile) }
    it { is_expected.to belong_to(:organization).optional }
  end

  describe 'validations' do
    it { is_expected.to validate_presence_of(:email) }
    it { is_expected.to validate_uniqueness_of(:email) }
    it { is_expected.to validate_length_of(:password).is_at_least(8) }
    it { is_expected.to have_secure_password }
  end

  describe 'scopes' do
    let!(:active_user) { create(:user, active: true) }
    let!(:inactive_user) { create(:user, active: false) }

    describe '.active' do
      it 'returns only active users' do
        expect(User.active).to contain_exactly(active_user, admin)
      end
    end

    describe '.recent' do
      it 'returns users created in last 7 days' do
        old_user = create(:user, created_at: 10.days.ago)
        new_user = create(:user, created_at: 1.day.ago)

        expect(User.recent).not_to include(old_user)
        expect(User.recent).to include(new_user)
      end
    end
  end

  describe '#full_name' do
    context 'when both names present' do
      let(:user) { build(:user, first_name: 'John', last_name: 'Doe') }

      it 'returns concatenated name' do
        expect(user.full_name).to eq('John Doe')
      end
    end

    context 'when only first name present' do
      let(:user) { build(:user, first_name: 'John', last_name: nil) }

      it 'returns first name' do
        expect(user.full_name).to eq('John')
      end
    end
  end

  describe '#authenticate' do
    it 'authenticates with valid password' do
      user = create(:user, password: 'correct_password')
      expect(user.authenticate('correct_password')).to eq(user)
    end

    it 'rejects invalid password' do
      user = create(:user, password: 'correct_password')
      expect(user.authenticate('wrong_password')).to be_falsey
    end
  end
end
```

---

## Example 3: Request Spec with Authentication

```ruby
# spec/requests/api/users_spec.rb
RSpec.describe 'Users API', type: :request do
  let(:token) { JWT.encode({ user_id: user.id }, Rails.application.secrets.secret_key_base) }
  let(:user) { create(:user) }
  let(:headers) { { 'Authorization' => "Bearer #{token}", 'Content-Type' => 'application/json' } }

  describe 'GET /api/users' do
    let!(:users) { create_list(:user, 3, active: true) }
    let!(:inactive) { create(:user, active: false) }

    it 'returns active users' do
      get '/api/users', headers: headers

      expect(response).to have_http_status(:ok)
      data = JSON.parse(response.body)['data']
      expect(data.length).to eq(3)
      expect(data.map { |u| u['id'] }).not_to include(inactive.id)
    end

    it 'includes pagination metadata' do
      get '/api/users', headers: headers

      meta = JSON.parse(response.body)['meta']
      expect(meta).to include('total', 'page', 'per_page')
    end
  end

  describe 'POST /api/users' do
    let(:valid_params) do
      {
        user: {
          first_name: 'Jane',
          last_name: 'Smith',
          email: 'jane@example.com',
          password: 'SecurePass123'
        }
      }
    end

    context 'with valid parameters' do
      it 'creates user and returns 201' do
        expect {
          post '/api/users', params: valid_params, headers: headers
        }.to change(User, :count).by(1)

        expect(response).to have_http_status(:created)
        data = JSON.parse(response.body)['data']
        expect(data['email']).to eq('jane@example.com')
      end
    end

    context 'with invalid parameters' do
      let(:invalid_params) { { user: { email: '', password: 'short' } } }

      it 'does not create user' do
        expect {
          post '/api/users', params: invalid_params, headers: headers
        }.not_to change(User, :count)

        expect(response).to have_http_status(:unprocessable_entity)
      end
    end
  end
end
```

---

## Example 4: Async HTTP with Fibers

```ruby
# lib/async_service.rb
require 'async'
require 'async/http/internet'

class AsyncService
  def self.fetch_user_data(user_id)
    Async do
      internet = Async::HTTP::Internet.new

      # Concurrent requests
      user_task = Async { internet.get("https://api.example.com/users/#{user_id}") }
      posts_task = Async { internet.get("https://api.example.com/users/#{user_id}/posts") }
      comments_task = Async { internet.get("https://api.example.com/users/#{user_id}/comments") }

      user_response = user_task.wait
      posts_response = posts_task.wait
      comments_response = comments_task.wait

      {
        user: JSON.parse(user_response.read),
        posts: JSON.parse(posts_response.read),
        comments: JSON.parse(comments_response.read)
      }
    end
  end
end

# Usage in controller
class UsersController < ApplicationController
  def profile
    @data = AsyncService.fetch_user_data(params[:id])
    render json: @data
  end
end
```

---

## Example 5: Pattern Matching in Controllers

```ruby
# app/controllers/orders_controller.rb
class OrdersController < ApplicationController
  def update
    @order = Order.find(params[:id])
    result = process_order_update(@order, order_params)

    # Pattern matching on result
    case result
    in { success: true, order: order }
      render json: { data: order }, status: :ok
    in { success: false, errors: errors, code: 'validation' }
      render json: { errors: errors }, status: :unprocessable_entity
    in { success: false, errors: errors, code: 'not_found' }
      render json: { errors: errors }, status: :not_found
    in { success: false, errors: errors }
      render json: { errors: errors }, status: :internal_server_error
    end
  end

  private

  def process_order_update(order, params)
    case order.status
    in 'pending'
      if order.update(params)
        { success: true, order: order }
      else
        { success: false, errors: order.errors, code: 'validation' }
      end
    in 'processing' | 'shipped'
      { success: false, errors: ['Cannot update in current status'], code: 'invalid_state' }
    in status
      { success: false, errors: ["Unknown status: #{status}"], code: 'unknown' }
    end
  end

  def order_params
    params.require(:order).permit(:status, :notes)
  end
end
```

---

## Example 6: Bundler Gemfile with Groups

```ruby
# Gemfile
source "https://rubygems.org"
git_source(:github) { |repo| "https://github.com/#{repo}.git" }

ruby "3.4.0"

# Web framework
gem "rails", "~> 8.0"
gem "puma", "~> 6.0"

# Database
gem "pg", "~> 1.5"
gem "redis", "~> 5.0"
gem "redis-rails"

# API & serialization
gem "active_model_serializers", "~> 0.10"
gem "jsonapi-serializer", "~> 2.2"

# Auth
gem "bcrypt", "~> 3.1.20"
gem "jwt", "~> 2.8"

# Async & concurrency
gem "async", "~> 2.10"
gem "async-http", "~> 0.60"

# Validation
gem "validates_email_format_of", "~> 1.7"

# Performance
gem "rack-mini-profiler", "~> 3.0"
gem "stackprof", "~> 0.2"

group :development do
  gem "rubocop", require: false
  gem "rubocop-rails", require: false
  gem "brakeman", require: false
end

group :development, :test do
  gem "debug", "~> 1.9"
  gem "factory_bot_rails", "~> 6.2"
  gem "rspec-rails", "~> 6.0"
end

group :test do
  gem "rspec-expectations", "~> 3.12"
  gem "webmock", "~> 3.19"
  gem "vcr", "~> 6.1"
  gem "simplecov", "~> 0.22", require: false
end

group :production do
  gem "rails_12factor"
  gem "sentry-rails"
end
```

---

## Example 7: Custom Rack Middleware

```ruby
# lib/middleware/request_id_tracker.rb
class RequestIdTracker
  def initialize(app)
    @app = app
  end

  def call(env)
    request_id = env['HTTP_X_REQUEST_ID'] || SecureRandom.uuid
    env['request_id'] = request_id

    status, headers, body = @app.call(env)

    [status, headers.merge('X-Request-ID' => request_id), body]
  end
end

# config/application.rb
config.middleware.insert_before 0, RequestIdTracker

# Usage in logs
class ApplicationController
  before_action :set_request_id

  private

  def set_request_id
    request.env['request_id'] = params[:request_id] if params[:request_id]
  end
end
```

---

## Example 8: RuboCop Configuration

```ruby
# .rubocop.yml
require:
  - rubocop-rails
  - rubocop-rspec

AllCops:
  TargetRubyVersion: 3.4
  NewCops: enable
  Exclude:
    - 'bin/**/*'
    - 'config/**/*'
    - 'vendor/**/*'

Style/StringLiterals:
  EnforcedStyle: single_quotes

Style/FrozenStringLiteralComment:
  Enabled: true

Rails/OutputSafety:
  Enabled: true

RSpec/ExampleLength:
  Max: 20

RSpec/MultipleExpectations:
  Max: 3

Metrics/LineLength:
  Max: 120
  IgnoredPatterns:
    - '^\s*#'

Metrics/MethodLength:
  Max: 20
```

---

## Example 9: Hotwire Turbo Stream Integration

```ruby
# app/controllers/todos_controller.rb
class TodosController < ApplicationController
  def create
    @todo = Todo.new(todo_params)

    respond_to do |format|
      if @todo.save
        format.turbo_stream do
          render turbo_stream: [
            turbo_stream.append('todos', @todo),
            turbo_stream.update('todo_form', partial: 'form', locals: { todo: Todo.new })
          ]
        end
      else
        format.turbo_stream do
          render turbo_stream: turbo_stream.replace(
            'todo_form',
            partial: 'form',
            locals: { todo: @todo }
          ), status: :unprocessable_entity
        end
      end
    end
  end

  def destroy
    @todo = Todo.find(params[:id])
    @todo.destroy

    respond_to do |format|
      format.turbo_stream do
        render turbo_stream: turbo_stream.remove(@todo)
      end
    end
  end

  private

  def todo_params
    params.require(:todo).permit(:title, :description, :completed)
  end
end
```

---

## Example 10: Performance Profiling with Stackprof

```ruby
# lib/tasks/profile.rake
namespace :profile do
  desc 'Profile user index with Stackprof'
  task :users => :environment do
    require 'stackprof'

    result = StackProf.run(mode: :cpu, interval: 1000) do
      100.times do
        User.includes(:posts)
            .where(active: true)
            .order(created_at: :desc)
            .limit(50)
            .to_a
      end
    end

    File.write('tmp/stackprof_users.dump', Marshal.dump(result))
    puts "Profile saved to tmp/stackprof_users.dump"
    puts "View with: stackprof tmp/stackprof_users.dump --text"
  end

  desc 'Generate flamegraph'
  task :flamegraph => :environment do
    puts "run: stackprof tmp/stackprof_users.dump --flamegraph > tmp/flamegraph.html"
  end
end
```

---

## Example 11: Pattern Matching with Fibers

```ruby
# Processing concurrent async operations with pattern matching
require 'async'

class DataProcessor
  def self.process_requests(requests)
    Async do
      tasks = requests.map do |req|
        Async { fetch_and_process(req) }
      end

      results = tasks.map(&:wait)

      # Pattern match on results
      results.each do |result|
        case result
        in { status: :success, data: }
          puts "Processed: #{data}"
        in { status: :error, message: }
          puts "Error: #{message}"
        in { status: :timeout }
          puts "Request timed out"
        end
      end
    end
  end

  private

  def self.fetch_and_process(request)
    response = Net::HTTP.get_response(URI(request[:url]))
    { status: :success, data: response.body }
  rescue => e
    { status: :error, message: e.message }
  end
end
```

---

## Example 12: Ractor-based Parallel Processing

```ruby
# lib/parallel_processor.rb
class ParallelProcessor
  def self.process_items_in_parallel(items, workers: 4)
    # Split items among workers
    batches = items.each_slice((items.size.to_f / workers).ceil).to_a

    # Create ractors for parallel processing
    ractors = batches.map do |batch|
      Ractor.new(batch) do |items|
        items.map { |item| expensive_operation(item) }
      end
    end

    # Collect results
    ractors.flat_map(&:take)
  end

  private

  def self.expensive_operation(item)
    # Simulate expensive computation
    sleep 0.1
    item * 2
  end
end

# Usage
items = (1..1000).to_a
results = ParallelProcessor.process_items_in_parallel(items, workers: 8)
```

---

**Total Examples**: 12 | **Focus**: Practical production-ready Ruby patterns for enterprise applications
