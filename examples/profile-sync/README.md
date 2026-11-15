# Sync Terminal profiles across Macs

Keep an exported profiles file in your dotfiles repository and import it whenever
an interactive Terminal shell starts. The example assumes your dotfiles repository
is checked out at `~/dotfiles` on each Mac; adjust the paths if yours differs.

## Set up the first Mac

Install a TTM version that supports the window and title fields, then configure
the profiles you want in Terminal's settings. Export them into your existing
dotfiles checkout:

```bash
mkdir -p ~/dotfiles/ttm
ttm export -p "Pro" -p "Ocean" -o ~/dotfiles/ttm/profiles.yaml
```

Choose your own profile names, or omit the `-p` flags to export all profiles. Copy
[startup.zsh](startup.zsh) from this example to `~/dotfiles/ttm/startup.zsh`.

Add this to `~/.zshrc` after your PATH configuration and before any TTM auto-apply
hooks:

```zsh
[[ -r "$HOME/dotfiles/ttm/startup.zsh" ]] && source "$HOME/dotfiles/ttm/startup.zsh"
```

Commit and push the two files using your normal dotfiles workflow:

```bash
git -C ~/dotfiles add ttm/profiles.yaml ttm/startup.zsh
git -C ~/dotfiles commit -m "Sync Terminal profiles on shell startup"
git -C ~/dotfiles push
```

## Set up another Mac

1. Install the same TTM version and any fonts referenced by the profiles.
2. Clone or update your dotfiles repository at `~/dotfiles`.
3. Add the same `source` line to that Mac's `~/.zshrc`.
4. Open a new Terminal shell, or import immediately:

   ```bash
   ttm import -i ~/dotfiles/ttm/profiles.yaml
   ```

Profiles with matching names are updated; missing profiles are created. Importing
does not select a profile. Use Terminal's settings to choose its default profile,
or use TTM's existing `ttm apply` shell integration for project-specific profiles.
For matching theme/environment mappings, also sync `~/.ttm/config.yaml` through
your dotfiles setup and keep project `.terminal-profile` files in their repos.

## Share later changes

After editing Terminal settings on one Mac, export the managed profiles again
**before opening another shell**; startup import otherwise reapplies the saved
YAML and overwrites changes to the supported properties:

```bash
ttm export -p "Pro" -p "Ocean" -o ~/dotfiles/ttm/profiles.yaml
git -C ~/dotfiles diff -- ttm/profiles.yaml
git -C ~/dotfiles add ttm/profiles.yaml
git -C ~/dotfiles commit -m "Update Terminal profiles"
git -C ~/dotfiles push
```

On another Mac, run `git -C ~/dotfiles pull`, then open a new Terminal shell or
run the import command above. You can also edit the YAML directly and sync it.
Startup import uses the local checkout; it does not fetch remote changes.

Normal import output is hidden, while errors remain visible. Each new shell
reimports every profile in the file. Removing an entry from YAML does not delete
the corresponding profile from Terminal.

This synchronizes the attributes TTM exports: four colors, font name and size,
rows and columns, and the exposed title settings. It does not synchronize every
Terminal preference, such as the ANSI palette, transparency, keyboard mappings,
shell commands, or the app's default/startup profile selection. Those need separate
setup. macOS may ask for permission for TTM to control Terminal.
