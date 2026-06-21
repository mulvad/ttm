# TTM - Terminal Theme Manager

TTM is a CLI utility that manages terminal themes based on project context using a three-layer architecture:

1. **Project metadata** (`.terminal-profile`) - Per-project YAML config
2. **Semantic environment mapping** - Maps environments to themes
3. **Terminal-specific implementation** - Apple Terminal profiles via AppleScript

## Installation

### Homebrew (Recommended)

```bash
brew tap mulvad/tap
brew install --cask ttm
```

### Build from Source

```bash
git clone https://github.com/mulvad/ttm.git
cd ttm
make install
```

## Configuration

### Global Config (`~/.ttm/config.yaml`)

Define your environments, themes, and optionally an environment variable for auto-detection:

```yaml
# Optional: detect environment from this env var (e.g., NODE_ENV, APP_ENV)
environment_variable: NODE_ENV

environments:
  production:
    theme: prod
    badge: "🔴 PROD"    # Optional: sets window title
  staging:
    theme: stage
    badge: "🟡 STAGE"
  development:
    theme: dev
    badge: "🟢 DEV"

themes:
  prod:
    profile: "Red Sands"
  stage:
    profile: "Ocean"
  dev:
    profile: "Grass"
```

### Project Config (`.terminal-profile`)

Place a `.terminal-profile` file in your project root. Multiple modes are supported:

**Mode 1: Using environment (theme + badge from config)**

```yaml
environment: production
```

**Mode 2: Using theme directly (no badge)**

```yaml
theme: prod
```

**Mode 3: Auto-detect from environment variable**

```yaml
environment: auto
```

When set to `auto`, TTM reads the configured `environment_variable` (e.g., `NODE_ENV`) and matches its value against your environments.

**Mode 4: Combined (explicit theme + badge from environment)**

```yaml
environment: production   # Gets badge from this
theme: dev                # But uses this theme
```

This lets you keep your preferred color scheme while still showing environment context in the window title.

## Usage

### Commands

```bash
# Apply the terminal profile for the current project
ttm apply

# Show current terminal and project status
ttm current

# Show the full resolution chain without applying
ttm resolve

# Export terminal profiles to a file
ttm export -o profiles.yaml

# Import terminal profiles from a file
ttm import -i profiles.yaml
```

### Example Output

```bash
$ ttm resolve
Resolution chain:

├── [project] /Users/me/myproject/.terminal-profile → environment: production
├── [environment] production → theme: prod, badge: 🔴 PROD
├── [theme] prod → prod
└── [profile] prod → Red Sands

Final profile: Red Sands

$ ttm apply
Applied profile: Red Sands
Set badge: 🔴 PROD
```

**With auto-detection:**

```bash
$ NODE_ENV=production ttm resolve
Resolution chain:

├── [project] /Users/me/myproject/.terminal-profile → environment: auto → production
├── [environment] production → theme: prod, badge: 🔴 PROD
├── [theme] prod → prod
└── [profile] prod → Red Sands

Final profile: Red Sands
```

## Shell Integration

To automatically change terminal profiles when changing directories, add this to your shell configuration:

### Zsh (`~/.zshrc`)

```zsh
# Disable oh-my-zsh auto title (required for TTM title management)
export DISABLE_AUTO_TITLE="true"

# If using Powerlevel10k, also disable its title management
POWERLEVEL9K_DISABLE_TERM_TITLE=true

# TTM: Auto-apply terminal theme on directory change
ttm_chpwd() {
  if command -v ttm &> /dev/null; then
    ttm apply 2>/dev/null
  fi
}

autoload -U add-zsh-hook
add-zsh-hook chpwd ttm_chpwd

# Apply on shell startup too
ttm_chpwd
```

### Bash (`~/.bashrc`)

```bash
# Disable auto title (required for TTM title management)
export DISABLE_AUTO_TITLE="true"

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

## Profile Export/Import

TTM can export and import terminal profiles for backup or sharing:

```bash
# Export all profiles
ttm export -o my-profiles.yaml

# Export specific profiles
ttm export -o my-profiles.yaml -p "Pro" -p "Ocean"

# Import profiles (creates new or updates existing)
ttm import -i my-profiles.yaml

# Import specific profiles
ttm import -i my-profiles.yaml -p "Pro"
```

Exported format:

```yaml
profiles:
  - name: Pro
    background_color:
      red: 0
      green: 0
      blue: 0
    text_color:
      red: 65535
      green: 65535
      blue: 65535
    bold_text_color:
      red: 65535
      green: 65535
      blue: 65535
    cursor_color:
      red: 35700
      green: 35700
      blue: 35700
    font_name: Menlo-Regular
    font_size: 12
```

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

### Test changes are working without creating a new release

```bash
git tag -d v0.1.0 && git push origin :refs/tags/v0.1.0 && git tag v0.1.0 && git push origin v0.1.0
```

## License

MIT
