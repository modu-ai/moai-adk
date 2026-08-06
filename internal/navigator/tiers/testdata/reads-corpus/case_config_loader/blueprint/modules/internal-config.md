# internal-config

Loads and validates configuration from disk.

The config loader (ConfigLoader) reads YAML from the user's project,
applies env-var overrides, and returns a typed Config struct. Defaults
live in defaults.go.
