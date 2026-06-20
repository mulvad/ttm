# Semantic Environment Example

This example demonstrates using semantic environments in the `.terminal-profile`.

## When to Use

Use semantic environments when:

- You want visual cues based on project risk level (production vs development)
- You want consistent theming across similar project types
- You want to change all production project themes by updating one config

## How It Works

The `.terminal-profile` in this directory contains:

```yaml
environment: production
```

TTM resolves this through the three-layer architecture:

```
.terminal-profile (environment: production)
        ↓
config.yaml environments.production.theme → "danger"
        ↓
config.yaml themes.danger.profile
        ↓
Terminal profile: "Red Sands"
```

## Benefits

1. **Semantic meaning**: "production" conveys intent, not just appearance
2. **Centralized control**: Change all production themes in one place
3. **Consistency**: All production projects share the same visual warning

## Testing

```bash
# From this directory
cd examples/semantic-environment

# See the full resolution chain
ttm resolve -c ../config.yaml

# Apply the profile
ttm apply -c ../config.yaml
```

## Changing All Production Themes

To change the theme for all production projects, update your global config:

```yaml
environments:
  production:
    theme: danger  # Change this to any theme name
```

All projects with `environment: production` will now use the new theme.
