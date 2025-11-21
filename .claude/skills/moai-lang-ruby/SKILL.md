---
name: moai-lang-ruby
description: Ruby 3.4+ best practices with RSpec 4, RuboCop 2, Bundler, and Rails
  8 patterns.
---

## Quick Reference (30 seconds)

# Ruby Web Development — Enterprise

## References (Latest Documentation)

_Documentation links updated 2025-11-22_

---

---

## Implementation Guide

## What It Does

Ruby 3.4+ best practices with RSpec 4, RuboCop 2, Bundler, and Rails 8 patterns.

**Key capabilities**:
- ✅ Best practices enforcement for language domain
- ✅ TRUST 5 principles integration
- ✅ Latest tool versions (2025-11-22)
- ✅ TDD workflow support
- ✅ Modern Ruby features (Pattern Matching, Fibers, YJIT)
- ✅ Rails 8 advanced patterns (Hotwire, Turbo, Stimulus)
- ✅ Performance optimization with Ruby 3.4 YJIT

---

## When to Use

**Automatic triggers**:
- Related code discussions and file patterns
- SPEC implementation (`/alfred:2-run`)
- Code review requests

**Manual invocation**:
- Review code for TRUST 5 compliance
- Design new features
- Troubleshoot issues

---

## Tool Version Matrix (2025-11-22)

| Tool | Version | Purpose | Status |
|------|---------|---------|--------|
| **Ruby** | 3.4.0 | Runtime | ✅ Current |
| **RSpec** | 4.0.0 | Testing | ✅ Current |
| **RuboCop** | 2.0.0 | Linting | ✅ Current |
| **Rails** | 8.0.0 | Web framework | ✅ Current |
| **Bundler** | 2.6.0 | Dependency manager | ✅ Current |
| **Hotwire** | 1.6.0 | Front-end | ✅ Current |

---

## Ruby 3.4+ Modern Features

### Pattern Matching (Enhanced in Ruby 3.2+)

```ruby
# Basic pattern matching
case user
in { name:, age: 18.. }
  "Adult: #{name}"
in { name:, age: }
  "Minor: #{name}, age #{age}"
else
  "Unknown user"
end

# Array pattern matching
case response
in [200, { data: }]
  puts "Success: #{data}"
in [404, *]
  puts "Not found"
in [500, { error: message }]
  puts "Error: #{message}"
end

# Pin operator for matching specific values
expected_status = 200
case response
in [^expected_status, body]
  process_success(body)
in [status, _]
  handle_error(status)
end

# Find pattern (Ruby 3.0+)
case [1, 2, 3, 4, 5]
in [*, 3, *rest]
  rest  # => [4, 5]
end

# Rightward assignment
user => { name:, email: }
puts name  # Extracts name and email
```

### Fiber Scheduling for Concurrency

```ruby
# Ruby 3.0+ Fiber Scheduler for async I/O
require 'async'

Async do
  # Concurrent HTTP requests
  responses = Async::HTTP::Internet.new.tap do |internet|
    tasks = [
      Async { internet.get('https://api.example.com/users') },
      Async { internet.get('https://api.example.com/posts') },
      Async { internet.get('https://api.example.com/comments') }
    ]
    
    Async::Barrier.new(tasks).wait
  end
end

# Custom Fiber Scheduler
class SimpleScheduler
  def initialize
    @readable = {}
    @writable = {}
  end
  
  def io_wait(io, events, timeout)
    fiber = Fiber.current
    
    if events & IO::READABLE
      @readable[io] = fiber
    end
    
    if events & IO::WRITABLE
      @writable[io] = fiber
    end
    
    Fiber.yield
  end
  
  def run
    # Event loop implementation
    loop do
      readable, writable = IO.select(@readable.keys, @writable.keys, [], 0)
      
      readable&.each { |io| @readable.delete(io)&.resume }
      writable&.each { |io| @writable.delete(io)&.resume }
      
      break if @readable.empty? && @writable.empty?
    end
  end
end
```

### Endless Method Definition

```ruby
# Ruby 3.0+ endless method syntax
def square(x) = x ** 2

def full_name(first, last) = "#{first} #{last}"

# Works with pattern matching
def greet(user) = case user
  in { name:, age: 18.. } then "Hello adult #{name}"
  in { name: } then "Hello #{name}"
  else "Hello stranger"
end
```

### Ractor for True Parallelism

```ruby
# Ruby 3.0+ Ractors for parallel execution
def parallel_process(items)
  ractors = items.map do |item|
    Ractor.new(item) do |i|
      # Process in isolated memory space
      heavy_computation(i)
    end
  end
  
  # Collect results
  ractors.map(&:take)
end

# Inter-Ractor communication
pipe = Ractor.new do
  loop do
    value = Ractor.receive
    puts "Received: #{value}"
  end
end

10.times { |i| pipe.send(i) }

# Shareable objects
Config = Ractor.make_shareable({ timeout: 30, retries: 3 })

ractor = Ractor.new do
  Config[:timeout]  # Can access without copying
end
```

---

## Metaprogramming Power

### Dynamic Method Definition

```ruby
class DynamicModel
  ATTRIBUTES = %i[name email age]
  
  # Define getter methods
  ATTRIBUTES.each do |attr|
    define_method(attr) do
      instance_variable_get("@#{attr}")
    end
  end
  
  # Define setter methods
  ATTRIBUTES.each do |attr|
    define_method("#{attr}=") do |value|
      instance_variable_set("@#{attr}", value)
    end
  end
  
  # Query methods
  ATTRIBUTES.each do |attr|
    define_method("#{attr}?") do
      !instance_variable_get("@#{attr}").nil?
    end
  end
end

user = DynamicModel.new
user.name = "Alice"
user.name?  # => true
```

### DSL Creation

```ruby
# Internal DSL for configuration
class ApiClient
  attr_reader :config
  
  def initialize(&block)
    @config = Configuration.new
    @config.instance_eval(&block) if block
  end
  
  class Configuration
    attr_accessor :base_url, :timeout, :headers
    
    def initialize
      @headers = {}
      @timeout = 30
    end
    
    def header(key, value)
      @headers[key] = value
    end
    
    def authentication(type, token)
      case type
      when :bearer
        header('Authorization', "Bearer #{token}")
      when :basic
        header('Authorization', "Basic #{token}")
      end
    end
  end
end

# Usage
client = ApiClient.new do
  base_url 'https://api.example.com'
  timeout 60
  authentication :bearer, ENV['API_TOKEN']
  header 'Accept', 'application/json'
end
```

### Method Missing Magic

```ruby
class SmartProxy
  def initialize(target)
    @target = target
  end
  
  def method_missing(method, *args, &block)
    if method.to_s.start_with?('find_by_')
      attribute = method.to_s.sub('find_by_', '')
      @target.find { |item| item[attribute.to_sym] == args.first }
    elsif @target.respond_to?(method)
      @target.public_send(method, *args, &block)
    else
      super
    end
  end
  
  def respond_to_missing?(method, include_private = false)
    method.to_s.start_with?('find_by_') || @target.respond_to?(method) || super
  end
end

users = [
  { name: 'Alice', age: 30 },
  { name: 'Bob', age: 25 }
]

proxy = SmartProxy.new(users)
proxy.find_by_name('Alice')  # => { name: 'Alice', age: 30 }
proxy.find_by_age(25)        # => { name: 'Bob', age: 25 }
```

### Refinements for Scoped Monkey Patching

```ruby
# Safe monkey patching with refinements
module StringExtensions
  refine String do
    def titleize
      split.map(&:capitalize).join(' ')
    end
    
    def truncate(length)
      self[0...length] + (length < size ? '...' : '')
    end
  end
end

class Article
  using StringExtensions  # Only affects this class
  
  def display_title
    @title.titleize
  end
  
  def preview
    @content.truncate(100)
  end
end

# Outside the class, String doesn't have these methods
"hello world".titleize  # => NoMethodError
```

---

## Rails 8.0 Modern Patterns

### Hotwire & Turbo Streams

```ruby
# Controller with Turbo Stream responses
class TasksController < ApplicationController
  def create
    @task = Task.new(task_params)
    
    respond_to do |format|
      if @task.save
        format.turbo_stream do
          render turbo_stream: turbo_stream.append(
            'tasks',
            partial: 'tasks/task',
            locals: { task: @task }
          )
        end
        format.html { redirect_to tasks_path }
      else
        format.turbo_stream do
          render turbo_stream: turbo_stream.replace(
            'task_form',
            partial: 'tasks/form',
            locals: { task: @task }
          )
        end
      end
    end
  end
  
  def destroy
    @task = Task.find(params[:id])
    @task.destroy
    
    respond_to do |format|
      format.turbo_stream { head :ok }
      format.html { redirect_to tasks_path }
    end
  end
end
```

### Stimulus Controller Integration

```ruby
# Stimulus controller in Rails view
<div data-controller="dropdown">
  <button data-action="click->dropdown#toggle">
    Menu
  </button>
  
  <div data-dropdown-target="menu" class="hidden">
    <a href="/profile">Profile</a>
    <a href="/settings">Settings</a>
  </div>
</div>
```

### Importmap for Asset Pipeline

```ruby
# config/importmap.rb
pin "application", preload: true
pin "@hotwired/turbo-rails", to: "turbo.min.js", preload: true
pin "@hotwired/stimulus", to: "stimulus.min.js", preload: true
pin "@hotwired/stimulus-loading", to: "stimulus-loading.js", preload: true

pin_all_from "app/javascript/controllers", under: "controllers"

# Use CDN for external libraries
pin "chart.js", to: "https://ga.jspm.io/npm:chart.js@4.4.0/dist/chart.js"
```

### Active Record Advanced Patterns

```ruby
class User < ApplicationRecord
  # Delegations
  delegate :street, :city, :zipcode, to: :address, prefix: true, allow_nil: true
  
  # Scopes
  scope :active, -> { where(active: true) }
  scope :created_after, ->(date) { where('created_at > ?', date) }
  scope :with_profile, -> { joins(:profile).includes(:profile) }
  
  # Virtual attributes
  def full_name
    "#{first_name} #{last_name}"
  end
  
  def full_name=(name)
    split = name.split(' ', 2)
    self.first_name = split.first
    self.last_name = split.last
  end
  
  # Callbacks with strict ordering
  before_validation :normalize_email
  before_save :encrypt_password, if: :password_changed?
  after_commit :send_welcome_email, on: :create
  
  # Custom validators
  validates :email, presence: true, uniqueness: true
  validate :email_format
  
  private
  
  def normalize_email
    self.email = email.downcase.strip if email.present?
  end
  
  def email_format
    unless email =~ /\A[\w+\-.]+@[a-z\d\-]+(\.[a-z\d\-]+)*\.[a-z]+\z/i
      errors.add(:email, 'is invalid')
    end
  end
end
```

---

## Rack & Middleware Ecosystem

### Custom Rack Middleware

```ruby
# Authentication middleware
class AuthenticationMiddleware
  def initialize(app)
    @app = app
  end
  
  def call(env)
    request = Rack::Request.new(env)
    
    # Skip auth for public paths
    if public_path?(request.path)
      return @app.call(env)
    end
    
    # Extract token
    token = extract_token(request)
    
    if token && valid_token?(token)
      user = User.find_by(auth_token: token)
      env['current_user'] = user
      @app.call(env)
    else
      [401, { 'Content-Type' => 'application/json' }, 
       [{ error: 'Unauthorized' }.to_json]]
    end
  end
  
  private
  
  def public_path?(path)
    ['/login', '/register', '/health'].include?(path)
  end
  
  def extract_token(request)
    if (header = request.get_header('HTTP_AUTHORIZATION'))
      header.split(' ').last
    end
  end
  
  def valid_token?(token)
    # Token validation logic
    token.present?
  end
end

# Register middleware
Rails.application.config.middleware.use AuthenticationMiddleware
```

### Request/Response Logging

```ruby
class RequestLogger
  def initialize(app)
    @app = app
  end
  
  def call(env)
    start_time = Time.now
    
    status, headers, body = @app.call(env)
    
    duration = ((Time.now - start_time) * 1000).round(2)
    
    log_request(env, status, duration)
    
    [status, headers, body]
  end
  
  private
  
  def log_request(env, status, duration)
    request = Rack::Request.new(env)
    
    Rails.logger.info({
      method: request.request_method,
      path: request.path,
      status: status,
      duration_ms: duration,
      ip: request.ip,
      user_agent: request.user_agent
    }.to_json)
  end
end
```

---

## Performance Optimization

### Ruby 3.4 YJIT Configuration

```ruby
# Enable YJIT (Just-In-Time compiler)
# Start Rails with YJIT
RUBY_YJIT_ENABLE=1 rails server

# Or enable programmatically
RubyVM::YJIT.enable

# Check YJIT status
if RubyVM::YJIT.enabled?
  puts "YJIT is running"
  puts RubyVM::YJIT.stats  # Performance statistics
end

# YJIT configuration options
ENV['RUBY_YJIT_EXEC_MEM_SIZE'] = '128'  # MB of memory for JIT code
```

### Profiling with Stackprof

```ruby
# Gemfile
gem 'stackprof'

# Profiling in development
require 'stackprof'

result = StackProf.run(mode: :cpu, interval: 1000) do
  # Code to profile
  User.includes(:posts).where(active: true).limit(100).to_a
end

# Save report
File.write('tmp/stackprof.dump', Marshal.dump(result))

# View report
# stackprof tmp/stackprof.dump --text

# Or use flamegraph
# stackprof tmp/stackprof.dump --flamegraph > tmp/flamegraph.html
```

### Query Optimization

```ruby
# Bad: N+1 query problem
users = User.all
users.each do |user|
  puts user.posts.count  # N queries
end

# Good: Eager loading
users = User.includes(:posts)
users.each do |user|
  puts user.posts.size  # No additional queries
end

# Counter cache to avoid counts
class Post < ApplicationRecord
  belongs_to :user, counter_cache: true
end

# Add posts_count column to users table
# rails g migration AddPostsCountToUsers posts_count:integer

# Now efficient:
user.posts_count  # From cached column

# Batch processing for large datasets
User.find_each(batch_size: 1000) do |user|
  user.process_data
end

# Pluck for specific columns
User.where(active: true).pluck(:id, :email)
# vs
User.where(active: true).map { |u| [u.id, u.email] }
```

---

## RSpec 4 Testing Patterns

### Modern RSpec Syntax

```ruby
# spec/models/user_spec.rb
RSpec.describe User, type: :model do
  # Factory setup
  let(:user) { build(:user) }
  let!(:admin) { create(:user, :admin) }
  
  describe 'validations' do
    it { is_expected.to validate_presence_of(:email) }
    it { is_expected.to validate_uniqueness_of(:email) }
    it { is_expected.to have_secure_password }
  end
  
  describe 'associations' do
    it { is_expected.to have_many(:posts).dependent(:destroy) }
    it { is_expected.to belong_to(:organization).optional }
  end
  
  describe '#full_name' do
    context 'when first and last name are present' do
      let(:user) { build(:user, first_name: 'John', last_name: 'Doe') }
      
      it 'returns the full name' do
        expect(user.full_name).to eq('John Doe')
      end
    end
    
    context 'when only first name is present' do
      let(:user) { build(:user, first_name: 'John', last_name: nil) }
      
      it 'returns just the first name' do
        expect(user.full_name).to eq('John')
      end
    end
  end
  
  describe '.active' do
    let!(:active_user) { create(:user, active: true) }
    let!(:inactive_user) { create(:user, active: false) }
    
    it 'returns only active users' do
      expect(User.active).to contain_exactly(active_user, admin)
    end
  end
end
```

### Request Specs with JSON API

```ruby
# spec/requests/api/users_spec.rb
RSpec.describe 'Users API', type: :request do
  let(:headers) { { 'Authorization' => "Bearer #{token}" } }
  let(:token) { generate_auth_token(user) }
  let(:user) { create(:user) }
  
  describe 'GET /api/users' do
    let!(:users) { create_list(:user, 3) }
    
    before { get '/api/users', headers: headers }
    
    it 'returns all users' do
      expect(response).to have_http_status(:ok)
      expect(json_response['data'].size).to eq(4)  # 3 + authenticated user
    end
    
    it 'returns proper JSON structure' do
      expect(json_response).to match(
        data: a_collection_containing_exactly(
          *users.map do |u|
            a_hash_including(
              id: u.id,
              type: 'user',
              attributes: a_hash_including(
                email: u.email,
                name: u.full_name
              )
            )
          end
        ),
        meta: a_hash_including(
          total: 4,
          page: 1
        )
      )
    end
  end
  
  describe 'POST /api/users' do
    let(:valid_params) do
      {
        user: {
          email: 'new@example.com',
          password: 'SecurePass123',
          first_name: 'Jane',
          last_name: 'Doe'
        }
      }
    end
    
    context 'with valid parameters' do
      it 'creates a new user' do
        expect {
          post '/api/users', params: valid_params, headers: headers
        }.to change(User, :count).by(1)
      end
      
      it 'returns created status' do
        post '/api/users', params: valid_params, headers: headers
        expect(response).to have_http_status(:created)
      end
    end
    
    context 'with invalid parameters' do
      let(:invalid_params) do
        { user: { email: '', password: '123' } }
      end
      
      it 'does not create a user' do
        expect {
          post '/api/users', params: invalid_params, headers: headers
        }.not_to change(User, :count)
      end
      
      it 'returns errors' do
        post '/api/users', params: invalid_params, headers: headers
        expect(response).to have_http_status(:unprocessable_entity)
        expect(json_response['errors']).to be_present
      end
    end
  end
end
```

### Shared Examples & Contexts

```ruby
# spec/support/shared_examples/authenticatable.rb
RSpec.shared_examples 'authenticatable' do |factory_name|
  let(:resource) { create(factory_name) }
  
  describe 'authentication' do
    it 'authenticates with valid credentials' do
      expect(resource.authenticate('password')).to eq(resource)
    end
    
    it 'fails with invalid password' do
      expect(resource.authenticate('wrong')).to be_falsey
    end
  end
end

# Usage in spec
RSpec.describe User do
  it_behaves_like 'authenticatable', :user
end

RSpec.describe Admin do
  it_behaves_like 'authenticatable', :admin
end
```

---

## Inputs

- Language-specific source directories
- Configuration files
- Test suites and sample data

## Outputs

- Test/lint execution plan
- TRUST 5 review checkpoints
- Migration guidance

## Failure Modes

- When required tools are not installed
- When dependencies are missing
- When test coverage falls below 85%

## Dependencies

- Access to project files via Read/Bash tools
- Integration with `moai-foundation-langs` for language detection
- Integration with `moai-foundation-trust` for quality gates

---

## Changelog

- **v3.0.0** (2025-11-22): Massive expansion with Ruby 3.4 features, Rails 8 patterns, metaprogramming, performance optimization
- **v2.0.0** (2025-10-22): Major update with latest tool versions, comprehensive best practices, TRUST 5 integration
- **v1.0.0** (2025-03-29): Initial Skill release

---

## Works Well With

- `moai-foundation-trust` (quality gates)
- `moai-alfred-code-reviewer` (code review)
- `moai-essentials-debug` (debugging support)
- `moai-domain-backend` (backend architecture)

---

## Best Practices

✅ **DO**:
- Use pattern matching for complex conditionals
- Leverage Ractor for CPU-intensive parallel tasks
- Apply Hotwire/Turbo for modern Rails UX
- Enable YJIT for 15-30% performance boost
- Write idiomatic Ruby (blocks, symbols, duck typing)
- Use RSpec for behavior-driven development
- Implement custom Rack middleware for cross-cutting concerns
- Profile with Stackprof before optimizing

❌ **DON'T**:
- Monkey patch core classes without refinements
- Ignore N+1 query problems
- Use `eval` or `instance_eval` carelessly
- Skip eager loading for associations
- Write overly clever metaprogramming
- Use deprecated Rails features (update_attributes, before_filter)
- Ignore RuboCop warnings
- Optimize without profiling data

---

## Advanced Patterns




---

## Context7 Integration

### Related Libraries & Tools
- [Rails](/rails/rails): Full-stack web framework
- [RSpec](/rspec/rspec-core): Testing framework
- [RuboCop](/rubocop/rubocop): Code analyzer
- [Hotwire](/hotwired/hotwire): Modern front-end
- [Bundler](/rubygems/bundler): Dependency manager

### Official Documentation
- [Documentation](https://www.ruby-lang.org/en/documentation/)
- [API Reference](https://ruby-doc.org/)
- [Rails Guides](https://guides.rubyonrails.org/)
- [RSpec Documentation](https://rspec.info/documentation/)

### Version-Specific Guides
Latest stable version: Ruby 3.4
- [Ruby 3.4 Release](https://www.ruby-lang.org/en/news/)
- [Pattern Matching](https://docs.ruby-lang.org/en/master/syntax/pattern_matching_rdoc.html)
- [YJIT Optimization](https://shopify.engineering/yjit-just-in-time-compiler-cruby)
