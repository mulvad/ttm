# TTM - Terminal Theme Manager

TTM is a CLI utility that manages terminal themes based on project context using a three-layer architecture:

1. **Project metadata** (`.terminal-profile`) - Per-project YAML config
2. **Semantic environment mapping** - Maps environments to themes
3. **Terminal-specific implementation** - Apple Terminal profiles via AppleScript

## Installation

```bash
# From source
go install github.com/mulvad/ttm/cmd/ttm@latest

# Or build locally
git clone https://github.com/mulvad/ttm.git
cd ttm
make install
```

## Configuration

### Global Config (`~/.ttm/config.yaml`)

Define your environments and themes:

```yaml
environments:
  production:
    theme: prod
  staging:
    theme: stage
  development:
    theme: dev

themes:
  prod:
    profile: "Red Sands"
  stage:
    profile: "Ocean"
  dev:
    profile: "Grass"
```

### Project Config (`.terminal-profile`)

Place a `.terminal-profile` file in your project root. You can specify either an environment or a theme directly:

**Using environment:**

```yaml
environment: production
```

**Using theme directly:**

```yaml
theme: prod
```

## Usage

### Commands

```bash
# Apply the terminal profile for the current project
ttm apply

# Show current terminal and project status
ttm current

# Show the full resolution chain without applying
ttm resolve
```

### Example Output

```bash
$ ttm resolve
Resolution chain:

├── [project] /Users/me/myproject/.terminal-profile → environment: production
├── [environment] production → prod
├── [theme] prod → prod
└── [profile] prod → Red Sands

Final profile: Red Sands
```

## Shell Integration

To automatically change terminal profiles when changing directories, add this to your shell configuration:

### Zsh (`~/.zshrc`)

```zsh
# TTM: Auto-apply terminal theme on directory change
ttm_chpwd() {
  if command -v ttm &> /dev/null; then
    ttm apply 2>/dev/null
  fi
}

# Add to chpwd hooks
autoload -U add-zsh-hook
add-zsh-hook chpwd ttm_chpwd

# Apply on shell startup too
ttm_chpwd
```

### Bash (`~/.bashrc`)

```bash
# TTM: Auto-apply terminal theme on directory change
ttm_prompt_command() {
  if [[ "$TTM_PREV_PWD" != "$PWD" ]]; then
    TTM_PREV_PWD="$PWD"
    if command -v ttm &> /dev/null; then
      ttm apply 2>/dev/null
    fi
  fi
}

PROMPT_COMMAND="ttm_prompt_command${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
```

## Resolution Flow

When you run `ttm apply`, the following resolution happens:

1. **Find project profile**: Walk up the directory tree to find `.terminal-profile`
2. **Load project config**: Parse the YAML to get environment or theme
3. **Resolve environment** (if applicable): Look up the environment in global config to get theme
4. **Resolve theme**: Look up the theme in global config to get terminal profile name
5. **Apply profile**: Use AppleScript to set the terminal profile

## Supported Terminals

Currently supports:

- **Apple Terminal** (macOS Terminal.app)

The architecture supports adding additional backends (iTerm2, Alacritty, etc.) in the future.

## Development

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Build binary
make build

# Install to GOPATH
make install
```

## License

MIT
