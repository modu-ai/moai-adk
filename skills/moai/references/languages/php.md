# Language: PHP

When to read: `internal/quality` sent you here because the project's detected languages include `php`. The project's own declarations (CI steps, `composer.json` `scripts`, `AGENTS.md`) win over every default below.

## Evidence and defaults

Prefer the `scripts` of `composer.json` (`composer run <script>`). Without a script, run the tools from `vendor/bin` after `composer install`:

| Kind | Default command | Use it only when |
| --- | --- | --- |
| format | `vendor/bin/php-cs-fixer fix --dry-run --diff` or `vendor/bin/phpcs` | `.php-cs-fixer.php`/`.php-cs-fixer.dist.php`, or `phpcs.xml`/`phpcs.xml.dist` |
| type check | `vendor/bin/phpstan analyse` or `vendor/bin/psalm` | `phpstan.neon`/`phpstan.neon.dist`, or `psalm.xml` |
| test | `vendor/bin/phpunit` | `phpunit.xml` or `phpunit.xml.dist` |
| test | `vendor/bin/pest` | `pestphp/pest` is in `composer.json` |
| manifest | `composer validate` | `composer.json` |

## Dependency existence

- `https://repo.packagist.org/p2/<vendor>/<package>.json` answers 404 for a package that does not exist; `composer show --all <vendor>/<package>` prints it.
- Add with `composer require <vendor>/<package>`; show the `composer.lock` entry.

## Traps

- `composer install` follows `composer.lock`; `composer update` rewrites it. A VERIFY run only installs.
- `php-cs-fixer fix` without `--dry-run` rewrites files.
- `markTestSkipped()` and `@group` filters hide tests; a new one is a finding.
- Platform requirements (`"php": ">=8.2"`, extensions) fail on another PHP version: that is `blocked`, not a pass.
