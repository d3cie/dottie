#!/usr/bin/env bash
set -euo pipefail

repository_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
embedded_dir="$repository_dir/internal/web/dist"
app_dir="$repository_dir/web/apps/app/build"
tracker_file="$repository_dir/web/apps/client/dist/tracker.js"

if [[ ! -f "$app_dir/index.html" ]]; then
  echo "Missing dashboard build at $app_dir" >&2
  exit 1
fi
if [[ ! -f "$tracker_file" ]]; then
  echo "Missing tracker build at $tracker_file" >&2
  exit 1
fi

find "$embedded_dir" -mindepth 1 -maxdepth 1 -delete
cp -R "$app_dir"/. "$embedded_dir"/
cp "$tracker_file" "$embedded_dir/tracker.js"

