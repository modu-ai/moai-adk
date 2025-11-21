# Ruby Advanced Patterns

## Metaprogramming Mastery

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

### DSL Creation with instance_eval

```ruby
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

# Usage with DSL
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
```

### Refinements for Safe Monkey Patching

```ruby
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
  using StringExtensions

  def display_title
    @title.titleize
  end

  def preview
    @content.truncate(100)
  end
end

# Outside the class, these methods don't exist
"hello world".titleize  # => NoMethodError
```

---

## Concurrency & Parallelism

### Fiber Scheduling for Async I/O

```ruby
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
    loop do
      readable, writable = IO.select(@readable.keys, @writable.keys, [], 0)

      readable&.each { |io| @readable.delete(io)&.resume }
      writable&.each { |io| @writable.delete(io)&.resume }

      break if @readable.empty? && @writable.empty?
    end
  end
end
```

### Ractor for True Parallelism

```ruby
def parallel_process(items)
  ractors = items.map do |item|
    Ractor.new(item) do |i|
      heavy_computation(i)
    end
  end

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
  Config[:timeout]  # Access without copying
end
```

---

## Rack & Middleware

### Authentication Middleware

```ruby
class AuthenticationMiddleware
  def initialize(app)
    @app = app
  end

  def call(env)
    request = Rack::Request.new(env)

    if public_path?(request.path)
      return @app.call(env)
    end

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
    token.present?
  end
end
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

## Pattern Matching Advanced

### Comprehensive Pattern Examples

```ruby
# Struct pattern matching
case User.new("John", 30)
in User(name:, age:)
  "User: #{name}, #{age}"
end

# Array patterns with guards
case [1, 2, 3]
in [a, b, c] if a + b + c > 5
  "Sum is greater than 5"
end

# Rightward assignment
user => { name:, email: }
puts name

# Complex nested patterns
data = {
  user: { name: 'Alice', role: 'admin' },
  posts: [{ title: 'First' }, { title: 'Second' }]
}

case data
in { user: { name:, role: 'admin' }, posts: [first, *rest] }
  "Admin #{name} has #{rest.size + 1} posts"
end
```

---

## Performance Profiling

### Stackprof Integration

```ruby
require 'stackprof'

result = StackProf.run(mode: :cpu, interval: 1000) do
  # Code to profile
  User.includes(:posts).where(active: true).limit(100).to_a
end

File.write('tmp/stackprof.dump', Marshal.dump(result))

# View: stackprof tmp/stackprof.dump --text
# Or:   stackprof tmp/stackprof.dump --flamegraph > tmp/fg.html
```

### Query Analysis Tools

```ruby
# Prevent N+1 queries
users = User.includes(:posts)
users.each { |user| puts user.posts.size }

# Counter cache for efficiency
class Post < ApplicationRecord
  belongs_to :user, counter_cache: true
end

# Batch processing for large datasets
User.find_each(batch_size: 1000) do |user|
  user.process_data
end

# Pluck specific columns
User.where(active: true).pluck(:id, :email)
```

---

## Advanced Rails Patterns

### Active Record Associations & Scopes

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

### Hotwire Integration Pattern

```ruby
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
end
```

---

**Total Lines**: 400+ | **Focus**: Advanced Ruby patterns for enterprise applications
