# Language: Ruby

When to read: `internal/quality` sent you here because the project's detected languages include `ruby`. The project's own declarations (CI steps, `Rakefile` tasks, `AGENTS.md`) win over every default below.

## Evidence and defaults

Run every command through Bundler (`bundle exec`) when a `Gemfile` exists, so the locked versions run.

| Kind | Default command | Use it only when |
| --- | --- | --- |
| format and lint | `bundle exec rubocop` | `.rubocop.yml` exists and `rubocop` is in the `Gemfile` |
| type check | `bundle exec srb tc` or `bundle exec steep check` | Sorbet (`sorbet/`) or Steep (`Steepfile`) is set up |
| test | `bundle exec rspec` | `spec/` and `rspec` in the `Gemfile` |
| test | `bundle exec rake test` | `test/` and a `test` task in the `Rakefile` |
| build | `gem build <name>.gemspec` | the project is a gem (`*.gemspec`) |

## Dependency existence

- `https://rubygems.org/api/v1/gems/<name>.json` answers 404 for a gem that does not exist; `gem search --remote --exact <name>` prints it.
- Add with `bundle add <name>`; show the `Gemfile.lock` entry.

## Traps

- Without `bundle exec`, the machine's newest gem version runs instead of the locked one.
- `rubocop -a` or `-A` rewrites files; a VERIFY run never rewrites.
- `skip`, `pending` and `xit` hide tests; a new one is a finding.
- The Ruby version comes from `.ruby-version` or the `Gemfile`; a different interpreter is `blocked`, not a pass.
