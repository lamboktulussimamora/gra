# GRA Framework Performance Benchmarks

This document provides benchmark results for the GRA framework's key components.

## Router Performance

| Benchmark | Operations | Ns/Op | Bytes/Op | Allocs/Op |
|-----------|-----------|-------|----------|-----------|
| SimpleRoute | 5,618,768 | 195.3 | 416 | 9 |
| ParameterizedRoute | 2,165,229 | 555.8 | 1,088 | 16 |
| ManyRoutes_Simple | 42,332 | 28,220 | 36,048 | 531 |
| ManyRoutes_WithParameter | 34,532 | 34,811 | 45,457 | 648 |
| DeepNestedParameters | 14,604 | 70,299 | 83,009 | 814 |

## Interpretation

1. **Simple Routes**: The framework can handle ~5.6 million operations per second for simple routes, which is very efficient.
2. **Parameterized Routes**: Handling dynamic parameters is about 2-3x slower than static routes, but still very fast at ~2.1 million ops/sec.
3. **Complex Routing**: When dealing with large numbers of routes or deeply nested parameters:
   - The router can still process 42k ops/sec with many simple routes
   - Parameter extraction reduces this to ~34k ops/sec
   - Complex nested parameters bring this down to ~14.6k ops/sec

## Memory Usage

Memory usage increases significantly with route complexity:
- Simple routes: 416 bytes with 9 allocations
- Parameterized routes: 1,088 bytes with 16 allocations
- Complex nested parameters: 83,009 bytes with 814 allocations

## Recommendations

1. For high-performance applications:
   - Keep route hierarchies shallow
   - Minimize the number of path parameters
   - Consider optimizing the router's path matching algorithm for further improvements

2. For routes with complex parameter patterns:
   - Consider implementing route caching
   - Monitor memory usage in production

## Conclusion

The GRA framework offers excellent performance for a lightweight Go web framework. Simple and parameterized routes are very efficient, while even complex routing scenarios maintain reasonable performance.

### Hardware Information

These benchmarks were run on:
- OS: macOS
- Architecture: arm64
- CPU: Apple M2
