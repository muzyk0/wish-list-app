#!/bin/bash

# Configuration
SOURCE_FILE="$HOME/.qwen/oauth_creds.json"
DEST_FILE="$HOME/.claude-code-router/config.json"
BACKUP_DIR="$HOME/.claude-code-router/backups"
PROVIDER_NAME="qwen"
DRY_RUN=false
MAX_BACKUPS=10  # Ограничение количества хранимых бэкапов

# Parse options
while getopts "n" opt; do
    case $opt in
        n) DRY_RUN=true ;;
        *) echo "Usage: $0 [-n] (dry run)"; exit 1 ;;
    esac
done

# Check dependencies
if ! command -v jq >/dev/null 2>&1; then
    echo "Error: jq is required but not installed" >&2
    exit 1
fi

# Check source file
if [[ ! -f "$SOURCE_FILE" ]]; then
    echo "Error: Source file not found: $SOURCE_FILE" >&2
    exit 1
fi

if [[ ! -r "$SOURCE_FILE" ]]; then
    echo "Error: Cannot read source file: $SOURCE_FILE" >&2
    exit 1
fi

# Check destination file
if [[ ! -f "$DEST_FILE" ]]; then
    echo "Error: Destination file not found: $DEST_FILE" >&2
    exit 1
fi

if [[ ! -w "$DEST_FILE" ]]; then
    echo "Error: Cannot write to destination file: $DEST_FILE" >&2
    exit 1
fi

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"
if [[ $? -ne 0 ]]; then
    echo "Error: Failed to create backup directory: $BACKUP_DIR" >&2
    exit 1
fi

# Extract access token
ACCESS_TOKEN=$(jq -r '.access_token' "$SOURCE_FILE" 2>/dev/null)
JQ_EXIT_CODE=$?

if [[ $JQ_EXIT_CODE -ne 0 ]]; then
    echo "Error: Failed to parse JSON from $SOURCE_FILE" >&2
    exit 1
fi

if [[ -z "$ACCESS_TOKEN" || "$ACCESS_TOKEN" == "null" ]]; then
    echo "Error: access_token not found or is null in $SOURCE_FILE" >&2
    exit 1
fi

# Create backup with timestamp
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/config.json.backup.$TIMESTAMP"

if cp "$DEST_FILE" "$BACKUP_FILE" 2>/dev/null; then
    echo "Created backup: $BACKUP_FILE"

    # Clean up old backups (keep only MAX_BACKUPS most recent)
    BACKUP_COUNT=$(ls -1 "$BACKUP_DIR"/config.json.backup.* 2>/dev/null | wc -l)
    if [[ $BACKUP_COUNT -gt $MAX_BACKUPS ]]; then
        echo "Cleaning up old backups (keeping $MAX_BACKUPS most recent)..."
        ls -1t "$BACKUP_DIR"/config.json.backup.* 2>/dev/null | tail -n +$(($MAX_BACKUPS + 1)) | while read old_backup; do
            echo "  Removing: $(basename "$old_backup")"
            rm "$old_backup"
        done
    fi
fi

# Update config
TEMP_FILE=$(mktemp) || { echo "Error: Failed to create temp file"; exit 1; }
trap 'rm -f "$TEMP_FILE"' EXIT

if ! jq --arg token "$ACCESS_TOKEN" --arg name "$PROVIDER_NAME" '
  .Providers |= map(
    if .name == $name then
      .api_key = $token
    else
      .
    end
  )
' "$DEST_FILE" > "$TEMP_FILE" 2>/dev/null; then
    echo "Error: Failed to update configuration" >&2
    exit 1
fi

# Validate generated JSON
if ! jq empty "$TEMP_FILE" 2>/dev/null; then
    echo "Error: Generated configuration is invalid JSON" >&2
    exit 1
fi

if $DRY_RUN; then
    echo "Dry run - configuration would be updated"
    echo "New token (first 10 chars): ${ACCESS_TOKEN:0:10}..."
    echo "Diff:"
    diff -u "$DEST_FILE" "$TEMP_FILE" || true
    exit 0
fi

# Apply changes
if ! mv "$TEMP_FILE" "$DEST_FILE"; then
    echo "Error: Failed to replace configuration file" >&2
    exit 1
fi

echo "Successfully updated $PROVIDER_NAME api_key in $DEST_FILE"

# Restart service if available
if command -v ccr >/dev/null 2>&1; then
    echo "Restarting claude-code-router..."
    if ccr restart; then
        echo "Service restarted successfully"
    else
        echo "Warning: Service restart failed" >&2
    fi
fi

echo "Script completed successfully"
echo "Backups are stored in: $BACKUP_DIR"
echo "To delete all backups: rm -rf $BACKUP_DIR/*"
