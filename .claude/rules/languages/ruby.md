---
paths:
  - "**/*.rb"
  - "**/Gemfile"
  - "**/Gemfile.lock"
  - "**/*.rake"
---

# Ruby Rules

Version: Ruby 3.3+

## Tooling

- Linting: rubocop
- Testing: RSpec or Minitest
- Package management: Bundler

## Best Practices (2026)

- Enable YJIT for 30-50% performance improvement over Ruby 2.7
- Use Hotwire (Turbo + Stimulus) as default frontend stack
- Use Turbo Streams for partial page updates without full requests
- Prefer incremental Rails upgrades (e.g., 6.1 → 7.0 → 7.2)
- Use Rails conventions - they solve real pain points

## Rails 7.2+ Features

- Hotwire is the default frontend framework
- Zeitwerk autoloading
- Parallel testing
- Multi-database support

## Hotwire Patterns

```ruby
# Turbo Stream response
respond_to do |format|
  format.turbo_stream do
    render turbo_stream: turbo_stream.replace(@user)
  end
end
```

```html
<!-- Turbo Frame for partial updates -->
<turbo-frame id="user_<%= @user.id %>">
  <%= render @user %>
</turbo-frame>
```

## Performance

- Use eager loading to prevent N+1 queries
- Use counter caches for frequently counted associations
- Use Russian Doll caching with cache keys

## Rails 8 Features (Preview)

- Solid Queue for background jobs
- Built-in authentication generator
- Ruby 3.3+ required for best performance

## MoAI Integration

- Use Skill("moai-lang-ruby") for detailed patterns
- Follow TRUST 5 quality gates
