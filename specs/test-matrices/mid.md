# Lynx Mid-Tier Test Matrix (concur, fluent, conf, fixture, obs)

Spec under review: `C:\Users\natha\.claude\plans\mongoose-spec-0002-framework-shape.md` §21.5, §21.6, §21.7, §21.8, §21.15, plus §23.13–§23.17 amendments. All references below use the §-numbers from that plan. "GLACIER_*" env-var names follow §22/§23.18. Build tag: `glacier_debug` (per §23.18). Module path: `github.com/nathanbrophy/glacier`.

---

## Package: `concur/`

### Test files

- `concur/mutex_test.go` :  Mutex/RWMutex correctness + ctx cancel + race.
- `concur/mutex_debug_test.go` :  `//go:build glacier_debug` :  hold-timeout slog event capture.
- `concur/mutex_bench_test.go` :  Lock/Unlock vs stdlib parity (NF1, §23.13).
- `concur/group_test.go` :  Group correctness (Go, TryGo, WaitDone, panic recovery, limit).
- `concur/group_panic_test.go` :  Go-after-WaitDone PANICS (§23.14).
- `concur/group_bench_test.go` :  Per-Go alloc budget (NF3).
- `concur/semaphore_test.go` :  Acquire/Release/TryAcquire + invalid permits.
- `concur/semaphore_watcher_test.go` :  Cancel-watcher leak guard (§23.14).
- `concur/semaphore_bench_test.go` :  Fast-path ≤ 50 ns/op zero allocs (§23.13).
- `concur/pool_test.go` :  `Pool[T]` round-trip + zero-on-empty.
- `concur/pool_bench_test.go` :  Allocation parity vs `sync.Pool`.
- `concur/once_test.go` :  Memoization, panic-doesn't-memoize, ctx pass-through-first-call-only.
- `concur/waitgroup_test.go` :  `WaitCtx` correctness.
- `concur/race_test.go` :  `//go:build race` :  combined race-detector matrix.
- `concur/lifecycle_doc_test.go` :  Documents Group has no Close (§23.16).
- `concur/example_test.go` :  Godoc examples for every primitive.

### Test matrix

| # | Name | Spec ref | Type | Description | Test helpers used |
|---|---|---|---|---|---|
| 1 | TestMutexLockUnlockBasic | §21.5 F2 | Unit (positive) | Single goroutine Lock→Unlock returns no error. | `assert.NoError`, `assert.True` |
| 2 | TestMutexLockCtxAlreadyCancelled | §21.5 E1 | Unit (negative) | LockCtx with cancelled ctx returns ErrCancelled immediately, never acquires. | `assert.ErrorIs`, `fixture.NewClock` |
| 3 | TestMutexLockCtxCancelledMidWait | §21.5 E2 | Unit (negative) | Mutex held by g1; g2 LockCtx; cancel ctx; g2 returns ErrCancelled wrapping ctx.Err. | `assert.ErrorIs`, `concur.WaitGroup` |
| 4 | TestMutexLockCtxTryLockBackoffWindow | §23.14 | Unit (concurrency lock-in) | LockCtx with quickly-cancelled ctx during backoff window verifies "best-effort cancellation" doc claim :  must observe ctx.Err within 2× configured backoff. | `fixture.NewClock`, `assert.Eventually` |
| 5 | TestMutexLockCtxNoLeakAfterCancel | §23.14 | Race + leak | Repeated cancelled LockCtx leaves no orphan goroutines. | `fixture.GuardLeaks`, `fixture.WatchGoroutines` |
| 6 | TestMutexUnlockUnheldPanics | stdlib parity | Unit (negative) | Unlock without prior Lock panics with stdlib message (production parity NF1). | `assert.Panics` (require) |
| 7 | TestMutexDebugHoldTimeoutEmitsSlog | §21.5 F1, E3 | Unit (`glacier_debug`) | Hold > timeout emits structured slog event with holder caller, waiter caller, elapsed. Lock not released by diagnostic. | `fixture.Capture`, custom slog handler sink, `assert.Contains` |
| 8 | TestMutexDebugProductionByteEquivalent | §21.5 NF1, §23.13 | Bench/Test | `unsafe.Sizeof(concur.Mutex{}) == unsafe.Sizeof(sync.Mutex{})` in non-debug builds. | `assert.Equal` |
| 9 | TestRWMutexParallelReadersExclusiveWriter | §21.5 F3 | Unit (positive) | N RLockers proceed concurrently; Lock waits until all RUnlock. | `concur.WaitGroup`, `assert.Equal` |
| 10 | TestRWMutexRLockCtxCancelled | §21.5 F3 | Unit (negative) | RLockCtx returns ErrCancelled when held exclusively. | `assert.ErrorIs` |
| 11 | TestRWMutexLockCtxCancelled | §21.5 F3 | Unit (negative) | LockCtx returns ErrCancelled when readers hold. | `assert.ErrorIs` |
| 12 | TestRWMutexUpgradePanicNotSupported | §21.5 (no upgrade) | Unit (edge) | Documenting test: Lock while RLock-held by same goroutine deadlocks (stdlib semantics). Test uses ctx-bounded wait. | `assert.ErrorIs` |
| 13 | TestGroupAllPass | §21.5 F4 | Unit (positive) | All goroutines succeed; WaitDone returns nil. | `assert.NoError` |
| 14 | TestGroupOneError | §21.5 F8 | Unit (negative) | One goroutine errors; WaitDone returns errs.Join with that single error. | `assert.ErrorIs`, `errs.Chain` |
| 15 | TestGroupAllErrorsCollected | §21.5 F8 | Unit (negative) | N goroutines all error; errs.Join collects every error (NOT first-wins). | `fluent.Count(errs.Chain(...))` (cross-pkg), `assert.Equal` |
| 16 | TestGroupCtxCancelDuringWait | §21.5 F8 | Unit (negative) | WaitDone returns ErrCancelled wrapping ctx.Err if ctx fires before goroutines. | `assert.ErrorIs`, `fixture.NewClock` |
| 17 | TestGroupPanicRecoveredAsPanicError | §21.5 F6, E7, F18 | Unit (negative) | Goroutine panics with non-error value; recovered as `*PanicError{Value:...}`; appended via errs.Join. | `assert.ErrorAs`, `errs.Chain` |
| 18 | TestGroupPanicErrorMessage | §21.5 F18 | Unit (positive) | `(&PanicError{Value:"boom"}).Error()` formats exactly `"concur: panic in group goroutine: <boom>"`. | `assert.Equal` |
| 19 | TestGroupWithLimitBlocks | §21.5 F5, E5 | Unit (positive) | WithLimit(2); 3rd Go blocks until first finishes. Verified via timing under fake clock. | `fixture.NewClock`, `assert.Eventually` |
| 20 | TestGroupWithLimitZeroReturnsOptionError | §21.5 E5 | Unit (negative, §23.14 amend) | NewGroup(WithLimit(0)) :  option validates, returns option-error per amended E5. | `assert.ErrorIs` |
| 21 | TestGroupWithLimitNegativeIsUnlimited | §21.5 E6 | Unit (positive) | WithLimit(-1) acts as default-unlimited (alias for WithUnlimited). | `assert.NoError` |
| 22 | TestGroupWithUnlimitedExplicit | §23.14 | Unit (positive) | Explicit `WithUnlimited()` overrides default `WithLimit(NumCPU*64)`. | `assert.NoError` |
| 23 | TestGroupDefaultLimitIsNumCPU64 | §23.14 | Unit (positive) | Internal default of `runtime.NumCPU() * 64` verified by introspection (or by observing back-pressure threshold). | `assert.Equal` |
| 24 | TestGroupTryGoAtCapReturnsFalse | §21.5 F7 | Unit (negative) | TryGo with limit at capacity returns false; not scheduled. | `assert.False` |
| 25 | TestGroupTryGoAfterWaitDoneReturnsFalse | §21.5 F7 | Unit (negative) | TryGo after WaitDone returns false. | `assert.False` |
| 26 | TestGroupGoAfterWaitDonePanics | §23.14 (Major B) | Unit (negative :  concurrency lock-in) | `Group.Go` after `WaitDone` PANICS with stable message `"concur: Go after WaitDone"` (matches `sync.WaitGroup.Add`-after-`Wait`). | `assert.PanicsWithMessage` |
| 27 | TestGroupGoAfterWaitDonePanicMessage | §23.14 | Unit (positive) | Asserts exact panic string for the Go-after-Wait case (locked register). | `assert.Equal` |
| 28 | TestGroupConcurrentGoFromManyGoroutines | §21.5 NF5 | Race | 1000 goroutines call Go; all errors collected; race-clean. | `concur.WaitGroup`, `fixture.GuardLeaks`, `-race` |
| 29 | PropertyGroupErrorCountEqualsGoroutineErrors | §21.5 F8 | Property | For N in [0, 200], with K errors and N-K nils, len(errs.Chain(WaitDone)) == K. | property-based generator over `fluent.Range`, `assert.Equal` |
| 30 | TestSemaphoreAcquireFastPath | §21.5 F9 | Unit (positive) | Acquire when permits available returns nil without blocking. | `assert.NoError` |
| 31 | TestSemaphoreAcquireSlowPathBlocks | §21.5 F9 | Unit (positive) | Acquire when no permits blocks until Release. | `concur.WaitGroup`, `assert.Eventually` |
| 32 | TestSemaphoreAcquireCancelled | §21.5 F10 | Unit (negative) | Acquire blocks; ctx cancels; returns ErrCancelled wrapping ctx.Err. | `assert.ErrorIs` |
| 33 | TestSemaphoreInvalidPermitsZero | §21.5 E8 | Unit (negative) | Acquire(ctx, 0) returns ErrInvalidPermits. | `assert.ErrorIs` |
| 34 | TestSemaphoreInvalidPermitsNegative | §21.5 F10 | Unit (negative) | Acquire(ctx, -1) returns ErrInvalidPermits. | `assert.ErrorIs` |
| 35 | TestSemaphoreInvalidPermitsExceedsCapacity | §21.5 E9 | Unit (negative) | Acquire(ctx, capacity+1) returns ErrInvalidPermits. | `assert.ErrorIs` |
| 36 | TestSemaphoreTryAcquireSuccess | §21.5 F11 | Unit (positive) | TryAcquire(n) with available permits returns true. | `assert.True` |
| 37 | TestSemaphoreTryAcquireFailure | §21.5 F11 | Unit (negative) | TryAcquire when not enough permits returns false; permits unchanged. | `assert.False` |
| 38 | TestSemaphoreReleaseZeroNoOp | §21.5 E10 | Unit (positive) | Release(0) is a permitted no-op; counter unchanged. | `assert.Equal` |
| 39 | TestSemaphoreOverReleasePanics | §21.5 E11 | Unit (negative) | Release(n) where total released > total acquired panics with `"concur: release: over-release"`. | `assert.PanicsWithMessage` |
| 40 | TestSemaphoreCtxWatcherNoLeak | §23.14 (Major B) | Race + leak | 10k Acquire→success cycles; verifies cancel-watcher cleanup via `defer cancel()` on per-acquire derived ctx. NO leaked goroutines. | `fixture.GuardLeaks(WatchGoroutines, WithDrainTimeout(500ms))` |
| 41 | TestSemaphoreCtxWatcherNoLeakOnCancel | §23.14 | Race + leak | 10k cancelled-Acquire cycles; cancel-watcher exits; no leaked goroutines. | `fixture.GuardLeaks` |
| 42 | TestSemaphoreManyGoroutinesAcquireRelease | §21.5 NF5 | Race | 1000 goroutines acquire/release; final counter == 0. | `-race`, `concur.WaitGroup` |
| 43 | FuzzSemaphoreAcquireRelease | §21.5 F9–F12 | Fuzz | Fuzz-driven sequence of Acquire/Release/TryAcquire with capacity in [1,32] :  invariant: counter never exceeds capacity, never goes negative, every panic is the documented over-release. | `assert.True`, `testing.F` |
| 44 | TestPoolGetPutRoundTripPreservesType | §21.5 F13, E12 | Unit (positive) | NewPool[*Buf](newFn).Get() → mutate → Put → Get returns same kind. | `assert.IsType` (typed via generics) |
| 45 | TestPoolGetEmptyNoNewReturnsZero | §21.5 E12 | Unit (negative) | NewPool[int](nil).Get() returns 0. | `assert.Equal[int]` |
| 46 | TestPoolConcurrentGetPut | §21.5 NF5 | Race | 1000 goroutines Get/Put; race-clean. | `-race` |
| 47 | TestOnceMemoizesValueAndError | §21.5 F14 | Unit (positive) | First call's `(value, error)` is memoized; subsequent calls return same. | `assert.Equal`, `assert.ErrorIs` |
| 48 | TestOnceFirstCallCtxOnly | §21.5 F14 | Unit (positive) | ctx of first call is the one threaded to fn; later callers' ctxs are ignored. | `assert.Equal` |
| 49 | TestOncePanicDoesNotMemoize | §21.5 E13 | Unit (negative) | First call panics in fn; panic propagates; Once not "completed"; second Do re-attempts and can succeed. | `assert.Panics`, `assert.NoError` |
| 50 | TestOnceConcurrentFirstCallWins | §21.5 NF5 | Race | 100 goroutines call Do simultaneously; fn invoked exactly once; all observe same memoized result. | `concur.WaitGroup`, `-race`, atomic counter |
| 51 | TestWaitGroupWaitCtxAlreadyZero | §21.5 E14 | Unit (positive) | WaitCtx returns nil immediately if counter == 0. | `assert.NoError` |
| 52 | TestWaitGroupWaitCtxCancelled | §21.5 F15 | Unit (negative) | Counter > 0; ctx cancels; returns ErrCancelled wrapping ctx.Err. | `assert.ErrorIs` |
| 53 | TestWaitGroupWaitCtxNormalCompletion | §21.5 F15 | Unit (positive) | Counter reaches zero before ctx cancel; returns nil. | `assert.NoError` |
| 54 | TestWaitGroupRaceAddDuringWait | §21.5 E15 | Race (documented) | Documents stdlib semantics :  racy Add during WaitCtx. | `-race` (expected to pass since stdlib WaitGroup is the source of behavior) |
| 55 | TestErrCancelledIsSentinelStable | §21.5 F16 | Unit (positive) | `errors.Is(wrapped, concur.ErrCancelled)` true; `wrapped == concur.ErrCancelled` false (wrapping). | `assert.ErrorIs` |
| 56 | TestErrCancelledWrapsContextCanceled | §21.5 F16 | Unit (positive) | `errors.Is(err, context.Canceled)` true through the chain. | `assert.ErrorIs` |
| 57 | TestErrCancelledWrapsDeadlineExceeded | §21.5 F16 | Unit (positive) | `errors.Is(err, context.DeadlineExceeded)` true through the chain. | `assert.ErrorIs` |
| 58 | TestErrInvalidPermitsSentinelStable | §21.5 F17 | Unit (positive) | `errors.Is(err, concur.ErrInvalidPermits)` after Acquire(ctx,0). | `assert.ErrorIs` |
| 59 | TestPanicErrorUnwrapToSynthesized | §21.5 F18 | Unit (positive) | PanicError.Unwrap returns a synthesized error reflecting Value. | `assert.NotNil`, `assert.ErrorAs` |
| 60 | TestErrFormatRegisterCompliance | §21.5 NF6, D15 | Unit (cross-cutting) | Every error string matches `^concur: [a-z]+(?:: [a-z ]+)*$` :  lowercase, no period, package prefix. | `assert.Regexp` |
| 61 | TestPackageDoesNotExposeChannels | §21.5 charter | Unit (architecture) | `go list -f '{{.Exports}}'` introspection: no exported `chan` types. (Exclusion lock-in.) | `internal/reflectx`-driven check |
| 62 | TestNoOnceErr | §21.5 D23 charter | Unit (architecture) | Verify `concur.OnceErr` does NOT exist (stdlib `sync.OnceValues` is the answer). | reflection / build-time guard |
| 63 | TestGroupHasNoCloseDocumented | §23.16 | Unit (lifecycle) | Documents that Group has no Close :  only Wait/WaitDone is the lifecycle. Compile-time check that `*Group` does NOT have `Close()`. | reflection check |
| 64 | TestConfPointerSnapshotOrdering | §23.14 (cross-link to conf) | Property | Verify the related conf atomicity invariant from concur side: an atomic.Pointer based snapshot accessor never returns torn state under contention. (Lives here because primitive-level invariant.) | `concur.Group`, `assert.Equal` |
| 65 | BenchmarkMutexLockUnlock | §21.5 NF1, §23.13 | Benchmark | Bytes/op == 0; ns/op within 5% of `sync.Mutex` baseline. Captured via paired benchmark. | `testing.B`, `benchstat` |
| 66 | BenchmarkRWMutexRLockRUnlock | §21.5 NF1 | Benchmark | Within 5% of `sync.RWMutex`. | `testing.B` |
| 67 | BenchmarkSemaphoreAcquireReleaseUncontended | §21.5 NF2, §23.13 | Benchmark | ≤ 50 ns/op, zero allocs (verified via `testing.AllocsPerRun`). | `testing.AllocsPerRun` |
| 68 | BenchmarkSemaphoreAcquireReleaseContended | §21.5 NF2 | Benchmark | Documents slow-path cost; baseline for regressions. | `testing.B` |
| 69 | BenchmarkGroupGoWithLimit | §21.5 NF3 | Benchmark | Per-Go alloc count within stated bound (1 closure + 1 recover frame). | `testing.AllocsPerRun` |
| 70 | BenchmarkGroupTryGo | §21.5 F7 | Benchmark | Tracks scheduling latency. | `testing.B` |
| 71 | BenchmarkPoolGetPut | §21.5 NF4 | Benchmark | Allocation-equivalent to `sync.Pool` (paired bench). | `testing.AllocsPerRun` |
| 72 | BenchmarkOnceDoFastPath | §21.5 F14 | Benchmark | Post-memoization Do is single atomic load. | `testing.B` |
| 73 | BenchmarkWaitGroupWaitCtxFastPath | §21.5 F15 | Benchmark | Counter==0 path is constant-time. | `testing.B` |
| 74 | TestConcurrentNoRaceCombined | §21.5 NF5 | Race (umbrella) | Single test exercising every primitive in interleaved goroutines under -race for 5s. | `fixture.GuardLeaks(WatchGoroutines)`, `-race` |

### Coverage target

- 100% line + branch on `mutex.go`, `mutex_debug.go` (under build tag), `group.go`, `semaphore.go`, `pool.go`, `once.go`, `waitgroup.go`, `panic.go`.
- 100% on every public symbol per F1–F18.
- ≥ 95% line coverage overall (some `runtime.Caller` debug paths only reachable under `glacier_debug`; covered by tag-gated tests).

### Edge cases not in spec (Lynx adds)

- **EX1**: `Group.Go` from inside `Go`'s own fn (re-entrant) :  does the panic-recovery deferred frame double-fire? Verify single recover.
- **EX2**: `Semaphore.Release(MaxInt64)` when only 1 acquired :  verify panic, not silent overflow.
- **EX3**: `Once[T].Do` where `fn` returns the zero value AND a non-nil error :  second call returns same `(zero, err)`; Once IS completed (memoization is unconditional once fn returns normally).
- **EX4**: `Pool[T]` with T = an interface :  verify zero is nil, not a typed-nil pitfall.
- **EX5**: `RWMutex.LockCtx` then `Unlock` (not RUnlock) :  accidental mismatch :  verify panic per stdlib semantics.
- **EX6**: `Group` with `WithLimit(1)` :  effectively serializes :  verify ordering not guaranteed but liveness holds.
- **EX7**: Two `Group`s sharing nothing :  verify their Semaphore-backed limits are independent.

### Special concerns

- **Debug-tag tests** for `Mutex`/`RWMutex` (`mutex_debug_test.go`) are run in a separate CI job with `-tags glacier_debug`. The hold-timeout test must not flake :  use a `fixture.NewClock`-driven internal clock injection (the spec must expose a test-only seam) OR use generous wall-clock timeouts (50 ms hold vs 10 ms threshold).
- **NF1 byte-equivalence** is tested via both `unsafe.Sizeof` AND a benchstat-driven alert: `BenchmarkMutexLockUnlock` is compared head-to-head with a sibling `BenchmarkStdlibMutex` that uses `sync.Mutex` directly; CI fails if delta > 5% (per §23.13 wording).
- **§23.16 Group lifecycle**: `Group` has NO `Close()`. The matrix includes a compile-time and reflection-time test that `*Group` does not expose `Close()` (test #63). This is documentation-as-test.
- **§23.14 atomic.Pointer indirection** for `conf.Load`: while the implementation lives in `conf/`, the *primitive correctness* property (atomic snapshot reader never sees a torn struct under load) is bench-tested in `concur/` test #64 because it's a concurrency invariant.

---

## Package: `fluent/`

### Test files

- `fluent/sources_test.go` :  F1–F10 source builders.
- `fluent/transformers_test.go` :  F11–F24 lazy Seq transformers.
- `fluent/transformers2_test.go` :  F25–F29 Seq2 transformers.
- `fluent/sort_test.go` :  F30–F33 sort family.
- `fluent/fallible_test.go` :  F34–F35 MapErr / FilterErr.
- `fluent/sinks_test.go` :  F36–F49 sinks.
- `fluent/properties_test.go` :  Property-based algebraic identities.
- `fluent/lines_fuzz_test.go` :  F7 Lines fuzz target (D31 fuzz gate).
- `fluent/words_fuzz_test.go` :  F8 Words fuzz target.
- `fluent/race_test.go` :  concurrent consumption.
- `fluent/bench_test.go` :  D35 + §23.13 perf gates.
- `fluent/example_test.go` :  Godoc examples.

### Test matrix

| # | Name | Spec ref | Type | Description | Test helpers used |
|---|---|---|---|---|---|
| 1 | TestFromSlice | §21.6 F1 | Unit (positive) | Slice → Seq → ToSlice round-trip preserves order and contents. | `assert.Equal` |
| 2 | TestFromNil | §21.6 E1 | Unit (edge) | `From(nil)` yields nothing. | `fluent.Count`, `assert.Equal` |
| 3 | TestFromMapKeysComplete | §21.6 F2 | Unit (positive) | FromMap yields every (k,v) exactly once (order undefined). | `fluent.ToMap`, `assert.Equal` |
| 4 | TestFromChanDrainsUntilClose | §21.6 F3, E2 | Unit (positive) | FromChan over closed channel yields buffered then ends; over open never-closed channel blocks (test uses ctx-bounded buffered). | `assert.Equal`, `concur.WaitGroup` |
| 5 | TestRangePositiveStep | §21.6 F4 | Unit (positive) | Range(0,5,1) yields 0,1,2,3,4. | `assert.Equal` |
| 6 | TestRangeNegativeStep | §21.6 F4 | Unit (positive) | Range(5,0,-1) yields 5,4,3,2,1. | `assert.Equal` |
| 7 | TestRangeEmptyEqualEnds | §21.6 E3 | Unit (edge) | Range(0,0,1) yields nothing. | `fluent.Count`, `assert.Equal` |
| 8 | TestRangeStepDirectionMismatch | §21.6 E4 | Unit (edge) | Range(0,5,-1) yields nothing. | `assert.Equal` |
| 9 | TestRangeZeroStepPanics | §21.6 E5 | Unit (negative) | Range(0,5,0) panics with `"fluent: Range: step must be non-zero"`. | `assert.PanicsWithMessage` |
| 10 | TestRepeatNZero | §21.6 F5 | Unit (edge) | Repeat(v,0) yields nothing. | `fluent.Count` |
| 11 | TestRepeatN | §21.6 F5 | Unit (positive) | Repeat(v,5) yields v exactly 5 times. | `assert.Equal` |
| 12 | TestGenerateStops | §21.6 F6, E15 | Unit (positive) | Generate fn returning (v,true) k times then (z,false) yields k items. | `fluent.Count`, `assert.Equal` |
| 13 | TestLinesNoTrailingNewline | §21.6 F7 | Unit (positive) | Lines yields strings without trailing `\n`. | `assert.Equal` |
| 14 | TestLinesEmpty | §21.6 F7 | Unit (edge) | Lines over empty reader yields nothing. | `fluent.Count` |
| 15 | TestLinesCRLF | §21.6 F7 | Unit (edge) | Lines handles `\r\n` line endings (drops both). | `assert.Equal` |
| 16 | TestLinesGiantLine | §21.6 F7 untrusted-input | Unit (negative) | Line > internal bufio cap returns scanner error per library register. | `assert.ErrorContains` |
| 17 | FuzzLines | §21.6 F7 (D31 fuzz gate) | Fuzz | Random byte streams over Lines: never panics; total bytes consumed ≤ input bytes; no UTF-8 corruption beyond input. | `testing.F` |
| 18 | TestWordsWhitespaceSeparation | §21.6 F8 | Unit (positive) | Words yields tokens split on Unicode whitespace. | `assert.Equal` |
| 19 | TestWordsEmpty | §21.6 F8 | Unit (edge) | Words over empty input yields nothing. | `fluent.Count` |
| 20 | FuzzWords | §21.6 F8 | Fuzz | Random byte streams: never panics; output count ≤ input length. | `testing.F` |
| 21 | TestSplitOnSep | §21.6 F9 | Unit (positive) | Split("a,b,c", ",") yields a,b,c. | `assert.Equal` |
| 22 | TestSplitNoMatch | §21.6 F9 | Unit (edge) | Split("abc", ",") yields ["abc"]. | `assert.Equal` |
| 23 | TestSplitEmptyString | §21.6 F9 | Unit (edge) | Split("", ",") yields one empty element (or none :  locked behavior to be in matrix). | `assert.Equal` |
| 24 | TestSplitEmptySeparatorPanics | §21.6 F9 | Unit (negative) | Split(s, "") panics with library register message. | `assert.Panics` |
| 25 | TestPairsToSeq2RoundTrip | §21.6 F10 | Unit (positive) | Pairs(...) then Entries returns input. | `assert.Equal` |
| 26 | TestMapBasic | §21.6 F11 | Unit (positive) | Map(From([1,2,3]), x*2) yields 2,4,6. | `assert.Equal` |
| 27 | TestMapZeroAllocPipeline | §21.6 NF2, §23.13 | Bench/Test | `Reduce(Map(Filter(From(s)), pred), 0, sum)` allocations per element == 0 for non-capturing fns in pre-composed pipelines. | `testing.AllocsPerRun` |
| 28 | TestFilterBasic | §21.6 F12 | Unit (positive) | Filter(From([1,2,3,4]), even) yields 2,4. | `assert.Equal` |
| 29 | TestFilterAllExclude | §21.6 F12 | Unit (edge) | Predicate always false → empty. | `fluent.Count` |
| 30 | TestTakeBasic | §21.6 F13 | Unit (positive) | Take(seq, 2) yields first 2. | `assert.Equal` |
| 31 | TestTakeNegative | §21.6 E6 | Unit (edge) | Take(seq, -1) yields nothing. | `fluent.Count` |
| 32 | TestTakeMoreThanAvailable | §21.6 F13 | Unit (edge) | Take(seq, 100) over 3-element seq yields 3. | `assert.Equal` |
| 33 | TestDropBasic | §21.6 F14 | Unit (positive) | Drop(seq, 2) yields all but first 2. | `assert.Equal` |
| 34 | TestDropMoreThanLen | §21.6 E7 | Unit (edge) | Drop more than length → empty. | `fluent.Count` |
| 35 | TestWindowBasic | §21.6 F15 | Unit (positive) | Window([1,2,3,4], 3) yields [1,2,3], [2,3,4]. | `assert.Equal` |
| 36 | TestWindowSizeZeroPanics | §21.6 E8 | Unit (negative) | Window(_, 0) panics with `"fluent: Window: size must be positive"`. | `assert.PanicsWithMessage` |
| 37 | TestWindowSeqShorterThanSize | §21.6 E9 | Unit (edge) | Window([1,2], 3) yields nothing. | `fluent.Count` |
| 38 | TestChunkBasic | §21.6 F16 | Unit (positive) | Chunk([1..6], 2) yields [1,2],[3,4],[5,6]. | `assert.Equal` |
| 39 | TestChunkSizeZeroPanics | §21.6 E8 | Unit (negative) | Chunk(_, 0) panics. | `assert.Panics` |
| 40 | TestChunkPartialLast | §21.6 E10 | Unit (edge) | Chunk([1..5], 2) yields [1,2],[3,4],[5]. | `assert.Equal` |
| 41 | TestDistinctComparable | §21.6 F17 | Unit (positive) | Distinct preserves first occurrence; dedups. | `assert.Equal` |
| 42 | TestDistinctEmpty | §21.6 F17 | Unit (edge) | Distinct over empty yields nothing. | `fluent.Count` |
| 43 | TestZipUnequalLengths | §21.6 E11, F18 | Unit (positive) | Zip yields min(len(a), len(b)). | `assert.Equal` |
| 44 | TestGroupByBasic | §21.6 F19 | Unit (positive) | GroupBy partitions by key fn. | `assert.Equal` |
| 45 | TestGroupByEmpty | §21.6 E12 | Unit (edge) | GroupBy over empty yields no groups. | `fluent.Count` |
| 46 | TestJoinInner | §21.6 F20 | Unit (positive) | Join yields only matching pairs; b materialized once. | `assert.Equal` |
| 47 | TestJoinNoMatches | §21.6 F20 | Unit (edge) | Disjoint keys → no pairs. | `fluent.Count` |
| 48 | TestLeftJoinIncludesUnmatched | §21.6 F21 | Unit (positive) | LeftJoin: every a yielded; b is zero on no-match. | `assert.Equal` |
| 49 | TestUnionDistinct | §21.6 F22 | Unit (positive) | Union(a,b) yields distinct elements from a, then b. | `assert.Equal` |
| 50 | TestIntersect | §21.6 F23 | Unit (positive) | Intersect distinct elements present in both. | `assert.Equal` |
| 51 | TestExcept | §21.6 F24 | Unit (positive) | Except: distinct in a not in b. | `assert.Equal` |
| 52 | TestMap2 | §21.6 F25 | Unit (positive) | Map2 transforms (k,v) → (k2,v2). | `assert.Equal` |
| 53 | TestFilter2 | §21.6 F26 | Unit (positive) | Filter2 retains pairs satisfying pred. | `assert.Equal` |
| 54 | TestKeysOf | §21.6 F27 | Unit (positive) | KeysOf yields keys in order. | `assert.Equal` |
| 55 | TestValuesOf | §21.6 F28 | Unit (positive) | ValuesOf yields values in order. | `assert.Equal` |
| 56 | TestEntries | §21.6 F29 | Unit (positive) | Entries yields KV{k,v}. | `assert.Equal` |
| 57 | TestPairsEntriesRoundTrip | §21.6 F50 | Unit (property-lite) | `Entries(Pairs(seq))` == `seq`. | `assert.Equal` |
| 58 | TestSortAscending | §21.6 F30 | Unit (positive) | Sort yields ascending order via slices.Sort. | `assert.Equal` |
| 59 | TestSortByKey | §21.6 F31 | Unit (positive) | SortBy(seq, keyFn) ordered by key. | `assert.Equal` |
| 60 | TestSortStableLess | §21.6 F32 | Unit (positive) | Stable sort preserves relative order of equal-key elements. | `assert.Equal` |
| 61 | TestSortDescending | §21.6 F33 | Unit (positive) | SortDesc reverse order. | `assert.Equal` |
| 62 | TestSortDoesNotMutateInput | §21.6 E13 | Unit (positive) | Original slice unchanged after Sort(From(s)). | `assert.Equal` |
| 63 | TestMapErrAllSuccess | §21.6 F34 | Unit (positive) | All (v,nil); caller gets every value, no errors. | `assert.NoError` (per yield) |
| 64 | TestMapErrShortCircuit | §21.6 F34 | Unit (positive) | Caller breaks on first non-nil error; iteration stops. | `assert.ErrorIs` |
| 65 | TestMapErrContinuePastErrors | §21.6 E14 | Unit (positive) | Caller continues; gets every (zero,err) pair. | `assert.Equal`, `fluent.Count` |
| 66 | TestFilterErrShortCircuit | §21.6 F35 | Unit (positive) | Same shape as MapErr. | `assert.ErrorIs` |
| 67 | TestReduceBasic | §21.6 F36 | Unit (positive) | Reduce sums correctly. | `assert.Equal` |
| 68 | TestReduceEmpty | §21.6 F36 | Unit (edge) | Reduce over empty returns zero. | `assert.Equal` |
| 69 | TestToSlice | §21.6 F37 | Unit (positive) | ToSlice round-trips From. | `assert.Equal` |
| 70 | TestToMap | §21.6 F38 | Unit (positive) | ToMap from Seq2 produces map. | `assert.Equal` |
| 71 | TestToMapDuplicateKeyLastWins | §21.6 F38 | Unit (edge) | Documented behavior on duplicate keys. | `assert.Equal` |
| 72 | TestCount | §21.6 F39 | Unit (positive) | Count == ToSlice length. | `assert.Equal` |
| 73 | TestFirstNonEmpty | §21.6 F40 | Unit (positive) | First yields first elem, true. | `assert.Equal`, `assert.True` |
| 74 | TestFirstEmpty | §21.6 E16 | Unit (edge) | First on empty → (zero, false). | `assert.False` |
| 75 | TestLastNonEmpty | §21.6 F41 | Unit (positive) | Last yields final elem. | `assert.Equal` |
| 76 | TestLastEmpty | §21.6 E16 | Unit (edge) | Last on empty → (zero, false). | `assert.False` |
| 77 | TestAnyTrue | §21.6 F42 | Unit (positive) | Any short-circuits on first match. | `assert.True` |
| 78 | TestAllTrue | §21.6 F43 | Unit (positive) | All true when every element matches. | `assert.True` |
| 79 | TestAllFalseShortCircuits | §21.6 F43 | Unit (positive) | All short-circuits on first false. | `assert.False` |
| 80 | TestSumIntegers | §21.6 F44 | Unit (positive) | Sum over `[]int` returns int total. | `assert.Equal` |
| 81 | TestSumEmpty | §21.6 E17 | Unit (edge) | Sum over empty returns zero T. | `assert.Equal` |
| 82 | TestAvgFloat | §21.6 F45 | Unit (positive) | Avg returns float64 mean. | `assert.InDelta` |
| 83 | TestAvgEmptyNaN | §21.6 E17 | Unit (edge) | Avg over empty returns NaN. | `assert.True(math.IsNaN(...))` |
| 84 | TestAvgIntegerOverflowAvoided | §21.6 E18 | Unit (positive) | Sum(int64s with totals exceeding int64) :  internal accumulator is float64; result correct within precision. | `assert.InDelta` |
| 85 | TestMin | §21.6 F46 | Unit (positive) | Min returns smallest. | `assert.Equal` |
| 86 | TestMinEmpty | §21.6 E16 | Unit (edge) | Min over empty → (zero, false). | `assert.False` |
| 87 | TestMax | §21.6 F47 | Unit (positive) | Max returns largest. | `assert.Equal` |
| 88 | TestMinByKey | §21.6 F48 | Unit (positive) | MinBy returns elem with smallest key. | `assert.Equal` |
| 89 | TestMaxByKey | §21.6 F49 | Unit (positive) | MaxBy returns elem with largest key. | `assert.Equal` |
| 90 | TestKVStructFields | §21.6 F50 | Unit (positive) | KV[K,V]{K, V} round-trips. | `assert.Equal` |
| 91 | TestNumberConstraintAccepted | §21.6 F51 | Unit (compile-time) | All numeric kinds satisfy Number; non-numeric does not (negative-compile test via build constraint). | compile-only |
| 92 | TestLazyNoSideEffectsBeforeIteration | §21.6 NF1 | Unit (positive) | Build pipeline; never iterate; assert side-effecting predicate's call count == 0. | `assert.Equal` (counter) |
| 93 | TestSortMaterializesOnce | §21.6 NF3 | Unit (positive) | Sort consumes seq exactly once (count source pulls). | `assert.Equal` |
| 94 | TestDistinctAllocatesHashSet | §21.6 NF4 | Unit (positive, alloc-aware) | Distinct allocates O(unique) :  bench captures number. | `testing.AllocsPerRun` |
| 95 | TestSortPanicsOnNonOrdered | §21.6 E19 | Unit (compile-time) | Sort over `[]struct{}` is a compile error (constraint violation). | compile-only |
| 96 | PropertyMapFilterDistributivity | §21.6 algebra (req #5) | Property | `fluent.Map(Filter(s, p), f)` ≡ `Filter(Map(s, f), q)` where q(y) = ∃x. f(x)=y ∧ p(x). For invertible f only; randomized inputs. | `assert.Equal`, randomized generator |
| 97 | PropertyMapFilterReducibleEqual | §21.6 algebra | Property | `Reduce(Map(s, f), zero, ⊕)` produces consistent result across pipeline orderings when ⊕ is associative. | `assert.Equal` |
| 98 | PropertyReduceInvertibleRoundTrip | §21.6 algebra (req #5) | Property | `Reduce(seq, zero, f)` followed by reverse op reconstructs sequence (for invertible f, e.g., `Reduce(seq, "", concat)` → `Split(_, sep)`). | `assert.Equal` |
| 99 | TestConcurrentConsumptionDistinctSeqs | §21.6 NF5 (race req) | Race | Each of 100 goroutines consumes its own freshly-built seq; race-clean. | `-race`, `concur.WaitGroup`, `fixture.GuardLeaks` |
| 100 | TestSharedSeqConsumedTwicePanicsOrIsSafe | §21.6 NF1 | Unit (documented behavior) | Documents whether `iter.Seq` consumed twice is safe (depends on source; chan source: empty 2nd time; slice source: full again). | `assert.Equal` |
| 101 | BenchmarkMapFilterReduce | §21.6 NF2, §23.13 | Benchmark | Per-element zero allocs verified for non-capturing functions in pre-composed pipelines. | `testing.AllocsPerRun` |
| 102 | BenchmarkSortFamily | §21.6 NF3 | Benchmark | Sort/SortBy/SortStable/SortDesc tracked vs `slices.Sort`. | `testing.B` |
| 103 | BenchmarkDistinctLargeCardinality | §21.6 NF4 | Benchmark | Allocation profile for hash-set-backed Distinct. | `testing.B` |
| 104 | BenchmarkLinesLargeFile | §21.6 F7 | Benchmark | Lines streaming from 1 MiB reader. | `testing.B` |
| 105 | BenchmarkJoinMaterialization | §21.6 E20 | Benchmark | Documents the `map[K][]B` materialization cost for Join. | `testing.B` |

### Coverage target

- 100% line on every public symbol F1–F51.
- ≥ 95% branch coverage on transformer chains; the rare error path of `Lines` (io.Reader returning non-EOF error) is fuzz-covered.

### Edge cases not in spec

- **EX1**: `Map(seq, f)` where `f` panics :  does the panic propagate at iteration time? (Document: yes, no recovery.)
- **EX2**: `Window` with overlapping windows shares backing slice or copies? (Lock to "copies" :  test).
- **EX3**: `Chunk` on infinite Generate seq with Take :  verify lazy composition produces finite result.
- **EX4**: `Distinct` over NaN floats :  Go map equality of NaN :  document (NaN != NaN, so NaNs all yield).
- **EX5**: `Zip` where one source panics mid-iteration :  Zip terminates; partner side-effect-free.
- **EX6**: `MapErr` where iteration count > input due to caller `continue` after error :  verify off-by-one.
- **EX7**: `ToMap` on Seq2 with zero elements :  returns empty non-nil map.

### Special concerns

- **§23.13 zero-alloc qualifier** ("for non-capturing functions in pre-composed pipelines") :  `BenchmarkMapFilterReduce` MUST use top-level functions, never closures over loop variables. A second benchmark `BenchmarkMapFilterReduceCapturing` documents the (allowed) closure-allocation cost.
- **No internal mongoose deps** (§21.6 NF8) :  verified architecturally by `import "fluent"; want_no_imports glacier/{option,errs,log,assert}`. Lint-style test in CI.
- **Property-based tests** use `pgregory.net/rapid` would normally be the choice, but per D4 zero deps :  implement small in-tree property generator (table-driven over enumerated input shapes).

---

## Package: `conf/`

### Test files

- `conf/register_test.go` :  F1, E1, E2.
- `conf/load_test.go` :  F2, F4, atomic load, precedence.
- `conf/decode_test.go` :  F3, E15.
- `conf/sources_file_test.go` :  F6, E5–E7, E17, E18.
- `conf/sources_env_test.go` :  F7, E8–E10.
- `conf/sources_flag_test.go` :  F8 (with mock FlagSource).
- `conf/sources_set_test.go` :  F9, E16.
- `conf/sources_defaults_test.go` :  F10.
- `conf/types_test.go` :  F12–F17 mapping.
- `conf/errors_test.go` :  F21–F23.
- `conf/atomicity_test.go` :  §23.14 atomic.Pointer indirection.
- `conf/concurrent_test.go` :  F4 NF4, E13, E14, race.
- `conf/lifecycle_test.go` :  §23.16 Loader.Close idempotency.
- `conf/load_fuzz_test.go` :  D31 + req #3 fuzz target on JSON parser (`internal/safejson`).
- `conf/withset_fuzz_test.go` :  req #3 fuzz on WithSet path coercion.
- `conf/properties_test.go` :  Load-Load idempotence.
- `conf/path_safety_test.go` :  Falcon §1.12 path traversal/symlink/UNC.
- `conf/bench_test.go` :  D35 + §23.13.
- `conf/example_test.go` :  Godoc examples.

### Test matrix

| # | Name | Spec ref | Type | Description | Test helpers used |
|---|---|---|---|---|---|
| 1 | TestRegisterReturnsTypedAccessor | §21.7 F1, §23.14 | Unit (positive) | `Register[T](path, defaults)` returns `func() *T` snapshot accessor (changed from stable pointer per §23.14). | `assert.NotNil`, generic-typed assertions |
| 2 | TestRegisterDuplicatePanics | §21.7 E1 | Unit (negative) | Second Register at same path panics with `"conf: register: path \"server\" already registered"`. | `assert.PanicsWithMessage` |
| 3 | TestRegisterEmptyPathIsRoot | §21.7 E2 | Unit (positive) | Register("", T) treats as root config. | `assert.Equal` |
| 4 | TestLoadDefaultsOnly | §21.7 E3 | Unit (positive) | No options → registered structs reflect defaults. | `assert.Equal` |
| 5 | TestLoadFile | §21.7 F6 | Unit (positive) | WithFile populates from JSON. | `fixture.NewFS` (or `t.TempDir`+JSON), `assert.Equal` |
| 6 | TestLoadFilePrecedenceOverDefaults | §21.7 F5 | Unit (positive) | File value beats defaults. | `assert.Equal` |
| 7 | TestLoadEnvPrecedenceOverFile | §21.7 F5 | Unit (positive) | Env wins over file. | `t.Setenv`, `assert.Equal` |
| 8 | TestLoadFlagPrecedenceOverEnv | §21.7 F5, F8 | Unit (positive) | Mock FlagSource wins over env. | `mock.Of[FlagSource]`, `assert.Equal` |
| 9 | TestLoadSetPrecedenceOverFlag | §21.7 F5, F9 | Unit (positive) | WithSet wins over all. | `assert.Equal` |
| 10 | TestEnvDoubleUnderscoreSeparator | §21.7 F7 | Unit (positive) | `APP__SERVER__PORT` → `server.port`. | `t.Setenv` |
| 11 | TestEnvDerivesFromJSONTag | §21.7 F12 | Unit (positive) | Field with `json:"max_conns"` reads `APP__DB__MAX_CONNS`. | `t.Setenv` |
| 12 | TestEnvFallsBackToFieldName | §21.7 F12 | Unit (positive) | No json tag → upper-snake of Go field name. | `t.Setenv` |
| 13 | TestNestedStructsViaPath | §21.7 F13 | Unit (positive) | `server.tls.cert_file` works for both file and env. | `t.Setenv`, `assert.Equal` |
| 14 | TestPointerFieldOptionalNil | §21.7 F14, E11 | Unit (positive) | `*Foo` absent → nil. | `assert.Nil` |
| 15 | TestPointerFieldOptionalAllocated | §21.7 F14, E12 | Unit (positive) | Any source mentions sub-key → ptr allocated and populated. | `assert.NotNil` |
| 16 | TestSliceFromJSONArray | §21.7 F15, E10 | Unit (positive) | JSON `["a","b"]` → `[]string{"a","b"}`. | `assert.Equal` |
| 17 | TestSliceFromEnvCommaSplit | §21.7 F15 | Unit (positive) | `APP__SERVERS=a,b,c` → `[]string{"a","b","c"}`. | `assert.Equal` |
| 18 | TestSliceFromEnvCustomSep | §21.7 F15 | Unit (positive) | `WithEnvSliceSep(":")` splits on `:`. | `assert.Equal` |
| 19 | TestMapFromJSON | §21.7 F16 | Unit (positive) | JSON object → `map[string]V`. | `assert.Equal` |
| 20 | TestMapFromEnvUnsupported | §21.7 F16 | Unit (negative) | env vars do not populate maps; documented; no-op without error if no other source. | `assert.Equal` |
| 21 | TestDurationFromEnv | §21.7 F17, E9 | Unit (positive) | `APP__TTL=30s` → 30 * time.Second. | `t.Setenv`, `assert.Equal` |
| 22 | TestTimeRFC3339FromEnv | §21.7 F17 | Unit (positive) | RFC3339 string parses. | `assert.Equal` |
| 23 | TestNumericParseError | §21.7 E8 | Unit (negative) | `APP__PORT=not-a-number` → DecodeError{Path:"server.port", Cause: strconv.ErrSyntax}. | `assert.ErrorAs`, `assert.ErrorIs` |
| 24 | TestBoolFromEnv | §21.7 F17 | Unit (positive) | `true`/`false`/`1`/`0` parse. | `assert.Equal` |
| 25 | TestUnknownJSONFieldRejected | §21.7 E5, F18 | Unit (negative) | DisallowUnknownFields → DecodeError. | `assert.ErrorAs` |
| 26 | TestFileTooLarge | §21.7 E6, F18, NF8 | Unit (negative) | File > 1 MiB → DecodeError{Cause: ErrFileTooLarge}. | `assert.ErrorIs` |
| 27 | TestDepthExceeded | §21.7 E7, F18, NF8 | Unit (negative) | JSON depth > 32 → DecodeError{Cause: ErrDepthExceeded}. | `assert.ErrorIs` |
| 28 | TestPathTraversalRejected | §21.7 E17, NF8 | Unit (negative :  Falcon §1.12) | File path containing `..` rejected. | `assert.ErrorContains` |
| 29 | TestSymlinkRejected | §21.7 E18, NF8 | Unit (negative) | Symlinked file rejected via Lstat. | `os.Symlink` setup, `assert.ErrorContains` |
| 30 | TestWindowsUNCPathRejected | §21.7 NF8 | Unit (negative :  Windows-only) | `\\server\share\` rejected. | build-tag windows test |
| 31 | TestNonRegularFileRejected | §21.7 NF8 | Unit (negative) | Named pipe / device rejected. | platform-conditional |
| 32 | TestFileMissing | §21.7 E4 | Unit (negative) | File-with-WithFile absent → DecodeError{Cause: fs.ErrNotExist}. | `assert.ErrorIs` |
| 33 | TestAtomicLoadFailureNoMutation | §21.7 NF3, E E13/E14 | Unit (positive) | Failure mid-Load: NO registered struct mutated. | `assert.Equal`, snapshot before/after |
| 34 | TestAtomicLoadAtomicPointerNoTorn | §23.14 (Major B) | Race | Concurrent readers via the snapshot accessor during Load :  no torn struct ever observed; readers see either pre-Load or post-Load entire struct. | `concur.Group`, `-race`, `assert.Equal` |
| 35 | TestConcurrentLoadSerializes | §21.7 E13, NF4 | Race | Two concurrent Loads: second blocks until first completes. | `concur.WaitGroup`, timing harness |
| 36 | TestConcurrentReadsDuringLoadRaceClean | §21.7 E14, NF4 | Race | 100 readers via snapshot accessor while 10 Loads cycle; -race clean. | `-race`, `fixture.GuardLeaks(WatchGoroutines)` |
| 37 | TestReLoadInPlace | §21.7 F4 | Unit (positive) | Second Load replaces contents; staged-and-replace. | `assert.Equal` |
| 38 | TestDecodeOneShot | §21.7 F3 | Unit (positive) | Decode[T] without registry. | `assert.Equal` |
| 39 | TestDecodeNoExportedFields | §21.7 E15 | Unit (edge) | Decode[T] where T has no exported fields → zero T, no error. | `assert.NoError` |
| 40 | TestMustLoadPanicsOnError | §21.7 F20 | Unit (negative) | MustLoad panics with library register on file error. | `assert.Panics` |
| 41 | TestMustLoadSucceedsNoOp | §21.7 F20 | Unit (positive) | MustLoad with valid sources does not panic. | `assert.NoError` (paired) |
| 42 | TestWithSetTypeMismatch | §21.7 E16 | Unit (negative) | WithSet("server.port", "string") for int field → DecodeError. | `assert.ErrorAs` |
| 43 | TestWithSetUnknownPath | §21.7 F9 | Unit (negative) | WithSet on unregistered path → DecodeError. | `assert.ErrorAs` |
| 44 | TestWithDefaultsLayer | §21.7 F10 | Unit (positive) | WithDefaults applies between defaults and file layer. | `assert.Equal` |
| 45 | TestWithLoggerLogsPerLayer | §21.7 F11 | Unit (positive) | Each layer applied logged at debug. | `fixture.Capture` + custom slog handler, `assert.Contains` |
| 46 | TestDecodeErrorMessageFormat | §21.7 F21 | Unit (positive) | `(*DecodeError)(...).Error()` formats as `"conf: decode <path>: <cause>"`. | `assert.Equal` |
| 47 | TestDecodeErrorUnwrap | §21.7 F21 | Unit (positive) | `errors.Is(de, de.Cause)` true. | `assert.ErrorIs` |
| 48 | TestErrLayerConflictSentinel | §21.7 F22 | Unit (positive) | Cross-layer type conflict surfaces ErrLayerConflict. | `assert.ErrorIs` |
| 49 | TestErrFileTooLargeSentinel | §21.7 F23 | Unit (positive) | Sentinel stable. | `assert.ErrorIs` |
| 50 | TestErrDepthExceededSentinel | §21.7 F23 | Unit (positive) | Sentinel stable. | `assert.ErrorIs` |
| 51 | TestErrFormatRegisterCompliance | §21.7 NF5, D15 | Unit (cross-cutting) | Every conf-emitted error matches library register regex. | `assert.Regexp` |
| 52 | TestValidateAfterLoad | §21.7 F19 | Unit (integration) | After Load, `option.Validate(*Cfg, ...)` enforces semantic invariants. | `option.Validate`, `assert.NoError` / `assert.ErrorIs` |
| 53 | TestOptionRequiredTLoadBearing | §23.17 | Unit (generics) | `option.Required[T any](name, getter func(*T) any)` :  T threaded through; compile error on T mismatch. | compile-only sample + runtime check |
| 54 | TestLoaderCloseIdempotent | §23.16 | Unit (lifecycle) | `Loader.Close()` idempotent: 2nd call returns nil; releases file watchers (when watch lands; v0 returns nil). | `assert.NoError` (paired) |
| 55 | TestLoaderCloseJoinsErrors | §23.16 | Unit (lifecycle) | Close errors from multiple resources joined via `errs.Join`. | `errs.Chain`, `fluent.Count` |
| 56 | TestMultipleRegistrationsAtomic | §21.7 NF3 | Unit (positive) | 50 registrations; one Load fails mid-way; NO struct mutated. | `assert.Equal` |
| 57 | TestRegisterAfterLoadCoveredByNextLoad | §21.7 F1 | Unit (positive) | Register post-Load is covered by next Load. | `assert.Equal` |
| 58 | FuzzLoadJSON | §21.7 F18 (req #3, D31) | Fuzz | JSON parser (the `internal/safejson` wrapper): random bytes → never panics; size-cap and depth-cap enforced; UTF-8 validation enforced. | `testing.F` |
| 59 | FuzzWithSetPathCoercion | req #3 | Fuzz | Attacker-shaped paths (`server..port`, `server.port.`, `..server`, with control chars) into WithSet :  never panics; always returns DecodeError or applies cleanly. | `testing.F` |
| 60 | PropertyLoadIdempotent | §21.7 (req #5) | Property | `Load` of identical sources yields identical struct values across runs (idempotence). Generates random configs. | `assert.Equal` |
| 61 | PropertyLoadLoadIdempotent | §21.7 F4 (req #5) | Property | Two consecutive Load calls with same sources produce identical results. | `assert.Equal` |
| 62 | PropertyPrecedenceMonotonic | §21.7 F5 | Property | Adding a higher-precedence source never reduces config completeness. | `assert.True` |
| 63 | TestEnvPrefixAllowlist | §21.7 NF8 (Falcon) | Unit (negative) | Without `WithEnvPrefix`, no env vars are read. With prefix `APP`, only `APP__*` are read; `MONGOOSE_*` (legacy) never auto-included. | `t.Setenv`, `assert.Equal` |
| 64 | TestNoIncludeDirectivesAccepted | §21.7 NF8 (Falcon) | Unit (negative) | JSON containing `$include` keys treated as ordinary unknown field → DecodeError per DisallowUnknownFields. | `assert.ErrorAs` |
| 65 | TestNoCommandStringFields | §21.7 NF8 (Falcon) | Unit (architectural) | Documentation/lint-test: type-driven, no field types coerce a raw command string. | `internal/reflectx` walk |
| 66 | BenchmarkLoadOneRegistration | §21.7 NF1 | Benchmark | Single registration, single source. | `testing.B` |
| 67 | BenchmarkLoadFiftyRegistrations | §21.7 NF1, req #2 | Benchmark | 50 registrations × file source :  D35 regression gate. | `testing.B`, benchstat |
| 68 | BenchmarkDecode | §21.7 NF1 | Benchmark | One-shot Decode[T] for typical 5-field struct. | `testing.B` |
| 69 | BenchmarkReflectionCacheHot | §21.7 NF1 | Benchmark | Verifies `internal/reflectx` cache: 2nd Load is faster than 1st. | `testing.B` |

### Coverage target

- 100% line on every public symbol per F1–F23.
- 100% on `errors.go` sentinel paths.
- ≥ 95% line on `load.go` (atomic commit path); ≥ 95% on `file.go` (size/depth/symlink rejection).

### Edge cases not in spec

- **EX1**: JSON file with BOM :  accepted or rejected? (Lock to: stripped silently if UTF-8 BOM; rejected for UTF-16 BOM.)
- **EX2**: Concurrent `Register` from multiple goroutines (NF4 says it's safe but expected init-only) :  verify map-write-during-init is mutex-guarded.
- **EX3**: Env var `APP__SERVER__PORT=` (empty value) :  does it count as "set" and override file? Lock to: yes, set.
- **EX4**: Field with both `json:"name,omitempty"` tag :  does omitempty affect Load (it should not; only affects Marshal). Verify.
- **EX5**: Path `server.tls.cert_file` in WithSet with leading/trailing dots :  rejected.
- **EX6**: File whose root JSON is a number/array (not object) :  rejected with DecodeError.
- **EX7**: `MustLoad` from goroutine :  panic propagation behavior documented (panics in goroutine kill program; same as user code).

### Special concerns

- **§23.14 atomic.Pointer[T] indirection**: the F1 amendment changes Register to return a `func() *T` accessor backed by `atomic.Pointer[*T]` swap on Load commit. Test #34 is the lock-in test verifying NO torn reads under contention. The matrix asserts the example code in spec was updated (matches §23.14 wording).
- **Falcon §1.12 path safety**: tests #28–#31 are the security-review gate. Each MUST run on every CI; `fixture.GuardLeaks(WatchTempDirs)` ensures test artifacts (symlinks) are cleaned up.
- **§21.7 NF8 untrusted-input register row**: the JSON parser wrapper is `internal/safejson` :  fuzz target #58 IS the D31 fuzz gate. Must be in `corpus/` with seed inputs covering deeply nested arrays, escaped Unicode, repeated keys, malformed UTF-8.
- **§23.16 Loader.Close**: even though v0 doesn't watch files, Close is implemented as idempotent no-op-or-cleanup. Tests #54, #55 lock this in.

---

## Package: `fixture/`

### Test files

- `fixture/golden_test.go` :  F1, F5, E1–E3.
- `fixture/snapshot_test.go` :  F2, E4, E16.
- `fixture/load_test.go` :  F3, F4, E5–E7.
- `fixture/clock_test.go` :  F7–F10, E8, E9.
- `fixture/mockfs_test.go` :  F11, E10.
- `fixture/capture_test.go` :  F12, E11.
- `fixture/guardleaks_test.go` :  F13–F20, E12–E15.
- `fixture/path_safety_test.go` :  Falcon §1.13.
- `fixture/concurrency_test.go` :  Capture process-wide lock + NewClock-while-timers-fire.
- `fixture/properties_test.go` :  Snapshot round-trip property.
- `fixture/lifecycle_test.go` :  §23.16 NewClock cleanup + GuardLeaks baseline.
- `fixture/bench_test.go` :  D35 + §23.13.
- `fixture/example_test.go` :  Godoc examples.
- `fixture/testdata/...` :  golden + snapshot fixtures.

### Test matrix

| # | Name | Spec ref | Type | Description | Test helpers used |
|---|---|---|---|---|---|
| 1 | TestGoldenCreateOnMissing | §21.8 E1 | Unit (positive, env-driven) | `GLACIER_GOLDEN_UPDATE=1` + missing file → file created with bytes; returns true. | `t.Setenv("GLACIER_GOLDEN_UPDATE","1")`, `assert.True` |
| 2 | TestGoldenMissingNoUpdateErrors | §21.8 E2 | Unit (negative) | Missing file + no env → t.Errorf with hint message; returns false. | mock TB recorder, `assert.Contains` |
| 3 | TestGoldenMatchPassesSilently | §21.8 F1 | Unit (positive) | Matching bytes → no error, returns true. | mock TB recorder |
| 4 | TestGoldenMismatchReportsDiff | §21.8 E3 | Unit (negative) | Mismatch → t.Errorf with line-by-line diff for textual; hex header for binary. | mock TB recorder, `assert.Contains` |
| 5 | TestGoldenWithRoot | §21.8 F5 | Unit (positive) | WithRoot redirects to alternate testdata dir. | `t.TempDir`, `assert.True` |
| 6 | TestSnapshotDeterministicAcrossRuns | §21.8 F2, NF3 | Unit (positive) | Two invocations produce byte-identical pretty-printed output. | `assert.Equal` |
| 7 | TestSnapshotIgnoreFields | §21.8 E16, F2 | Unit (positive) | `assert.IgnoreFields("CreatedAt")` excludes from comparison and from persisted snapshot. | `assert.IgnoreFields` |
| 8 | TestSnapshotIgnoreOrder | §21.8 F2 | Unit (positive) | `assert.IgnoreOrder` honored for slice fields. | `assert.IgnoreOrder` |
| 9 | TestSnapshotIgnoreCase | §21.8 F2 | Unit (positive) | IgnoreCase for string fields. | `assert.IgnoreCase` |
| 10 | TestSnapshotWithDelta | §21.8 F2 | Unit (positive) | WithDelta for floats. | `assert.WithDelta` |
| 11 | TestSnapshotMissingCreates | §21.8 F2 | Unit (positive) | `GLACIER_GOLDEN_UPDATE=1` + missing snapshot → created. | `t.Setenv` |
| 12 | TestSnapshotPrettyPrintStableMapKeys | §21.8 NF3 | Unit (positive) | Map keys sorted in pretty-printed output. | `assert.Equal` |
| 13 | TestSnapshotPrettyPrintLineEndingsLF | §21.8 NF3 | Unit (positive) | Output uses `\n` regardless of platform. | `assert.NotContains(out, "\r\n")` |
| 14 | TestLoadReadsTestdata | §21.8 F3 | Unit (positive) | `Load(t, "name")` reads `testdata/name`. | `assert.Equal` |
| 15 | TestLoadMissingErrors | §21.8 E5 | Unit (negative) | Missing file → t.Errorf, returns nil. | mock TB |
| 16 | TestLoadJSONUnmarshalsTyped | §21.8 F4 | Unit (positive) | `LoadJSON[Foo](t, "name")` returns typed value. | `assert.Equal` (typed) |
| 17 | TestLoadJSONExtensionAuto | §21.8 F4 | Unit (positive) | "name" auto-suffixes with `.json` if file absent. | `assert.Equal` |
| 18 | TestLoadJSONMalformed | §21.8 E6 | Unit (negative) | Bad JSON → t.Errorf with parse error; returns zero T. | mock TB |
| 19 | TestPathTraversalRejectedLoad | §21.8 E7, NF6 | Unit (negative :  Falcon §1.13) | `Load(t, "../etc/passwd")` rejected. | `assert.ErrorContains` (via mock TB) |
| 20 | TestPathTraversalRejectedGolden | §21.8 NF6 | Unit (negative) | Golden refuses `..` paths. | mock TB |
| 21 | TestAbsolutePathRejected | §21.8 NF6 | Unit (negative) | Absolute path rejected. | mock TB |
| 22 | TestNonRegularFileRejected | §21.8 NF6 | Unit (negative) | Symlink/device rejected via Lstat. | platform-conditional |
| 23 | TestRealClockNow | §21.8 F8 | Unit (positive) | `Real()` Now ≈ time.Now() within tolerance. | `assert.WithDelta` |
| 24 | TestNewClockNowReturnsStart | §21.8 F10 | Unit (positive) | NewClock(t, T0) :  Now() == T0. | `assert.Equal` |
| 25 | TestNewClockAdvanceMovesNow | §21.8 F9 | Unit (positive) | Advance(d) :  Now() == T0+d. | `assert.Equal` |
| 26 | TestNewClockSetTimeForward | §21.8 F9 | Unit (positive) | SetTime moves Now forward. | `assert.Equal` |
| 27 | TestNewClockSetTimeBackward | §21.8 E8 | Unit (positive) | SetTime backward allowed (test author responsibility documented). | `assert.Equal` |
| 28 | TestNewClockTimerFiresOnAdvance | §21.8 E9, F7 | Unit (positive) | NewTimer(d); Advance(d) → channel receives once. | `concur.WaitGroup`, `assert.Eventually` |
| 29 | TestNewClockTimerFiresOnAdvancePast | §21.8 E9 | Unit (positive) | Advance(2d) on Timer(d) → fires once. | `assert.Eventually` |
| 30 | TestNewClockTickerFiresMultiple | §21.8 F7 | Unit (positive) | NewTicker(d); Advance(3d) → 3 ticks delivered. | `assert.Equal` |
| 31 | TestNewClockTimerStopBeforeFire | §21.8 F7 | Unit (positive) | Stop() returns true; channel never receives. | `assert.True` |
| 32 | TestNewClockTimerResetWhileActive | §21.8 F7 | Unit (positive) | Reset(d2) reschedules. | `assert.Equal` |
| 33 | TestNewClockSleepBlocksUntilAdvance | §21.8 F7 | Unit (positive) | Sleep(d) blocks until Advance(d) on another goroutine. | `concur.WaitGroup` |
| 34 | TestNewClockAfterChannelDelivery | §21.8 F7 | Unit (positive) | After(d) channel receives on Advance(d). | `assert.Eventually` |
| 35 | TestNewClockCleanupViaTCleanup | §23.16 | Unit (lifecycle) | Cleanup registered via t.Cleanup; pending timers stopped at test end. | mock TB tracking Cleanup |
| 36 | TestNewClockAdvanceWhileTimersFireRace | req #4 race | Race | 100 timers + concurrent Advance from another goroutine → -race clean; correct delivery count. | `-race`, `concur.WaitGroup`, `fixture.GuardLeaks` |
| 37 | TestNewClockAdvanceCleanupRace | req #4 race | Race | t.Cleanup runs while a goroutine calls Advance → no race, no panic. | `-race` |
| 38 | TestMockFSRead | §21.8 F11 | Unit (positive) | NewFS({"a":bytes}); fs.ReadFile("a") returns bytes. | `assert.Equal` |
| 39 | TestMockFSReadDir | §21.8 F11 | Unit (positive) | ReadDir lists entries. | `assert.Equal` |
| 40 | TestMockFSReadFileInterface | §21.8 F11 | Unit (positive) | Returned FS satisfies `fs.ReadFileFS` AND `fs.ReadDirFS`. | type assertion + `assert.True` |
| 41 | TestMockFSConflictPanics | §21.8 E10 | Unit (negative) | NewFS with both `/foo` and `/foo/bar` (file vs dir conflict) panics with `"fixture: NewFS: conflict at path \"/foo\": both file and directory"`. | `assert.PanicsWithMessage` |
| 42 | TestMockFSNestedPaths | §21.8 F11 | Unit (positive) | Path `a/b/c.txt` accessible. | `assert.Equal` |
| 43 | TestMockFSReadOnly | §21.8 F11 | Unit (positive) | No write methods on returned FS. | reflection check |
| 44 | TestCaptureRoundTrip | §21.8 F12 | Unit (positive) | Capture(t, fn) returns fn's stdout/stderr. | `assert.Equal` |
| 45 | TestCaptureRestoresStreams | §21.8 NF4 | Unit (positive) | After Capture returns, os.Stdout / os.Stderr restored. | identity check, `assert.Equal` |
| 46 | TestCaptureProcessWideLockSerializes | §21.8 NF7, E11, req #4 race | Race | Two parallel `Capture` calls :  they serialize; no interleaving. | `concur.WaitGroup`, `-race`, `fixture.GuardLeaks` |
| 47 | TestCaptureFromGoroutine | §21.8 E11 | Unit (documented) | Capture lock holds even when fn writes from goroutine; ALL stdout goes to buffer. | `concur.WaitGroup`, `assert.Contains` |
| 48 | TestGuardLeaksTempDirCatchesLeak | §21.8 F14 | Unit (negative) | Test creates `mongoose-XXXXX` (or `glacier-XXXXX`) dir without removing → cleanup reports. | mock TB tracker |
| 49 | TestGuardLeaksTempDirIgnoresUnrelated | §21.8 E14 | Unit (positive) | Non-glacier-prefix temp dirs do NOT trigger. | mock TB tracker |
| 50 | TestGuardLeaksGoroutineCatchesLeak | §21.8 F15 | Unit (negative) | Test spawns goroutine that doesn't terminate → cleanup reports with stack. | `concur.WaitGroup` (deliberately unfinished) |
| 51 | TestGuardLeaksGoroutineDrainTimeout | §21.8 F15, E12 | Unit (positive) | WithDrainTimeout extends wait; allows legitimate-async cleanup. | `assert.Eventually` |
| 52 | TestGuardLeaksGoroutineFiltersFalsePositives | §21.8 F15 | Unit (positive) | Network resolver / GC sweepers / cleanup goroutine itself filtered. | mock TB |
| 53 | TestGuardLeaksEnvCatchesLeak | §21.8 F16 | Unit (negative) | Test sets env var without unsetting → cleanup reports. | mock TB |
| 54 | TestGuardLeaksEnvIgnoresUnrelated | §21.8 F16 | Unit (positive) | Env vars set BEFORE GuardLeaks ignored. | mock TB |
| 55 | TestGuardLeaksFDsCatchesLeak | §21.8 F17 | Unit (negative :  Linux/macOS) | Open file without close → cleanup reports leaked FD. | platform-conditional |
| 56 | TestGuardLeaksFDsNoOpOnWindows | §21.8 E13 | Unit (positive :  Windows) | No-op with debug log on Windows. | `fixture.Capture`, `assert.Contains` |
| 57 | TestGuardLeaksWatchAll | §21.8 F18 | Unit (positive) | WatchAll == all four Watches. | introspection |
| 58 | TestGuardLeaksStrictHaltsTest | §21.8 F19 | Unit (negative) | Strict() makes leaks call t.Fatalf. | mock TB tracker |
| 59 | TestGuardLeaksStrictRenamed | §23.15 | Unit (positive) | The renamed `StrictLeaks` (per §23.15) IS the public API; `Strict` exists too OR is aliased :  verify naming. | reflection |
| 60 | TestGuardLeaksParallelSubtest | §21.8 E15 | Unit (positive) | Subtest with own GuardLeaks has own baseline. | `t.Run`, mock TB |
| 61 | TestGuardLeaksBaselineCleanupRace | req #4 race | Race | Baseline recorded synchronously; no race with subsequent ops. | `-race` |
| 62 | TestSnapshotRoundTripProperty | req #5 property | Property | `Snapshot[T](v)`: `Load` of the snapshot file produces representation equal to `v` modulo IgnoreFields. Random struct generator. | property generator, `assert.Equal` |
| 63 | TestEnvVarRenameToGlacier | §22, §23.18 | Unit (positive) | `GLACIER_GOLDEN_UPDATE=1` activates update mode. (`MONGOOSE_GOLDEN_UPDATE` is gone.) | `t.Setenv`, `assert.True` |
| 64 | TestSnapshotFormatterDeterministicWithMaps | §21.8 NF3 | Unit (positive) | Map with 100 keys produces identical output across 100 runs. | `concur.Group`, `assert.Equal` |
| 65 | TestSnapshotFormatterUnicodeStable | §21.8 NF3 | Unit (positive) | UTF-8 strings preserve byte identity. | `assert.Equal` |
| 66 | TestSnapshotPathRejectsTraversal | §21.8 NF6 | Unit (negative) | Snapshot name `../oops` rejected. | mock TB |
| 67 | TestErrFormatRegisterCompliance | §21.8 NF8 | Unit (cross-cutting) | Internal errors use library register; user-facing test output uses CLI register (capitalized, period-terminated). | `assert.Regexp` |
| 68 | BenchmarkGoldenBytes | §21.8 NF1 | Benchmark | Compare-and-pass path for 4 KiB goldens. | `testing.B` |
| 69 | BenchmarkSnapshotStruct | §21.8 NF1, §23.13 | Benchmark | Snapshot of 50-field struct :  pretty-print is the cost. | `testing.B`, `testing.AllocsPerRun` |
| 70 | BenchmarkSnapshotMap100Keys | req #2 | Benchmark | Deterministic-formatter perf with 100-key map (sort cost). | `testing.B` |
| 71 | BenchmarkCaptureSmallOutput | §21.8 NF1 | Benchmark | Capture of small fn cost (lock + redirect). | `testing.B` |
| 72 | BenchmarkNewClockAdvanceManyTimers | §21.8 NF1 | Benchmark | NewClock with 1000 pending timers; Advance triggers all. | `testing.B` |

### Coverage target

- 100% line on every public symbol per F1–F20.
- 100% on `path_safety.go` Lstat / canonicalize / `..` rejection paths.
- ≥ 95% line overall; FD-watching code (Linux/macOS specific) covered on those platforms.

### Edge cases not in spec

- **EX1**: `Snapshot[T]` of a value containing `time.Time` :  the formatter must NOT include time-of-day (NF3); verify by snapshotting a struct with a `time.Time` field set to `time.Now()` and checking determinism after sleeping.
- **EX2**: `Capture` with fn that panics :  streams MUST still be restored. Test via deferred restore check.
- **EX3**: `NewClock` with a stopped Timer that's then Reset :  fires correctly on Advance.
- **EX4**: `GuardLeaks` ordering with multiple Watch* options :  does ordering matter? (Lock to: no.)
- **EX5**: `LoadJSON[T]` for T = `map[string]any` :  works; no struct tags involved.
- **EX6**: `MockFS` `Stat` on a directory :  reports IsDir correctly.
- **EX7**: `Golden` with empty `got` byte slice :  distinct from missing file; matched against empty golden file.

### Special concerns

- **§23.15 naming**: `StrictLeaks` is the new public name (renamed from `Strict()` to disambiguate from mock/httpmock semantics). Test #59 verifies the public name.
- **Capture process-wide lock** (NF7): the lock is the single most-invasive piece of fixture/. Test #46 explicitly runs two parallel `Capture` calls and verifies serialization. CI gates this under `-race -parallel=4`.
- **§23.16 NewClock cleanup**: test #35 mock-TB-tracks that `t.Cleanup` was called. Cleanup STOPS pending timers (preventing goroutine leaks).
- **Path-safety tests** (#19–#22) are Falcon-required; they must pass on every CI run. macOS `/dev/fd` and Linux `/proc/self/fd` paths are covered by tests #55. Windows test #56 verifies the no-op + log behavior.
- **§22/§23.18 env-var rename**: `MONGOOSE_GOLDEN_UPDATE` is now `GLACIER_GOLDEN_UPDATE`. Test #63 locks this. No fallback to old name (clean rebrand).

---

## Package: `obs/`

### Test files

- `obs/init_test.go` :  F1–F3.
- `obs/metrics_test.go` :  F4–F8.
- `obs/traces_test.go` :  F9–F13.
- `obs/log_integration_test.go` :  F14 (cross-package: log + obs).
- `obs/per_package_instrumentation_test.go` :  F15–F19 (uses httpc/cli/conf with WithTracing/WithMetrics).
- `obs/attributes_test.go` :  F20–F21.
- `obs/otlp_fake_collector_test.go` :  fake OTLP collector via httpmock.
- `obs/url_parse_fuzz_test.go` :  req #3: OTLP endpoint URL parsing fuzz.
- `obs/concurrency_test.go` :  req #4: span/counter from concurrent goroutines.
- `obs/lifecycle_test.go` :  §23.16 Provider.Shutdown idempotency.
- `obs/properties_test.go` :  req #5: trace/span ID propagation through `log.With(ctx)`.
- `obs/bench_test.go` :  §23.13 perf gates.
- `obs/example_test.go` :  Godoc examples.

### Edge case list (Lynx-authored :  obs is newer per req #6)

| # | Edge case | Behavior |
|---|---|---|
| EO1 | `Init` called twice | Second Init returns error `"obs: init: already initialized"`; Default unchanged. |
| EO2 | `Init` with no exporter and `OTEL_EXPORTER_OTLP_ENDPOINT` unset | Falls back to no-op exporter; not an error; logs at info. |
| EO3 | `Init` with malformed OTLP endpoint URL | Returns typed error `*InitError{Cause: url.Error}`. |
| EO4 | `Provider.Shutdown(ctx)` with already-cancelled ctx | Best-effort flush; returns ErrCancelled wrapping ctx.Err. |
| EO5 | `Shutdown` called twice | Second call returns nil (idempotency per §23.16). |
| EO6 | `Counter[int64]("name").Add(ctx, 1)` before Init | No-op; Default is nil; Add is safe (early return). |
| EO7 | `StartSpan(ctx, "")` empty name | Span created with name `"<unnamed>"`; logged at warn. |
| EO8 | `(*Span).End()` called twice | Second End is a no-op (idempotency). |
| EO9 | `(*Span).RecordError(nil)` | No-op. |
| EO10 | `SpanFromContext(ctx)` with no active span | Returns a non-nil "no-op" span (calls on it are safe no-ops). |
| EO11 | `TraceIDFromContext(ctx)` with no span | Returns (zero, false). |
| EO12 | `Counter[int64]` and `Counter[float64]` with same name | OTEL semantics: error / replaced; locked behavior in matrix (last-wins or error). |
| EO13 | `WithUnit("ms")` paired with histogram of bytes | Documented; obs does not validate unit/value coherence. |
| EO14 | `Attribute` with control characters in key | Sanitized per OTEL semantic conventions or rejected. |
| EO15 | Concurrent `StartSpan` from 100 goroutines | Race-clean; each gets independent span. |
| EO16 | Span left un-Ended at goroutine exit | Garbage-collected eventually; no panic; warn at debug. |
| EO17 | `log.With(ctx, ...)` when ctx has no span | trace_id/span_id NOT appended (absence, not zero). |
| EO18 | `Init` with custom sampler that always rejects | All spans no-op; metrics still record. |
| EO19 | `OTEL_EXPORTER_OTLP_ENDPOINT` with non-localhost (untrusted) | Allowed (OTEL semantics); logged at info; documented per untrusted-input register. |
| EO20 | OTLP exporter unable to connect at startup | Init does NOT block; exports happen async; failures logged. |
| EO21 | Counter increment with attrs containing nil string | Coerced to empty string. |
| EO22 | Span set status `StatusOk` after `StatusError` | Last write wins per OTEL semantics. |
| EO23 | Provider.Shutdown joining errors from Meter and Tracer | Joined via errs.Join (per §23.16). |

### Test matrix

| # | Name | Spec ref | Type | Description | Test helpers used |
|---|---|---|---|---|---|
| 1 | TestInitDefaults | §21.15 F1 | Unit (positive) | Init with no opts → Default set; resource attrs include service info. | `assert.NotNil`, `assert.Equal` |
| 2 | TestInitTwiceErrors | EO1 | Unit (negative) | Second Init returns error. | `assert.ErrorContains` |
| 3 | TestInitNoEndpointFallsBackToNoOp | EO2 | Unit (positive) | No env, no opt → no-op exporter; not an error. | `t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT","")`, `assert.NoError` |
| 4 | TestInitMalformedEndpoint | EO3 | Unit (negative) | Bad URL → InitError. | `t.Setenv`, `assert.ErrorAs` |
| 5 | TestWithExporterOption | §21.15 F2 | Unit (positive) | Custom exporter accepted; receives spans. | fake exporter, `assert.Equal` |
| 6 | TestWithSamplerOption | §21.15 F2 | Unit (positive) | Custom sampler controls span recording. | always-reject sampler, `assert.Equal` |
| 7 | TestWithResourceAttribute | §21.15 F2 | Unit (positive) | Resource attr appears on every span. | fake exporter, `assert.Contains` |
| 8 | TestWithMetricsInterval | §21.15 F2 | Unit (positive) | Metrics flushed at custom interval. | `fixture.NewClock`, fake exporter |
| 9 | TestWithLoggerInternal | §21.15 F2 | Unit (positive) | Internal events log via custom slog logger. | `fixture.Capture` + slog handler, `assert.Contains` |
| 10 | TestDefaultPackageVarSet | §21.15 F3 | Unit (positive) | After Init, `obs.Default` set. | `assert.NotNil` |
| 11 | TestPackageLevelCounterUsesDefault | §21.15 F3 | Unit (positive) | `obs.Counter[int64]("x").Add(ctx, 1)` after Init records on Default. | fake exporter, `assert.Equal` |
| 12 | TestCounterInt64Add | §21.15 F4 | Unit (positive) | int64 counter Add increments. | fake exporter |
| 13 | TestCounterFloat64Add | §21.15 F4 | Unit (positive) | float64 counter Add increments. | fake exporter |
| 14 | TestCounterAddBeforeInitNoOp | EO6 | Unit (positive) | Counter created before Init; Add is no-op. | `assert.NoError` |
| 15 | TestCounterDuplicateNameDifferentT | EO12 | Unit (negative or last-wins) | Counter[int64]("x") and Counter[float64]("x") :  locked behavior asserted. | locked semantics |
| 16 | TestHistogramRecord | §21.15 F5 | Unit (positive) | Histogram[float64].Record observes value. | fake exporter |
| 17 | TestGaugeSet | §21.15 F6 | Unit (positive) | Gauge.Set updates value; subsequent Set replaces. | fake exporter |
| 18 | TestMetricWithDescription | §21.15 F7 | Unit (positive) | Description appears in instrument metadata. | fake exporter |
| 19 | TestMetricWithUnit | §21.15 F7 | Unit (positive) | Unit string applied. | fake exporter |
| 20 | TestNumericTypeConstraintCompiles | §21.15 F8 | Unit (compile-time) | int64 and float64 satisfy NumericType; string does not. | compile-only |
| 21 | TestStartSpanReturnsCtx | §21.15 F9 | Unit (positive) | StartSpan returns derived ctx with span attached. | `obs.SpanFromContext`, `assert.NotNil` |
| 22 | TestSpanEnd | §21.15 F10 | Unit (positive) | End() exports span via configured exporter. | fake exporter, `assert.Equal` |
| 23 | TestSpanEndIdempotent | EO8 | Unit (positive) | End() twice :  second is no-op. | fake exporter, `assert.Equal(count==1)` |
| 24 | TestSpanRecordError | §21.15 F10 | Unit (positive) | RecordError adds an exception event. | fake exporter, `assert.Contains` |
| 25 | TestSpanRecordErrorNilNoOp | EO9 | Unit (positive) | RecordError(nil) :  no event. | fake exporter |
| 26 | TestSpanSetStatusOk | §21.15 F10 | Unit (positive) | SetStatus(Ok, "") | fake exporter |
| 27 | TestSpanSetStatusError | §21.15 F10 | Unit (positive) | SetStatus(Error, "msg") | fake exporter |
| 28 | TestSpanSetStatusOkAfterError | EO22 | Unit (positive) | Last write wins. | fake exporter |
| 29 | TestSpanAddEvent | §21.15 F10 | Unit (positive) | Event appears with attrs. | fake exporter, `assert.Contains` |
| 30 | TestSpanSetAttribute | §21.15 F10 | Unit (positive) | Attribute on span. | fake exporter |
| 31 | TestSpanWithSpanKindServer | §21.15 F11 | Unit (positive) | Span kind exported correctly. | fake exporter, `assert.Equal` |
| 32 | TestSpanWithAttributesOption | §21.15 F11 | Unit (positive) | Initial attrs applied. | fake exporter |
| 33 | TestSpanFromContextNonNil | §21.15 F12 | Unit (positive) | After StartSpan, SpanFromContext returns the same span. | identity check |
| 34 | TestSpanFromContextNoActiveReturnsNoOp | EO10 | Unit (positive) | No span in ctx → returns non-nil no-op span; calls safe. | `assert.NotNil`, no-panic |
| 35 | TestTraceIDFromContext | §21.15 F13 | Unit (positive) | After StartSpan, TraceIDFromContext returns valid id. | `assert.True` |
| 36 | TestTraceIDFromContextNoSpan | EO11 | Unit (positive) | (zero, false). | `assert.False` |
| 37 | TestSpanIDFromContext | §21.15 F13 | Unit (positive) | After StartSpan, valid SpanID. | `assert.True` |
| 38 | TestEmptySpanNameWarns | EO7 | Unit (positive) | Empty name → `<unnamed>`; warn log. | `fixture.Capture`, `assert.Contains` |
| 39 | TestLogIntegrationAddsTraceIDSpanID | §21.15 F14, req #5 property | Unit (positive) | `log.With(ctx, ...)` after StartSpan adds `trace_id` and `span_id` slog attrs to records. | `fixture.Capture` + custom slog handler, `assert.Contains` |
| 40 | TestLogIntegrationNoSpanNoAttrs | EO17 | Unit (positive) | log.With(ctx, ...) with no span :  no trace attrs. | `fixture.Capture`, `assert.NotContains` |
| 41 | PropertyTraceSpanIDPropagation | req #5 property | Property | For random ctx-derivation depths and span-nesting depths, the active span's ids are exactly what `log.With` exposes. | property generator |
| 42 | TestHttpcWithTracingEmitsClientSpan | §21.15 F15 | Integration | httpc client with WithTracing → each request emits client-kind span with `http.method`, `http.url`, `http.status_code`, `http.response.size`, `glacier.retry_attempt`, name `"HTTP <METHOD> <PATH>"`. | `httpmock`, fake OTLP exporter, `assert.Equal` |
| 43 | TestHttpcWithMetricsEmitsCounterAndHistogram | §21.15 F16 | Integration | Counter `glacier.httpc.requests`, histograms `glacier.httpc.duration_ms`, `glacier.httpc.response_size_bytes`. | fake exporter, `assert.Contains` |
| 44 | TestCliWithMetrics | §21.15 F17 | Integration | Counter `glacier.cli.invocations`, histogram `glacier.cli.duration_ms`. | fake exporter |
| 45 | TestConfWithMetrics | §21.15 F18 | Integration | Counter `glacier.conf.loads`, histogram `glacier.conf.load_duration_ms`, gauge `glacier.conf.registered_paths`. | fake exporter |
| 46 | TestPerPackageOptOutZeroOverhead | §21.15 NF1 | Bench | Without WithTracing/WithMetrics, no spans/counters emitted; overhead == zero. | `testing.AllocsPerRun` |
| 47 | TestAttributeStringConstructor | §21.15 F20 | Unit (positive) | String(k,v) builds Attribute. | `assert.Equal` |
| 48 | TestAttributeIntConstructor | §21.15 F20 | Unit (positive) | Int(k,v). | `assert.Equal` |
| 49 | TestAttributeFloatConstructor | §21.15 F20 | Unit (positive) | Float(k,v). | `assert.Equal` |
| 50 | TestAttributeBoolConstructor | §21.15 F20 | Unit (positive) | Bool(k,v). | `assert.Equal` |
| 51 | TestAttributeStringSliceConstructor | §21.15 F20 | Unit (positive) | StringSlice(k,v). | `assert.Equal` |
| 52 | TestStandardKeyConstantsStable | §21.15 F21 | Unit (positive) | KeyHTTPMethod == "http.method", etc. :  frozen string values. | `assert.Equal` |
| 53 | TestStandardKeyConstantsOTELConformant | req obs OTEL-specific | Unit (positive) | All key constants follow OTEL semantic conventions where they overlap (http.*); custom (glacier.*) consistently namespaced. | regex check |
| 54 | TestProviderShutdownFlushes | §23.16, §21.15 F1 | Unit (lifecycle) | Shutdown flushes pending spans and metrics through exporter before returning. | fake exporter, `assert.Equal` |
| 55 | TestProviderShutdownIdempotent | §23.16, EO5 | Unit (lifecycle) | Two Shutdown calls :  second returns nil. | `assert.NoError` (paired) |
| 56 | TestProviderShutdownJoinsErrors | §23.16, EO23 | Unit (lifecycle) | Shutdown returns errs.Join over Meter+Tracer shutdown errors. | `errs.Chain`, `fluent.Count` |
| 57 | TestProviderShutdownCancelledCtx | EO4 | Unit (negative) | Cancelled ctx → ErrCancelled-class error after best-effort flush. | `assert.ErrorIs` |
| 58 | TestConcurrentSpanStartEnd | req #4 race | Race | 100 goroutines StartSpan/End simultaneously :  race-clean. | `concur.WaitGroup`, `-race`, `fixture.GuardLeaks` |
| 59 | TestConcurrentCounterAdd | req #4 race | Race | 100 goroutines Counter.Add :  final value correct, no race. | `concur.Group`, `-race`, `assert.Equal` |
| 60 | TestSpanLeftUnEndedNoLeak | EO16 | Unit (lifecycle) | Span goes out of scope without End :  finalizer or weakref handles cleanup; no goroutine leak. | `fixture.GuardLeaks(WatchGoroutines)` |
| 61 | FuzzOTLPEndpointURLParse | req #3, §21.15 F1 | Fuzz | `OTEL_EXPORTER_OTLP_ENDPOINT` value (and `WithExporter` URL handling) :  random byte URLs never panic; either parse cleanly or return InitError. | `testing.F`, seed corpus |
| 62 | TestOTLPExporterRoundTripViaFakeCollector | req obs OTEL-specific | Integration | Fake OTLP collector via `httpmock` (HTTP/protobuf endpoint) :  counter/histogram/gauge round-trip via OTLP exporter; spans exported. | `httpmock`, fake collector, `assert.Equal` |
| 63 | TestSpansEmitOTELConformantAttributeKeys | req obs OTEL-specific | Integration | Emitted spans pass OTEL semantic-convention validation (http.* prefix for HTTP attrs, etc.). | OTEL conformance checker |
| 64 | TestUntrustedRemoteEndpointDocumented | EO19 | Unit (positive :  Falcon §1.x untrusted-input row) | Non-localhost endpoint accepted; logged at info; documented in untrusted-input register. | `fixture.Capture`, `assert.Contains` |
| 65 | TestErrFormatRegisterCompliance | §21.15 NF3 | Unit (cross-cutting) | Errors emitted by obs use library register: `obs: <action>: <cause>`. | `assert.Regexp` |
| 66 | BenchmarkSpanStartEnd | §21.15 NF1, §23.13 (req #2) | Benchmark | StartSpan + End ≤ 1 µs/op. | `testing.B` |
| 67 | BenchmarkCounterAdd | §21.15 NF1, §23.13 (req #2) | Benchmark | Counter.Add ≤ 200 ns/op. | `testing.B` |
| 68 | BenchmarkHistogramRecord | §21.15 NF1 | Benchmark | Same envelope; documented. | `testing.B` |
| 69 | BenchmarkGaugeSet | §21.15 NF1 | Benchmark | Same envelope; documented. | `testing.B` |
| 70 | BenchmarkPerPackageDisabledOverhead | §21.15 NF1 | Benchmark | When `WithTracing` not set, overhead == 0 allocs/op. | `testing.AllocsPerRun` |

### Coverage target

- 100% line on every public symbol per F1–F21.
- 100% on `Provider.Shutdown` flush + idempotency paths (§23.16).
- ≥ 90% line overall (some OTEL-SDK-internal error paths are wrapper-only and difficult to exercise; fuzz #61 covers URL parsing).

### Special concerns

- **Heaviest dependency footprint in the framework** (§21.15 NF6): `go.opentelemetry.io/otel` + 7 sub-packages + grpc + protobuf transitively. CI `vuln-scan` and `license-scan` (D31) MUST gate this. Falcon's six-question checklist sign-off recorded.
- **Per-package instrumentation tests** (§21.15 F15–F18) require integration setups that import `httpc/`, `cli/`, `conf/`. These tests live in `obs/per_package_instrumentation_test.go` and are guarded by `//go:build integration` to keep unit-test runs fast.
- **OTLP fake collector** (test #62): leverages `httpmock` to fake the OTLP HTTP endpoint. The gRPC OTLP exporter requires a different setup :  the matrix uses HTTP-OTLP for unit tests; gRPC-OTLP is exercised separately in a long-running test.
- **req #5 property** (test #41): trace/span ID propagation through `log.With(ctx)` is an integration property :  verified across random ctx-tree shapes.
- **req #3 fuzz** (test #61): the OTLP endpoint URL parser is the untrusted-input boundary; CI runs 30s/PR, 10min/nightly per D31.
- **§23.16 Shutdown idempotency** (tests #55–#56): explicit lock-in.
- **No obs → log import** (§21.15 F14): obs uses its own injected slog logger via `WithLogger`; log/ optionally imports obs/ for the trace-context bridge (one-way edge). Architectural test asserts this in `obs/imports_test.go`.

---

## Cross-package integration tests for mid-tier

Located in `internal/integration_test/midtier/` (separate test package, `//go:build integration`).

| # | Name | Description | Packages |
|---|---|---|---|
| X1 | TestConfLoadFiltersViaFluent | `conf.Load` with 50 registrations; fluent's `Filter`+`GroupBy` selects subset of registered structs by path prefix. | conf + fluent |
| X2 | TestObsSpanWrapsConcurGroupGo | `obs.StartSpan`; pass ctx into `concur.Group`; each goroutine reads SpanFromContext and emits child spans. Verify trace-tree topology. | obs + concur |
| X3 | TestFixtureSnapshotOfConfDecodeResult | `Decode[Config]`; `fixture.Snapshot[Config]` of result; verify deterministic. | fixture + conf |
| X4 | TestFixtureGoldenForFluentToSliceOutput | Pipeline output captured to bytes via fluent.Lines + ToSlice; persisted via Golden. | fixture + fluent |
| X5 | TestObsCounterIncrementedFromConcurGroup | Many goroutines via concur.Group; each calls `Counter.Add`; final aggregated count correct. | obs + concur |
| X6 | TestConfReloadDoesNotRaceWithFluentReader | Reader iterating `fluent.From(snapshot.Items)` while `conf.Load` runs concurrently :  atomic.Pointer indirection means reader sees snapshot consistent. | conf + fluent + concur |
| X7 | TestFixtureNewClockDrivesObsMetricsInterval | `fixture.NewClock` injected into obs's `WithMetricsInterval` test path; deterministic flush timing. | fixture + obs |
| X8 | TestGuardLeaksCatchesObsExporterGoroutine | If `Provider.Shutdown` is skipped, `WatchGoroutines()` reports the OTLP-exporter background goroutine. | fixture + obs |
| X9 | TestConcurSemaphoreLimitsHttpcViaObs | concur.Semaphore-backed rate limiter; obs WithMetrics tracks histogram of wait time. | concur + obs |
| X10 | TestFluentLinesOverFixtureMockFS | fluent.Lines reading from a file inside `fixture.NewFS`; never escapes the FS. | fluent + fixture |
| X11 | TestConfFlagSourceFromMockedCli | Mock `FlagSource` populated by simulated cli flags; conf.Load uses it. | conf + mock (kernel adjacent) |
| X12 | TestTraceCorrelatedLogsThroughLogPackage | obs.StartSpan; `log.From(ctx).Info(...)`; output contains trace_id+span_id. | obs + log |

---

## Test infrastructure dependencies

Mid-tier packages CAN import all kernel packages per D12; this matrix uses that freely (req #10).

| Package | Kernel deps used in tests | Other deps |
|---|---|---|
| `concur` | `assert`, `assert/require`, `option`, `errs` | `fixture` (GuardLeaks, NewClock for hold-timeout); `term` (color in failure messages :  req #10) |
| `fluent` | `assert`, `assert/require` | `fixture` (Capture for Lines/Words tests, GuardLeaks for race tests); no `option`/`errs` because fluent is dep-free |
| `conf` | `assert`, `assert/require`, `option`, `errs`, `log` | `fixture` (NewFS, GuardLeaks, t.Setenv via stdlib); `mock` (FlagSource); `internal/reflectx` (for cache invariant tests) |
| `fixture` | `assert`, `assert/require`, `option`, `errs`, `log` | self-tests use `fixture` recursively (acceptable since this is mid-tier per req #10) :  must avoid the "asserts asserting themselves" cycle by writing the lowest-layer self-test in plain stdlib `testing` |
| `obs` | `assert`, `assert/require`, `option`, `errs`, `log` | `fixture` (Capture, NewClock, GuardLeaks); `httpmock` (fake OTLP collector); `mock` (fake exporter/sampler); `concur` (concurrency tests); `fluent` (no :  only used for cross-pkg integration) |

**Test-helper dogfooding (req #10):** every test uses `assert/`, `assert/require/`, `fixture/`, `mock/`, `httpmock/`. No `testify`. No hand-rolled equality. `term/` may be imported for color in failure messages. The "test helpers testing themselves" problem is resolved at the kernel tier by Lynx-kernel matrix; mid-tier tests use the helpers as a normal consumer would.

---

## Sign-off conditions

The mid-tier section of spec 0002 reaches `accepted` only when ALL of the following hold:

1. **Implementation completeness** :  every public symbol in §21.5, §21.6, §21.7, §21.8, §21.15 is exercised by at least one positive, one negative, and one edge test in this matrix.
2. **`go test ./...` exit code 0** :  across `concur/`, `fluent/`, `conf/`, `fixture/`, `obs/` on Linux, macOS, and Windows.
3. **`go test -race ./...` exit code 0** :  every concurrent path covered (req #4 list complete: Mutex+RWMutex+LockCtx, Group.Go+WaitDone, Semaphore concurrent, Pool, Once, WaitGroup.WaitCtx, conf.Load Snapshot accessor, fluent operators concurrent contexts, fixture.Capture, NewClock.Advance with timers, GuardLeaks baseline+cleanup, obs span/counter concurrent).
4. **Coverage thresholds (req #12)** :  ≥ 90% line per package; **100% on public API** (every exported symbol has a covering test that asserts at least one observable behavior).
5. **Benchmark gates (D35 + §23.13)** :  every gate in req #2 met:
   - `concur.Mutex` Lock/Unlock within 5% of stdlib `sync.Mutex` (test #65, #8).
   - `concur.Semaphore` fast-path uncontended ≤ 50 ns/op zero allocs (test #67).
   - `fluent` per-element zero-alloc for non-capturing functions in pre-composed pipelines (test #101).
   - `conf.Load` with 50 registrations :  baselined; regression gate ±5% (test #67).
   - `fixture.Snapshot` deterministic-formatter perf :  baselined (test #69, #70).
   - `obs` span-start/end ≤ 1 µs (test #66); counter Add ≤ 200 ns (test #67).
6. **Fuzz gates (D31 + req #3)** :  `FuzzLoadJSON`, `FuzzWithSetPathCoercion`, `FuzzLines`, `FuzzWords`, `FuzzOTLPEndpointURLParse`, `FuzzSemaphoreAcquireRelease` all run 30s in PR CI, 10min nightly. Zero crashes, zero un-recovered panics across 24 hours of continuous nightly fuzzing.
7. **Property-based tests pass (req #5)** :  Group error-count, fluent algebraic identities, Reduce invertibility, conf.Load idempotence, fixture.Snapshot round-trip, obs trace-id propagation. Each runs 1000 generated cases per CI.
8. **Concurrency lock-ins (§23.14)** :  tests #26 (Go-after-WaitDone PANICS), #34 (atomic.Pointer no torn read), #40+#41 (semaphore cancel-watcher no leak), #4 (LockCtx try-lock-backoff) all pass.
9. **Lifecycle audit (§23.16)** :  Group has no Close (test #63); conf.Loader.Close idempotent (#54, #55); fixture.NewClock cleanup via t.Cleanup (#35); obs.Provider.Shutdown idempotent + flushes + joins (#54–#57).
10. **Generics fixes (§23.17)** :  `option.Required[T]` T-load-bearing test (conf #53); `option.Apply[T]` variadic mode last-wins + `Strict()` collects all errors (covered in option/ kernel matrix, but conf tests assert the conf-level usage works).
11. **Spec-traceability (req #11)** :  every test name in this matrix carries a comment block citing its spec section / decision ID (`// spec: §21.5 F4` etc.). Lint-test verifies presence.
12. **Test-helper dogfooding (req #10)** :  `grep -r "github.com/stretchr/testify"` returns zero matches across `concur/`, `fluent/`, `conf/`, `fixture/`, `obs/` test files. CI lint-gate.
13. **OTEL conformance (req #13)** :  obs spans' attribute keys validated against OTEL semantic conventions (test #63); counter/histogram/gauge round-trip via fake collector (test #62); `obs.SpanFromContext` retrieval verified (test #33, #34).
14. **Security gates (Falcon-required)** :  path-safety tests pass on all platforms (conf #28–#31, fixture #19–#22); untrusted-input register rows for conf JSON parser, fixture path canonicalization, obs OTLP endpoint URL all have fuzz coverage.
15. **Cross-package integration tests** :  X1–X12 all pass on Linux CI under the `integration` build tag.
16. **No-flake gate** :  every concurrency test passes 100/100 runs on `-count=100`. Property tests pass 10/10 runs with new random seeds.
17. **Lynx final sign-off** :  Lynx (this agent) reviews the implementation diff against this matrix; any uncovered symbol blocks `accepted`.

---

**Lynx's verdict, conditional on implementation:** the mid-tier matrix as specified above gives 100% release confidence. A developer who runs `go test -race -count=10 ./...` plus the nightly fuzz suite plus the integration tests does NOT need to verify anything by hand. If any item above proves un-implementable, the matrix is incomplete and Lynx withholds sign-off until the gap is filled or the spec is amended via §23-style addendum.

`ʕ•ᴥ•ʔ` Mid-tier sign-off pending implementation.