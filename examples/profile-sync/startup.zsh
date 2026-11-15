# Source from ~/.zshrc after PATH configuration and before TTM auto-apply hooks.
# Keep profiles.yaml in the same dotfiles checkout on each Mac.
if [[ "$TERM_PROGRAM" == "Apple_Terminal" ]] &&
   command -v ttm >/dev/null 2>&1 &&
   [[ -r "$HOME/dotfiles/ttm/profiles.yaml" ]]; then
  ttm import -i "$HOME/dotfiles/ttm/profiles.yaml" >/dev/null
fi
