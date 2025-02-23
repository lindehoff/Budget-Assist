1. **CLI Framework & Structure**

```go
- Primary Package: github.com/spf13/cobra
- Pattern: Root command with nested subcommands
- Best Practices:
  - Commands defined in separate files
  - Use PersistentPreRun for cross-command setup
  - SilenceUsage/SilenceErrors for clean error handling
  - Strict flag validation with MarkFlagRequired()
  - Viper integration for config management (github.com/spf13/viper)
```

2. **Error Handling**

```go
- Custom Error Types: 
  type RedisOperationError struct { ... }
- Error Wrapping: fmt.Errorf("failed... %w", err)
- Error Logging:
  logger.Error("context", slog.Any("error", err), slog.String("stack_trace", debug.Stack()))
- Validation: github.com/go-playground/validator/v10
```

3. **Logging**

```go
- Standard Package: log/slog (Structured logging)
- Patterns:
  - Central logger instance injected through context
  - JSON format in production
  - Rich context with slog.Any() for debugging
  - Log command lifecycle events (start/finish/error)
```

4. **Dependency Management**

```go
- Interface-based Design:
  type Client interface { SetJSON(...) }
- Dependency Injection:
  func NewRedisClient(host string, port int) (redisexporter.Client, error)
- Mock Implementations:
  internal/aws/mock_s3_client.go
```

5. **Testing Practices**

```go
- Table-driven tests
- Custom mock implementations (MockValidateClient)
- Test helpers for complex structures
- Strict error type checking
- Coverage for:
  - Error conditions
  - Boundary cases
  - Serialization/deserialization
  - API client interactions
```

6. **Concurrency & Context**

```go
- Context propagation through all layers
- signal.NotifyContext for graceful shutdown
- checkContextCancelled() pattern in long-running operations
- Concurrent-safe logging
```

7. **Documentation Standards**

```go
- Package-level godocs explaining responsibility
- Public types/methods documented
- Error reasons clearly specified
- Examples in test files
```

8. **CI/CD & Packaging**

```dockerfile
- Multi-stage Docker builds
- Static binaries with CGO_ENABLED=0
- Alpine base images
- Non-root container user
```

9. **Style Enforcement**

```go
- Slog context ordering: 
  logger.Info("message", slog.String("key", val), slog.Int("count", n))
- Error message style: "failed to <action> [for <resource>]: <cause>"
- Receiver names: Single letter (c *RedisClient)
- Line length: 100-120 chars
```

11. **Serialization Patterns**

```go
- JSON: encoding/json with MarshalIndent
- YAML: gopkg.in/yaml.v3 with custom cleaning logic
- Strict type conversion between API/domain models
```
