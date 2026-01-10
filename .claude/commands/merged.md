---
description: "PR is successfully merged and closed (post-merge cleanup)"
allowed-tools: ["Bash"]
---

The pull request: $ARGUMENTS was successfully merged and closed...

### Your Task
- [ ] Clear the conversation
- [ ] Switch to the main branch

!git fetch origin main && git checkout main && git pull origin main
/clear
