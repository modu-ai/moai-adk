---
name: moai-lang-ruby
description: Ruby 3.4+ best practices with RSpec 4, RuboCop 2, Bundler, and Rails 8 patterns
version: 1.0.0
modularized: true
last_updated: 2025-11-22
compliance_score: 70
auto_trigger_keywords:
  - lang
  - ruby
category_tier: 1
---

## Quick Reference (30 seconds)

# Ruby Web Development — Enterprise

**Primary Focus**: Ruby 3.4+ with Rails 8, modern concurrency (Fibers, Ractors), and performance optimization
**Best For**: Web applications, APIs, real-time features, background jobs
**Key Tools**: Rails 8, RSpec 4, RuboCop 2, Bundler 2.6, Hotwire 1.6
**Auto-triggers**: Ruby, Rails, RSpec, RuboCop, Bundler, Gemfile

| Version | Release | Support |
|---------|---------|---------|
| Ruby 3.4.0 | Dec 2024 | Dec 2028 |
| Rails 8.0.0 | Nov 2024 | Nov 2026 |
| RSpec 4.0.0 | 2024 | Active |
| Bundler 2.6 | 2025 | Active |

---

## Quick Start

### Three-Level Learning Path

**Level 1: Fundamentals** (Read examples.md)
- Ruby 3.4 core features (pattern matching, endless methods)
- Rails 8 basics (models, controllers, associations)
- RSpec testing essentials
- Examples: See `examples.md` for 10+ practical patterns

**Level 2: Advanced Patterns** (See reference.md)
- Metaprogramming (define_method, method_missing, refinements)
- Concurrency (Fibers, Ractors, async patterns)
- Performance optimization (YJIT, query optimization)
- API details: See `reference.md`

**Level 3: Production Deployment** (Consult backend/performance skills)
- Docker & Kubernetes deployment
- Database optimization and monitoring
- Performance profiling (Stackprof)
- Scaling strategies

---

## What It Does

Ruby 3.4+ with RSpec, Rails, and modern concurrency features.

**Key capabilities**:
- ✅ Pattern matching and endless methods
- ✅ Fiber scheduling for async I/O
- ✅ Ractor parallelism for CPU-bound work
- ✅ Rails 8 Hotwire/Turbo integration
- ✅ Metaprogramming with refinements
- ✅ Performance optimization with YJIT

---

## When to Use

**Automatic triggers**:
- Rails/Ruby project discussions
- Code review and TRUST compliance
- TDD implementation (`/alfred:2-run`)

**Manual invocation**:
- Code quality review
- Feature design and architecture
- Troubleshooting errors

---

## Ruby 3.4+ Modern Features (Highlights)

### Pattern Matching
```ruby
case response
in [200, { data: }]
  "Success: #{data}"
in [404, *]
  "Not found"
in [500, { error: message }]
  "Error: #{message}"
end
```

### Fiber Scheduling
```ruby
Async do
  tasks = [
    Async { fetch('https://api.example.com/users') },
    Async { fetch('https://api.example.com/posts') }
  ]
  Async::Barrier.new(tasks).wait
end
```

### Ractor for Parallelism
```ruby
ractors = items.map { |item| Ractor.new(item) { |i| process(i) } }
results = ractors.map(&:take)
```

---

## Rails 8 Key Features

### Hotwire & Turbo Streams
```ruby
respond_to do |format|
  format.turbo_stream do
    render turbo_stream: turbo_stream.append('tasks', partial: 'task')
  end
end
```

### Active Record Best Practices
```ruby
class User < ApplicationRecord
  scope :active, -> { where(active: true) }
  delegate :street, to: :address, prefix: true
  validates :email, presence: true, uniqueness: true
end
```

---

## Testing with RSpec 4

### Model Testing
```ruby
RSpec.describe User, type: :model do
  let(:user) { build(:user) }

  it { is_expected.to validate_presence_of(:email) }
  it { is_expected.to have_many(:posts) }
end
```

### Request Testing
```ruby
RSpec.describe 'Users API', type: :request do
  describe 'GET /api/users' do
    before { get '/api/users' }
    it { expect(response).to have_http_status(:ok) }
  end
end
```

---

## Performance Essentials

### YJIT Configuration
```ruby
# Enable with: RUBY_YJIT_ENABLE=1 rails server
RubyVM::YJIT.enable if RUBY_YJIT_ENABLED
```

### Query Optimization
```ruby
# Good: Eager loading
User.includes(:posts).to_a

# Efficient: Batch processing
User.find_each(batch_size: 1000) { |u| u.process }
```

---

## Installation & Setup

```bash
# Install Ruby 3.4
rbenv install 3.4.0

# Create Gemfile with essential gems
bundle add rails bundler rubocop rspec-rails
bundle add async async-http  # For concurrent IO
bundle add ractors            # For parallelism

# Generate Rails app
rails new myapp
cd myapp && bundle install
```

---

## Production Best Practices

1. **Use pattern matching** for cleaner conditionals
2. **Enable YJIT** for 15-30% performance boost
3. **Implement eager loading** to prevent N+1 queries
4. **Profile with Stackprof** before optimizing
5. **Use Fibers** for I/O-bound concurrency
6. **Apply Ractors** for CPU-bound parallelism
7. **Leverage Hotwire** for modern UX
8. **Enforce RuboCop** for code consistency

---

## Works Well With

- `moai-domain-backend` (backend architecture)
- `moai-essentials-debug` (debugging support)
- `moai-essentials-perf` (performance profiling)
- `moai-foundation-trust` (TRUST compliance)

---

## Learn More

- **Examples**: See `examples.md` for Rails, RSpec, Fibers, Ractors patterns
- **Reference**: See `reference.md` for API details and configuration
- **Modules**: See `modules/` for advanced patterns
- [Ruby 3.4 Docs](https://www.ruby-lang.org/en/documentation/)
- [Rails 8 Guides](https://guides.rubyonrails.org/)

---

## Context7 Integration

### Related Libraries & Tools
- [Rails](/rails/rails): Full-stack web framework
- [RSpec](/rspec/rspec-core): Testing framework
- [Bundler](/rubygems/bundler): Dependency manager
- [RuboCop](/rubocop/rubocop): Code analyzer
- [Hotwire](/hotwired/hotwire): Modern front-end

### Official Documentation
- [Ruby 3.4](https://www.ruby-lang.org/en/documentation/)
- [Rails Guides](https://guides.rubyonrails.org/)
- [RSpec Documentation](https://rspec.info/documentation/)
- [Bundler Guide](https://bundler.io/)

---

**Status**: Production Ready | **Version**: 4.0.0 | **Updated**: 2025-11-22