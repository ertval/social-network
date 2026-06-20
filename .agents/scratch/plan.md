# Plan: Issue #3 - Rename Go module to social-network

## 🚀 Checklist

- [ ] Update `go.mod` to rename module to `social-network`
- [ ] Update `.golangci.yml` to set gci prefix to `social-network`
- [ ] Rename import paths in all `.go` files from `github.com/arnald/forum` to `social-network`
- [ ] Rename module references in documentation and audit files
- [ ] Run `go mod tidy`
- [ ] Run `go vet ./...` to verify Go code soundness
- [ ] Run `go build ./...` to build Go files
- [ ] Run `make test` to verify tests pass
- [ ] Run `make ci` to run full CI pipeline
- [ ] Update knowledge graph: `graphify update .`
- [ ] Verify branch rules and create PR

## 🛠️ Step Detail

### 1. Update module declaration
Run `go mod edit -module social-network` or edit `go.mod` directly.

### 2. Update lint config
Update `prefix(github.com/arnald/forum)` to `prefix(social-network)` in `.golangci.yml`.

### 3. Replace all import paths
Use `sed` or a Go script to replace `github.com/arnald/forum` with `social-network` in all `.go` files:
```bash
sed -i 's|github.com/arnald/forum|social-network|g' $(grep -rl 'github.com/arnald/forum' --include='*.go' .)
```

### 4. Verify & Tidy
Run `go mod tidy` followed by compilation and test checks.
