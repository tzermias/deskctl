# Build, Release & CI

## Local builds

`make build` runs a goreleaser snapshot build (`goreleaser build --clean
--snapshot`) for Linux only, with `CGO_ENABLED=0`. Binaries go to:

```
./dist/deskctl_linux_<ARCH>/deskctl   # ARCH: amd64_v1 | arm_6 | arm64_v8.0
```

`Version`, `Commit` and `BuildTime` in `cmd/` are injected through ldflags
(see `.goreleaser.yaml`). Don't hard-code them.

## CI (GitHub Actions)

- `build.yml`: runs on push/PR to `main`. Runs fmt, golangci-lint v2.1,
  vet, test, then a goreleaser snapshot build.
- `release.yml`: runs on `v*` tags and publishes a GitHub release
  via goreleaser.
