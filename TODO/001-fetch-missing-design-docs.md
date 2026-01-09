# TODO 001: Fetch missing design docs from restic

Several `x/sim3/design/*.md` documents reference `15000.md` and
`13130.md`, but neither file exists in the current repo checkout or git
history.

Action

- Locate `15000.md` and `13130.md` in restic backups and restore them.
- Decide the intended destination paths (likely `x/sim3/design/`).
- Re-check any references in `x/sim3/design/*.md` and update as needed.

