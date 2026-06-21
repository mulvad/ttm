# Combined Mode Example

This example demonstrates using both environment and theme together.

## When to Use

Use combined mode when:

- You have a preferred color scheme you always want to use
- But you still want visual environment indicators (badges in window title)
- You want the best of both worlds: consistent styling + environment awareness

## How It Works

The `.terminal-profile` in this directory contains:

```yaml
environment: production   # Gets badge from this
theme: dark               # But uses this theme
```

TTM resolves this by:

1. Looking up the environment to get the badge
2. Using the explicitly specified theme (ignoring the environment's theme)
3. Applying the theme's profile + the environment's badge

```
environment: production → badge: "🔴 PROD"
theme: dark → profile: "Pro"
        ↓
Terminal profile: "Pro" + Window title: "🔴 PROD - project"
```

## Testing

```bash
# From this directory
cd examples/combined-mode

# See the resolution
ttm resolve -c ../config.yaml

# Apply it
ttm apply -c ../config.yaml
```

## Comparison with Other Modes

| Mode              | Profile      | Window Title      |
| ----------------- | ------------ | ----------------- |
| `theme: dark`     | Pro          | (unchanged)       |
| `environment: production` | Red Sands | 🔴 PROD - project |
| Combined (this)   | Pro          | 🔴 PROD - project |

## Benefits

1. **Consistent styling**: Always use your preferred color scheme
2. **Environment awareness**: Always know what environment you're in
3. **Flexible**: Mix and match as needed per project
