# Performance Baselines

This document captures the initial performance baselines for Treacherest, established during Phase 1 implementation.

## Date: 2025-06-20
## Test Environment
- CPU: AMD Ryzen 9 3900X 12-Core Processor
- OS: Linux
- Go Version: As specified in flake.nix

## Performance Baselines

### Core Operations

| Operation | Target | Actual | Status |
|-----------|--------|--------|--------|
| Room Creation | < 50ms | 0.01ms | ✅ PASS |
| Join Room | < 50ms | 0.01ms | ✅ PASS |
| SSE Broadcast (10 players) | < 10ms | 5μs | ✅ PASS |
| Concurrent SSE (100 clients) | Support 100+ | 100 clients in 10.6ms | ✅ PASS |
| Memory per Room (16 players) | < 1MB | 17.75KB | ✅ PASS |

### Detailed Benchmark Results

#### Room Operations
```
BenchmarkRoomCreation         0.01045 ms/op
BenchmarkJoinRoom            0.01042 ms/op
```

#### SSE Broadcasting Performance
```
BenchmarkSSEBroadcast/1_clients      0 μs/broadcast
BenchmarkSSEBroadcast/5_clients      2 μs/broadcast
BenchmarkSSEBroadcast/10_clients     5 μs/broadcast
BenchmarkSSEBroadcast/20_clients     9 μs/broadcast
BenchmarkSSEBroadcast/50_clients    17 μs/broadcast
```

#### Concurrent SSE Connections
```
BenchmarkConcurrentSSEClients/10_concurrent    10 clients/iteration (10.17ms)
BenchmarkConcurrentSSEClients/50_concurrent    50 clients/iteration (10.25ms)
BenchmarkConcurrentSSEClients/100_concurrent  100 clients/iteration (10.61ms)
BenchmarkConcurrentSSEClients/200_concurrent  200 clients/iteration (11.77ms)
```

#### Memory Usage
```
BenchmarkMemoryPerRoom/2_players    2.244 KB/room (1.122 KB/player)
BenchmarkMemoryPerRoom/4_players    4.484 KB/room (1.121 KB/player)
BenchmarkMemoryPerRoom/8_players    9.045 KB/room (1.131 KB/player)
BenchmarkMemoryPerRoom/16_players  17.75 KB/room (1.109 KB/player)
```

#### Event Bus Performance
```
BenchmarkEventBusPublish       14,226 ns/publish
```

#### Store Operations
```
BenchmarkStoreOperations/CreateRoom    282.4 ns/op
BenchmarkStoreOperations/GetRoom       16.18 ns/op
BenchmarkStoreOperations/UpdateRoom    26.92 ns/op
```

## Analysis

All performance targets have been exceeded by significant margins:

1. **Room Operations**: Both room creation and joining are extremely fast at ~0.01ms, well under the 50ms target.

2. **SSE Broadcasting**: Scales linearly with the number of clients. Even with 50 clients, broadcast time is only 17μs, far below the 10ms target.

3. **Concurrent Connections**: The system handles 200 concurrent SSE connections with minimal overhead, completing operations in ~11ms.

4. **Memory Efficiency**: Memory usage is very efficient at ~1.1KB per player, meaning a full 16-player room uses less than 18KB (target was < 1MB).

## Recommendations for Future Optimization

1. **Current Performance is Excellent**: No immediate optimization needed for Phase 2.

2. **Monitoring Points**:
   - Watch SSE broadcast times as game state complexity increases
   - Monitor memory usage when game mechanics are added
   - Track event bus performance with more event types

3. **Scaling Considerations**:
   - Current architecture can easily handle 1000+ concurrent players
   - Event bus might need optimization if event frequency increases significantly
   - Consider connection pooling for very high player counts (10,000+)

## Running Benchmarks

To reproduce these benchmarks:
```bash
cd /workspace/nix/app
go test -bench=. ./internal/handlers -run=^$ -benchtime=1s
```

For quick benchmarks during development:
```bash
go test -bench=. ./internal/handlers -run=^$ -benchtime=10ms
```