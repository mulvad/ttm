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

# Optional: fallback profile when no .terminal-profile is found
# (TTM automatically remembers your original profile and restores it,
# but you can set this as a fallback)
default_profile: "Basic"

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
# Apply the terminal theme for the current project (silent by default)
ttm apply

# Apply with verbose output
ttm apply -v

# Show current terminal and project status
ttm current

# Show the full resolution chain without applying
ttm resolve

# List all available themes
ttm themes

# List all available environments
ttm environments

# Verify your TTM setup is working correctly
ttm verify

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
# (silent by default - theme is applied)

$ ttm apply -v
Applied environment: production
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
  ttm apply
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
    ttm apply
  fi
}

PROMPT_COMMAND="ttm_prompt_command${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
```

## Resolution Flow

When you run `ttm apply`, the following resolution happens:

1. **Find project profile**: Walk up the directory tree to find `.terminal-profile`
2. **Load project config**: Parse the YAML to get environment and/or theme
3. **Resolve environment** (if applicable): Look up the environment in global config to get theme and badge
4. **Determine final theme**:
   - If explicit `theme` is specified → use it (explicit theme wins)
   - Otherwise → use environment's theme
5. **Resolve theme**: Look up the theme in global config to get terminal profile name
6. **Apply profile**: Use AppleScript to set the terminal profile

### Resolution Order (Precedence)

When both `environment` and `theme` are specified in `.terminal-profile`:

```yaml
environment: emacs  # Gets badge from here
theme: safe         # Gets profile from here (theme wins)
```

The resolution follows this precedence:
1. **Explicit theme** wins over environment's theme (but environment's badge is used)
2. **Environment badge** is always used if environment is specified
3. **Auto-detection** works: `environment: auto` detects the env var and applies the same rules

**Examples:**

| Config | EDITOR | Result |
|--------|--------|--------|
| `environment: emacs` | - | Theme from emacs env, badge 🟢 EMACS |
| `theme: safe` | - | Theme safe, no badge |
| `environment: emacs`<br>`theme: safe` | - | Theme safe (explicit wins), badge 🟢 EMACS |
| `environment: auto` | emacs | Theme from emacs env, badge 🟢 EMACS |
| `environment: auto`<br>`theme: safe` | emacs | Theme safe (explicit wins), badge 🟢 EMACS (from env) |

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
    window:
      rows: 40
      columns: 120
    title:
      custom: "Development"
      show_custom: true
      show_device_name: false
      show_shell_path: false
      show_window_size: false
      show_profile_name: false
```

The optional `window` and `title` sections control the profile properties exposed
by Terminal's AppleScript API. Rows and columns must be positive integers. Omitted
fields leave existing settings unchanged; explicit `false` disables a title
component, and `custom: ""` clears the profile's custom title. Existing YAML files
without these sections continue to work. These sections do not cover every option
in Terminal's Window and Tab settings panels.

Terminal couples custom-title visibility to its text: setting `show_custom: false`
also clears the custom title through its scripting API, even when `custom` is
omitted. If both are supplied, `show_custom: false` takes precedence.

Dimensions are profile defaults for new windows and tabs; importing a profile does
not explicitly resize existing windows. `ttm apply` preserves the profile's title
components. An environment badge temporarily overrides the active tab's custom
title and enables its display, while leaving the other title components alone.
Clearing the badge restores the profile's custom title and its display preference.

To import managed profiles at each interactive Terminal shell startup, place this
in `~/.zshrc` after PATH configuration and before the TTM auto-apply setup:

```zsh
if [[ "$TERM_PROGRAM" == "Apple_Terminal" ]] &&
   command -v ttm >/dev/null 2>&1 &&
   [[ -r "$HOME/.ttm/profiles.yaml" ]]; then
  ttm import -i "$HOME/.ttm/profiles.yaml" >/dev/null
fi
```

Each startup imports all profiles in that file, updating profiles with matching
names. Errors remain visible. YAML export/import carries colors, font settings,
and the size/title fields above; it does not preserve the full native profile.

For a shared setup across Macs, see the [profile sync example](examples/profile-sync/README.md).
It includes a reusable startup script, a dotfiles workflow, and instructions for
sharing later changes without overwriting unsaved profile edits.

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

To check profile export/import against a real Terminal.app on macOS:

```bash
TTM_TERMINAL_INTEGRATION=1 go test ./internal/terminal -run TestProfileSettingsTerminalIntegration -v
```

This opt-in test creates a temporary profile, checks its size and title settings
and saved preferences, then deletes it. macOS may request permission to control
Terminal.

### Test changes are working without creating a new release

```bash
git tag -d v0.1.0 && git push origin :refs/tags/v0.1.0 && git tag v0.1.0 && git push origin v0.1.0
```

## License

MIT
