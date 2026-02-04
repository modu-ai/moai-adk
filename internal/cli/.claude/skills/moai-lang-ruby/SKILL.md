---
name: moai-lang-ruby
description: >
  Ruby 3.3+ development specialist covering Rails 7.2, ActiveRecord,
  Hotwire/Turbo, and modern Ruby patterns. Use when developing Ruby APIs,
  web applications, or Rails projects.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob Bash(ruby:*) Bash(gem:*) Bash(bundle:*) Bash(rake:*) Bash(rspec:*) Bash(rubocop:*) Bash(rails:*) mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.1.0"
  category: "language"
  status: "active"
  updated: "2026-01-11"
  modularized: "true"
  tags: "language, ruby, rails, activerecord, hotwire, turbo, rspec"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["Ruby", "Rails", "ActiveRecord", "Hotwire", "Turbo", "RSpec", ".rb", "Gemfile", "Rakefile", "config.ru"]
  languages: ["ruby"]
---

## Quick Reference (30 seconds)

Ruby 3.3+ Development Specialist - Rails 7.2, ActiveRecord, Hotwire/Turbo, RSpec, and modern Ruby patterns.

Auto-Triggers: Files with .rb extension, Gemfile, Rakefile, config.ru, Rails or Ruby discussions

Core Capabilities:

- Ruby 3.3 Features: YJIT production-ready, pattern matching, Data class, endless methods
- Web Framework: Rails 7.2 with Turbo, Stimulus, and ActiveRecord
- Frontend: Hotwire including Turbo and Stimulus for SPA-like experiences
- Testing: RSpec with factories, request specs, and system specs
- Background Jobs: Sidekiq with ActiveJob
- Package Management: Bundler with Gemfile
- Code Quality: RuboCop with Rails cops
- Database: ActiveRecord with migrations, associations, and scopes

### Quick Patterns

Rails Controller Pattern:

Create UsersController inheriting from ApplicationController. Add before_action for set_user calling only on show, edit, update, and destroy actions. Define index method assigning User.all to instance variable. Define create method creating new User with user_params. Use respond_to block with format.html redirecting on success or rendering new with unprocessable_entity status, and format.turbo_stream for Turbo responses. Add private set_user method finding by params id. Add user_params method requiring user and permitting name and email.

ActiveRecord Model Pattern:

Create User model inheriting from ApplicationRecord. Define has_many for posts with dependent destroy and has_one for profile with dependent destroy. Add validates for email with presence, uniqueness, and format using URI::MailTo::EMAIL_REGEXP. Add validates for name with presence and length minimum 2 maximum 100. Define active scope filtering where active is true. Define recent scope ordering by created_at descending. Create full_name method returning first_name and last_name joined with space and stripped.

RSpec Test Pattern:

Create RSpec.describe for User model type. In describe validations block, add expectations for validate_presence_of and validate_uniqueness_of for email. In describe full_name block, use let to build user with first_name John and last_name Doe. Add it block expecting user.full_name to eq John Doe.

---

## Implementation Guide (5 minutes)

### Ruby 3.3 New Features

YJIT Production-Ready:

YJIT is enabled by default in Ruby 3.3 providing 15 to 20 percent performance improvement for Rails applications. Enable by running ruby with yjit flag or setting RUBY_YJIT_ENABLE environment variable to 1. Check status by calling RubyVM::YJIT.enabled? method.

Pattern Matching with case/in:

Create process_response method taking response parameter. Use case with response and in for pattern matching. Match status ok with data extracting data variable and puts success message. Match status error with message extracting msg variable. Match status with guard condition checking pending or processing. Use else for unknown response.

Data Class for Immutable Structs:

Create User using Data.define with name and email symbols. Add block defining greeting method returning hello message with name. Create user instance with keyword arguments. Access name property and call greeting method.

Endless Method Definition:

Create Calculator class with add, multiply, and positive? methods using equals sign syntax for single expression methods.

### Rails 7.2 Patterns

Application Setup in Gemfile:

Set source to rubygems.org. Add rails version constraint for 7.2, pg for 1.5, puma for 6.0 or later, turbo-rails, stimulus-rails, and sidekiq for 7.0. In development and test group add rspec-rails for 7.0, factory_bot_rails, faker, and rubocop-rails with require false. In test group add capybara and shoulda-matchers.

Model with Concerns:

Create Sluggable module extending ActiveSupport::Concern. In included block add before_validation for generate_slug on create and validates for slug with presence and uniqueness. Define to_param returning slug. Add private generate_slug method setting slug from parameterized title if title present and slug blank. Create Post model including Sluggable with belongs_to user, has_many comments with dependent destroy, has_many_attached images. Add validations and published scope.

Service Objects Pattern:

Create UserRegistrationService with initialize accepting user_params. Define call method creating User, using ActiveRecord::Base.transaction to save user, create profile, and send welcome email. Return Result with success true and user. Rescue RecordInvalid returning Result with success false and errors. Add private methods for create_profile and send_welcome_email. Define Result as Data.define with success, user, and errors, adding success? and failure? predicate methods.

### Hotwire Turbo and Stimulus

Turbo Frames Pattern:

In index view, use turbo_frame_tag with posts id and iterate posts rendering each. In post partial, use turbo_frame_tag with dom_id for post, containing article with h2 link and truncated content paragraph.

Turbo Streams Pattern:

In controller create action, build post from current_user.posts. Use respond_to with format.turbo_stream and format.html for redirect or render based on save success. In create.turbo_stream.erb view, use turbo_stream.prepend for posts with post, and turbo_stream.update for new_post form partial.

Stimulus Controller Pattern:

In JavaScript controller file, import Controller from hotwired/stimulus. Export default class extending Controller with static targets array for input and submit. Define connect method calling validate. Define validate method checking all input targets have values and setting submit target disabled accordingly.

### RSpec Testing Basics

Factory Bot Patterns:

In factories file, define factory for user with sequence for email, Faker::Name.name for name, and password123 for password. Add admin trait setting role to admin symbol. Add with_posts trait with transient posts_count defaulting to 3, using after create callback to create_list posts for user.

---

## Advanced Implementation (10+ minutes)

For comprehensive coverage including:

- Production deployment patterns for Docker and Kubernetes
- Advanced ActiveRecord patterns including polymorphic, STI, and query objects
- Action Cable real-time features
- Performance optimization techniques
- Security best practices
- CI/CD integration patterns
- Complete RSpec testing patterns

See:

- modules/advanced-patterns.md for production patterns and advanced features
- modules/testing-patterns.md for complete RSpec testing guide

---

## Context7 Library Mappings

- rails/rails for Ruby on Rails web framework
- rspec/rspec for RSpec testing framework
- hotwired/turbo-rails for Turbo for Rails
- hotwired/stimulus-rails for Stimulus for Rails
- sidekiq/sidekiq for background job processing
- rubocop/rubocop for Ruby style guide enforcement
- thoughtbot/factory_bot for test data factories

---

## Works Well With

- moai-domain-backend for REST API and web application architecture
- moai-domain-database for SQL patterns and ActiveRecord optimization
- moai-workflow-testing for DDD and testing strategies
- moai-essentials-debug for AI-powered debugging
- moai-foundation-quality for TRUST 5 quality principles

---

## Troubleshooting

Common Issues:

Ruby Version Check:

Run ruby with version flag for 3.3 or later. Check YJIT status by running ruby -e with puts RubyVM::YJIT.enabled? command.

Rails Version Check:

Run rails with version flag for 7.2 or later. Run bundle exec rails about for full environment information.

Database Connection Issues:

- Check config/database.yml configuration
- Ensure PostgreSQL or MySQL service is running
- Run rails db:create if database does not exist

Asset Pipeline Issues:

Run rails assets:precompile to compile assets. Run rails assets:clobber to clear compiled assets.

RSpec Setup Issues:

Run rails generate rspec:install for initial setup. Run bundle exec rspec with specific file path for single spec. Run bundle exec rspec with format documentation for verbose output.

Turbo and Stimulus Issues:

Run rails javascript:install:esbuild for JavaScript setup. Run rails turbo:install for Turbo installation.

---

Last Updated: 2026-01-11
Status: Active (v1.1.0)
