# Ruby 3.3+ CLI Reference — Tool Command Matrix

**Framework**: Ruby 3.3+ + RSpec 6.0+ + RuboCop 1.70+ + Bundler 2.6+

---

## Ruby Runtime Commands

| Command | Purpose | Example |
|---------|---------|---------|
| `ruby --version` | Check Ruby version | `ruby --version` → `ruby 3.3.6` |
| `ruby script.rb` | Run Ruby script | `ruby script.rb` |
| `ruby -e "code"` | Execute Ruby code inline | `ruby -e "puts RUBY_VERSION"` |
| `ruby -w script.rb` | Run with warnings enabled | `ruby -w script.rb` |
| `ruby -c script.rb` | Check syntax without execution | `ruby -c script.rb` |
| `ruby -I lib script.rb` | Add lib to load path | `ruby -I lib script.rb` |
| `ruby --help` | Show Ruby help | `ruby --help` |
| `irb` | Start interactive Ruby shell | `irb` |
| `irb -r ./lib/app` | Load file in IRB | `irb -r ./lib/app` |

---

## Gem Management (Bundler 2.6+)

| Command | Purpose | Example |
|---------|---------|---------|
| `bundle install` | Install dependencies | `bundle install` |
| `bundle update` | Update all gems | `bundle update` |
| `bundle update gem_name` | Update specific gem | `bundle update rspec` |
| `bundle exec` | Run command in bundle context | `bundle exec rspec` |
| `bundle add gem_name` | Add gem to Gemfile | `bundle add rspec` |
| `bundle remove gem_name` | Remove gem from Gemfile | `bundle remove rspec` |
| `bundle show gem_name` | Show gem installation path | `bundle show rails` |
| `bundle list` | List installed gems | `bundle list` |
| `bundle outdated` | Show outdated gems | `bundle outdated` |
| `bundle clean` | Remove unused gems | `bundle clean` |
| `bundle init` | Create new Gemfile | `bundle init` |
| `bundle config` | Show bundle configuration | `bundle config` |
| `bundle gem gem_name` | Generate new gem | `bundle gem my_gem` |
| `gem list` | List installed gems | `gem list` |
| `gem install gem_name` | Install gem globally | `gem install bundler` |
| `gem uninstall gem_name` | Uninstall gem | `gem uninstall bundler` |

---

## Testing Framework (RSpec 6.0+)

| Command | Purpose | Example |
|---------|---------|---------|
| `rspec` | Run all tests | `rspec` |
| `rspec spec/` | Run tests in directory | `rspec spec/` |
| `rspec spec/file_spec.rb` | Run specific test file | `rspec spec/calculator_spec.rb` |
| `rspec spec/file_spec.rb:10` | Run test at line 10 | `rspec spec/calculator_spec.rb:10` |
| `rspec --format documentation` | Verbose output | `rspec --format documentation` |
| `rspec --format progress` | Dots progress (default) | `rspec --format progress` |
| `rspec --fail-fast` | Stop on first failure | `rspec --fail-fast` |
| `rspec --tag focus` | Run tagged tests only | `rspec --tag focus` |
| `rspec --tag ~slow` | Skip slow tests | `rspec --tag ~slow` |
| `rspec --profile` | Show slowest tests | `rspec --profile 10` |
| `rspec --only-failures` | Run only failed tests | `rspec --only-failures` |
| `rspec --next-failure` | Run next failed test | `rspec --next-failure` |
| `rspec --seed 12345` | Run with specific seed | `rspec --seed 12345` |
| `rspec --order random` | Randomize test order | `rspec --order random` |

**SimpleCov Integration (Coverage)**:
```ruby
# spec/spec_helper.rb
require 'simplecov'
SimpleCov.start do
  add_filter '/spec/'
  minimum_coverage 85
end
```

**Run with Coverage**:
```bash
rspec
open coverage/index.html
```

---

## Code Quality Linter (RuboCop 1.70+)

| Command | Purpose | Example |
|---------|---------|---------|
| `rubocop` | Check all files | `rubocop` |
| `rubocop lib/` | Check directory | `rubocop lib/` |
| `rubocop file.rb` | Check specific file | `rubocop lib/calculator.rb` |
| `rubocop -a` | Auto-correct safe issues | `rubocop -a` |
| `rubocop -A` | Auto-correct all issues | `rubocop -A` |
| `rubocop --only Style` | Check specific department | `rubocop --only Style` |
| `rubocop --except Style` | Skip specific department | `rubocop --except Style` |
| `rubocop --format offenses` | Show offense counts | `rubocop --format offenses` |
| `rubocop --format html -o report.html` | Generate HTML report | `rubocop --format html -o report.html` |
| `rubocop --list-target-files` | List files to check | `rubocop --list-target-files` |
| `rubocop --regenerate-todo` | Update .rubocop_todo.yml | `rubocop --regenerate-todo` |
| `rubocop --auto-gen-config` | Generate config from codebase | `rubocop --auto-gen-config` |

**Common Cop Departments**:
- **Style**: Code style conventions
- **Layout**: Code formatting
- **Lint**: Suspicious code patterns
- **Metrics**: Code complexity
- **Naming**: Naming conventions
- **Security**: Security issues
- **Performance**: Performance anti-patterns

**.rubocop.yml Configuration**:
```yaml
AllCops:
  TargetRubyVersion: 3.3
  NewCops: enable
  Exclude:
    - 'vendor/**/*'
    - 'db/schema.rb'

Metrics/MethodLength:
  Max: 50

Metrics/AbcSize:
  Max: 15

Metrics/CyclomaticComplexity:
  Max: 10

Style/Documentation:
  Enabled: false

Layout/LineLength:
  Max: 120
```

---

## Rails Commands (Rails 7.1+)

| Command | Purpose | Example |
|---------|---------|---------|
| `rails new app_name` | Create new Rails app | `rails new my_app` |
| `rails server` | Start development server | `rails server` or `rails s` |
| `rails console` | Start Rails console | `rails console` or `rails c` |
| `rails generate` | Generate code | `rails generate model User` |
| `rails generate model` | Generate model | `rails generate model User name:string` |
| `rails generate controller` | Generate controller | `rails generate controller Users` |
| `rails generate migration` | Generate migration | `rails generate migration AddEmailToUsers` |
| `rails db:create` | Create database | `rails db:create` |
| `rails db:migrate` | Run migrations | `rails db:migrate` |
| `rails db:rollback` | Rollback last migration | `rails db:rollback` |
| `rails db:seed` | Seed database | `rails db:seed` |
| `rails db:reset` | Reset database | `rails db:reset` |
| `rails routes` | Show all routes | `rails routes` |
| `rails test` | Run tests | `rails test` |
| `rails test:system` | Run system tests | `rails test:system` |

---

## Type Checking (Sorbet/RBS)

### Sorbet (Gradual Type Checker)

| Command | Purpose | Example |
|---------|---------|---------|
| `srb init` | Initialize Sorbet | `srb init` |
| `srb tc` | Type check files | `srb tc` |
| `srb tc --typed=strict` | Strict type checking | `srb tc --typed=strict` |

**Type Annotations**:
```ruby
# typed: strict
class Calculator
  extend T::Sig

  sig { params(a: Integer, b: Integer).returns(Integer) }
  def add(a, b)
    a + b
  end
end
```

### RBS (Ruby Signature)

| Command | Purpose | Example |
|---------|---------|---------|
| `rbs validate` | Validate RBS files | `rbs validate` |
| `rbs parse` | Parse RBS file | `rbs parse sig/calculator.rbs` |
| `rbs prototype rb` | Generate RBS from Ruby | `rbs prototype rb lib/calculator.rb` |

**RBS Signature File** (`sig/calculator.rbs`):
```rbs
class Calculator
  def add: (Integer, Integer) -> Integer
  def subtract: (Integer, Integer) -> Integer
end
```

---

## Performance Profiling

| Command | Purpose | Example |
|---------|---------|---------|
| `ruby -r profile script.rb` | Profile script execution | `ruby -r profile script.rb` |
| `ruby -r benchmark script.rb` | Benchmark code | `ruby -r benchmark script.rb` |

**Benchmark Example**:
```ruby
require 'benchmark'

Benchmark.bm do |x|
  x.report("Array.map") { (1..100000).map { |n| n * 2 } }
  x.report("Array.each") { a = []; (1..100000).each { |n| a << n * 2 } }
end
```

**Memory Profiler**:
```ruby
require 'memory_profiler'

report = MemoryProfiler.report do
  # Code to profile
  (1..100000).map { |n| n * 2 }
end

report.pretty_print
```

---

## Project Structure Setup

**Gemfile Template**:
```ruby
source 'https://rubygems.org'

ruby '3.3.6'

gem 'rake', '~> 13.0'

group :development, :test do
  gem 'rspec', '~> 6.0'
  gem 'rubocop', '~> 1.70'
  gem 'rubocop-rspec', '~> 3.3'
  gem 'simplecov', require: false
  gem 'pry-byebug'
end
```

**Directory Structure**:
```
my_project/
├── lib/
│   └── my_project.rb
├── spec/
│   ├── spec_helper.rb
│   └── my_project_spec.rb
├── Gemfile
├── Gemfile.lock
├── Rakefile
└── .rubocop.yml
```

---

## Combined Workflow (Quality Gate)

**Before Commit** (all must pass):

```bash
#!/bin/bash
set -e

echo "Running quality gate checks..."

# 1. Run tests with coverage (≥85%)
echo "1. Running tests..."
bundle exec rspec

# 2. Check coverage
echo "2. Checking coverage..."
if [ -f coverage/.last_run.json ]; then
  coverage=$(ruby -rjson -e "puts JSON.parse(File.read('coverage/.last_run.json'))['result']['line']")
  if (( $(echo "$coverage < 85" | bc -l) )); then
    echo "❌ Coverage $coverage% is below 85%"
    exit 1
  fi
  echo "✅ Coverage: $coverage%"
fi

# 3. Run RuboCop
echo "3. Running RuboCop..."
bundle exec rubocop

# 4. Type check (if using Sorbet)
# echo "4. Running type check..."
# bundle exec srb tc

echo "✅ All quality gates passed!"
```

**Save as** `bin/quality-gate.sh`:
```bash
chmod +x bin/quality-gate.sh
./bin/quality-gate.sh
```

---

## CI/CD Integration (GitHub Actions)

**.github/workflows/test.yml**:
```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        ruby-version: ['3.2', '3.3']

    steps:
      - uses: actions/checkout@v4

      - name: Set up Ruby
        uses: ruby/setup-ruby@v1
        with:
          ruby-version: ${{ matrix.ruby-version }}
          bundler-cache: true

      - name: Install dependencies
        run: bundle install

      - name: Run tests
        run: bundle exec rspec

      - name: Check coverage
        run: |
          if [ -f coverage/.last_run.json ]; then
            ruby -rjson -e "
              result = JSON.parse(File.read('coverage/.last_run.json'))
              coverage = result['result']['line']
              puts \"Coverage: #{coverage}%\"
              exit 1 if coverage < 85
            "
          fi

      - name: RuboCop
        run: bundle exec rubocop

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage/coverage.json
```

---

## Debugging Tools

| Command | Purpose | Example |
|---------|---------|---------|
| `pry` | Start Pry REPL | `pry` |
| `binding.pry` | Set breakpoint in code | Add to code: `binding.pry` |
| `byebug` | Debugger | Add to code: `byebug` |
| `ruby -rdebug script.rb` | Debug with ruby debugger | `ruby -rdebug script.rb` |

**Pry Commands**:
```ruby
# In pry session:
cd Object        # Navigate into object
ls              # List methods and variables
whereami        # Show current context
show-source     # Show method source code
show-doc        # Show method documentation
exit            # Exit pry session
```

---

## Environment Management (rbenv/rvm)

### rbenv

| Command | Purpose | Example |
|---------|---------|---------|
| `rbenv install -l` | List available Ruby versions | `rbenv install -l` |
| `rbenv install 3.3.6` | Install Ruby version | `rbenv install 3.3.6` |
| `rbenv global 3.3.6` | Set global Ruby version | `rbenv global 3.3.6` |
| `rbenv local 3.3.6` | Set local Ruby version | `rbenv local 3.3.6` |
| `rbenv versions` | List installed versions | `rbenv versions` |
| `rbenv which ruby` | Show path to ruby binary | `rbenv which ruby` |

### rvm

| Command | Purpose | Example |
|---------|---------|---------|
| `rvm list known` | List available Ruby versions | `rvm list known` |
| `rvm install 3.3.6` | Install Ruby version | `rvm install 3.3.6` |
| `rvm use 3.3.6` | Use Ruby version | `rvm use 3.3.6` |
| `rvm use 3.3.6 --default` | Set default Ruby version | `rvm use 3.3.6 --default` |
| `rvm list` | List installed versions | `rvm list` |
| `rvm gemset create name` | Create gemset | `rvm gemset create myapp` |
| `rvm gemset use name` | Use gemset | `rvm gemset use myapp` |

---

## TRUST 5 Principles Integration

### T - Test First (RSpec 6.0+)
```bash
bundle exec rspec --format documentation
```

### R - Readable (RuboCop 1.70+)
```bash
bundle exec rubocop
```

### U - Unified Types (Sorbet/RBS)
```bash
bundle exec srb tc
# or
rbs validate
```

### S - Security
```bash
bundle audit check
gem install bundler-audit
bundle audit
```

### T - Trackable (@TAG)
```bash
rg '@(CODE|TEST|SPEC):' -n lib/ spec/ --type ruby
```

---

**Version**: 0.1.0
**Created**: 2025-10-22
**Framework**: Ruby 3.3+ CLI Tools Reference
