# CONTRIBUTING

## Development Workflow

### Setting Up Local Environment

```bash
# Clone repository
git clone <repo-url>
cd stepdeploy-lab

# Install dependencies
make vendor

# Build services
make build

# Or use Docker
make docker-build
make docker-up

# Verify services are running
make healthcheck
```

### Making Changes

1. **Create feature branch**
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make changes**
   - Edit source files
   - Keep changes focused on single feature/fix
   - Add tests for new functionality

3. **Test changes**
   ```bash
   make test
   make lint
   make docker-build
   make docker-up
   make healthcheck
   ```

4. **Format code**
   ```bash
   make fmt
   ```

5. **Commit changes**
   ```bash
   git add .
   git commit -m "feat: Add new feature"
   git push origin feature/my-feature
   ```

6. **Create pull request**
   - Link to relevant issues
   - Describe changes clearly
   - Ensure CI/CD pipeline passes

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation
- `style:` Code style (no logic change)
- `refactor:` Code refactoring
- `test:` Test additions
- `chore:` Build/tooling changes

**Example:**
```
feat(ca-service): Add certificate validation

- Implement detailed certificate validation
- Add tests for validation logic
- Update documentation

Fixes #123
```

## Testing

### Unit Tests
```bash
# Run all tests
make test

# Run specific service tests
cd services/ca-service && go test ./...

# Run with coverage
go test -cover ./...
```

### Integration Tests
```bash
# Start services
make docker-up

# Run integration tests
make enroll

# Check traces
curl http://localhost:8080/traces | jq .
```

### Failure Injection Testing
```bash
# Test specific failure scenario
make inject-failure FAILURE=expired-oauth-token

# Verify failure is handled
make enroll

# Check incident was created
curl http://localhost:8080/incidents | jq .

# Reset for next test
make reset
```

## Code Standards

### Go Code Style
- Follow Go conventions from `effective go`
- Use `gofmt` for formatting
- Run `golangci-lint` before committing
- Write unit tests for new code

### Go Packages Used
- `github.com/google/uuid` - UUID generation
- `github.com/gorilla/mux` - HTTP routing
- Standard library: `crypto/x509`, `net/http`, `encoding/json`, `time`

### Error Handling
```go
// Always handle errors
if err != nil {
    logger.Error("Operation failed: %v", err)
    return nil, err
}

// Log with context
logger.Error("Failed to sign certificate for %s: %v", clientCN, err)
```

### Logging
```go
// Use structured logging
logger.Debug("Starting enrollment for device %s", deviceID)
logger.Info("Certificate signed: %s", fingerprint)
logger.Warn("Token expires in %d seconds", timeToExpiry)
logger.Error("Failed to connect to CA service: %v", err)
```

## Documentation

### Update Documentation When
- Adding new endpoints
- Changing configuration
- Adding failure scenarios
- Changing architecture
- Updating dependencies

### Files to Update
- `README.md` - Quick start and overview
- `docs/architecture.md` - System design changes
- `docs/runbooks/*.md` - Troubleshooting guides
- `docs/escalation_playbook.md` - Escalation procedures
- Service-specific `Dockerfile` comments
- Go code comments for public functions

## Troubleshooting

### Common Issues

**Build Fails**
```bash
# Clean and rebuild
make clean
make build

# Check Go version
go version  # Should be 1.22+
```

**Docker Issues**
```bash
# Rebuild without cache
make docker-build

# Check logs
docker compose logs <service>

# Reset environment
make reset
```

**Tests Fail**
```bash
# Run with verbose output
make test

# Check for race conditions
go test -race ./...
```

## Review Process

1. **Automated Checks**
   - GitHub Actions CI pipeline
   - Code linting
   - Build validation
   - Test coverage

2. **Code Review**
   - Minimum 1 approval required
   - Comments addressed before merge
   - Tests must pass

3. **Merge**
   - Squash and merge to main branch
   - Delete feature branch
   - Update CHANGELOG.md

## Release Process

```bash
# Create release branch
git checkout -b release/v1.2.0

# Update version numbers
# Update CHANGELOG

# Create tag
git tag -a v1.2.0 -m "Release v1.2.0"
git push origin release/v1.2.0
git push origin v1.2.0

# Deploy to production
make terraform-apply
```

## Getting Help

- Check existing issues and pull requests
- Review runbooks in `docs/runbooks/`
- Check architecture documentation
- Reach out to team leads
- File new issues with detailed reproduction steps

## Code of Conduct

- Be respectful in all interactions
- Provide constructive feedback
- Help others learn and grow
- Report issues through proper channels
