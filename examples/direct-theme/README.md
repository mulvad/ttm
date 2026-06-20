# Direct Theme Example

This example demonstrates specifying a theme directly in the `.terminal-profile`.

## When to Use

Use direct theme specification when:

- You want a specific visual style for a project
- The project doesn't fit into a semantic environment category
- You prefer explicit control over the terminal appearance

## How It Works

The `.terminal-profile` in this directory contains:

```yaml
theme: retro
```

This bypasses environment resolution and maps directly to the theme in your global config:

```
.terminal-profile (theme: retro)
        ↓
config.yaml themes.retro.profile
        ↓
Terminal profile: "Homebrew"
```

## Testing

```bash
# From this directory
cd examples/direct-theme

# See resolution chain
ttm resolve -c ../config.yaml

# Apply the profile
ttm apply -c ../config.yaml
```
