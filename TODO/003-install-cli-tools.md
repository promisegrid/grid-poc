# TODO 003: Install CLI tools that help Codex do repo/spec work

## Goal

Install a small set of local CLI tools that improve search, review, and
document/PDF processing.

## Suggested tools

- Search/navigation: `rg` (ripgrep), `fd`, `fzf`, `bat` (or `batcat`), `tree` (or `eza`)
- Data inspection: `jq`, `yq`
- Git review: `delta`, `tig`
- Docs/PDFs: `pdftotext`/`pdfinfo` (poppler), `qpdf`, `pandoc`
- CBOR/CID: `ipfs` (kubo) for CID/DAG tooling; optional CBOR inspector

## Next step

Decide package manager and install set (e.g., `apt`, `dnf`, `pacman`), then
update this TODO with the exact command(s) used for reproducibility.
