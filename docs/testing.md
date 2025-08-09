# Testing

## Unit tests
```bash
# Run frontend and backend unit tests
make test

# Backend only
cd bridge && go test ./...

# Frontend only
cd ui && npm test
```

## Integration tests
`run_tests.sh` exercises the backend with coverage, fuzzing, race detection and benchmarks.
```bash
./run_tests.sh            # Linux/macOS
# or
pwsh run_tests.ps1        # Windows
```

## End-to-end (E2E) tests
Requires a built UI (`make ui`) and a running Cockpit instance.
```bash
cd ui
npm run e2e               # headless Cypress run
```
