# .zshrc
alias sync-rules='~/workspace/agent-rules/scripts/rsync-rules.sh'

```bash
# sync Claude setup
sync-rules claude
sync-rules agents

# sync Cursor setup
sync-rules cursor
```