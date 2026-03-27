# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go build -o go-presigner .   # build binary
go run . <args>              # run without building
go vet ./...                 # lint
go test ./...                # run all tests
go install .                 # install to $GOPATH/bin
```

## Architecture

A zero-dependency Go CLI that generates AWS Signature V4 presigned GET URLs for S3-compatible storage. Config is persisted in `~/.go-presigner/config` (INI format, like `~/.aws/credentials`).

### Files

| File | Responsibility |
|------|---------------|
| `main.go` | Subcommand routing, `--profile` extraction, usage text |
| `config.go` | Read/write `~/.go-presigner/config`; INI parser/writer; `LoadConfig`, `SaveConfig`, `SetConfigField`, `PrintConfig` |
| `configure.go` | Interactive `configure` command with masked prompts |
| `presign.go` | `presign` subcommand — merges stored config with flag overrides, validates, calls signer |
| `presigner.go` | Pure SigV4 signing logic — `CreatePresignedURL(PresignParams)` |

### CLI structure

```
go-presigner [--profile <name>]
  configure                    # interactive prompts
  configure list               # show config (secret masked)
  configure set <field> <val>  # non-interactive single field
  presign <key> [--expires-in N] [--bucket B] [--signing-host H] [--cdn-host H] [--region R]
```

### Key design decisions

- **`signing-host` vs `cdn-host`**: The signature is computed using `signing-host` (what S3 validates against), but the returned URL uses `cdn-host`. This allows presigned URLs to be served through a CDN or proxy without breaking the signature.
- **Priority for config values**: CLI flag → stored config → default. Implemented in `presign.go:runPresign` via `firstNonEmpty`.
- **Credentials fallback**: `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` env vars are checked before stored config.
- **Profile isolation**: `--profile` is stripped from `os.Args` before subcommand routing; all config reads/writes are profile-scoped.
- **Config file permissions**: Written with `0600` (owner read/write only) to protect credentials.
