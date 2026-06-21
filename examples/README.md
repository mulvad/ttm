# TTM Examples

This directory contains examples demonstrating the different approaches for configuring terminal profiles.

## Setup

First, copy the example config to your TTM config directory:

```bash
mkdir -p ~/.ttm
cp config.yaml ~/.ttm/config.yaml
```

Or use the `-c` flag to reference the example config directly.

## Examples

### 1. [Direct Theme](./direct-theme/)

Specify a theme directly when you want explicit control over appearance:

```yaml
# .terminal-profile
theme: retro
```

Resolution: `theme → profile`

### 2. [Semantic Environment](./semantic-environment/)

Specify an environment for risk-based theming with window title badges:

```yaml
# .terminal-profile
environment: production
```

Resolution: `environment → theme + badge → profile`

### 3. [Auto Environment](./auto-environment/)

Auto-detect environment from an environment variable (e.g., `NODE_ENV`):

```yaml
# .terminal-profile
environment: auto
```

Resolution: `$NODE_ENV → environment → theme + badge → profile`

### 4. [Combined Mode](./combined-mode/)

Use environment for the badge but override the theme:

```yaml
# .terminal-profile
environment: production   # Gets badge from this
theme: dark               # But uses this theme
```

Resolution: `environment (badge only) + explicit theme → profile`

## Comparison

| Approach         | Use Case                          | Badge | Theme Control       |
| ---------------- | --------------------------------- | ----- | ------------------- |
| Direct Theme     | Specific visual style             | No    | Per-project         |
| Semantic Env     | Risk-based theming                | Yes   | Centralized         |
| Auto Environment | Container/CI environments         | Yes   | From env var        |
| Combined Mode    | Keep theme, show env context      | Yes   | Per-project         |

## Testing the Examples

```bash
# Build ttm first
cd /path/to/ttm
make build

# Test direct theme
cd examples/direct-theme
../../ttm resolve -c ../config.yaml

# Test semantic environment
cd ../semantic-environment
../../ttm resolve -c ../config.yaml

# Test auto environment
cd ../auto-environment
NODE_ENV=production ../../ttm resolve -c ../config.yaml

# Test combined mode
cd ../combined-mode
../../ttm resolve -c ../config.yaml
```
