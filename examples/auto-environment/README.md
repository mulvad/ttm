# Auto Environment Example

This example demonstrates automatic environment detection from an environment variable.

## When to Use

Use auto environment detection when:

- You're working in Docker containers where `NODE_ENV` or similar is set
- Your CI/CD pipeline sets environment variables
- You want the terminal to automatically reflect the current runtime environment

## How It Works

The `.terminal-profile` in this directory contains:

```yaml
environment: auto
```

TTM resolves this by:

1. Reading `environment_variable` from your global config (e.g., `NODE_ENV`)
2. Checking the value of that env var (e.g., `production`)
3. Matching it against your configured environments
4. Applying the corresponding theme and badge

```
$NODE_ENV = "production"
        ↓
config.yaml environments.production
        ↓
theme: danger, badge: "🔴 PROD"
        ↓
Terminal profile: "Red Sands" + Window title: "🔴 PROD - project"
```

## Configuration Required

Your global config (`~/.ttm/config.yaml`) must specify the env var:

```yaml
environment_variable: NODE_ENV   # or APP_ENV, RAILS_ENV, etc.

environments:
  production:
    theme: danger
    badge: "🔴 PROD"
  staging:
    theme: caution
    badge: "🟡 STAGE"
  development:
    theme: safe
    badge: "🟢 DEV"
```

## Testing

```bash
# From this directory
cd examples/auto-environment

# Test with different env values
NODE_ENV=production ttm resolve -c ../config.yaml
NODE_ENV=staging ttm resolve -c ../config.yaml
NODE_ENV=development ttm apply -c ../config.yaml
```

## Benefits

1. **Zero configuration per project**: Just use `environment: auto` everywhere
2. **Dynamic**: Terminal theme changes when env var changes
3. **Container-friendly**: Works with Docker, Kubernetes, etc.
4. **CI/CD integration**: Automatically shows correct environment in pipelines
