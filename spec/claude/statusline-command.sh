#!/bin/bash

# Read JSON input from stdin
input=$(cat)

# Extract model display name
model=$(echo "$input" | jq -r '.model.display_name // "Claude"')

# Extract context token usage directly from the API response
context_size=$(echo "$input" | jq -r '.context_window.context_window_size // 0')
used_percent=$(echo "$input" | jq -r '.context_window.used_percentage // 0' | cut -d'.' -f1)

context_display=""

if [ "$context_size" -gt 0 ]; then
   # Calculate tokens from the pre-calculated percentage (most accurate)
   display_percent=${used_percent:-0}
   total_used=$((display_percent * context_size / 100))

   # Determine color based on usage percentage
   # Green: 0-60%, Yellow: 61-85%, Red: 86-100%
   if [ "$display_percent" -le 60 ]; then
       context_color="\033[32m"  # Green
   elif [ "$display_percent" -le 85 ]; then
       context_color="\033[33m"  # Yellow
   else
       context_color="\033[31m"  # Red
   fi

   # Format tokens with k suffix for cleaner display
   if [ $total_used -ge 1000 ]; then
       used_k=$((total_used / 1000))
       context_size_k=$((context_size / 1000))
       context_display=$(printf "${context_color}%dk/%dk (%d%%)\033[0m" "$used_k" "$context_size_k" "$display_percent")
   else
       context_display=$(printf "${context_color}%d/%d (%d%%)\033[0m" "$total_used" "$context_size" "$display_percent")
   fi
else
   context_display="--"
fi


# Extract current directory
cwd=$(echo "$input" | jq -r '.workspace.current_dir // .cwd')
dir_display=$(basename "$cwd")


# Get git branch (skip locks for performance)
git_branch=""
if [ -d "$cwd/.git" ] || git -C "$cwd" rev-parse --git-dir > /dev/null 2>&1; then
   branch=$(git -C "$cwd" --no-optional-locks branch --show-current 2>/dev/null)
   if [ -n "$branch" ]; then
       git_branch=" | ⎇ $branch"
   fi
fi


# Output formatted status line with icons and colors
# Note: Status line is displayed with dimmed colors in terminal
# Colors: cyan for model, context uses dynamic green/yellow/red, dark blue for directory, light green for git branch
printf "\033[36m🤖 %s\033[0m | 📊 %s | \033[34m📁 %s\033[0m\033[92m%s\033[0m" "$model" "$context_display" "$dir_display" "$git_branch"
