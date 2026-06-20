# TTM Examples

This directory contains examples demonstrating the two approaches for configuring terminal profiles.

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

Specify an environment when you want consistent theming based on project type:

```yaml
# .terminal-profile
environment: production
```

Resolution: `environment → theme → profile`

## Comparison

| Approach             | Use Case              | Flexibility         |
| -------------------- | --------------------- | ------------------- |
| Direct Theme         | Specific visual style | Per-project control |
| Semantic Environment | Risk-based theming    | Centralized control |

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
```
