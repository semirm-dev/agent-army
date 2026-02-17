# .zshrc
alias sync-rules='~/workspace/agent-rules/scripts/rsync-rules.sh'

```bash
# sync Claude setup (includes agents/)
sync-rules claude

# sync Cursor setup
sync-rules cursor
```