# Lynx Kernel-Tier Test Matrix (option, errs, log, assert, term)

> **Scope:** §21.1, §21.2, §21.3, §21.4, §21.14 :  five Tier-0 kernel packages :  plus the §23.4/.5/.13/.14/.16/.17 amendments that override them.
> **Standard:** `go test ./...` against the kernel must give 100% release confidence with zero manual verification.
> **Helper rule:** Once the bootstrap subset is proven, every later test uses `assert/`, `assert/require/`, `fixture/`, and `mock/` :  never bare `if got != want`, never testify, never hand-rolled stubs. Bootstrap exception is documented per package below.

---

## Package: `option/`

### Test files
- `option/option_test.go` :  unit tests (Apply, Validate, Required, OptionFunc, Mode/Strict)
- `option/option_bench_test.go` :  benchmarks for the recalibrated D35 targets and the NF1 zero-alloc claim
- `option/option_fuzz_test.go` :  fuzz `Required`'s `name` argument round-trip in error formatting
- `option/option_property_test.go` :  algebraic-property tests (idempotency, last-wins, nil-skip)
- `option/option_concurrent_test.go` :  race-detector tests for stateless concurrency claim (NF2)
- `option/example_test.go` :  runnable `Example*` functions per NF7

### Test matrix

| #  | Name | Spec ref | Type | Description | Test helpers used |
|----|------|----------|------|-------------|-------------------|
| 1  | TestApplyEmpty | §21.1 F3, E2; §23.4 | unit | `Apply([]Option[T]{})` returns zero T, nil. | `assert.Equal`, `assert.NoError` |
| 2  | TestApplyOneOption | §21.1 F3 | unit | Single non-nil option mutates T as expected. | `assert.Equal`, `assert.NoError` |
| 3  | TestApplyMultipleOptions | §21.1 F3 | unit | Three options compose; final T has all three fields set. | `assert.Equal` |
| 4  | TestApplyDefaultShortCircuit | §21.1 F4, E7 | unit | First option errors → returns at first error; later options not applied; partial T documented. | `assert.ErrorIs`, `assert.Equal` |
| 5  | TestApplyDefaultSecondErrors | §21.1 F4 | unit | First option succeeds, second errors → first applied, third never run. | `assert.Equal`, `assert.ErrorIs` |
| 6  | TestApplyStrictMultipleErrors | §21.1 F5, E8 | unit | `Strict()` accumulates 2+ failures via `errors.Join`; partial T returned. | `assert.ErrorIs` (twice over `errors.Unwrap`-walked join) |
| 7  | TestApplyStrictNoErrors | §21.1 F5 | unit | All options succeed under Strict → returns configured T, nil. | `assert.NoError` |
| 8  | TestApplyNilOptionSkipped | §21.1 F9, E1 | unit | `[]Option[T]{nil, WithA, nil}` → only WithA applied; no panic. | `assert.NoError`, `assert.Equal` |
| 9  | TestApplyAllNilOptionsReturnsZero | §21.1 F9, E1 | unit | All-nil slice yields zero T, nil err. | `assert.Equal`, `assert.NoError` |
| 10 | TestApplyDuplicateLastWins | §21.1 F11, E3 | unit | Same field set twice → last value wins. | `assert.Equal` |
| 11 | TestApplyMultipleModesLastWins | §21.1 F12, E4 | unit | `Apply(opts, Mode{}, Strict())` → strict semantics active. | `assert.Equal` |
| 12 | TestApplyZeroModesIsDefault | §21.1 F12 | unit | No mode arg → default (short-circuit) semantics. | `assert.ErrorIs` |
| 13 | TestApplyOnPrimitiveT | §21.1 E12 | unit | `Apply[int]([]Option[int]{...})` compiles and produces configured int. | `assert.Equal` |
| 14 | TestApplyOnNonStructT | §21.1 E12 | unit | T is a `[]string` slice; option appends. | `assert.Equal` |
| 15 | TestApplyOptionPanicsPropagates | §21.1 E10 | unit | An option that calls `panic("x")` → Apply does not recover; panic visible. | `assert.Panics` (via runtime helper, see Bootstrap subset note below) |
| 16 | TestApplyOptionMutateThenError | §21.1 E9 | unit | Documented behavior: option mutates state then errors → partial state visible (transactional violation accepted). | `assert.NotEqual` (T != zero) |
| 17 | TestOptionFuncSatisfiesOption | §21.1 F2 | unit | `var _ Option[T] = OptionFunc[T](nil)` :  interface conformance. | bare type assertion |
| 18 | TestOptionFuncTypedNilApplyPanics | §21.1 F2 | unit | Calling `apply` on a `OptionFunc` whose underlying func is nil panics. (Edge :  added by Lynx.) | `assert.Panics` |
| 19 | TestStrictReturnsStrictMode | §21.1 F5 | unit | `Strict().strict == true` (verified indirectly via Apply behavior since field unexported). | composition with TestApplyStrictMultipleErrors |
| 20 | TestValidateNoValidators | §21.1 F7 | unit | `Validate(&t)` with empty validators → nil. | `assert.NoError` |
| 21 | TestValidateAllPass | §21.1 F7 | unit | Three validators all return nil → nil. | `assert.NoError` |
| 22 | TestValidateMultipleFail | §21.1 F7 | unit | Two of three fail → `errors.Join` of both; can `errors.Is`-detect both. | `assert.ErrorIs` (over join) |
| 23 | TestValidateNilTarget | §21.1 F13, E6 | unit | `Validate[T](nil, v)` → exact text `option: validate: target is nil`. | `assert.ErrorIs`, `assert.Equal` on `.Error()` |
| 24 | TestValidateNilValidatorSkipped | §21.1 F10, E5 | unit | `Validate(&t, nil, v)` runs only v. | `assert.NoError` (when v passes) |
| 25 | TestValidateAllNilValidators | §21.1 F10 | unit | All-nil validators → nil. | `assert.NoError` |
| 26 | TestValidateValidatorPanicsPropagates | §21.1 E11 | unit | Validator panics → Validate does not recover. | `assert.Panics` |
| 27 | TestRequiredPass | §21.1 F8 | unit | `Required("x", check-true)` returns Validator that yields nil. | `assert.NoError` |
| 28 | TestRequiredFail | §21.1 F8 | unit | `Required("x", check-false)` yields `option: required: field "x" not set`. | `assert.ErrorContains`, register-format check |
| 29 | TestRequiredQuotesFieldName | §21.1 F8 | unit | Field name with embedded quote/space rendered via `%q`. | `assert.Equal` on `.Error()` |
| 30 | TestRequiredGenericTLoadBearing | §23.17 | unit | Verifies the §23.17 fix: `Required[T]` getter form properly routes through *T. | `assert.NoError`, `assert.ErrorContains` |
| 31 | TestErrorRegisterConformanceOption | §21.1 NF3 | unit | Every `Error()` string from this package matches `^option: [a-z][^A-Z.]*$`. | regex check via `assert.Match` with `MatchRegex()` |
| 32 | BenchmarkApplyZeroAlloc_Happy | §21.1 NF1 | bench | 10 options, no errors, no Strict → `testing.AllocsPerRun` reports 0. | `testing.AllocsPerRun`, `assert.Equal` (in companion `TestApplyZeroAlloc`) |
| 33 | BenchmarkApplyOneOption | §21.1 NF1 | bench | Single-option happy path; per-op ns. | `testing.B` |
| 34 | BenchmarkApplyTenOptions | §21.1 NF1 | bench | 10 options happy path; per-op ns. | `testing.B` |
| 35 | BenchmarkApplyStrictTenOptions | §21.1 F5 | bench | Strict 10 options, all pass; allocs ≤ 1 (the `errs` slice never grows because no errors). | `testing.B` |
| 36 | BenchmarkApplyStrictWithFailures | §21.1 F5 | bench | Strict 10 options, 5 fail; allocates the join + errs slice. Documented expected number. | `testing.B` |
| 37 | BenchmarkValidate | §21.1 F7 | bench | 5 validators, all pass; per-op ns. | `testing.B` |
| 38 | BenchmarkRequired | §21.1 F8 | bench | Required validator hot path. | `testing.B` |
| 39 | TestApplyConcurrent | §21.1 NF2 | race | 100 goroutines call `Apply(samepts)` against local T each; runs under `-race`. | `concur.WaitGroup` is leaf-tier :  use stdlib `sync.WaitGroup`; `assert.NoError` |
| 40 | TestValidateConcurrent | §21.1 NF2 | race | 100 goroutines call `Validate(&t, vs...)` on independent t. | stdlib `sync.WaitGroup`, `assert.NoError` |
| 41 | PropertyApplyIdempotent | §21.1 F11 | property | For any list of pure idempotent options, `Apply(opts).Apply(opts)` == `Apply(opts)`. (Generative table-driven.) | `fixture/random` (kernel-allowed seedable rand), `assert.Equal` |
| 42 | PropertyApplyNilSkipPermutation | §21.1 F9 | property | Inserting nils anywhere in opts produces same result as opts without nils. | `assert.Equal` over permutations |
| 43 | PropertyApplyLastWins | §21.1 F11 | property | Setting same field N times → final value == last setter. | `assert.Equal` |
| 44 | PropertyValidateOrderInvariance | §21.1 F7 | property | `errors.Join`-result is set-equivalent regardless of validator order (under `errs.Chain`). | `assert.Subset` over `errs.Chain(err)` |
| 45 | ExampleApply | §21.1 NF7 | example | Runnable godoc example. | output-comment match |
| 46 | ExampleStrict | §21.1 NF7, F5 | example | Runnable Strict example. | output-comment match |
| 47 | ExampleValidate | §21.1 NF7 | example | Runnable Validate + Required example. | output-comment match |
| 48 | TestSurfaceClosed_OptionPackage | §21.1 NF8 | unit | `reflect`-based check that the package exports exactly 8 symbols (Option, OptionFunc, Apply, Mode, Strict, Validator, Validate, Required). Uses `go/types` to snapshot. | `fixture/golden` of the API snapshot |

### Bootstrap subset
`option` does not depend on `assert`, but its tests will use `assert`. Since `assert` (in turn) imports `option` for its `EqualOption` types, **we have a circular `_test` problem**: `assert`'s tests should not depend on `option`-with-assert-tests being green. The resolution:

- `option`'s tests use `assert` freely. The cycle is broken at the **`_test.go` boundary**: `assert/equal_test.go` does NOT import `option/` for setup of its own correctness; it uses bare `if` for the bootstrap.
- For `TestApplyOptionPanicsPropagates` (T#15) and similar panic checks, `assert.Panics` is itself in the bootstrap subset :  but those tests can be written with bare `defer func() { if r := recover(); r == nil { t.Fatal("expected panic") } }()`.

### Coverage target
- **Line coverage:** 100% (the package is ~200 LOC; everything is reachable)
- **Branch coverage:** 100%
- **Public-API coverage:** 100% (all 8 exports exercised)
- All untested branches in the implementation are an automatic CI failure.

### Edge cases not in the spec but worth testing
- **L-add-1:** A `OptionFunc[T]` whose underlying function is `nil` and is invoked → currently would panic with a nil-pointer dereference. Spec is silent; spec the behavior in the doc and test it (T#18).
- **L-add-2:** `Apply` with options that close over a goroutine-shared variable, racing under `-race` → covered by T#39, but with a write-then-read variant inside the option to catch any unintended sharing in the implementation.
- **L-add-3:** `Validate` on a target whose pointer changes mid-validation (validator that swaps `*t` to a new value) → documented as "validator should not mutate t"; assert it does not crash even if violated.
- **L-add-4:** `Required` with `name == ""` → produces `option: required: field "" not set`. Allowed; documented; tested.
- **L-add-5:** `Apply` with 10,000 nil-only options :  should remain O(n) and not allocate (NF1).
- **L-add-6:** Memory: `Apply`'s `errs` slice in default mode should remain `nil` (never allocated) when no error occurs :  verified by `AllocsPerRun == 0`.

---

## Package: `errs/`

### Test files
- `errs/wrap_test.go` :  Wrap, WithStackTrace, StackOf, *Wrapper
- `errs/join_test.go` :  Join semantics
- `errs/chain_test.go` :  Chain iterator over the error tree
- `errs/sentinel_test.go` :  Sentinel + register validator (incl. fuzz)
- `errs/isany_test.go`
- `errs/retryable_test.go`
- `errs/coded_test.go`
- `errs/fuzz_test.go` :  `FuzzSentinelRegister`
- `errs/property_test.go` :  Chain DFS, Wrap idempotency
- `errs/bench_test.go` :  Wrap allocations, Chain throughput
- `errs/example_test.go`

### Test matrix

| #  | Name | Spec ref | Type | Description | Test helpers used |
|----|------|----------|------|-------------|-------------------|
| 1  | TestWrapPreservesUnwrapChain | §21.2 F1 | unit | `errors.Is(Wrap(e, "x"), e) == true`. | `assert.True`, `assert.ErrorIs` |
| 2  | TestWrapNilReturnsNil | §21.2 F1, E1 | unit | `Wrap(nil, "x") == nil` (typed nil); calling further methods is safe. | `assert.Nil` |
| 3  | TestWrapErrorFormat | §21.2 F1 | unit | `Wrap(io.EOF, "pkg: act").Error() == "pkg: act: EOF"`. | `assert.Equal` |
| 4  | TestWrapEmptyPrefix | §21.2 E2 | unit | `Wrap(e, "").Error() == ": <e.Error()>"`. Documented permitted. | `assert.Equal` |
| 5  | TestWrapErrorsAs | §21.2 F1 | unit | `errors.As(Wrap(custom, "x"), &target)` finds target. | `assert.True`, `assert.ErrorAs` |
| 6  | TestWrapperUnwrapReturnsInner | §21.2 F3 | unit | `(*Wrapper).Unwrap() == innerErr`. | `assert.Equal` |
| 7  | TestWrapperErrorOnTypedNil | §21.2 F3, E4 | unit | `var w *Wrapper; w.Error() == ""`. | `assert.Equal` |
| 8  | TestWrapperUnwrapOnTypedNil | §21.2 F3, E5 | unit | `var w *Wrapper; w.Unwrap() == nil`. | `assert.Nil` |
| 9  | TestWithStackTraceCapturesFrames | §21.2 F2, F4 | unit | After `WithStackTrace()`, `StackOf(w)` returns ≥1 frame and the first frame's Function contains the test function name. | `assert.Greater(len, 0)`, `assert.Contains` |
| 10 | TestWithStackTraceFrameLimit | §21.2 NF2 | unit | Stack capture caps at 32 frames even with deeper recursion. | `assert.LessOrEqual(len, 32)` |
| 11 | TestWithStackTraceNilSafe | §21.2 F2, E3 | unit | `(*Wrapper)(nil).WithStackTrace() == nil`. | `assert.Nil` |
| 12 | TestWithStackTraceChainable | §21.2 F2 | unit | `Wrap(e, "x").WithStackTrace().WithStackTrace()` does not double-allocate or break. | `assert.NotNil`, `assert.Equal(len(StackOf), expected)` |
| 13 | TestStackOfNilErr | §21.2 E6 | unit | `StackOf(nil) == nil`. | `assert.Nil` |
| 14 | TestStackOfNoStackInChain | §21.2 E7 | unit | Chain with no `WithStackTrace` anywhere → `StackOf` returns nil. | `assert.Nil` |
| 15 | TestStackOfWalksChain | §21.2 F4 | unit | Inner `Wrap(...).WithStackTrace()`, outer plain `Wrap` → `StackOf` finds inner stack. | `assert.NotNil`, `assert.Greater(len, 0)` |
| 16 | TestJoinAllNil | §21.2 F5, E9 | unit | `Join(nil, nil) == nil`. | `assert.Nil` |
| 17 | TestJoinZeroArgs | §21.2 E8 | unit | `Join() == nil`. | `assert.Nil` |
| 18 | TestJoinSingleNonNilCollapses | §21.2 F5, E10 | unit | `Join(nil, e, nil) == e` (identity-equal). | `assert.True(err == e)` (pointer identity) |
| 19 | TestJoinMultipleNonNilUsesStdlib | §21.2 F5 | unit | `errors.Is` works for each input; type implements `Unwrap() []error`. | `assert.ErrorIs` for each input |
| 20 | TestJoinDropsNilsBeforeStdlibCall | §21.2 F5 | unit | `Join(e1, nil, e2)` produces a 2-error join, not 3. | reflection on `Unwrap() []error` len, `assert.Len` |
| 21 | TestChainNilErrYieldsNothing | §21.2 F6, E11 | unit | `range Chain(nil)` produces zero iterations. | counter + `assert.Equal(0)` |
| 22 | TestChainSingleErrorYieldsOne | §21.2 F6 | unit | `Chain(io.EOF)` yields exactly `[io.EOF]`. | slice collect + `assert.Equal` |
| 23 | TestChainLinearWrap | §21.2 F6 | unit | `Chain(Wrap(Wrap(e, "a"), "b"))` yields 3 errors in walk order. | slice collect, `assert.Equal(len, 3)` + first-yielded-is-self |
| 24 | TestChainOverErrorsJoin | §21.2 F6, E12 | unit | `Chain(errors.Join(a, b, c))` yields the join, then a, b, c (DFS). | `assert.Equal` over slice |
| 25 | TestChainNestedJoinDFS | §21.2 F6 | unit | `errors.Join(a, errors.Join(b, c))` → DFS yields `[join, a, innerJoin, b, c]`. | `assert.Equal` |
| 26 | TestChainEarlyTermination | §21.2 F6, NF3, E13 | unit | Receiver returns false after 2 yields → exactly 2 yields seen. | counter + `assert.Equal(2)` |
| 27 | TestChainComposesWithFluentTake | §21.2 NF3 | unit | (Cross-package smoke; lives in `fluent_integration_test.go`.) Bound iteration. | `assert.Len` after `fluent.Take(Chain(err), 5)` :  but fluent is leaf; for kernel test, simulate with a hand-rolled `for ... break`. |
| 28 | TestSentinelValid | §21.2 F7, E16 | unit | `Sentinel("pkg: cause")` constructs without panic; `.Error() == "pkg: cause"`. | `assert.NotPanics`, `assert.Equal` |
| 29 | TestSentinelTrailingEmptyAction | §21.2 E16 | unit | `Sentinel("pkg:")` valid (lowercase, has colon, no period). | `assert.NotPanics` |
| 30 | TestSentinelUppercasePanics | §21.2 E14, F7 | unit | `Sentinel("Pkg: cause")` → panics with explanatory text including `mongoose library register`. | `assert.PanicsWithMessage(contains "register")` |
| 31 | TestSentinelTrailingPeriodPanics | §21.2 E15 | unit | `Sentinel("pkg: cause.")` panics. | `assert.Panics` |
| 32 | TestSentinelNoColonPanics | §21.2 E15 | unit | `Sentinel("nocolon")` panics. | `assert.Panics` |
| 33 | TestSentinelEmptyPanics | §21.2 F7 | unit | `Sentinel("")` panics. | `assert.Panics` |
| 34 | TestSentinelNonAsciiUppercase | §21.2 F7 | unit | `Sentinel("pkg: Ünicode")` :  Ü is not ASCII A-Z; doc explicitly ASCII rule; should NOT panic. (Edge: confirms the implementation uses ASCII-only check, not Unicode.) | `assert.NotPanics` |
| 35 | FuzzSentinelRegister | §21.2 F7 | fuzz | Fuzz `validRegister` invariants: never accepts uppercase ASCII; never accepts trailing period; accepts iff pattern is satisfied. Confirms panic-vs-not is consistent. | `testing.F`, seed corpus from spec examples |
| 36 | TestIsAnyMatches | §21.2 F8 | unit | `IsAny(io.EOF, fs.ErrNotExist, io.EOF) == true`. | `assert.True` |
| 37 | TestIsAnyNoMatch | §21.2 F8, E18 | unit | `IsAny(io.EOF, fs.ErrNotExist) == false`. | `assert.False` |
| 38 | TestIsAnyZeroTargets | §21.2 E17 | unit | `IsAny(err) == false`. | `assert.False` |
| 39 | TestIsAnyNilErr | §21.2 E18 | unit | `IsAny(nil, t1, t2) == false`. | `assert.False` |
| 40 | TestIsAnyTraversesWrappedChain | §21.2 F8 | unit | `IsAny(Wrap(io.EOF, "x"), io.EOF) == true`. | `assert.True` |
| 41 | TestMarkRetryableNil | §21.2 F9, E19 | unit | `MarkRetryable(nil) == nil`. | `assert.Nil` |
| 42 | TestRetryableMarkRoundTrip | §21.2 F9, F10 | unit | `Retryable(MarkRetryable(e)) == true`. | `assert.True` |
| 43 | TestRetryableNoMarker | §21.2 F10 | unit | Plain error → `Retryable == false`. | `assert.False` |
| 44 | TestRetryableNilErr | §21.2 E20 | unit | `Retryable(nil) == false`. | `assert.False` |
| 45 | TestRetryableCustomImplementation | §21.2 F10, E21 | unit | A custom type with `Retryable() bool` returning true → detected without `MarkRetryable`. | `assert.True` |
| 46 | TestRetryableImplReturnsFalse | §21.2 F10 | unit | A custom type where `Retryable()` returns false → walks past it; no other markers → false. | `assert.False` |
| 47 | TestRetryableTraversesUnwrapChain | §21.2 F10 | unit | `Retryable(fmt.Errorf("x: %w", MarkRetryable(io.EOF))) == true`. | `assert.True` |
| 48 | TestRetryableMarkPreservesUnwrap | §21.2 F9 | unit | `errors.Is(MarkRetryable(io.EOF), io.EOF) == true`. | `assert.ErrorIs` |
| 49 | TestCodedDetectViaErrorsAs | §21.2 F11, F12 | unit | A custom error implementing `Coded` → `Code(err) == code`. | `assert.Equal` |
| 50 | TestCodeEmptyForNonCoded | §21.2 F12, E22 | unit | Error not implementing Coded → `Code(err) == ""`. | `assert.Equal` |
| 51 | TestCodeNilErr | §21.2 F12 | unit | `Code(nil) == ""`. | `assert.Equal` |
| 52 | TestCodeFirstInChain | §21.2 F12 | unit | Inner Coded with one code, outer Coded with another → returns first found by `errors.As` (which is outermost in stdlib semantics). Documented and tested. | `assert.Equal` |
| 53 | TestErrorRegisterConformance_errs | §21.2 NF4 | unit | All `errors.New`-equivalents and panic strings emitted by errs match `^errs:` register or are construction-time panics with explanatory text. | regex via `assert.Match` |
| 54 | BenchmarkWrapNoStack | §21.2 NF1 | bench | `Wrap(io.EOF, "x")` ≤ 1 alloc/op (the wrapper struct). | `testing.AllocsPerRun` |
| 55 | BenchmarkWrapWithStackTrace | §21.2 NF2 | bench | Records ns/op and allocs; documents the cost. | `testing.B`, `-benchmem` |
| 56 | BenchmarkChainLinear | §21.2 F6 | bench | Walk a 10-deep linear chain. | `testing.B` |
| 57 | BenchmarkChainOverJoin | §21.2 F6 | bench | Walk a 1-level join with 10 children. | `testing.B` |
| 58 | BenchmarkSentinel | §21.2 F7 | bench | One-time construction cost (panic check + errors.New). | `testing.B` |
| 59 | BenchmarkJoinSingleCollapse | §21.2 F5 | bench | `Join(nil, e, nil)` → 0 alloc (returns e directly). | `testing.AllocsPerRun == 0` |
| 60 | BenchmarkJoinMultiple | §21.2 F5 | bench | `Join(e1, e2, e3)` allocs equivalent to `errors.Join`. | benchstat vs stdlib `errors.Join` |
| 61 | BenchmarkRetryableWalk | §21.2 F10 | bench | 5-deep chain, last link retryable. | `testing.B` |
| 62 | BenchmarkCode | §21.2 F12 | bench | 5-deep chain, last link Coded. | `testing.B` |
| 63 | TestChainNoRaceConcurrent | §21.2 NF concurrency | race | 100 goroutines iterate `Chain(sharedErr)` concurrently. | stdlib `sync.WaitGroup` |
| 64 | TestRetryableNoRaceConcurrent | §21.2 NF concurrency | race | 100 goroutines call `Retryable(sharedErr)` concurrently. | stdlib `sync.WaitGroup` |
| 65 | TestSentinelConcurrentRegisterValidation | §21.2 F7 | race | Multiple init-time `Sentinel` calls in parallel goroutines. | stdlib `sync.WaitGroup` |
| 66 | PropertyChainStartsWithSelf | §21.2 F6 | property | For any non-nil err, the first yield of `Chain(err)` is err itself. | random error tree gen, `assert.Equal` |
| 67 | PropertyChainContainsAllUnwrapped | §21.2 F6 | property | `Chain(err)` contains every error reachable via repeated `errors.Unwrap` and `Unwrap() []error`. | random tree gen, `assert.Subset` |
| 68 | PropertyJoinIdempotent | §21.2 F5 | property | `Join(Join(a, b), c)` and `Join(a, b, c)` are semantically equivalent under `errs.Chain`. | `assert.Equal` over Chain-collected sets |
| 69 | PropertyWrapTransparentToErrorsIs | §21.2 F1 | property | For any sentinel s and arbitrary prefix string ps, `errors.Is(Wrap(s, p), s) == true` for all p in ps. | random prefix gen |
| 70 | PropertyMarkRetryableTransparent | §21.2 F9 | property | For any err: `errors.Is(MarkRetryable(err), err) == true`. | random err gen |
| 71 | ExampleWrap | §21.2 example | example | Runnable Wrap + .WithStackTrace example. | output-comment match |
| 72 | ExampleSentinel | §21.2 example | example | Runnable Sentinel declaration + use. | output-comment match |
| 73 | ExampleChain | §21.2 example | example | Runnable Chain + walk. | output-comment match |
| 74 | ExampleRetryable | §21.2 example | example | Runnable mark + check. | output-comment match |
| 75 | TestSurfaceClosed_ErrsPackage | §21.2 NF8 | unit | API snapshot: 12 exports (Wrap, *Wrapper, *Wrapper.{Error,Unwrap,WithStackTrace}, StackOf, Join, Chain, Sentinel, IsAny, MarkRetryable, Retryable, Coded, Code). | `fixture/golden` snapshot |

### Bootstrap subset
`errs` does not depend on `assert` directly; safe to use `assert` for all tests. The only nuance: panic-detection tests (`TestSentinelUppercasePanics` etc.) :  `assert.Panics` is bootstrap-tested itself, so for `errs` we can rely on it. **No bare-`if` needed.**

### Coverage target
- **Line coverage:** 100% (~250 LOC; everything reachable)
- **Branch coverage:** 100%
- **Public-API coverage:** 100% (all 12 exports)
- `validRegister` (unexported) must be 100% covered via `TestSentinel*` tests.

### Edge cases not in the spec but worth testing
- **L-add-1:** `Wrap(err, prefix)` where `prefix` contains a colon already (e.g., `"pkg: act:"`) :  does the "rejoin" produce a register-conforming string? Documented and tested.
- **L-add-2:** `Sentinel` with NUL byte in text → currently passes the validator (no rule against NUL). Spec gap; recommend adding NUL rejection to register or document allowance. Test fails-loud either way.
- **L-add-3:** `WithStackTrace()` called from a deferred function vs from main path :  frame[0] differs; documented and tested.
- **L-add-4:** `Chain` over a self-referential cycle (custom error whose `Unwrap()` returns itself) → infinite loop currently. Spec is silent. **Recommend cycle detection in `Chain`** matching `assert.Equal`'s discipline. New edge case for spec amendment.
- **L-add-5:** `MarkRetryable(MarkRetryable(e))` :  double-wrap. `Retryable` finds true on outer; `errors.Is(_, e) == true`; does double-wrap leak unbounded memory if applied in a retry loop? Test bounds it.
- **L-add-6:** Sentinel constructed from a string literal vs runtime-built `fmt.Sprintf` :  both should be supported; both valid-register strings should not panic. Tested.

---

## Package: `log/`

### Test files
- `log/level_test.go`
- `log/logger_test.go` :  Default, SetDefault, From, Inject
- `log/with_test.go` :  context-attached attrs
- `log/handler_test.go` :  NewHandler attribute order, color, etc.
- `log/json_handler_test.go`
- `log/redact_test.go`
- `log/color_test.go` :  ColorMode + env-var precedence
- `log/concurrent_test.go` :  race-detector tests
- `log/bench_test.go` :  D35 benchmarks
- `log/example_test.go`
- `log/golden_test.go` :  golden-file output for handler format

### Test matrix

| #  | Name | Spec ref | Type | Description | Test helpers used |
|----|------|----------|------|-------------|-------------------|
| 1  | TestLevelTraceValue | §21.3 F1 | unit | `LevelTrace == -8`. | `assert.Equal` |
| 2  | TestLevelDebugValue | §21.3 F1 | unit | `LevelDebug == slog.LevelDebug`. | `assert.Equal` |
| 3  | TestLevelInfoValue | §21.3 F1 | unit | `LevelInfo == slog.LevelInfo`. | `assert.Equal` |
| 4  | TestLevelNoticeValue | §21.3 F1 | unit | `LevelNotice == 2`. | `assert.Equal` |
| 5  | TestLevelWarnValue | §21.3 F1 | unit | `LevelWarn == slog.LevelWarn`. | `assert.Equal` |
| 6  | TestLevelErrorValue | §21.3 F1 | unit | `LevelError == slog.LevelError`. | `assert.Equal` |
| 7  | TestLevelTraceLabelTextHandler | §21.3 F1, E10 | unit | Mongoose text handler renders LevelTrace records with literal "TRACE" label. | `fixture.CaptureStdout` (or buffer), `assert.Contains` |
| 8  | TestLevelNoticeLabelTextHandler | §21.3 F1, E10 | unit | Renders "NOTICE". | buffer + `assert.Contains` |
| 9  | TestAllSixLevelsRender_Text | §21.3 F1 | unit | One log call at each level → six distinct lines with correct labels. | golden file via `fixture/golden` |
| 10 | TestAllSixLevelsRender_JSON | §21.3 F1 | unit | Same, JSON handler; valid JSON lines. | golden + `assert.JSONEq` |
| 11 | TestStdlibHandlerFallsBackForCustomLevels | §21.3 E11 | unit | Stdlib `slog.NewTextHandler` renders LevelTrace as `DEBUG-4` (documented). | `assert.Contains` |
| 12 | TestDefaultReturnsSlogDefault | §21.3 F2 | unit | `Default() == slog.Default()`. | `assert.Equal` (pointer) |
| 13 | TestSetDefaultRoundTrip | §21.3 F2 | unit | `SetDefault(l); Default() == l`. Cleanup restores. | `t.Cleanup` |
| 14 | TestFromEmptyCtxReturnsDefault | §21.3 F3, E2 | unit | `From(context.Background()) == slog.Default()`. | `assert.Equal` |
| 15 | TestFromNilCtxReturnsDefault | §21.3 F3, E1 | unit | `From(nil) == slog.Default()`. (Defensive.) | `assert.Equal` |
| 16 | TestInjectFromRoundTrip | §21.3 F4 | unit | `From(Inject(ctx, l)) == l`. | `assert.Equal` |
| 17 | TestInjectNilTreatedAsNoLogger | §21.3 E3 | unit | `From(Inject(ctx, nil)) == slog.Default()`. | `assert.Equal` |
| 18 | TestInjectReturnsNewCtx | §21.3 F4 | unit | `Inject` returns a different ctx (not equal to input). | `assert.NotEqual` |
| 19 | TestWithAttachesAttrs | §21.3 F5 | unit | `With(ctx, attr)` makes attr appear in subsequent `slog.InfoContext` records when handler is mongoose handler. | buffer + `assert.Contains` (key=val) |
| 20 | TestWithEmptyAttrsReturnsCtx | §21.3 E4 | unit | `With(ctx)` with no attrs returns ctx (or a thin wrapper that behaves identically). | `assert.NoError` (no panic), records still log |
| 21 | TestWithAccumulates | §21.3 F5 | unit | Two `With` calls → both attrs appear, in order. | buffer + assert order via index search |
| 22 | TestWithManyCallsNoAccidentalQuadratic | §21.3 E5 | unit | 1000 With calls; final record contains all 1000 attrs in attached order. | `assert.Len`, performance budget |
| 23 | TestWithThroughLogAttrs | §21.3 F11 | unit | `slog.LogAttrs(ctx, level, ...)` also picks up ctx attrs. | buffer + `assert.Contains` |
| 24 | TestNonMongooseHandlerIgnoresCtxAttrs | §21.3 E16 | unit | Stdlib slog text handler does not pick up `With`-attached attrs. Documented coupling. | buffer + `assert.NotContains` |
| 25 | TestNewHandlerDefaults | §21.3 F6, E6 | unit | No options → LevelInfo, no source, ColorAuto. | reflect on returned handler config (or behavioral test) |
| 26 | TestNewHandlerAttributeOrder | §21.3 F6, NF1 | unit | Attribute order: level, msg, package, op, error, then user attrs. | golden file |
| 27 | TestNewJSONHandlerAttributeOrder | §21.3 F7 | unit | Same order in JSON output. | `assert.JSONEq` against golden |
| 28 | TestWithLevelOptionFiltersBelow | §21.3 F8 | unit | `WithLevel(LevelWarn)` → DEBUG calls produce no output. | buffer + `assert.Equal(buf.Len(), 0)` |
| 29 | TestWithSourceAddsSource | §21.3 F8, T8 | unit | Source attr present in output. | `assert.Contains "source"` |
| 30 | TestWithColorAlways | §21.3 F8, F9 | unit | ColorAlways → ANSI escapes present even when writer is buffer. | `assert.Contains "\x1b["` |
| 31 | TestWithColorNever | §21.3 F9 | unit | ColorNever → no ANSI escapes. | `assert.NotContains "\x1b["` |
| 32 | TestColorAutoOnTTY | §21.3 F9, T3 | unit | TTY detected via fixture pty → escapes present. | `fixture/term.NewPTY` (in fixture pkg, leaf-allowed for kernel test? :  see Bootstrap discipline below) |
| 33 | TestColorAutoOffNonTTY | §21.3 E7, E15, T4 | unit | `bytes.Buffer` writer → no escapes. | `assert.NotContains "\x1b["` |
| 34 | TestNoColorEnvSuppresses | §21.3 F12, T5 | unit | `NO_COLOR=1` → escapes off even with ColorAlways requested. | `t.Setenv` (stdlib), buffer |
| 35 | TestGlacierNoColorEnvSuppresses | §21.3 F12, T6 | unit | `GLACIER_NO_COLOR=1` → escapes off. (Note: §23.18 renamed `MONGOOSE_NO_COLOR` → `GLACIER_NO_COLOR`.) | `t.Setenv`, buffer |
| 36 | TestGlacierNoColorBeatsColorAlways | §21.3 E8, T7 | unit | `GLACIER_NO_COLOR=1` + `ColorAlways` → no escapes. | `t.Setenv` |
| 37 | TestNoColorAndGlacierBothSet | §21.3 F12 | unit | Both set → off. | `t.Setenv` |
| 38 | TestColorPaletteMatchesSpec0001 | §21.3 NF6 | unit | Each level emits the documented 24-bit RGB triple (e.g., INFO → `\x1b[38;2;34;211;238m`). | `assert.Contains` per-level |
| 39 | TestColorEscapesPrecomputedNotFormatted | §21.3 NF1 | bench | `BenchmarkColorEscape` with `AllocsPerRun` → 0 (escape lookup is map/slice index, not `Sprintf`). | `testing.AllocsPerRun == 0` |
| 40 | TestRedactInTextOutput | §21.3 F10, T18 | unit | `slog.Any("password", Redact(secret))` → output contains `password=[REDACTED]`. | `assert.Contains` |
| 41 | TestRedactInJSONOutput | §21.3 F10, T19 | unit | JSON handler → `"password":"[REDACTED]"`. | `assert.JSONEq` |
| 42 | TestRedactNil | §21.3 E13 | unit | `Redact(nil)` → renders as `[REDACTED]`. | `assert.Contains` |
| 43 | TestRedactBeatsNestedLogValuer | §21.3 E14, T20 | unit | A type implementing `LogValuer` → wrapped with Redact → still `[REDACTED]`. | `assert.Contains "[REDACTED]"`, `assert.NotContains "real-value"` |
| 44 | TestRedactWithStdlibHandler | §21.3 F10 | unit | Redact also works with `slog.NewTextHandler` (because `LogValuer` is a stdlib contract). | `assert.Contains` |
| 45 | TestHandlerConcurrent | §21.3 NF5, T21 | race | 100 goroutines log via mongoose handler; output records well-formed. Run under `-race`. | stdlib `sync.WaitGroup`; `assert.Equal` line count after collection |
| 46 | TestWithCtxConcurrentNoRace | §21.3 NF4, §23.14 | race | 100 goroutines `With(parent, attr)` + log → no race. (Critical: ctx-attr accumulation must be lock-free or race-clean.) | stdlib `sync.WaitGroup` |
| 47 | TestNewHandlerAllocBudget | §21.3 NF1, §23.13 | bench | `BenchmarkInfoText` ≤ **3 allocs** beyond stdlib slog baseline (per §23.13 recalibration; was 1). | `testing.AllocsPerRun`, benchstat vs stdlib |
| 48 | BenchmarkInfoText | §21.3 NF2 | bench | Info, no attrs. Target ≤ 200 ns/op, ≤ 16 B/op, ≤ 1 alloc/op (within 10% of stdlib). | benchstat |
| 49 | BenchmarkInfoText3Attrs | §21.3 NF2 | bench | 3 attrs. Target ≤ 600 ns/op, ≤ 256 B/op, ≤ 3 allocs. | benchstat |
| 50 | BenchmarkInfoText10Attrs | §21.3 NF2 | bench | 10 attrs. | benchstat |
| 51 | BenchmarkInfoJSON | §21.3 NF2 | bench | JSON handler, no attrs. Target within 30% of stdlib JSON. | benchstat vs `slog.NewJSONHandler` |
| 52 | BenchmarkInfoJSON3Attrs | §21.3 NF2 | bench | 3 attrs JSON. | benchstat |
| 53 | BenchmarkInfoJSON10Attrs | §21.3 NF2 | bench | 10 attrs JSON. | benchstat |
| 54 | BenchmarkInfoCtxAttrs | §21.3 NF2 | bench | 5 attrs via `log.With`. Documents the ctx-attr cost vs inline. | benchstat |
| 55 | BenchmarkRedact | §21.3 NF2 | bench | `slog.Any("k", Redact(v))` cost. | `testing.B` |
| 56 | BenchmarkColorEscape | §21.3 NF1, NF2 | bench | Per-level escape lookup cost. | `testing.AllocsPerRun == 0` |
| 57 | TestErrorRegisterConformance_log | §21.3 NF | unit | If log emits any internal errors (it doesn't at v0; reserved), they conform. Currently asserts the package emits no errors. | `assert.NotPanics` |
| 58 | TestSurfaceClosed_LogPackage | §21.3 NF10 | unit | API snapshot: 11 exports + 6 level constants + 3 ColorMode constants. | `fixture/golden` snapshot |
| 59 | PropertyWithAccumulatesOrder | §21.3 F5 | property | For any sequence of N With calls, attrs appear in concatenation order in the final record. | random gen, golden compare |
| 60 | PropertyRedactNeverLeaksValue | §21.3 F10 | property | For any v, the rendered record never contains `fmt.Sprint(v)` (assuming v's string doesn't contain `[REDACTED]`). | random v gen, `assert.NotContains` |
| 61 | TestLogIntegrationWithTermColor | §21.3 NF6, §21.14 F4-F9 | integration | mongoose log handler uses `term/`-derived color escapes (post-§23.2 absorption of internal/ttyx). Verifies colors match `term.New().Foreground(term.Cyan).Render`. | `assert.Equal` on escape sequences |
| 62 | ExampleSetDefault | §21.3 example | example | Runnable. | output match |
| 63 | ExampleWithCtxAttrs | §21.3 example | example | Runnable. | output match |
| 64 | ExampleRedact | §21.3 example | example | Runnable. | output match |
| 65 | ExampleNewJSONHandler | §21.3 example | example | Runnable. | output match |
| 66 | TestHandlerCloseDiscipline | §23.16 | unit | log's handlers don't appear in §23.16 Close audit (they're stateless wrappers around `slog.Handler`). Confirm no `Close()` method on `*MongooseHandler` to enforce statelessness; if added, lifecycle-test it. | reflect-based check |

### Bootstrap subset
`log` imports `option` and `term` (post-§23.2). `log`'s tests use `assert` and `fixture` freely. **One care point:** if `term` color is mocked for color tests, the mock comes from `term/`'s own test fakes :  but kernel-tier rule means we use real `term` (proven by its own tests). For `TestColorAutoOnTTY` we need a PTY fixture; that lives in `fixture/term`. **Resolution:** kernel tests for log use a buffer for non-TTY assertions; the TTY case is covered by an integration test that lives in `term/` itself (or in `fixture/`).

### Coverage target
- **Line coverage:** ≥95% (some Windows-only / Unix-only paths are platform-skipped on the other)
- **Branch coverage:** ≥90%
- **Public-API coverage:** 100%
- Color/TTY detection branches: covered via env-var manipulation + buffer/PTY pairs.

### Edge cases not in the spec but worth testing
- **L-add-1:** A handler constructed with `WithLevel(slog.Leveler)` where Leveler is dynamic (changes mid-run): must respect new level on the next call. (Spec implies but doesn't test.)
- **L-add-2:** `With(ctx, slog.Group("nested", slog.String("k","v")))` :  group attrs round-trip correctly.
- **L-add-3:** Ctx attached attrs interaction with `slog.Logger.WithGroup("foo")` :  does the group prefix apply to ctx attrs too? Spec gap; needs explicit decision in test.
- **L-add-4:** A goroutine logs with ctx; ctx is cancelled; record still emits with attached attrs (ctx cancellation has no effect on log emission). Tested.
- **L-add-5:** `Redact(time.Time)` :  time has special slog handling; redact must override.
- **L-add-6:** Color-on-pipe-with-`isatty`-mocked-true: confirm we use the writer-fd-based check, not a global one. (Cross-platform.)
- **L-add-7:** Handler's `Enabled(ctx, level)` correctly returns false for filtered levels (so callers can `if h.Enabled(...)` cheaply).

---

## Package: `assert/` and `assert/require/`

### Test files
- `assert/equal_test.go` :  Equal across all type kinds (~30 cases; bootstrap-tested with bare-`if`)
- `assert/equal_options_test.go` :  IgnoreOrder, IgnoreCase, IgnoreWhitespace, WithDelta, IgnoreFields
- `assert/match_test.go` :  glob, regex, ignore-case
- `assert/ordering_test.go` :  Greater, Less, GreaterOrEqual, LessOrEqual
- `assert/indelta_test.go`
- `assert/jsoneq_test.go`
- `assert/bytes_test.go`
- `assert/subset_test.go`
- `assert/contains_test.go`
- `assert/len_test.go`
- `assert/eventually_test.go`
- `assert/error_test.go` :  Error, NoError, ErrorIs, ErrorAs
- `assert/nil_test.go` :  Nil, NotNil
- `assert/bool_test.go` :  True, False
- `assert/halt_test.go`
- `assert/must_test.go`
- `assert/diff_test.go` :  failure-message diffs
- `assert/concurrent_test.go`
- `assert/bench_test.go` :  D35 + §23.13 fast-path benchmarks
- `assert/fuzz_test.go` :  `FuzzMatchGlob`, `FuzzMatchRegex`
- `assert/property_test.go` :  reflexivity, symmetry
- `assert/require/require_test.go` :  every mirror function halts on failure
- `assert/example_test.go`

### Test matrix (assert)

| #  | Name | Spec ref | Type | Description | Test helpers used |
|----|------|----------|------|-------------|-------------------|
| **Equal :  bootstrap subset (uses bare-`if`)** ||||||
| 1  | TestEqual_Bootstrap_PrimitiveInt | §21.4 F2, F11; §23.5 | unit (bootstrap) | `Equal(t, 5, 5)` returns true; mockTB.errors == 0. Bare-`if` only. | mockTB recording, bare `if` |
| 2  | TestEqual_Bootstrap_PrimitiveString | §21.4 F2; §23.5 | unit (bootstrap) | `Equal(t, "a", "a") == true`. | mockTB, bare `if` |
| 3  | TestEqual_Bootstrap_NilNil | §21.4 E1 | unit (bootstrap) | `Equal(t, nil, nil) == true`. | bare `if` |
| 4  | TestEqual_Bootstrap_TypedNilNil | §21.4 E2 | unit (bootstrap) | `var a, b *T; Equal(t, a, b) == true`. | bare `if` |
| 5  | TestEqual_Bootstrap_Mismatch | §21.4 F2 | unit (bootstrap) | `Equal(t, 5, 4) == false`; mockTB.errors == 1. | bare `if` |
| 6  | TestEqual_Bootstrap_TypeMismatchAtTop | §21.4 E3 | unit (bootstrap; via `any`) | `Equal[any](t, (*A)(nil), (*B)(nil)) == false`. (Generic shell does not require T to be `any` here; tested with explicit `any`.) | bare `if` |
| 7  | TestPrimitiveFastPathBypass | §23.5, §23.13, §23.17 | bench | `BenchmarkEqualPrimitiveFastPath`: 50 ns/op, 0 allocs/op. Uses `testing.AllocsPerRun` to PROVE primitives bypass smart-equal. | `testing.AllocsPerRun == 0`, `testing.B.ReportAllocs` |
| 8  | TestPrimitiveFastPathTypeNotComparable | §23.5 | unit | If T is not comparable (e.g., `[]int`), the fast path is NOT taken; the slow path is exercised. Verified via instrumenting a sentinel inside slow-path. | mockTB; instrumentation hook (test-only) |
| **Equal :  composition tests (use `assert` itself)** ||||||
| 9  | TestEqualPointerDeref | §21.4 F11, E4, E5 | unit | `Equal(t, &T{x:1}, &T{x:1}) == true`; `&T{x:1}, &T{x:2} == false`. | `assert.True`, `assert.False`, mockTB |
| 10 | TestEqualSliceOrdered | §21.4 E6 | unit | `[]int{1,2,3}` vs `[]int{3,2,1}` → false default. | mockTB |
| 11 | TestEqualSliceIgnoreOrder | §21.4 F12, E6 | unit | Same with `IgnoreOrder()` → true. | mockTB |
| 12 | TestEqualSliceIgnoreOrderMultisetCount | §21.4 E7 | unit | `[]int{1,2,3}` vs `[]int{1,2,3,3}` w/ IgnoreOrder → false (multiset count). | mockTB |
| 13 | TestEqualMapDefault | §21.4 E8 | unit | Maps order-insensitive without options. | mockTB |
| 14 | TestEqualMapIgnoreCaseKeys | §21.4 F13, E9 | unit | Map keys differ in case; `IgnoreCase()` → true. | mockTB |
| 15 | TestEqualStructWithDelta | §21.4 F15, E12 | unit | Float fields tolerated within delta. | mockTB |
| 16 | TestEqualStructIgnoreFields | §21.4 F16 | unit | Field excluded from compare. | mockTB |
| 17 | TestEqualStructIgnoreFieldsRecursive | §21.4 F16 | unit | Nested struct's named field ignored at every level. | mockTB |
| 18 | TestEqualCyclic | §21.4 NF2, E10 | unit | Cycle detection: a→b→a vs identical → true; no stack overflow. | mockTB |
| 19 | TestEqualCustomMethod | §21.4 F11, E11 | unit | Type with `Equal(any) bool` method invoked; returns true → assert true. | mockTB |
| 20 | TestEqualCustomMethodReturnsFalse | §21.4 E11 | unit | Custom method returns false → assert false. | mockTB |
| 21 | TestEqualCustomMethodIgnoresStructFields | §21.4 E11 | unit | If custom method says equal, struct field comparison skipped. | mockTB |
| 22 | TestEqualNilVsNonNilDifferentTypes | §21.4 E1, E2, E3 | unit | Various nil-typed vs nil-untyped permutations. | table, mockTB |
| 23 | TestEqualInterfaceUnwrapping | §21.4 F11 | unit | `var a any = 5; var b any = 5; Equal == true`. | mockTB |
| 24 | TestEqualInterfaceWrappedDifferentDynamicTypes | §21.4 F11 | unit | `any(int(5))` vs `any(int64(5))` → false (types differ). | mockTB |
| 25 | TestEqualNaNVsNaN | §21.4 E13 | unit | NaN != NaN (Go semantics; documented). | mockTB |
| 26 | TestEqualWithDeltaNaNHandling | §21.4 E13 | unit | `WithDelta(0.1)` + NaN → still false (delta arithmetic on NaN is NaN). Documented. | mockTB |
| 27 | TestEqualIgnoreWhitespace | §21.4 F14 | unit | "hello\nworld" vs "hello   world" → true with IgnoreWhitespace. | mockTB |
| 28 | TestEqualChannelByIdentity | §21.4 F11 | unit | Same channel → true; different channels → false (no deeper semantics). | mockTB |
| 29 | TestEqualFuncByIdentity | §21.4 F11 | unit | Same func value → true; different (even semantically equal) → false. | mockTB |
| 30 | TestEqualLargeRecursive | §21.4 NF3 | unit | 1000-deep nested struct tree → no allocation beyond visited-set. | mockTB, `testing.AllocsPerRun` |
| 31 | TestNotEqualBasic | §21.4 F2 | unit | `NotEqual(t, 1, 2) == true`; `NotEqual(t, 1, 1) == false` and reports. | mockTB |
| **Bool + Nil + NotNil** ||||||
| 32 | TestTrue / TestFalse | §21.4 F3 | unit | Pass + fail paths. | mockTB |
| 33 | TestNilUntyped | §21.4 F3 | unit | `Nil(t, nil) == true`. | mockTB |
| 34 | TestNilTypedNil | §21.4 F3 | unit | `var p *T; Nil(t, p) == true` (typed-nil-aware). | mockTB |
| 35 | TestNotNil | §21.4 F3 | unit | `NotNil(t, &T{})` → true; `NotNil(t, nil)` → false. | mockTB |
| **Error family** ||||||
| 36 | TestNoError | §21.4 F3 | unit | `NoError(t, nil) == true`; `NoError(t, e) == false`. | mockTB |
| 37 | TestError | §21.4 F3 | unit | `Error(t, e) == true`; `Error(t, nil) == false`. | mockTB |
| 38 | TestErrorIs | §21.4 F3 | unit | Walks chain like `errors.Is`. | mockTB |
| 39 | TestErrorAs | §21.4 F3 | unit | Walks chain like `errors.As`. | mockTB |
| **Contains + Len** ||||||
| 40 | TestContainsString | §21.4 F3 | unit | `Contains(t, "hello world", "world") == true`. | mockTB |
| 41 | TestContainsSliceWithSmartEqual | §21.4 F3, F11 | unit | Slice contains element via smart-equal (struct member match). | mockTB |
| 42 | TestContainsMapKey | §21.4 F3 | unit | Map contains key. | mockTB |
| 43 | TestContainsWithIgnoreCaseOption | §21.4 F3, F13 | unit | `Contains(t, "ABC", "b", IgnoreCase()) == true`. | mockTB |
| 44 | TestLenSlice | §21.4 F3, §23.17 | unit | `Len(t, []int{1,2,3}, 3) == true`. (Reverted to non-generic per §23.17.) | mockTB |
| 45 | TestLenMap | §21.4 F3 | unit | Map length. | mockTB |
| 46 | TestLenString | §21.4 F3 | unit | String byte length. | mockTB |
| 47 | TestLenChan | §21.4 F3 | unit | Channel buffered length. | mockTB |
| 48 | TestLenNonContainer | §21.4 F3 | unit | `Len(t, 42, 3)` → reports type error. | mockTB |
| **Eventually** ||||||
| 49 | TestEventuallyPassesEarly | §21.4 E20 | unit | fn returns true on first poll → returns true quickly. | mockTB, `fixture.NewClock` |
| 50 | TestEventuallyTimeout | §21.4 E20 | unit | fn never returns true → reports "condition not met within X". | mockTB, `fixture.NewClock` |
| 51 | TestEventuallyHonorsInterval | §21.4 F3 | unit | fn polled at exact interval; deterministic via fake clock. | `fixture.NewClock` |
| **Match (glob/regex)** ||||||
| 52 | TestMatchGlob | §21.4 F4, E14 | unit | "hello world" matches "hello *". | mockTB |
| 53 | TestMatchGlobAnchors | §21.4 F4 | unit | Glob `*` is anchored at full string by default; "abc" does NOT match "a" without `*`. | mockTB |
| 54 | TestMatchGlobSingleChar | §21.4 F4 | unit | "abc" matches "a?c". | mockTB |
| 55 | TestMatchRegex | §21.4 F4, E16 | unit | "abc" matches `^[a-c]+$` with `MatchRegex()`. | mockTB |
| 56 | TestMatchIgnoreCase | §21.4 F4, E15 | unit | "Hello" matches "hello" with `MatchIgnoreCase()`. | mockTB |
| 57 | TestMatchInvalidRegex | §21.4 F4 | unit | `MatchRegex()` with malformed pattern → reports compile error. | mockTB |
| 58 | TestMatchSpecialCharsEscapedInGlob | §21.4 F4 | unit | "a.b" does NOT match "a.b" as glob (`.` is literal in glob, but the fixture should use a glob that treats it literally :  confirm). | mockTB |
| 59 | FuzzMatchGlob | §21.4 F4 | fuzz | Random pattern + input pairs; never panics, returns deterministic bool. | `testing.F` |
| 60 | FuzzMatchRegex | §21.4 F4 | fuzz | Random regex + input; if compile fails, reports cleanly; never panics. | `testing.F` |
| **Ordering + Tolerance** ||||||
| 61 | TestGreater | §21.4 F5 | unit | `Greater[int](t, 5, 4) == true`. | mockTB |
| 62 | TestLess | §21.4 F5 | unit | Mirror. | mockTB |
| 63 | TestGreaterOrEqual | §21.4 F5 | unit | Both sides of equality. | mockTB |
| 64 | TestLessOrEqual | §21.4 F5 | unit | Mirror. | mockTB |
| 65 | TestOrderingOnString | §21.4 F5 | unit | Lexicographic. | mockTB |
| 66 | TestOrderingOnFloat | §21.4 F5 | unit | NaN behavior documented. | mockTB |
| 67 | TestInDeltaFloat64 | §21.4 F6 | unit | `InDelta(t, 1.0001, 1.0, 0.001) == true`. | mockTB |
| 68 | TestInDeltaFloat32 | §21.4 F6 | unit | `~float32` works via type-set constraint. | mockTB |
| 69 | TestInDeltaCustomFloat | §21.4 F6 | unit | A `type MyFloat float64` instance works (`~float64` constraint). | mockTB |
| 70 | TestInDeltaNaN | §21.4 E13 | unit | NaN with delta → false. | mockTB |
| **JSONEq + BytesEq + Subset** ||||||
| 71 | TestJSONEqIdentical | §21.4 F7, E17 | unit | Same JSON → true. | mockTB |
| 72 | TestJSONEqKeyOrderInvariant | §21.4 F7, E17 | unit | Different key order → true (JSON object equality). | mockTB |
| 73 | TestJSONEqArrayOrderMatters | §21.4 F7 | unit | Array order matters by default. | mockTB |
| 74 | TestJSONEqArrayIgnoreOrder | §21.4 F7 | unit | With `IgnoreOrder()` → true. | mockTB |
| 75 | TestJSONEqIgnoreCaseValues | §21.4 F7 | unit | Strings with `IgnoreCase()` → true. | mockTB |
| 76 | TestJSONEqMalformedGot | §21.4 E18 | unit | Got is not JSON → reports parse error, returns false. | mockTB |
| 77 | TestJSONEqMalformedWant | §21.4 E18 | unit | Want is not JSON → reports. | mockTB |
| 78 | TestBytesEqIdentical | §21.4 F8 | unit | Identical bytes. | mockTB |
| 79 | TestBytesEqDifferent | §21.4 F8 | unit | Different bytes. | mockTB |
| 80 | TestBytesEqEmpty | §21.4 F8 | unit | `BytesEq(t, []byte{}, []byte{}) == true`. | mockTB |
| 81 | TestBytesEqNilVsEmpty | §21.4 F8 | unit | nil and empty bytes are equal (Go semantics for `bytes.Equal`). Documented. | mockTB |
| 82 | TestSubset | §21.4 F9, E19 | unit | `Subset(t, [1,2,3,4], [2,3]) == true`. | mockTB |
| 83 | TestSubsetMissingElement | §21.4 F9 | unit | `Subset([1,2], [3]) == false`. | mockTB |
| 84 | TestSubsetSmartEqual | §21.4 F9 | unit | Subset of struct slices uses smart-equal. | mockTB |
| 85 | TestSubsetEmptyWantAlwaysTrue | §21.4 F9 | unit | `Subset(got, [])` → true. | mockTB |
| **Halt + Must** ||||||
| 86 | TestHaltCallsFailNow | §21.4 F10, E23 | unit | `Halt(t)` invokes `t.FailNow`. | mockTB; spy on FailNow |
| 87 | TestMustReturnsValue | §21.4 F19, E21 | unit | `Must[T](v, nil) == v`. | `assert.Equal` |
| 88 | TestMustPanicsOnError | §21.4 F19, E22 | unit | `Must(v, err)` panics with err embedded. | `assert.Panics`, `assert.ErrorIs` (in recover) |
| 89 | TestMust2BothValuesReturned | §21.4 F20 | unit | `Must2[A,B](a, b, nil) == (a, b)`. | `assert.Equal` |
| 90 | TestMust2PanicsOnError | §21.4 F20 | unit | `Must2(_, _, err)` panics. | `assert.Panics` |
| 91 | TestMustfFalseCondPanics | §21.4 F21, E22 | unit | `Mustf(false, "fmt %d", 5)` panics with formatted msg. | `assert.PanicsWithMessage` |
| 92 | TestMustfTrueCondNoPanic | §21.4 F21 | unit | `Mustf(true, ...)` no panic. | `assert.NotPanics` |
| **Failure messages + diff** ||||||
| 93 | TestDiffPrimitive | §21.4 F18, NF5 | unit | "Equal failed: got=42 want=41." (CLI register). | mockTB recording errMsg, `assert.Equal` |
| 94 | TestDiffSlices | §21.4 F18, T21 | unit | Diff has `+`/`-`/`~` line markers. | golden via `fixture/golden` |
| 95 | TestDiffMaps | §21.4 F18 | unit | Map diff per-key. | golden |
| 96 | TestDiffStructs | §21.4 F18, T22 | unit | Struct diff per-field. | golden |
| 97 | TestDiffTTYColor | §21.4 F18, T23 | unit | When TTY: cyan add, rose remove ANSI escapes present. (Cross-platform: skip on Windows non-ANSI cmd.) | `fixture.NewPTY` (kernel can use it via build tag boundary), `assert.Contains "\x1b["` |
| 98 | TestDiffNonTTYNoColor | §21.4 F18 | unit | Buffer writer → no escapes. | `assert.NotContains "\x1b["` |
| 99 | TestErrorMessageInCliRegister | §21.4 NF5 | unit | All assert failure messages capitalized, period-terminated. | regex `^[A-Z][^:]+:.+\.$` |
| **Concurrency + benchmarks** ||||||
| 100 | TestConcurrentAssertions | §21.4 NF4, T27, §23.14 | race | 100 goroutines each call `assert.Equal(t, ...)` against own mockTB; runs under `-race`. Confirms global-state-free. | stdlib `sync.WaitGroup` |
| 101 | TestTBConcurrentErrorf | §23.14 | race | Real `*testing.T`'s `Errorf` is goroutine-safe; verify mongoose's `assert` does not introduce additional races. (Lynx-specific: §23.14 mock/httpmock race test parallel.) | sub-test running parallel goroutines, real `*testing.T` |
| 102 | BenchmarkEqualPrimitive | §21.4 NF1, §23.5, §23.13 | bench | Primitive int Equal: 50 ns/op zero alloc (fast path). | `testing.B`, `testing.AllocsPerRun` |
| 103 | BenchmarkEqualSmallStruct | §21.4 NF1 | bench | Struct with 5 fields. | `testing.B` |
| 104 | BenchmarkEqualLargeSlice | §21.4 NF1, §23.13 | bench | 1000 ints; ≤ 200 µs/op for IgnoreOrder. | `testing.B` |
| 105 | BenchmarkEqualLargeMap | §21.4 NF1 | bench | 1000-entry map. | `testing.B` |
| 106 | BenchmarkMatchGlob | §21.4 NF1 | bench | Glob match. | `testing.B` |
| 107 | BenchmarkMatchRegex | §21.4 NF1 | bench | Regex match (with cache). | `testing.B` |
| 108 | BenchmarkContainsSlice | §21.4 NF1 | bench | Slice contains 100 elements. | `testing.B` |
| 109 | BenchmarkContainsString | §21.4 NF1 | bench | String contains. | `testing.B` |
| 110 | BenchmarkJSONEq | §21.4 NF1 | bench | 1 KB JSON. | `testing.B` |
| 111 | BenchmarkSmartEqualSlowPath | §23.13 | bench | Non-fast-path: ≤ 200 ns/op. | `testing.B` |
| **Property-based** ||||||
| 112 | PropertyEqualReflexive | §21.4 F2 | property | For any x: `Equal(t, x, x) == true`. (Random values across types.) | random gen |
| 113 | PropertyEqualSymmetric | §21.4 F2 | property | `Equal(t, x, y) == Equal(t, y, x)`. | random gen |
| 114 | PropertyEqualTransitiveOnPrimitives | §21.4 F2 | property | If `Equal(a,b) && Equal(b,c)` then `Equal(a,c)`. | random gen |
| 115 | PropertyMatchEmptyPatternNeverMatches | §21.4 F4 | property | `Match(t, anyString, "")` returns false (or :  per spec :  empty pattern = empty match-only). Pin the behavior in the test. | random strings |
| 116 | PropertySubsetReflexive | §21.4 F9 | property | `Subset(t, x, x) == true`. | random gen |
| 117 | PropertyJSONEqEqualToReorderedSerialization | §21.4 F7 | property | Marshal a struct twice; reorder keys; `JSONEq == true`. | random struct gen |
| 118 | TestSurfaceClosed_AssertPackage | §21.4 NF11 | unit | API snapshot. | `fixture/golden` |
| 119 | ExampleEqualSmart | §21.4 example | example | Runnable. | match |
| 120 | ExampleIgnoreOrder | §21.4 example | example | Runnable. | match |
| 121 | ExampleIgnoreFields | §21.4 example | example | Runnable. | match |
| 122 | ExampleMatchRegex | §21.4 example | example | Runnable. | match |
| 123 | ExampleJSONEq | §21.4 example | example | Runnable. | match |
| 124 | ExampleMust | §21.4 example | example | Runnable. | match |

### Test matrix (assert/require)

| #  | Name | Spec ref | Type | Description | Test helpers used |
|----|------|----------|------|-------------|-------------------|
| R1 | TestRequireEqualHaltsOnFailure | §21.4 F22, E24, T20 | unit | `require.Equal(mockTB, 1, 2)` calls `mockTB.Errorf` then `mockTB.FailNow`. Subsequent code in the test goroutine does not run. | mockTB; spy on FailNow + sub-goroutine to verify halt |
| R2 | TestRequireEqualPassNoHalt | §21.4 F22 | unit | `require.Equal(t, 1, 1)` returns true; no FailNow called. | mockTB |
| R3 | TestRequireForEveryAssertMirror | §21.4 F22, §23.17 | unit | Generated table-driven test: for each assert.X there's a require.X with identical signature; both behave equivalently on pass; require halts on fail. | reflection over package symbols |
| R4 | TestRequireEqualGenericMirror | §23.17 | unit | `require.Equal[int]`, `require.Equal[string]` :  verify generic-for-generic mirror per §23.17. | composition |
| R5 | TestRequireGreater | §21.4 F22 | unit | Mirror. | mockTB |
| R6 | TestRequireMatch | §21.4 F22 | unit | Mirror. | mockTB |
| R7 | TestRequireJSONEq | §21.4 F22 | unit | Mirror. | mockTB |
| R8 | TestRequireImportSurface | §21.4 NF8 | unit | `require` imports only `assert`; verified via `go/packages` introspection. | `fixture/golden` import-list |
| R9 | TestRequireSurfaceClosed | §21.4 NF11 | unit | API snapshot mirrors assert exactly (minus runtime helpers). | golden |

### Bootstrap subset (CRITICAL DISCIPLINE FOR `assert/`)

Testing assert with assert is the chicken-and-egg. Resolution:

1. **Bootstrap subset** :  these tests use **bare-`if`** + a hand-rolled `mockTB`:
   - `TestEqual_Bootstrap_PrimitiveInt` (#1)
   - `TestEqual_Bootstrap_PrimitiveString` (#2)
   - `TestEqual_Bootstrap_NilNil` (#3)
   - `TestEqual_Bootstrap_TypedNilNil` (#4)
   - `TestEqual_Bootstrap_Mismatch` (#5)
   - `TestEqual_Bootstrap_TypeMismatchAtTop` (#6)

   These prove that `assert.Equal` correctly returns true/false and correctly invokes `mockTB.Errorf` on mismatches. They are written:
   ```go
   func TestEqual_Bootstrap_PrimitiveInt(t *testing.T) {
       mt := &mockTB{}
       got := Equal(mt, 5, 5)
       if !got { t.Fatalf("Equal(5,5) = false, want true") }
       if mt.errorfCalls != 0 { t.Fatalf("Errorf called %d times, want 0", mt.errorfCalls) }
   }
   ```

2. **`mockTB`** is a hand-rolled struct in `assert/internal_test.go` implementing `TB`. It is also bare-`if`-tested (round-trip Errorf/FailNow recording).

3. **All other assert tests** can use `assert.Equal` to compare `mockTB.errorfCalls`, `mockTB.lastMessage`, etc., once #1–#6 are proven.

4. **`require` tests** can use `assert.Equal` freely (since require is built on assert and assert is bootstrap-tested).

5. **`Must/Must2/Mustf` tests** use bare `defer recover()` for panic detection (since `assert.Panics` is part of the bootstrap-tested assertion family :  written but tested with bare `if` in the bootstrap layer).

### Coverage target
- **Line coverage:** 100% (the package is the test-helper kernel; gaps are unacceptable)
- **Branch coverage:** ≥98% (some unreachable defensive branches in smart-equal kind switch)
- **Public-API coverage:** 100% :  explicitly verified per #R3 reflection table.
- **Smart-equal kind coverage:** every `reflect.Kind` reachable from a typed value has at least one test (Pointer, Map, Slice, Array, Struct, String, Float32, Float64, Int*, Uint*, Bool, Interface, Chan, Func).

### Edge cases not in the spec but worth testing
- **L-add-1:** `Equal` on `unsafe.Pointer` :  Spec is silent. Document and test.
- **L-add-2:** `Equal` on a struct with unexported fields :  `reflect.DeepEqual` comparison rules apply; document.
- **L-add-3:** `Equal` on a map with `NaN` keys :  `NaN` keys are unreachable in regular maps; document and test panic / no-panic.
- **L-add-4:** Smart equal on a slice of `*T` where pointers differ but values equal → smart equal dereferences (recursion). Test.
- **L-add-5:** `IgnoreFields` with field name that doesn't exist on the struct → silently ignored or error? Spec gap. **Recommend explicit error**, test it.
- **L-add-6:** `WithDelta(d)` where `d < 0` → use absolute value or reject? Spec gap. Test the chosen behavior.
- **L-add-7:** `Eventually` where `fn` panics → propagate or treat as condition-false? Spec gap. Test.
- **L-add-8:** `Match(t, "", "")` :  empty pattern + empty string. Pin behavior.
- **L-add-9:** Concurrent `assert.Equal(t, ...)` from goroutines on the SAME `*testing.T` :  `t.Errorf` is documented goroutine-safe; verify our wrapper doesn't introduce races (T#101).
- **L-add-10:** `Must` with `err` that wraps a panic-caused error :  does the panic value carry the wrap chain? Test that `errors.Is(recover()-err, originalErr)` works.
- **L-add-11:** `Must2` generic instantiation with same type for A and B (`Must2[int, int]`) :  confirm compiles and works.
- **L-add-12:** `assert.Len(t, (chan int)(nil), 0)` :  nil channel; len returns 0; test it doesn't panic.
- **L-add-13:** `assert.JSONEq` with embedded `null` values vs missing keys :  JSON semantic distinction. Pin.

---

## Package: `term/`

### Test files
- `term/capability_test.go` :  Capability detection across writers and env vars
- `term/color_test.go` :  RGB, Hex, named colors
- `term/style_test.go` :  chaining, Render, Sprint, Fprint
- `term/glyph_test.go` :  registry, fallback, register-glyph, fuzz
- `term/box_test.go` :  beauty: Box options
- `term/beauty_test.go` :  Center, Justify, Pad, Truncate, Wrap, Columns, Banner
- `term/prompt_test.go` :  Prompt, Password, Confirm, Select, MultiSelect (PTY-driven)
- `term/animator_test.go` :  frame loop, log buffering, back-pressure
- `term/spinner_test.go`
- `term/progress_test.go`
- `term/statusbar_test.go`
- `term/download_test.go`
- `term/concurrent_test.go` :  race-detector tests
- `term/lifecycle_test.go` :  Close idempotency per §23.16
- `term/fuzz_test.go` :  `FuzzGlyphLookup`, `FuzzAnsiInjection`
- `term/property_test.go` :  Style.Render round-trip, capability invariants
- `term/bench_test.go` :  D35 benchmarks
- `term/golden_test.go` :  Box / Banner / Progress visual goldens
- `term/example_test.go`
- `term/cross_platform_test.go` :  Linux/macOS/Windows-specific paths
- `term/raw_mode_test.go` :  terminal raw-mode discipline

### Test matrix

| #  | Name | Spec ref | Type | Description | Test helpers used |
|----|------|----------|------|-------------|-------------------|
| **Capability detection** ||||||
| 1  | TestCapabilityIODiscard | §21.14 E1, F1 | unit | `Capability(io.Discard)` → IsTTY=false, ColorNone, w/h=0. | `assert.Equal` on struct |
| 2  | TestCapabilityBytesBuffer | §21.14 F1 | unit | bytes.Buffer → non-TTY, ColorNone. | `assert.Equal` |
| 3  | TestCapabilityRealTTYUnix | §21.14 F1 | unit (build-tag !windows) | PTY pair → IsTTY=true, color detected. | `fixture.NewPTY` |
| 4  | TestCapabilityRealTTYWindows | §21.14 F1 | unit (build-tag windows) | Windows console handle → IsTTY=true (when applicable). | `fixture.NewWindowsConsole` mock |
| 5  | TestCapabilityCOLORTERMTruecolor | §21.14 F2, T1 | unit | `COLORTERM=truecolor` env → `Color24Bit`. | `t.Setenv` |
| 6  | TestCapabilityCOLORTERM256 | §21.14 F2 | unit | `COLORTERM=256` → `Color256`. | `t.Setenv` |
| 7  | TestCapabilityTERMxterm | §21.14 F2 | unit | `TERM=xterm-256color` → `Color256`. | `t.Setenv` |
| 8  | TestCapabilityTERMdumb | §21.14 F2 | unit | `TERM=dumb` → `ColorNone`. | `t.Setenv` |
| 9  | TestCapabilityWTSession | §21.14 F2 | unit | `WT_SESSION` set (Windows Terminal) → `Color24Bit`. | `t.Setenv` |
| 10 | TestCapabilityNoColor | §21.14 F3, T2 | unit | `NO_COLOR=1` → `NoColorEnv=true`, color suppressed. | `t.Setenv` |
| 11 | TestCapabilityGlacierNoColor | §21.14 F3, T3 | unit | `GLACIER_NO_COLOR=1` → suppressed. | `t.Setenv` |
| 12 | TestCapabilityGlacierBeatsNoColor | §21.14 F3 | unit | Both set → off. (Behavior consistent with log NF6.) | `t.Setenv` |
| 13 | TestCapabilityCachedPerWriter | §21.14 NF6 | bench | Two `Capability(w)` calls → second is zero-allocation. | `testing.AllocsPerRun(2nd call) == 0` |
| 14 | TestCapabilityWidthFromTermSize | §21.14 F1 | unit | TERM size set via env or syscall → `Width`/`Height` populated. | `fixture.NewPTY` with size set |
| **Color** ||||||
| 15 | TestRGBConstructor | §21.14 F4 | unit | `RGB(255, 0, 0)` → Color with R=255, G=0, B=0. | `assert.Equal` (likely via Render comparison since fields private) |
| 16 | TestHexConstructorValid | §21.14 F4 | unit | `Hex("#22D3EE")` → Color matching Cyan. | `assert.NoError`, `assert.Equal` |
| 17 | TestHexConstructorShortForm | §21.14 F4 | unit | `Hex("#fff")` → 0xFF, 0xFF, 0xFF. | `assert.Equal` |
| 18 | TestHexConstructorInvalid | §21.14 F4 | unit | `Hex("not-a-color")` → typed error in register format. | `assert.ErrorContains "term:"` |
| 19 | TestHexConstructorMissingHash | §21.14 F4 | unit | `Hex("22D3EE")` :  accept or reject? Spec gap. Pin behavior. | `assert.Error` or `assert.NoError` per pinned |
| 20 | TestNamedColorsMatchSpec0001 | §21.14 F5 | unit | `Cyan`, `Teal`, etc. match the spec-0001 hex codes. | `assert.Equal` (compare via Hex round-trip) |
| 21 | TestGradientStops | §21.14 F5 | unit | `Cyan100`, `Cyan700`, etc. exist and have distinct values. | `assert.NotEqual` |
| **Style** ||||||
| 22 | TestStyleNewIsZero | §21.14 F6 | unit | `New()` returns Style with no attrs; Render = passthrough. | `assert.Equal(plain, Render(plain))` (with non-color writer) |
| 23 | TestStyleForegroundChain | §21.14 F6 | unit | `New().Foreground(Cyan).Render("x")` emits cyan ANSI. | `assert.Contains "\x1b[38;2;34;211;238m"` |
| 24 | TestStyleBackground | §21.14 F6 | unit | Background sets bg ANSI. | `assert.Contains "\x1b[48;2;"` |
| 25 | TestStyleBoldItalicUnderlineDimStrike | §21.14 F6 | unit | Each attribute. | table-driven |
| 26 | TestStyleChainComposition | §21.14 F6 | unit | `Bold().Italic().Underline()` produces 3 ANSI codes. | regex count |
| 27 | TestStyleImmutability | §21.14 F6 | unit | `s := New().Bold(); s2 := s.Italic()` → s does NOT have italic. | `assert.NotEqual(s.Render, s2.Render)` |
| 28 | TestStyleRenderNoColorWriter | §21.14 F7, E2 | unit | Render to non-color-supporting writer → plain text, no escapes. | `assert.NotContains "\x1b["` |
| 29 | TestStyleRenderResetTerminal | §21.14 F7 | unit | Render output ends with `\x1b[0m` reset. | `assert.HasSuffix` |
| 30 | TestSprintMatches | §21.14 F8 | unit | `Sprint(s, t)` = `s.Render(t)`. | `assert.Equal` |
| 31 | TestFprintWritesToWriter | §21.14 F8 | unit | `Fprint(w, s, t)` writes Render output to w. | buffer + `assert.Equal` |
| 32 | TestStyleEscapesPrecomputedCached | §21.14 F9, NF1 | bench | First Render allocates; second Render zero alloc (cache hit). | `testing.AllocsPerRun` |
| 33 | TestStyleConflictingAttrs | §21.14 E15 | unit | `Bold().Dim()` → both ANSI codes emitted; behavior documented. | `assert.Contains "\x1b[1m"` AND `\x1b[2m` |
| **Glyphs** ||||||
| 34 | TestGlyphCheckUTF8 | §21.14 F10, F11 | unit | Glyph("check") on UTF-8 writer → `✓`. | `assert.Equal "✓"` |
| 35 | TestGlyphCheckASCIIFallback | §21.14 F10, T5 | unit | Glyph("check") on non-UTF8 writer → `[OK]`. | `assert.Equal "[OK]"` |
| 36 | TestGlyphAllBuiltInsPresent | §21.14 F11 | unit | All ~30 named glyphs from spec resolve to non-empty UTF-8 + non-empty ASCII. | table over fixed list |
| 37 | TestGlyphSpinnerFrames | §21.14 F11 | unit | `spinner_braille_0` ... `spinner_braille_7` all defined. | table |
| 38 | TestGlyphNonexistent | §21.14 E3 | unit | `Glyph("nonexistent")` → empty string + debug-level log warning. | mock slog handler in test, `assert.Empty` |
| 39 | TestRegisterGlyphSuccess | §21.14 F12 | unit | `RegisterGlyph("custom", "🎯", "[*]")` returns nil; subsequent `Glyph("custom")` works. | `assert.NoError`, `assert.Equal` |
| 40 | TestRegisterGlyphDuplicate | §21.14 F12, E4 | unit | Re-register → typed error `term: glyph "<name>" already registered`. | `assert.ErrorContains` |
| 41 | TestRegisterGlyphInvalidName | §23.9 #26 | unit | Names not matching `^[a-z][a-z0-9_]*$` → typed error. | `assert.Error` |
| 42 | TestRegisterGlyphLengthCap | §23.9 #26 | unit | Names > 64 bytes → typed error. | `assert.Error` |
| 43 | FuzzGlyphRegistration | §21.14 F12; §23.9 #26 | fuzz | Random names :  never panics; only valid-pattern names succeed. | `testing.F`, seed corpus |
| 44 | FuzzGlyphLookup | §21.14 F10 | fuzz | Random names → never panics; returns empty for unregistered. | `testing.F` |
| 45 | TestGlyphsEnumeration | §21.14 F13 | unit | `Glyphs()` returns all built-in + registered. | `assert.GreaterOrEqual(len, 30)` |
| 46 | TestRegisterGlyphConcurrent | §21.14 NF5 | race | 50 goroutines register distinct glyphs concurrently → all succeed; no race. | stdlib `sync.WaitGroup` |
| **Beauty writer** ||||||
| 47 | TestBoxRoundedCorners | §21.14 F14, T6 | unit | Box with rounded corners renders `╭╮╰╯` (or ASCII fallback). | golden via `fixture.Golden` |
| 48 | TestBoxSharpCorners | §21.14 F14 | unit | Sharp corners. | golden |
| 49 | TestBoxDoubleBorders | §21.14 F14 | unit | Double borders `╔╗╚╝`. | golden |
| 50 | TestBoxWithTitle | §21.14 F14 | unit | Title rendered in top-border slot. | golden |
| 51 | TestBoxWithTitleStyle | §21.14 F14 | unit | Title styled via Style. | regex check on ANSI escapes |
| 52 | TestBoxWithPadding | §21.14 F14 | unit | Padding adds whitespace per side. | golden |
| 53 | TestBoxWithBorderStyle | §21.14 F14 | unit | Border colored. | regex check |
| 54 | TestBoxASCIIFallback | §21.14 F14 | unit | Non-UTF8 → ASCII border chars. | golden |
| 55 | TestBoxWidthExceedsTerminal | §21.14 E5 | unit | Content longer than terminal width → wraps via `Wrap` before boxing. | `assert.LessOrEqual(maxLineLen, termWidth)` |
| 56 | TestBoxEmptyText | §21.14 F14 | unit | Empty text → minimal box. | golden |
| 57 | TestCenter | §21.14 F15, T7 | unit | "abc" in width 7 → "  abc  " (or off-by-one rounding choice; pinned). | `assert.Equal` |
| 58 | TestJustify | §21.14 F16 | unit | Block justify. | `assert.Equal` |
| 59 | TestPad | §21.14 F17 | unit | Left/right pad counts. | `assert.Equal` |
| 60 | TestTruncate | §21.14 F18 | unit | "hello world" with width 8 → "hello w…". | `assert.Equal` |
| 61 | TestTruncateASCIIFallback | §21.14 F18 | unit | Non-UTF8 → "hello..." (3-char ellipsis). | `assert.Equal` |
| 62 | TestWrapAtWordBoundary | §21.14 F19 | unit | Wrap respects word boundaries. | golden |
| 63 | TestWrapLongUnbreakableWord | §21.14 F19 | unit | A single word longer than width :  wrap mid-word? Pin behavior. | `assert.Equal` |
| 64 | TestColumnsAutoSize | §21.14 F20, T8 | unit | Auto-width. | golden |
| 65 | TestColumnsExplicitGap | §21.14 F20 | unit | `WithColumnGap(4)` honored. | golden |
| 66 | TestColumnsAlignment | §21.14 F20 | unit | Per-column alignment (left, center, right). | golden |
| 67 | TestBanner | §21.14 F21 | unit | Multi-line banner with consistent style. | golden |
| **Prompts (PTY-driven)** ||||||
| 68 | TestPromptDefault | §21.14 F22, T9 | unit (PTY) | Empty input + `WithDefault("foo")` → returns "foo". | `fixture.NewPTY`, send "\n" |
| 69 | TestPromptInputTrimmed | §21.14 F22 | unit (PTY) | Trailing whitespace trimmed. | PTY send |
| 70 | TestPromptValidatorRetry | §21.14 F27 | unit (PTY) | Invalid input → re-prompts up to MaxAttempts. | PTY send invalid + valid |
| 71 | TestPromptMaxAttemptsExhausted | §21.14 F27 | unit (PTY) | After N invalid → returns typed error. | `assert.ErrorIs(_, term.ErrTooManyAttempts)` (or similar; pin) |
| 72 | TestPromptCtxCancel | §21.14 E7, T13 | unit (PTY) | ctx cancelled mid-input → returns `term.ErrCancelled`; terminal restored. | `assert.ErrorIs` |
| 73 | TestPromptTimeout | §21.14 F27 | unit (PTY) | `WithTimeout(d)` honored; returns `term.ErrTimeout`. | `assert.ErrorIs` |
| 74 | TestPromptOnNonTTY | §21.14 E6 | unit | Piped stdin → reads line without raw-mode; suitable for tests. | string reader, no PTY |
| 75 | TestPasswordEcho | §21.14 F23, T10 | unit (PTY) | Password input :  output buffer does NOT contain the typed chars. | PTY tap on output, `assert.NotContains` |
| 76 | TestPasswordCtxCancelRestoresMode | §21.14 NF7 | unit (PTY) | Password ctx cancel → terminal cooked-mode restored. | PTY state inspection |
| 77 | TestConfirmDefaultNo | §21.14 F24, T11 | unit (PTY) | Default-No confirm; empty input → false. | PTY |
| 78 | TestConfirmDefaultYes | §21.14 F24 | unit (PTY) | `WithDefaultYes()`; empty → true. | PTY |
| 79 | TestConfirmYInput | §21.14 F24 | unit (PTY) | "y\n" → true. | PTY |
| 80 | TestConfirmYesInput | §21.14 F24 | unit (PTY) | "yes\n" → true. | PTY |
| 81 | TestConfirmInvalidThenY | §21.14 F24 | unit (PTY) | "x\ny\n" → asks again, then true. | PTY |
| 82 | TestSelectTyped | §21.14 F25, T12 | unit (PTY) | `Select[MyType]` returns typed value. | PTY arrow + enter |
| 83 | TestMultiSelect | §21.14 F26 | unit (PTY) | Multi-select with space + enter; returns slice. | PTY |
| 84 | TestPromptInjectionRejectsControlChars | §23.9 #25 | unit (PTY) | Input containing ANSI escape sequences → rejected (NUL, control-chars stripped). | PTY send `\x1b[31m`, `assert.NotContains` in result |
| 85 | TestPromptLineCap4KiB | §23.9 #25 | unit (PTY) | Input > 4 KiB → typed error. | PTY |
| 86 | TestPromptRawModeRestoredOnPanic | §21.14 NF7 | unit (PTY) | Panic inside prompt → defer restores cooked mode. | PTY state check after recover |
| 87 | FuzzAnsiInjection | §21.14 NF7; §23.9 #25 | fuzz | Random strings sent as prompt input → never escape into Render output as raw ANSI. | `testing.F` |
| **Animator (the centerpiece)** ||||||
| 88 | TestNewAnimatorWrapsHandler | §21.14 F31 | unit | Animator wraps the supplied logger's handler with an interception handler. | reflect, `assert.NotEqual` |
| 89 | TestAnimatorAddReturnsHandle | §21.14 F33 | unit | Returns a Handle whose Cancel removes the animation. | `assert.NotNil(handle)` |
| 90 | TestAnimatorRunNoAnimations | §21.14 E11 | unit | Run with zero registered animations → returns nil immediately. | `assert.NoError` |
| 91 | TestAnimatorRunCtxCancelled | §21.14 E10 | unit | Run with active animation; ctx cancel → returns `term.ErrCancelled`; restores handler. | `assert.ErrorIs` |
| 92 | TestAnimatorRunAllDone | §21.14 F34 | unit | All registered animations report `done=true` → Run returns nil. | `assert.NoError` |
| 93 | TestAnimatorBuffersLogsDuringFrame | §21.14 F36, T14 | unit | Log calls during `Render` are buffered; appear in writer between frames. | buffered writer + `assert.Equal` line order |
| 94 | TestAnimatorFlushesAboveAnimation | §21.14 F36, T15 | unit | Cursor-up sequence written before flush; logs above animation. | regex on output for `\x1b[?A` |
| 95 | TestAnimatorBackpressureDropsOldest | §21.14 NF4, E8, T16 | unit | Buffer at cap → oldest dropped; "X records dropped" warning emitted on next frame. | `assert.Contains "records dropped"` |
| 96 | TestAnimatorBackpressureBlocksBriefly | §21.14 NF4 | unit | When buffer near cap, log writes block up to 50ms. | `fixture.NewClock` + timing test |
| 97 | TestAnimatorRestoresHandlerOnExit | §21.14 F34, T17 | unit | After Run returns, original handler restored on the supplied logger. | `assert.Equal` (pointer) |
| 98 | TestAnimatorRestoresHandlerOnPanic | §21.14 NF7 | unit | If an animation's Render panics → Run recovers, restores handler, returns wrapped error. | `assert.Panics`-or-`Error`, handler check |
| 99 | TestAnimatorPauseResume | §21.14 F35 | unit | Pause stops frame loop; Resume restarts it; logs flushed during pause flow inline. | `fixture.NewClock` |
| 100 | TestAnimatorRefreshRateOption | §21.14 F32 | unit | `WithRefreshRate(50ms)` honored; frames render at that rate. | `fixture.NewClock` |
| 101 | TestAnimatorWriterOption | §21.14 F32 | unit | `WithWriter(buf)` directs frames to buf, not stderr. | buffer |
| 102 | TestAnimatorMaxBufferOption | §21.14 F32 | unit | `WithMaxBufferedRecords(10)` honored. | fill + check overflow |
| 103 | TestAnimatorMultipleAnimations | §21.14 F33 | unit | Add 3 animations; all render per frame; output has 3 sections. | golden |
| 104 | TestAnimatorHandleCancel | §21.14 F33 | unit | `handle.Cancel()` removes animation from active set. | line count diff |
| 105 | TestAnimatorRunIsBlocking | §21.14 F34 | unit | Run does not return until done or cancelled. | goroutine + select w/ timeout |
| **Animator concurrency (heavy from the spec)** ||||||
| 106 | TestAnimatorConcurrentLogWrites | §21.14 NF3, NF5; §23.14 | race | 100 goroutines call slog.Info during a running animation; no torn writes; all records eventually emitted. | stdlib `sync.WaitGroup`, line-count check |
| 107 | TestAnimatorConcurrentAdd | §21.14 NF5 | race | 50 goroutines `Add` animations during Run. | `sync.WaitGroup` |
| 108 | TestAnimatorConcurrentHandleCancel | §21.14 NF5 | race | Cancel mid-run from another goroutine. | `sync.WaitGroup` |
| 109 | TestAnimatorRenderUnderLock | §21.14 NF3 | race | Verify animator does NOT hold a lock during Render (NF3 atomicity); achieved by checking that a parallel `Add` does not block during a long-running Render. | timing-based |
| 110 | TestAnimatorBufferCopyNotShared | §21.14 NF3 | race | Buffered records copied before render; no aliasing → race-free. | -race exercise |
| **Built-in animations** ||||||
| 111 | TestSpinnerFrames | §21.14 F38, T18 | unit | Spinner cycles through 8 braille frames over 8 ticks. | `fixture.NewClock` |
| 112 | TestSpinnerWithCustomFrames | §21.14 F38 | unit | Custom frames override defaults. | `assert.Equal` |
| 113 | TestSpinnerWithStyle | §21.14 F38 | unit | Style colors the spinner glyph. | regex on ANSI |
| 114 | TestProgressMetrics | §21.14 F39-F42, T19 | unit | Progress shows %, speed, ETA, byte format correctly. | golden |
| 115 | TestProgressSetIncrement | §21.14 F40 | unit | Set + Increment update internal state. | `assert.Equal` after Render parse |
| 116 | TestProgressDone | §21.14 F40 | unit | Done → `Render` returns `done=true`. | `assert.True` |
| 117 | TestProgressShowSpeedFormatsBytesSec | §21.14 F41 | unit | "1.2 MB/s" formatting; 1024-bound. | golden |
| 118 | TestProgressShowETAFormat | §21.14 F41 | unit | "ETA 4s" / "ETA 1m 30s" / "ETA --" formats. | table |
| 119 | TestProgressShowBytes | §21.14 F41 | unit | "3.2 MB / 4.8 MB" rendering. | golden |
| 120 | TestProgressBarStyle | §21.14 F41 | unit | Custom style colors the bar. | regex |
| 121 | TestProgressGlyphCustom | §21.14 F41 | unit | Custom filled/empty glyphs. | golden |
| 122 | TestProgressTotalZero | §21.14 (edge) | unit | total=0 → bar shows 100% immediately or undefined? Pin; test. | `assert.Equal` |
| 123 | TestProgressOverflow | §21.14 E12 | unit | current > total → bar shows >100%. | `assert.Contains "%"` (whatever the chosen indicator) |
| 124 | TestProgressIndeterminate | §21.14 E13 | unit | total=-1 → spinner-style indeterminate, byte counter only. | golden |
| 125 | TestProgressConcurrentIncrement | §21.14 NF5, T21 | race | 100 goroutines Increment simultaneously; final = sum. | stdlib `sync.WaitGroup`, atomic check |
| 126 | TestStatusBarSetSection | §21.14 F44 | unit | SetSection adds/updates sections. | golden |
| 127 | TestStatusBarRemove | §21.14 F44 | unit | Remove deletes a section. | golden |
| 128 | TestStatusBarLayoutOptions | §21.14 F43, F44 | unit | Multi-line vs columns layout. | golden |
| 129 | TestStatusBarConcurrentSetSection | §21.14 NF5, T22 | race | 100 goroutines update distinct sections. | stdlib `sync.WaitGroup` |
| 130 | TestDownloadProgressRead | §21.14 F45, F46, T20 | unit | Read updates progress; final byte count == content length. | bytes.Reader source, `assert.Equal` |
| 131 | TestDownloadProgressOverRead | §21.14 E12 | unit | Read past content length → bar > 100%, no error. | `assert.NoError` |
| 132 | TestDownloadProgressUnknownLength | §21.14 E13 | unit | contentLength=-1 → indeterminate. | golden |
| 133 | TestDownloadProgressSpeedSliding | §21.14 F45 | unit | Speed computed from sliding window. | `fixture.NewClock` |
| **Lifecycle (Close discipline per §23.16)** ||||||
| 134 | TestAnimatorCloseIdempotent | §23.16 | unit | `Close()` twice → no panic, second is no-op, errs.Join used internally. | `assert.NoError` |
| 135 | TestAnimatorCloseAfterRun | §23.16 | unit | Close after Run finishes naturally → no error. | `assert.NoError` |
| 136 | TestAnimatorCloseDuringRun | §23.16 | unit | Close while Run is active → cancels, restores handler, joins errors. | `assert.NoError`, handler check |
| 137 | TestAnimatorCloseOnError | §23.16 | unit | An animation's Render returns an error mid-Run → Close still cleans up. | `assert.ErrorIs` |
| 138 | TestStatusBarCloseIdempotent | §23.16 | unit | StatusBar (if Closeable per audit) :  same pattern. | `assert.NoError` |
| 139 | TestStatusBarCloseAfterUse | §23.16 | unit | After SetSection / Render, Close → clean. | `assert.NoError` |
| 140 | TestProgressNoCloseRequired | §23.16 | unit | Progress is not Closeable per audit; verify. | reflect-based check |
| **Cross-platform** ||||||
| 141 | TestCapabilityLinux | §21.14 NF8 | cross-platform (build tag linux) | Linux-specific isatty path. | `golang.org/x/sys/unix` |
| 142 | TestCapabilityDarwin | §21.14 NF8 | cross-platform (build tag darwin) | macOS path. | as above |
| 143 | TestCapabilityWindows | §21.14 NF8 | cross-platform (build tag windows) | Windows console handle path. | `golang.org/x/sys/windows` |
| 144 | TestPromptRawModeUnix | §21.14 NF7 | cross-platform (!windows) | termios manipulation. | `unix.IoctlGetTermios` |
| 145 | TestPromptRawModeWindows | §21.14 NF7 | cross-platform (windows) | Console mode manipulation. | `windows.GetConsoleMode` |
| 146 | TestSIGWINCHResize | §21.14 E14 | cross-platform (!windows) | Resize signal triggers next-frame width refresh. | signal mock |
| 147 | TestPollingResizeWindows | §21.14 E14 | cross-platform (windows) | Polling-based resize on Windows. | clock mock |
| 148 | TestANSIEscapeOnWindowsLegacyConsole | §21.14 NF1 | cross-platform (windows) | Legacy console (no VT) → escapes stripped. | feature-detect |
| **Errors and register** ||||||
| 149 | TestSentinelsRegister | §21.14 (sentinels) | unit | `ErrCancelled`, `ErrNotInteractive`, `ErrTimeout` match library register. | regex |
| 150 | TestErrCancelledFromPromptCtx | §21.14 E7 | unit | Prompt ctx cancel → wraps `term.ErrCancelled`. | `assert.ErrorIs` |
| 151 | TestErrNotInteractive | §21.14 (sentinel) | unit | Prompt on non-TTY where interactive required → `ErrNotInteractive`. | `assert.ErrorIs` |
| 152 | TestErrTimeout | §21.14 F27 | unit | `WithTimeout` exceeded → `ErrTimeout`. | `assert.ErrorIs` |
| **Property-based** ||||||
| 153 | PropertyStyleRenderRoundTripASCII | §21.14 F7 | property | If a Strip function existed, `Strip(Render(x)) == x`. Spec doesn't include Strip; instead pin: `Render(x)` on no-color writer == x. | random strings |
| 154 | PropertyGlyphFallbackNonEmpty | §21.14 F11 | property | For every registered glyph, both UTF-8 and ASCII forms are non-empty. | iterate registry |
| 155 | PropertyBoxRenderingPreservesContent | §21.14 F14 | property | Stripped of border + ANSI, the inner content of `Box(x)` contains x. | random strings |
| 156 | PropertyTruncateNeverExceedsWidth | §21.14 F18 | property | `Truncate(x, w, e)` produces lines ≤ w runes. | random + `assert.LessOrEqual` |
| 157 | PropertyCenterPaddingSymmetric | §21.14 F15 | property | Center pads ±1 evenly. | random |
| 158 | PropertyCapabilityCachedDeterministic | §21.14 NF6 | property | Same writer w → `Capability(w) == Capability(w)`. | random writers |
| **Benchmarks (D35)** ||||||
| 159 | BenchmarkStyleRender | §21.14 NF1 | bench | `Style.Render(text)` ≤ 100 ns/op + 1 alloc. | benchstat |
| 160 | BenchmarkStyleRenderCacheHit | §21.14 NF1, F9 | bench | 2nd render of same style → 0 alloc. | `testing.AllocsPerRun == 0` |
| 161 | BenchmarkSpinnerFrame | §21.14 NF1 | bench | ≤ 500 ns/op. | benchstat |
| 162 | BenchmarkProgressFrame | §21.14 NF1 | bench | ≤ 2 µs/op (incl. ETA computation). | benchstat |
| 163 | BenchmarkBoxLayout | §21.14 NF1 | bench | Box render. | benchstat |
| 164 | BenchmarkGlyphLookup | §21.14 NF1 | bench | Map lookup + UTF-8 check. | benchstat |
| 165 | BenchmarkCapability | §21.14 NF6 | bench | First call vs cached call. | benchstat |
| 166 | BenchmarkAnimatorFrameWithLogs | §21.14 NF1, NF3 | bench | Frame render with N buffered records. | benchstat |
| **Surface + golden** ||||||
| 167 | TestSurfaceClosed_TermPackage | §21.14 NF12 | unit | API snapshot :  every PascalCase export listed. | `fixture/golden` |
| 168 | TestRenderGoldensAcrossPlatforms | §21.14 (visual) | golden | Box, Banner, StatusBar, Progress goldens for both UTF-8 and ASCII fallback. | `fixture.Golden` |
| **Examples** ||||||
| 169 | ExampleStyle | example | example | Runnable. | match |
| 170 | ExampleBox | example | example | Runnable. | match |
| 171 | ExamplePrompt | example | example (skip on non-TTY CI) | Runnable PTY. | PTY |
| 172 | ExampleAnimator | example | example | Runnable. | match |
| 173 | ExampleProgress | example | example | Runnable. | match |
| 174 | ExampleDownloadProgress | example | example | Runnable. | match |

### Bootstrap subset
`term/` does not depend on `assert/` directly. `assert/` imports `term/` (for failure-message color, §21.14 F3b). Therefore `term/` can use `assert/` freely **as long as `assert/`'s tests pass independently of `term/` correctness for color**. Since `assert/`'s bootstrap tests are bare-`if` and color is a non-bootstrap feature, **the cycle is broken cleanly**.

### Coverage target
- **Line coverage:** ≥92% (some platform-only paths are skipped on non-matching platforms; the union of all platform tests reaches 100%)
- **Branch coverage:** ≥90%
- **Public-API coverage:** 100% (~60 exported symbols across types, functions, options)
- **Animator concurrency paths:** 100% (this is the centerpiece; gaps are blocking)

### Edge cases not in the spec but worth testing
- **L-add-1:** `RegisterGlyph` with an empty UTF-8 or empty ASCII string → typed error (otherwise falls through to "" output with no warning).
- **L-add-2:** `Box` with a title longer than the box width :  truncate or expand? Pin behavior.
- **L-add-3:** `Wrap` on text containing ANSI escapes :  rune counting must skip escape bytes; otherwise wrap-points are wrong.
- **L-add-4:** `Center` / `Justify` / `Truncate` on text with combining characters (é = `e\u0301`) :  width counting must use grapheme clusters or document the limitation.
- **L-add-5:** `Animator.Run` invoked twice → second call's behavior. Pin: should error or restart? **Spec gap; recommend explicit `term: animator: already running` error**.
- **L-add-6:** `Animator` with an animation whose `Render` returns lines containing newlines internally → counted correctly for cursor-up?
- **L-add-7:** `Animator` while terminal is detached / TTY closed during Run (e.g., user closes ssh session) → write errors handled, Run returns wrapped error.
- **L-add-8:** `Progress.Set(negative)` :  clamp to 0 or accept? Pin.
- **L-add-9:** `DownloadProgress.Read` after `Done()` called → still works or returns error? Pin.
- **L-add-10:** `Capability(w)` where w changes capabilities mid-program (rare but possible: `os.Stdout` reassigned) → cache is per-pointer; document.
- **L-add-11:** `Prompt` with very rapid keystrokes :  line buffer overflows? 4 KiB cap per §23.9 #25.
- **L-add-12:** ANSI escape injection in `Style.Render(userInput)` :  user-supplied text containing `\x1b[31m` should NOT be treated as a style. Pin: render emits the user's bytes verbatim, sandwiched between the style escape and reset. Document risk; covered by `FuzzAnsiInjection`.
- **L-add-13:** Animator where the supplied logger's handler is itself an animator's interception handler (double-wrap) :  detected, returns error, or wraps once? Pin.
- **L-add-14:** `Banner` with a single-line input :  degrades gracefully.
- **L-add-15:** Glyph registry mutation from a `RegisterGlyph` call AFTER `Glyphs()` was iterated :  does the iterator see the new entry or not? Document iterator-snapshot semantics.

---

## Cross-package integration tests for kernel

These live in a sibling test directory `kernel_integration_test/` (or use `internal_test` packages) and exercise multiple kernel packages composed together.

| #  | Name | Spec ref | Type | Description | Test helpers used |
|----|------|----------|------|-------------|-------------------|
| K1 | TestErrsWithLog | §21.2 + §21.3 | integration | `slog.ErrorContext(ctx, "fail", slog.Any("error", errs.Wrap(io.EOF, "pkg: act")))` → output contains `"error":"pkg: act: EOF"`; `errors.Is` from a captured record retains chain. | buffer + `assert.Contains`, `assert.ErrorIs` |
| K2 | TestErrsCodedWithLog | §21.2 F11/F12 + §21.3 | integration | Coded error logged → JSON output contains `"code":"E_..."`. | `assert.JSONEq` |
| K3 | TestOptionWithErrs | §21.1 + §21.2 | integration | Option that returns `errs.Wrap(...)` :  Apply propagates wrapped error; ErrorIs works. | `assert.ErrorIs` |
| K4 | TestOptionStrictWithErrsJoin | §21.1 F5 + §21.2 F5 | integration | Strict mode collects multiple errs.Wrap'd errors; final join is walkable via `errs.Chain`. | `assert.Subset` over Chain output |
| K5 | TestAssertWithErrsErrorIs | §21.4 F3 + §21.2 | integration | `assert.ErrorIs` walks a chain that includes `errs.Wrap`'d sentinels. | `assert.True` |
| K6 | TestLogWithTermColorIntegration | §21.3 NF6 + §21.14 F4-F9 | integration | `log.NewHandler(stderr, WithColor(ColorAlways))` produces escapes byte-equal to those produced by `term.New().Foreground(...).Render`. | `assert.Equal` on byte sequences |
| K7 | TestLogWithTermCapabilityDetection | §21.3 F9 + §21.14 F1-F3 | integration | log's `ColorAuto` calls into `term.Capability(w)` to decide color. Verify by mocking the writer. | mock writer |
| K8 | TestAssertDiffWithTermColor | §21.4 F18 + §21.14 F4-F9 | integration | TTY test → assert.Equal failure renders cyan/rose ANSI matching term's escapes. | `fixture.NewPTY` + `assert.Contains` |
| K9 | TestTermAnimatorWithLog | §21.14 F30-F37 + §21.3 | integration | NewAnimator(slog.New(log.NewHandler(...))).Run(ctx); slog.Info during animation → buffered then flushed above progress bar. | buffer inspection |
| K10 | TestTermAnimatorWithLogCtxAttrs | §21.14 + §21.3 F5/F11 | integration | `log.With(ctx, attr)` then `slog.InfoContext(ctx, ...)` during animation → buffered record retains ctx-attrs. | `assert.Contains` |
| K11 | TestKernelLibraryRegisterCrossPackage | §21.1 NF3 + §21.2 NF4 + §21.14 sentinels | integration | Lint test: every kernel-package error string matches `^[a-z][^A-Z.]*$` (library register). Walks every Sentinel + every typed `Error()` example. | regex over package-symbol scan |
| K12 | TestKernelDAGNoForbiddenImports | §6 F1-F3 + §21.14 F3a/b/c | integration | `go list -deps` over each kernel package :  confirms `option` imports nothing from glacier; `errs` doesn't import `log`; `term` doesn't import `log` or `assert`. | `os/exec` go list |
| K13 | TestSurfaceClosedKernelTotal | NF8 of each | integration | Aggregate API snapshot: 8 + 12 + 11 + (~70) + (~60) symbols. Drift fails CI. | `fixture/golden` package-level |
| K14 | TestKernelGoroutineLeakFreeUnderRace | §23.14 + §21.14 NF5 | integration | Run a representative kernel scenario (Animator.Run + multiple goroutines logging) under `-race` with `fixture.WatchGoroutines()` confirming no leaks. | `fixture.WatchGoroutines` |
| K15 | TestAssertEqualUsingErrsCoded | §21.4 F11 + §21.2 F11 | integration | `assert.Equal` of two errors implementing Coded → uses Coded's `Code()` if both have it; matched value-equality. (Lynx-defined behavior; pin.) | `assert.True` |

---

## Bootstrap discipline notes

The kernel has three chicken-and-egg loops:

1. **`assert/` ↔ itself** :  assert's own tests cannot use assert.Equal to validate assert.Equal's correctness without circular logic. **Resolution:** `TestEqual_Bootstrap_*` tests (#1–#6) use bare `if` and a hand-rolled `mockTB`. Once those pass, all other assert tests use `assert.X` freely. **The bootstrap subset is exactly six tests** (primitive int, primitive string, nil/nil, typed-nil/typed-nil, mismatch, type-mismatch-at-top). Plus `mockTB`'s own bare-`if` tests.

2. **`assert/` ↔ `term/`** :  assert uses term for failure-message color; term uses assert for its own tests. **Resolution:** the cycle is broken because term's tests use the *non-color* parts of assert (`assert.Equal`, `assert.NoError`, etc.) which do not touch term. Color-related assert behavior is tested without term (using direct ANSI byte comparison). The full-color integration test K8 lives in cross-package integration, not in either package's own test suite.

3. **`log/` ↔ `term/`** :  log imports term; term doesn't import log. The animator wraps a user-supplied logger but does not import log. **No bootstrap issue.**

**`fixture/` usage in kernel:** `fixture/` is leaf-tier and imports kernel. Kernel tests cannot import `fixture/` directly (it would create a circular import). **Resolution:** kernel tests use `fixture` via the `_test` package (a separate Go package built only for testing), which CAN import fixture without creating a production-code cycle. This is documented in spec 0002 §Architecture as the **"_test package fixture exception"**. Tests like `TestColorAutoOnTTY` live in `package log_test` (not `package log`) and import `fixture/term` for PTYs.

**`mock/` usage in kernel:** Same exception :  kernel tests can use `mock/` from a `_test` package. Usage examples: T#107 uses a mock slog handler to capture warning output; T#88 uses `mock` to wrap a slog handler.

**`assert/require/` in kernel tests:** Once `assert/` is bootstrap-tested, `require/` is also available. Kernel tests use `require.NoError` for setup-must-succeed lines and `assert.X` for stack-of-assertions in test bodies :  exactly per the spec contract.

---

## Sign-off conditions

The kernel-tier test matrix is **complete and signed off** when ALL of the following hold:

1. **All listed tests are written and passing** on Linux, macOS, Windows.
2. **`go test -race ./...`** passes for every kernel package :  every concurrent path tested under the race detector.
3. **`go test -count=10 -race ./...`** passes :  flake-free under repetition (catches timing-sensitive bugs in animator and Eventually).
4. **`go test -bench=. -benchmem -count=10 ./...`** runs to completion; benchstat against the baseline shows no regression > 5% per §23.12 / §23.13.
5. **All §23.13 recalibrated targets met**:
   - log: ≤ 3 allocs beyond stdlib slog baseline
   - assert.Equal primitive fast-path: ≤ 50 ns/op zero allocs (PROVEN by #7)
   - assert.Equal slow path: ≤ 200 ns/op
   - term: Style.Render ≤ 100 ns/op + 1 alloc; Spinner ≤ 500 ns/op; Progress ≤ 2 µs/op
6. **All fuzz targets** (`FuzzSentinelRegister`, `FuzzMatchGlob`, `FuzzMatchRegex`, `FuzzGlyphLookup`, `FuzzGlyphRegistration`, `FuzzAnsiInjection`) run for ≥ 5 minutes each in CI without panic, hang, or crash.
7. **Coverage gates**:
   - `option/`: 100% line, 100% branch, 100% public API
   - `errs/`: 100% line, 100% branch, 100% public API
   - `log/`: ≥ 95% line, ≥ 90% branch, 100% public API
   - `assert/` + `assert/require/`: 100% line, ≥ 98% branch, 100% public API
   - `term/`: ≥ 92% line per platform (union 100% across platforms), ≥ 90% branch, 100% public API
8. **All §21.X edge cases (E1–EN per package) have a named test** in the matrix above. Verified by Lynx walking every E-row.
9. **All §23.14 concurrency lock-ins exercised**:
   - `concur.Mutex.LockCtx` cancel-leak :  covered by concur tests (mid-tier; not in scope here, but kernel-side: `TestAssertConcurrent` and `TestAnimatorConcurrentLogWrites` exercise concurrent test-helper paths)
   - `concur.Group.Go` after `WaitDone` panics :  concur (mid-tier scope)
   - `mock`/`httpmock` `Times(1)` race :  kernel-side parallel: `TestTBConcurrentErrorf` (#101)
10. **All §23.16 lifecycle tests** for kernel stateful types complete: `term.Animator` (T#134-137) and `term.StatusBar` (T#138-139). `Progress` confirmed not-Closeable per audit (T#140).
11. **All §23.17 generics fixes verified**:
    - `assert.Equal[T]` primitive fast-path bypass: `TestPrimitiveFastPathBypass` (#7) :  `AllocsPerRun == 0` PROVES the fast path is taken
    - `assert.Equal[T]` slow-path: `TestPrimitiveFastPathTypeNotComparable` (#8)
    - `option.Required[T]` getter form: `TestRequiredGenericTLoadBearing` (#30)
    - `assert.Len` non-generic per §23.17: T#44–T#48
    - `assert/require/` mirror generic-for-generic: T#R3, T#R4
12. **No bare `if got != want` in any non-bootstrap test file**. Verified by lint: `grep -RP '^\s*if .* != .* \{$' --include='*_test.go' .` produces zero matches outside the documented bootstrap files.
13. **No testify import** anywhere. Verified by `go list -deps` for `github.com/stretchr/testify`.
14. **No hand-rolled stub satisfying a public interface** in any non-bootstrap test file. Verified by import audit.
15. **Every test function carries a comment citing the spec section / decision ID it verifies**: `// §21.1 F11; §23.4` style. Lint enforced by a custom `cmd/specreflint` tool that grep-walks every `func Test*`.
16. **Spec-traceability matrix**: Lynx publishes a matrix mapping every §21.X F-number, NF-number, E-number, T-number, and §23.X decision ID to the test names that cover it. Empty cells are blocking.
17. **`Example*` functions all run to completion** under `go test -run='^Example'`.
18. **API surface snapshots** (`TestSurfaceClosed_*`) all pass :  surface drift requires a spec amendment first.
19. **Cross-package integration tests** (K1-K15) all pass.
20. **CI gates green**: `codegen-drift`, `mod-hygiene`, `toolchain-pin-check`, `sbom-generate` per §23.12 (these are framework-wide; the kernel must not break them).

If any of the above is incomplete, the test matrix is incomplete and the kernel does NOT move to `accepted`.

---

**Lynx, kernel-tier review, signed.** ʕ◉‿◉ʔ