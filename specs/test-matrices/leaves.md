# Lynx Leaf-Tier Test Matrix (cli, mock, httpmock, httpc)

**Scope:** §21.9 `cli/` (+ `cli/gen/`), §21.10 `mock/`, §21.11 `httpmock/`, §21.13 `httpc/`. §21.12 `sandbox/` is OUT (dropped per §23.1).
**Authority:** plan §21.9–§21.13 functional/non-functional/edge tables, plus §23.6, §23.7, §23.8, §23.9, §23.11, §23.13, §23.14, §23.15, §23.16, §23.17.
**Sign-off rule:** `go test -race ./...` against the implementation must give **100% release confidence with zero manual testing**. Any test below that is missing or skipped fails Lynx sign-off.
**Naming convention:** every test name carries a trailing comment in source citing spec ref (e.g. `// §21.9 F4 / E2`).

---

## Package: `cli/` (and `cli/gen/` codegen)

### Test files

Runtime:
- `cli/cli_test.go` — App/New/Register/Lookup/Close
- `cli/run_test.go` — Run/Main/dispatch/argv
- `cli/flags_test.go` — flag parsing, defaults, env, short, deprecated, choices, required, validate
- `cli/help_test.go` — help page rendering, doc-from-godoc
- `cli/banner_test.go` — banner ANSI, NO_COLOR/GLACIER_NO_COLOR, TTY detection
- `cli/register_test.go` — programmatic registration, alias, parent
- `cli/options_test.go` — every Option family (incl. §23.6 additions)
- `cli/error_test.go` — typed errors, FlagParseError (per §23.15 rename), ErrUnknownCommand/Flag/Cancelled/Panic
- `cli/concurrency_test.go` — concurrent `Run` (-race)
- `cli/signal_test.go` — SIGINT/SIGTERM (Unix), CTRL_BREAK_EVENT (Windows; build-tagged)
- `cli/lifecycle_test.go` — App.Close idempotency (§23.16)
- `cli/import_audit_test.go` — package import surface assertion
- `cli/example_test.go` — godoc Example_ functions for headline use cases
- `cli/cli_fuzz_test.go` — `FuzzArgvParse`
- `cli/cli_bench_test.go` — `BenchmarkRunSimpleCommand`, `BenchmarkRunSubcommand5Levels`, `BenchmarkParseFlags20`

Codegen (`cli/gen/`):
- `cli/gen/markers_test.go` — every marker type's grammar regex (§23.8)
- `cli/gen/discover_test.go` — interface discovery via `go/packages`
- `cli/gen/emit_test.go` — generated Go file emission (golden-file)
- `cli/gen/parse_test.go` — doc-comment parsing, help-text extraction
- `cli/gen/check_test.go` — `--check` drift detection
- `cli/gen/safety_test.go` — `strconv.Quote` invariant, identifier validation, output path canonicalization
- `cli/gen/generate_test.go` — `Generate(opts)` end-to-end on testdata modules
- `cli/gen/concurrency_test.go` — concurrent `Generate` invocations
- `cli/gen/property_test.go` — idempotency, determinism property tests
- `cli/gen/marker_fuzz_test.go` — `FuzzMarkerParse` (per §23.10)
- `cli/gen/testdata/...` — golden modules (simple, nested, multi-root-error, cycle-error, all-markers, mock-marker)
- `cli/gen/golden/...` — expected `zz_generated_*.go` outputs

### Test matrix

| # | Name | Spec ref | Type | Description | Helpers |
|---|---|---|---|---|---|
| 1 | TestRegisterFromInterface | §21.9 F1/F2 | unit | Register a struct satisfying `Run(ctx) error`; `Lookup` returns it | assert |
| 2 | TestRegisterReturnsErrOnDuplicateName | §21.9 F2/§21.9 NQ | unit | Register two commands with same name → typed error | assert/require |
| 3 | TestRegisterRejectsNonCommand | §21.9 F1 | unit | Pass struct without `Run(ctx) error` → typed registration error | assert |
| 4 | TestNewWithDefaultOptions | §21.9 F3 | unit | `New()` returns App with banner-on default (per §23.15 banner default) | assert |
| 5 | TestNewWithVersion | §21.9 F3 | unit | `WithVersion("v1.2.3")` sets version surfaced by `--version` | assert |
| 6 | TestNewWithoutBanner | §21.9 E11/§23.15 | unit | `WithoutBanner()` suppresses banner; verify stdout has no logo bytes | assert, fixture.CaptureStdout |
| 7 | TestNewWithStdoutStderrCapture | §21.9 F3 | unit | Custom writers receive output | assert, fixture |
| 8 | TestNewWithLogger | §21.9 F3 | unit | Custom slog logger receives lifecycle events | assert, mock (slog handler) |
| 9 | TestRunDispatchesArgs | §21.9 F4 (orig #2) | unit | Argv routes to correct command | assert |
| 10 | TestRunUnknownCommand | §21.9 F4 | unit | `ErrUnknownCommand` returned and `errors.Is` works | assert |
| 11 | TestRunUnknownFlag | §21.9 F4 | unit | `ErrUnknownFlag` returned for `--bogus` | assert |
| 12 | TestRunFlagParseError | §23.15 (renamed) | unit | Bad value (`--port abc` for int) returns `*FlagParseError` (was `ParseError`) | assert/require ErrorAs |
| 13 | TestFlagParseErrorUnwrap | §21.9 F4 | unit | `FlagParseError.Unwrap()` returns underlying cause | assert |
| 14 | TestMainCallsOSExitOnError | §21.9 F5 | unit | `Main` formats error to stderr CLI register and exits non-zero (use exit-trap shim) | fixture.CaptureStderr |
| 15 | TestMainExitZeroOnSuccess | §21.9 F5 | unit | Successful handler → exit 0 | fixture |
| 16 | TestAutoHelpFlagShort | §21.9 F6 | unit | `-h` prints usage; exits 0 | assert, fixture |
| 17 | TestAutoHelpFlagLong | §21.9 F6 | unit | `--help` prints usage | assert |
| 18 | TestAutoVersionFlagShort | §21.9 F6 | unit | `-v` prints wordmark + version | assert |
| 19 | TestAutoVersionFlagLong | §21.9 F6 | unit | `--version` prints version | assert |
| 20 | TestAutoVersionFromBuildInfo | §21.9 F6 | unit | When `WithVersion` not set, `runtime/debug.ReadBuildInfo` is consulted | assert |
| 21 | TestBannerNoArgs | §21.9 F7 (orig #5) | unit | Bare invocation prints banner | fixture.CaptureStdout, assert.Contains |
| 22 | TestBannerOnTTYUsesANSI | §21.9 F7/NF3 | unit | TTY → 24-bit ANSI escapes present | assert.Regex, fixture.PtyEmulator |
| 23 | TestBannerOffTTYNoANSI | §21.9 F7 | unit | non-TTY → no ANSI escapes | assert |
| 24 | TestBannerSuppressedByGlacierNoColor | §21.9 F11 | unit | `GLACIER_NO_COLOR=1` strips ANSI | fixture.SetEnv |
| 25 | TestBannerSuppressedByNoColor | §21.9 F11 | unit | `NO_COLOR=1` strips ANSI | fixture.SetEnv |
| 26 | TestBannerPrecomputedNoSprintf | §21.9 NF3 | unit | Lock byte slices precomputed; one alloc max in render | testing.AllocsPerRun |
| 27 | TestColorTTY | §21.9 F11 (orig #7) | unit | Color in TTY | fixture |
| 28 | TestSIGINTCancelsHandler | §21.9 F8/E10 (orig #9) | integration | SIGINT cancels ctx; handler returns; exit code 130 | fixture, assert (Unix only build tag) |
| 29 | TestSIGTERMCancelsHandler | §21.9 F8 | integration | SIGTERM cancels ctx | fixture (Unix only) |
| 30 | TestWindowsCtrlBreakCancels | §21.9 F8 / X-platform | integration | `CTRL_BREAK_EVENT` cancels ctx (build-tagged windows) | fixture |
| 31 | TestConfFlagSourceLookup | §21.9 F9 (orig #10) | integration | `app.Lookup("port")` returns parsed flag value as string | assert |
| 32 | TestConfFlagSourceUnknownReturnsFalse | §21.9 F9 | unit | `Lookup` returns "", false on unknown name | assert |
| 33 | TestSubcommandTreeDepthFive | §21.9 F10/NF1 | unit | Nested 5 levels resolves correctly | assert |
| 34 | TestSubcommandUnknownPath | §21.9 F10 | unit | Path `root.foo.unknown` → ErrUnknownCommand | assert |
| 35 | TestAddCommandProgrammatic | §21.9 F10 | unit | `App.AddCommand(parent, child)` works | assert |
| 36 | TestErrorFormatterCLIRegister | §21.9 F12/NF5 | unit | Error from handler is capitalized, period-terminated | assert.Regex |
| 37 | TestLibraryRegisterErrorWrapped | §21.9 NF5 | unit | Internal cli error wrapped before display | assert |
| 38 | TestPanicCaughtAsErrPanic | §21.9 E9 (orig #27) | unit | Handler panic → `cli.ErrPanic{Value any}`, non-zero exit | assert/require ErrorAs |
| 39 | TestPanicStackTraceIncluded | §21.9 E9 | unit | Stack trace in CLI-register output (debug build tag) | assert |
| 40 | TestCtxCancelledBeforeHandler | §21.9 E10 | unit | Pre-cancelled ctx → ErrCancelled; handler not invoked | assert |
| 41 | TestRequiredMissingTypedError | §21.9 §marker required (orig #23) | unit | Missing `+glacier:required` flag → typed RequiredError | assert |
| 42 | TestChoicesEnforced | §21.9 §marker choices (orig #24) | unit | Value outside choices set → typed ChoicesError | assert |
| 43 | TestDeprecatedFlagWarn | §21.9 §marker deprecated (orig #25) | unit | Use of deprecated flag → log warning to stderr; still works | mock (slog handler) |
| 44 | TestValidateFunctionInvoked | §21.9 §marker validate | unit | Validate fn runs after binding before Run; rejects bad value | assert, mock |
| 45 | TestWithFlagShort | §23.6 | unit | `WithFlagShort("port", 'p')` registers short alias | assert |
| 46 | TestWithFlagShortRejectsMultiByte | §23.8 marker `^[a-zA-Z]$` | unit | non-ASCII rune → registration error | assert |
| 47 | TestWithFlagEnv | §23.6 | unit | `WithFlagEnv("port", "GLACIER_PORT")` overrides default from env | fixture.SetEnv |
| 48 | TestWithFlagEnvBadName | §23.8 env regex | unit | env name not matching `^[A-Z][A-Z0-9_]*$` → registration error | assert |
| 49 | TestWithFlagHelp | §23.6 | unit | `WithFlagHelp` overrides godoc-extracted help | assert |
| 50 | TestWithRoot | §23.6 | unit | `WithRoot()` marks command as root | assert |
| 51 | TestWithRootDuplicateRejected | §23.6 / E2 | unit | Two `WithRoot()` calls → ErrMultipleRoots | assert |
| 52 | TestWithAlias | §21.9 F-reg / orig 21.9 | unit | `WithAlias` adds alternative name; both names dispatch | assert |
| 53 | TestWithParentResolves | §21.9 F-reg | unit | `WithParent("root.cmd")` parents correctly | assert |
| 54 | TestWithParentUnresolved | §21.9 E5 | unit | Parent path doesn't exist → typed unresolved-parent error | assert |
| 55 | TestAppCloseFlushesAndIsIdempotent | §23.16 | lifecycle | `App.Close()` twice safely; `errs.Join`-merged errors from logger flush | assert/require, mock |
| 56 | TestAppCloseClosesOwnedTransports | §23.16 | lifecycle | Owned http transport (when used through cli helpers) closed | mock (ioCloser) |
| 57 | TestConcurrentAppRunRaceFree | §21.9 NF4/E13 (orig #26, -race) | concurrency | Two goroutines call `Run` on same App; no race | assert; `go test -race` |
| 58 | TestConcurrentRegisterDuringConstruction | §21.9 NF4 | concurrency | Concurrent `Register` calls before `Run` serialize via mutex | assert; -race |
| 59 | TestImportSurfaceStdlibOnly | §21.9 NF7 | audit | Runtime `cli` package imports only stdlib + framework kernel; no `go/packages` | go-list-deps audit |
| 60 | FuzzArgvParse | §23.10 | fuzz | Control-char, NUL, oversize, malformed UTF-8 args don't panic | testing.F |
| 61 | TestArgvNULRejected | §23.9 row 25 (cross-cut) | unit | NUL byte in arg → typed parse error | assert |
| 62 | TestArgvOversizeCapped | §23.9 | unit | Single arg > 64 KiB → typed error | assert |
| 63 | BenchmarkRunSimpleCommand | §23.13 / NF1 | bench | ≤ 5 µs/op excluding handler | testing.B |
| 64 | BenchmarkRunSubcommand5Levels | §23.13 | bench | depth-5 dispatch overhead bounded | testing.B |
| 65 | BenchmarkParseFlags20 | §23.13 | bench | 20 flag parse with type coercion | testing.B |
| 66 | ExampleApp_Run | §21.9 examples | doc-test | Headline example compiles and runs | godoc |
| 67 | ExampleApp_Main | §21.9 examples | doc-test | Main entrypoint example | godoc |

### Codegen-specific test rows (cli/gen/)

| # | Name | Spec ref | Type | Description | Helpers |
|---|---|---|---|---|---|
| 68 | TestMarkerCommandName_Allowed | §23.8 regex | unit | `name=serve` accepted | assert |
| 69 | TestMarkerCommandName_RejectUppercase | §23.8 regex `^[a-z]...$` | unit | `name=Serve` rejected with typed error | assert |
| 70 | TestMarkerCommandName_RejectTrailingDash | §23.8 | unit | `name=serve-` rejected | assert |
| 71 | TestMarkerCommandName_OversizeRejected | §23.8 max 32 | unit | 33-char name rejected | assert |
| 72 | TestMarkerParentDottedSegments | §23.8 | unit | `parent=root.foo.bar` parsed; `..`-style rejected | assert |
| 73 | TestMarkerAlias | §23.8 | unit | Alias same regex as name | assert |
| 74 | TestMarkerDefault_String | §21.9 F23 (orig #11) | unit | `default "hello"` emitted via `strconv.Quote` | golden |
| 75 | TestMarkerDefault_Int | §21.9 F23 | unit | `default 8080` emitted as `int(8080)` | golden |
| 76 | TestMarkerDefault_TypeMismatch | §21.9 F23 | unit | `default abc` for int field → typed error | assert |
| 77 | TestMarkerShort_Single | §23.8 short regex (orig #12) | unit | `short p` accepted | assert |
| 78 | TestMarkerShort_NonASCII | §23.8 | unit | `short é` rejected | assert |
| 79 | TestMarkerShort_MultiChar | §23.8 | unit | `short port` rejected | assert |
| 80 | TestMarkerEnv_Allowed | §23.8 env regex (orig #13) | unit | `env GLACIER_PORT` accepted | assert |
| 81 | TestMarkerEnv_Lowercase | §23.8 | unit | `env glacier_port` rejected | assert |
| 82 | TestMarkerRequired | §21.9 F16 | unit | `+glacier:required` marker registered | assert |
| 83 | TestMarkerChoices_Pipe | §23.8 | unit | `choices a\|b\|c` parsed; each piece regex-validated | assert |
| 84 | TestMarkerChoices_Max32 | §23.8 | unit | 33 choices rejected | assert |
| 85 | TestMarkerChoices_BadChar | §23.8 | unit | `choices A\|b` (uppercase) rejected | assert |
| 86 | TestMarkerDeprecated | §21.9 F16 | unit | `deprecated message="use --new"` parsed | assert |
| 87 | TestMarkerValidate_FuncRefResolved | §23.8 / orig #20 | unit | `validate myValidate` resolves via go/types; signature check enforced | assert |
| 88 | TestMarkerValidate_BadIdent | §23.8 ident regex | unit | `validate 1bad` rejected | assert |
| 89 | TestMarkerValidate_WrongSig | §23.8 | unit | Resolved fn has wrong signature → typed error | assert |
| 90 | TestMarkerCommandApp | §23.8 (`app=` modifier) | unit | `+glacier:command app=myApp` targets named App; default `cli.Default` | golden |
| 91 | TestMarkerUnknownWarn | §21.9 E4 (orig #14) | unit | Unknown marker → warning, gen succeeds | mock(logger) |
| 92 | TestMarkerLintError | §21.9 E4 (orig #15) | unit | `--lint` upgrades unknown to error | assert |
| 93 | TestMarkerOnNonField_Warn | §21.9 E3 | unit | Marker above non-field → warning, ignored | mock(logger) |
| 94 | TestStrconvQuoteInvariant | §23.8 | unit | All emitted string literals are `strconv.Quote`d (verify no raw `"%s"` formatting in emitter) | grep-of-source-test |
| 95 | TestNoFmtSprintfInEmitter | §23.8 | audit | Emitter source forbids `fmt.Sprintf("\"%s\"", ...)` | static-source-grep |
| 96 | TestOutputPathRejectsDotDot | §23.8/§23.9 row 18 | unit | Output path containing `..` → typed error | assert |
| 97 | TestOutputPathRejectsAbsolute | §23.8 | unit | Absolute path → typed error | assert |
| 98 | TestOutputPathRejectsUNC | §23.8 / X-platform | unit | `\\?\C:\...` rejected on Windows | assert (build-tag) |
| 99 | TestOutputPathUnderModuleRoot | §23.8 | unit | Resolved output sits under module root | assert |
| 100 | TestDiscoverCommandImplementers | §21.9 F19 | integration | go/packages walk finds Run(ctx) error structs | golden testdata |
| 101 | TestDiscoverSameModuleOnly | §21.9 F19 | unit | Third-party packages excluded from discovery | assert |
| 102 | TestCodegenSimple | §21.9 (orig #16) | golden | Simple struct → expected `zz_generated_cli.go` | fixture.GoldenFile |
| 103 | TestCodegenNested | §21.9 (orig #17) | golden | Nested command tree | fixture.GoldenFile |
| 104 | TestCodegenAllMarkers | §21.9 examples | golden | All marker types in one struct → emits all `WithFlag*` per §23.6 | fixture |
| 105 | TestCodegenMultipleRoots | §21.9 E2 (orig #18) | unit | Two `+glacier:root` → typed error naming both | assert.ErrorContains |
| 106 | TestCodegenZeroRoots | §21.9 F20 | unit | Zero roots → typed error | assert |
| 107 | TestCodegenCycle | §21.9 E6 (orig #19) | unit | Parent A→B→A → cycle error | assert |
| 108 | TestCodegenNameCollision | §21.9 F21 | unit | Two commands same dotted path → typed error | assert |
| 109 | TestCodegenChoicesOnNonStringCoercible | §21.9 E8 | unit | choices on `time.Duration` → type-mismatch error | assert |
| 110 | TestCodegenMarkerOnNonField | §21.9 E3 | unit | Comment block above non-field → warning | mock(logger) |
| 111 | TestCheckModeNoDriftPasses | §23.8/§23.12 codegen-drift (orig #21) | unit | Source matches generated → exit 0 | assert |
| 112 | TestCheckModeDriftDetected | §23.8/§23.12 | unit | Source modified, gen stale → exit non-zero with diff | assert.ErrorContains |
| 113 | TestCheckModeMissingFile | §23.8 | unit | Generated file deleted → drift error | assert |
| 114 | TestGenerateIdempotent | §23.8 / property | property | `Generate(Generate(src))` byte-identical (run twice) | rapid/manual |
| 115 | TestGenerateDeterministic | §23.8 / property | property | Same input → identical bytes across N=20 runs | property |
| 116 | TestGenerateRespectsOutputName | §21.9 Options.OutputName | unit | Custom name honored | assert |
| 117 | TestGeneratedFileHeader | §21.9 NF8 | unit | First line is `// Code generated by glaciergen; DO NOT EDIT.` | assert |
| 118 | TestGeneratedFileSortsLast_zz | §21.9 conventions | unit | Filename starts `zz_` | assert |
| 119 | TestGeneratedFileCompiles | §21.9 NF8 / mock-marker E20 | golden | Run `go build` on emitted golden directory; succeeds | fixture.RunGo |
| 120 | TestMockInterfaceMarkerEmitsWrapper | §21.10 F19/F20 | golden | `+glacier:mock` on interface → typed `<I>Mock` struct emitted | fixture.GoldenFile |
| 121 | TestMockMarkerNameOverride | §21.10 F24 | golden | `+glacier:mock name=CustomMock` honored | fixture |
| 122 | TestMockMarkerEmittedCodeCompiles | §21.10 F20-22 | integration | Generated mock wrapper compiles; OnGet typed | fixture.RunGo |
| 123 | TestConcurrentGenerateDifferentModules | §21.9 NF4 | concurrency | Two goroutines generating different modules; no race | -race |
| 124 | TestConcurrentGenerateSameModuleSerializes | §21.9 NF4 | concurrency | Same module guarded by file lock; no torn writes | -race, fixture.NewFS |
| 125 | FuzzMarkerParse | §23.10/§23.12 | fuzz | Control chars, NUL, oversize, code-injection-shaped marker payloads | testing.F |
| 126 | TestMarkerInjectionAttempt_Quotes | §23.8 | unit | `default value="; rm -rf /` → emitted via `strconv.Quote`, no shell-meaning leakage | golden |
| 127 | TestMarkerInjectionAttempt_Newlines | §23.8 | unit | Marker with embedded newline → grammar rejects | assert |
| 128 | TestMarkerInjectionAttempt_GoCode | §23.8 | unit | `default \"a\"); evil()` quoted, not interpreted | golden |
| 129 | TestGoPackagesLoadModeRestricted | §23.9 row 19 | unit | LoadMode is exactly `LoadFiles\|LoadImports\|LoadTypes`; no LoadSyntax beyond | assert (introspect emitter) |
| 130 | TestModulePrefixEnforced | §23.9 row 19 | unit | Discovered packages must share module-prefix; cross-module rejected | assert |
| 131 | BenchmarkGenerateSmallModule | §23.12 perf | bench | Sanity bench (build-time tooling, not §23.13-targeted) | testing.B |

### Codegen-specific concerns

- **Golden files** live under `cli/gen/testdata/` (input modules) and `cli/gen/golden/` (expected output). Regeneration via `GLACIER_GOLDEN_UPDATE=1` env var (per §23.18). Goldens are diffed byte-for-byte; trailing-newline normalization explicit.
- **Compile-validity:** beyond byte-identity, every golden is `go build`-verified inside the test (run `go build ./...` against an emitted directory under a temp `testdata` GOPATH; `fixture.RunGo` helper).
- **`strconv.Quote` invariant** verified two ways: (a) golden bytes contain quoted strings; (b) static check on `cli/gen/emit.go` source forbids `fmt.Sprintf("\"%s\"", ...)` patterns.
- **Output path canonicalization** routed through `internal/safefile` (per §23.10); tests assert that.
- **Idempotency** (run-twice byte-identical) is the litmus that emitter does not depend on map iteration order, time, hostname, or any nondeterministic source.

### Coverage target

- **Line coverage:** 90% minimum on `cli/`, 92% minimum on `cli/gen/` (more branchy paths).
- **Public API coverage:** 100% — every exported symbol (App, Command, New, Register, Run, Main, Lookup, Close, every `With*` option, every error type, Generate, Options) has at least one happy-path test plus one edge-case test.

### Edge cases not in spec

- **Argv with `--` separator** (POSIX double-dash for "end of options"). Spec doesn't say; behavior should be: everything after `--` is positional. Add `TestArgvDoubleDashSeparator`.
- **Empty argv** (`Run(ctx, []string{})`) — spec doesn't say; treat as bare invocation (banner). Add `TestRunEmptyArgv`.
- **Argv with embedded `=`** in flag (`--port=8080` vs `--port 8080`) — both should work. Add `TestFlagEqualsAndSpaceForms`.
- **godoc with markers but no prose** — what becomes the help text? (Empty string vs default-derived). Add `TestHelpFromCommentEmpty`.
- **Marker on embedded struct field** — does it cascade? Spec is silent. Add `TestMarkerOnEmbeddedField` (recommend: not supported, typed error).
- **Multiple `//go:embed` content failures at startup** — surface to consumer? Recommend: `init()`-time panic in CLI register.

### Special concerns

- **The `cli.Default` global** (§21.9 example) is mutable state; tests must not leak registrations across tests. Provide `cli.resetDefaultForTest()` build-tagged helper used by every test.
- **§NF1 perf gates** are §23.13-recalibrated targets; benchmarks must run on per-OS baseline (per §23.12 benchstat-flake mitigation).
- **Cross-platform signal handling** is the trickiest leaf; build-tagged tests for Unix vs Windows; CI matrix must run both.

---

## Package: `mock/`

### Test files

- `mock/of_test.go` — `Of[T]`, panics on non-interface, type caching
- `mock/mock_test.go` — Mock[T] methods, Interface(), Close (alias for Verify per §23.16)
- `mock/expect_test.go` — Expectation builder fluent chain
- `mock/matchers_test.go` — every Matcher constructor (now `Matcher[T]` per §23.17)
- `mock/return_test.go` — Return / ReturnSeq / Do
- `mock/times_test.go` — Times / AtLeast / AtMost / AnyTimes / Never
- `mock/strict_test.go` — Strict / StrictFatal / Lenient (per §23.15 naming)
- `mock/concurrency_test.go` — concurrent calls (-race), Times(1) race per §23.14
- `mock/lifecycle_test.go` — Cleanup auto-Verify, Close idempotency
- `mock/dispatch_test.go` — reflect-based MakeFunc dispatch correctness
- `mock/bench_test.go` — benchmarks
- `mock/property_test.go` — property tests
- `mock/example_test.go` — godoc examples
- `mock/safety_test.go` — no `unsafe` audit, no on-disk emission audit
- `mock/codegen_integration_test.go` — uses cli/gen test harness for `+glacier:mock` markers (or links to cli/gen/testdata mock golden cases)

### Test matrix

| # | Name | Spec ref | Type | Description | Helpers |
|---|---|---|---|---|---|
| 1 | TestOfBasic | §21.10 F1 (orig #1) | unit | `Of[Iface]` returns mock satisfying interface | assert |
| 2 | TestOfNonInterfacePanics | §21.10 E1 (orig #2) | unit | `Of[ConcreteType]` panics with structured message | assert.Panics |
| 3 | TestOfRegistersCleanup | §21.10 F1 | unit | Mock auto-registers t.Cleanup → Verify runs | assert, fixture |
| 4 | TestOfTypeCacheReused | §21.10 NF1 | unit | Second `Of[T]` reuses cached `reflect.Type` | testing.AllocsPerRun |
| 5 | TestInterfaceReturnsSatisfyingValue | §21.10 F2 | unit | `m.Interface()` value satisfies T | assert |
| 6 | TestInterfaceMethodsRoutedToMock | §21.10 F2 | unit | Calls on Interface() reach the expectation engine | assert |
| 7 | TestOnCallReturnSimple | §21.10 F3/F9 (orig #3) | unit | `OnCall("Get").Return(v, nil)` → call returns those | assert |
| 8 | TestOnCallMethodNotInInterface | §21.10 F3 | unit | `OnCall("NoSuchMethod")` panics at registration | assert.Panics |
| 9 | TestOnCallMethodNameRegex | §23.9 row 20 | unit | Method name matches `^[A-Z][A-Za-z0-9_]*$`; control chars rejected | assert.Panics |
| 10 | TestOnCallMethodNameOversize | §23.9 row 20 | unit | 65-byte method name rejected | assert.Panics |
| 11 | TestCallsToReturnsRecordedCalls | §21.10 F4 | unit | `CallsTo("Get")` returns timestamped Calls | assert |
| 12 | TestCallsToOrderingPreserved | §21.10 F4 | unit | Calls returned in arrival order | assert |
| 13 | TestUnmatchedCallsLenient | §21.10 F5 / E4 (orig #10) | unit | Lenient mode records unmatched | assert |
| 14 | TestUnmatchedCallsStrictAlwaysEmpty | §21.10 F5 | unit | Strict mode → UnmatchedCalls empty (call failed test) | assert |
| 15 | TestVerifyAutoInvokedAtCleanup | §21.10 F6 (orig #22) | unit | Cleanup runs Verify; failure surfaces in t.Errorf | mock(TB) |
| 16 | TestVerifyMidTestCheckpoint | §21.10 F6 | unit | Manual `Verify()` mid-test reports current state | assert |
| 17 | TestStrictDefault | §21.10 F7/§23.15 | unit | Default options are Strict (use `mock.StrictDefault` per §23.15) | assert |
| 18 | TestStrictUnmatched_TErrorf | §21.10 E2 (orig #9) | unit | Unmatched call → t.Errorf with method, args, expected list | mock(TB) |
| 19 | TestStrictFatalHalts | §21.10 E3 (orig #11) | unit | StrictFatal → t.Fatalf | mock(TB), goroutine-recover harness |
| 20 | TestLenientMode | §21.10/§23.15 (orig #10) | unit | Lenient records unmatched silently; consumer asserts | assert |
| 21 | TestStrictUnmatchedReturnsZeroValues | §21.10 E2 | unit | Unmatched in strict still returns zero T to consumer (test continues) | assert |
| 22 | TestExpectationWithMatchersGeneric | §23.17 / orig #4 | unit | `With(Eq[string]("u"))` typed; mismatch type-args is compile error | assert + compile-only check |
| 23 | TestEqTyped | §21.10 F13 (orig #5) | unit | `Eq[string]("u-42")` matches "u-42", not 42 | assert |
| 24 | TestEqMismatchType | §21.10 E8 | unit | Eq[string] vs runtime int → no match (lenient) / fail (strict) | assert |
| 25 | TestAnyMatchesAnyT | §21.10 F14 (orig #6) | unit | `Any[T]()` matches any value | assert |
| 26 | TestPredCustomPredicate | §21.10 F15 (orig #7) | unit | `Pred[User](func(u User) bool { ... })` → typed predicate | assert |
| 27 | TestNilMatcher | §21.10 F16 | unit | `Nil()` matches nil value | assert |
| 28 | TestNotNilMatcher | §21.10 F16 | unit | `NotNil()` matches non-nil | assert |
| 29 | TestMatchFnUntypedFallback | §21.10 F17 | unit | `MatchFn(func(any) bool)` works as escape hatch | assert |
| 30 | TestRefSmartEqual | §21.10 F18 (orig #8) | unit | `Ref[T](v, IgnoreOrder())` uses assert.Equal smart-equal | assert |
| 31 | TestRefIgnoreFields | §21.10 F18 | unit | `Ref` honors `assert.IgnoreFields` option | assert |
| 32 | TestExpectationReturnArityMismatch | §21.10 E5 (orig #17) | unit | `Return(x)` for method returning (T, error) → panic at registration | assert.Panics |
| 33 | TestExpectationReturnTypeMismatch | §21.10 E6 | unit | Wrong-type Return value → panic | assert.Panics |
| 34 | TestExpectationDoFn | §21.10 F11 (orig #16) | unit | `Do(fn)` invoked with recorded args; returns used | assert |
| 35 | TestExpectationDoFnWrongSignature | §21.10 E7 | unit | Wrong fn signature → panic at registration | assert.Panics |
| 36 | TestExpectationReturnSeq | §21.10 F10 (orig #14) | unit | Sequence advances per call | assert |
| 37 | TestReturnSeqCycleDefault | §21.10 E12 / SeqCycle | unit | After exhaustion, cycles to start | assert |
| 38 | TestReturnSeqExhaustMode | §21.10 E13 / SeqExhaust (orig #15) | unit | After exhaustion in SeqExhaust → t.Errorf | mock(TB) |
| 39 | TestExpectationTimes_Exact | §21.10 F12 (orig #12) | unit | Times(2): 2 calls → Verify pass; 1 → fail; 3 → unmatched(strict) | assert |
| 40 | TestExpectationAtLeast | §21.10 F12 | unit | AtLeast(3): 3 ok, 4 ok, 2 fail | assert |
| 41 | TestExpectationAtMost | §21.10 F12 | unit | AtMost(2): 0/1/2 ok, 3 fail at 3rd call | assert |
| 42 | TestExpectationAnyTimes | §21.10 F12 | unit | AnyTimes always passes | assert |
| 43 | TestExpectationNever | §21.10 F12 / E11 | unit | Never violated → Verify reports unexpected | mock(TB) |
| 44 | TestFirstRegisteredMatchWins | §21.10 E9 | unit | Two matching expectations: first wins | assert |
| 45 | TestVerifyReportsAllUnmetInOneError | §21.10 NF6 | unit | One t.Errorf consolidates all unmet expectations with structured diff | mock(TB), assert.Contains |
| 46 | TestConcurrentCallsRecordedCorrectly | §21.10 E14/NF7 (orig #18, -race) | concurrency | 1000 goroutines call mocked method; all recorded | -race |
| 47 | TestTimes1RaceFix | §23.14 | concurrency | match-AND-increment-AND-respond is single critical section; Times(1) under contention → exactly one match | -race |
| 48 | TestNoUnsafeImports | §21.10 NF3 | audit | `mock` package source contains no `unsafe.Pointer` or `reflect.Value.UnsafeAddr` | go-list-deps + grep |
| 49 | TestNoOnDiskEmissionAtRuntime | §21.10 NF4 | audit | Runtime path opens no files for write | fixture.WatchFiles |
| 50 | TestReflectMakeFuncDispatch | §21.10 NF1 | unit | Synthesized methods invoke through MakeFunc correctly | assert |
| 51 | TestMockInBenchmarkB | §21.10 E16 | unit | `*testing.B` satisfies `assert.TB`; Verify runs at b.Cleanup | assert |
| 52 | TestThirdPartyInterfaceMockable | §21.10 E15 | unit | Mock of e.g. `io.Reader` works at runtime | assert |
| 53 | TestMockClose_AliasForVerify | §23.16 | lifecycle | `Mock[T].Close()` runs Verify exactly once | assert |
| 54 | TestMockCloseIdempotent | §23.16 | lifecycle | Calling Close twice → no double-Verify; second is no-op | assert |
| 55 | TestMockClose_BeforeCleanup | §23.16 | lifecycle | Manual Close pre-empts t.Cleanup; Cleanup is no-op | assert, fixture |
| 56 | TestSlogLogValuerHonoredInFailureMessage | §23.11 | unit | Failure message renders `slog.LogValuer`-redacted args as `[REDACTED]` | mock(TB), log.Redact |
| 57 | TestFailureMessageRegisterCLI | §21.10 NF8 | unit | t.Errorf output is capitalized, period-terminated | assert.Regex |
| 58 | TestInternalPanicRegisterLibrary | §21.10 NF8 | unit | Wrong-arity panic message is library-register format | assert.Regex |
| 59 | TestRuntimeAndCodegenMix | §21.10 (orig #23) | integration | Same test uses both runtime `Of[Iface]` and codegen `IfaceMock` | assert |
| 60 | PropertyTimes_n_Calls_Verifies | §21.10 F12 / property | property | For random n in [0,20], n calls → Verify pass; n+1 → fail | rapid |
| 61 | PropertyExpectationOrderingMatters | §21.10 E9 / property | property | First-registered-wins property holds across N=100 random orderings | rapid |
| 62 | PropertyMatcherStringIsStable | §21.10 NF8 / property | property | `Matcher.String()` deterministic for same matcher | property |
| 63 | TestMockUsesAssertHelpersInTests | dogfooding | meta | This test asserts the `mock_test` package imports `assert` (dogfooding) | go-list-deps |
| 64 | BenchmarkMockCallNoArgs | §21.10 NF1 / §23.13 | bench | ≤ 6 allocs/op (relaxed per §23.13) | testing.B |
| 65 | BenchmarkMockCall3Args | §21.10 NF1 | bench | ≤ 6 allocs/op | testing.B |
| 66 | BenchmarkMockCallWithMatchers | §21.10 NF1 | bench | matcher iteration cost bounded | testing.B |
| 67 | BenchmarkMockOf | §21.10 NF1 | bench | Cache hot path | testing.B |
| 68 | BenchmarkCodegenPathAllocs | §23.13 | bench | Codegen path ≤ 2 allocs/op (tighter than reflect) | testing.B |
| 69 | ExampleOf_Repo | §21.10 examples | doc-test | Headline example | godoc |

### Coverage target

- **Line coverage:** 92% minimum (mock has many tightly-coupled paths and reflect dispatch should be exhaustively tested).
- **Public API coverage:** 100% — every exported symbol including all matcher constructors, all option constructors, all expectation methods.

### Edge cases not in spec

- **Generic interface (e.g., `Repo[T any]`)** as T — spec is silent. Recommend test `TestOfGenericInterface` (likely works; verify).
- **Method with named return values** — does reflect see them? Add `TestNamedReturns`.
- **Variadic method** (`func(args ...string) error`) — how is variadic arg matched? Add `TestVariadicMethod`.
- **Method returning a function type** — Return value of function-type field. Add `TestReturnFunctionType`.
- **Concurrent expectation registration during dispatch** — should panic per §23.14 (mirrors WaitGroup.Add-after-Wait). Add `TestRegisterDuringDispatchPanics`.
- **Nil interface T** at compile-time impossible, but `Of[any]()` — should panic (any is not an interface in the satisfiable sense). Add `TestOfAnyType`.

### Special concerns

- **§NF7 concurrency** is the §23.14 lock-in: the `Times(1)` race fix requires one critical section. Test 47 above is the explicit guard.
- **§NF1 alloc target** relaxed to 6 per §23.13 — tests must not regress to 2; if optimization gets us to 4, that's fine but not enforced.
- **Reflect type cache** must not leak across processes (tests call `mock.resetTypeCacheForTest()` build-tagged helper to isolate).

---

## Package: `httpmock/`

### Test files

- `httpmock/transport_test.go` — Transport, RoundTrip, OnRequest, recording
- `httpmock/stub_test.go` — Stub fluent chain (Method/Path/PathPrefix/Regex/Query/Header/Body)
- `httpmock/responders_test.go` — every responder constructor
- `httpmock/sequence_test.go` — Sequence / SequenceCycle / SequenceExhaust
- `httpmock/body_matchers_test.go` — BodyExact/BodyJSON[T]/BodyContains/BodyMatchFn
- `httpmock/fixtures_test.go` — LoadFixtures, schema validation, size cap, path traversal
- `httpmock/concurrency_test.go` — concurrent RoundTrip (-race), Times(1) race
- `httpmock/lifecycle_test.go` — Close idempotency, NewWithT cleanup
- `httpmock/import_audit_test.go` — no `net.Dial` / no `http.DefaultTransport` import
- `httpmock/property_test.go` — property tests
- `httpmock/fuzz_test.go` — `FuzzFixture`
- `httpmock/bench_test.go` — benchmarks
- `httpmock/example_test.go` — godoc examples
- `httpmock/testdata/fixtures/` — golden fixture JSONs + malformed/oversize variants

### Test matrix

| # | Name | Spec ref | Type | Description | Helpers |
|---|---|---|---|---|---|
| 1 | TestNewReturnsTransport | §21.11 F1 | unit | `New()` returns *Transport implementing http.RoundTripper | assert |
| 2 | TestNewWithT_RegistersCleanup | §21.11 F7 | unit | `NewWithT(t)` registers Verify at t.Cleanup | assert |
| 3 | TestRoundTripBasicMatch | §21.11 F2 (orig #1) | unit | OnRequest+Method+Path → response | assert |
| 4 | TestRoundTripNeverDials | §21.11 NF3 / E14 (orig #23) | unit | Verify no real-network call by package-import audit + runtime hook | go-list-deps audit |
| 5 | TestStubFirstWins | §21.11 E2 (orig #2) | unit | First registered stub matches | assert |
| 6 | TestStubMethodCaseInsensitive | §21.11 F9 | unit | "get" matches "GET" | assert |
| 7 | TestStubPathExact | §21.11 F10 | unit | Path("/users/42") matches | assert |
| 8 | TestStubPathPrefix | §21.11 F11 (orig #4) | unit | PathPrefix("/users/") matches /users/42 | assert |
| 9 | TestStubRegex | §21.11 F12 (orig #3) | unit | `Path(`/users/(\\d+)`).Regex()` matches /users/42 | assert |
| 10 | TestStubRegexInvalidCompile | §21.11 F12 | unit | Invalid regex → panic at registration with library-register message | assert.Panics |
| 11 | TestStubPathPrefixAndRegexMutuallyExclusive | §21.11 F12 | unit | Calling both → panic at registration | assert.Panics |
| 12 | TestStubQuery | §21.11 F13 (orig #6) | unit | Query("page", "2") matches | assert |
| 13 | TestStubHeader | §21.11 F14 (orig #5) | unit | Header("X-Foo", "bar") matches | assert |
| 14 | TestStubHeaderCaseInsensitiveName | §21.11 F14 | unit | header name matched canonically | assert |
| 15 | TestBodyExact | §21.11 F15 (orig #7) | unit | `BodyExact([]byte(...))` matches exactly | assert |
| 16 | TestBodyJSONSmartEqual | §21.11 F15 (orig #8) | unit | `BodyJSON[T](want, IgnoreFields("CreatedAt"))` works | assert |
| 17 | TestBodyJSONTypeParam | §23.17 | unit | `BodyJSON[User]` carries T compile-time | compile-only check |
| 18 | TestBodyContains | §21.11 F15 (orig #9) | unit | Substring match | assert |
| 19 | TestBodyMatchFn | §21.11 F15 | unit | Predicate match | assert |
| 20 | TestRespondJSON | §21.11 F18 (orig #10) | unit | `JSON[T](200, body)` marshals + Content-Type set | assert |
| 21 | TestRespondJSONFrom | §21.11 F19 | unit | `JSONFrom[T](200, reader)` reads + validates | assert |
| 22 | TestRespondJSONFromMalformed | §21.11 F19 | unit | Malformed JSON in reader → panic at registration | assert.Panics |
| 23 | TestRespondStatus | §21.11 F20 (orig #11) | unit | Empty body, status set | assert |
| 24 | TestRespondBody | §21.11 F21 (orig #12) | unit | Raw bytes + content type | assert |
| 25 | TestRespondStream | §21.11 F22 (orig #13) | unit | Streaming reader honored; closed after | assert, fixture |
| 26 | TestRespondStreamReaderError | §21.11 E6 | unit | Reader errors mid-read → wrapped transport error | assert |
| 27 | TestRespondError | §21.11 F23 (orig #14) | unit | `Error(err)` surfaces err to client | assert |
| 28 | TestSequenceBasic | §21.11 F24 (orig #15) | unit | rs[0], rs[1], ... | assert |
| 29 | TestSequenceCycle | §21.11 E9 | unit | Default cycle wraps to start | assert |
| 30 | TestSequenceExhaust | §21.11 E10 (orig #16) | unit | SequenceExhaust → ErrNoRouteMatch after exhaustion | assert |
| 31 | PropertySequenceCycleSixCalls | property | property | Sequence(a,b,c).Cycle: first 6 calls = [a,b,c,a,b,c] | rapid |
| 32 | TestStubTimes | §21.11 F16 (orig #17) | unit | Times(2): exactly 2 calls expected | assert |
| 33 | TestStubAtLeast | §21.11 F16 | unit | AtLeast(2) | assert |
| 34 | TestStubAtMost | §21.11 F16 | unit | AtMost(2) | assert |
| 35 | TestStubNever | §21.11 F16 | unit | Never violated → Verify fail | mock(TB) |
| 36 | TestStubAnyTimes | §21.11 F16 | unit | AnyTimes ok | assert |
| 37 | TestStubMissingRespond | §21.11 E4 | unit | Stub w/o Respond → ScriptError on RoundTrip | assert |
| 38 | TestStubRespondCalledTwiceLastWins | §21.11 E3 | unit | Last Respond overrides | assert |
| 39 | TestStrictDefaultUnmatchedReturnsErrNoRouteMatch | §21.11 NF5 / orig #18 | unit | Unmatched → ErrNoRouteMatch | assert |
| 40 | TestLenientUnmatched404 | §21.11 F8 / orig #19 | unit | LenientMode (per §23.15) returns 404 | assert |
| 41 | TestWithDefaultStatus | §21.11 F8 | unit | Configurable lenient default status | assert |
| 42 | TestWithLogger | §21.11 F8 | unit | Logger receives stub-match events | mock(slog handler) |
| 43 | TestRequestsTo | §21.11 F4 | unit | RequestsTo("/users/42") returns matching requests | assert |
| 44 | TestRequestsToWildcard | §21.11 F4 | unit | "/users/*" wildcard support | assert |
| 45 | TestAllRequests | §21.11 F5 | unit | Returns every request in arrival order | assert |
| 46 | TestVerifyAllStubsTimesMet | §21.11 F6 / E7 | unit | Verify reports unmet | mock(TB) |
| 47 | TestVerifyAtCleanup_NewWithT | §21.11 F7 | unit | Auto-Verify at cleanup | mock(TB) |
| 48 | TestLoadFixturesBasic | §21.11 F25 (orig #20) | integration | Reads testdata/httpmock/<name>.json, registers stubs | fixture |
| 49 | TestLoadFixturesMalformedJSON | §21.11 E11 | unit | Malformed → t.Errorf with parse error; no stubs registered | fixture |
| 50 | TestLoadFixturesTooLarge | §21.11 E12 (orig #21) | unit | >16 MiB → typed error | fixture (16 MiB+ file) |
| 51 | TestLoadFixturesPathTraversal | §21.11 E13 / §23.10 (orig #22) | unit | `../etc/passwd` → typed error | fixture |
| 52 | TestLoadFixturesAbsolutePath | §23.10 | unit | Absolute path rejected | fixture |
| 53 | TestLoadFixturesUNCRejected | §23.10 / X-platform | unit | Windows UNC rejected | fixture (build-tag) |
| 54 | TestLoadFixturesUnknownFields | §21.11 NF6 / §23.10 | unit | DisallowUnknownFields → reject extra keys | fixture |
| 55 | TestLoadFixturesDepthCap | §23.10 | unit | JSON depth >32 → reject | fixture |
| 56 | TestLoadFixturesUTF8Validated | §23.10 | unit | Malformed UTF-8 → reject | fixture |
| 57 | TestNoNetworkImports | §21.11 NF3 (orig #23) | audit | Package imports neither `net.Dial` nor `http.DefaultTransport` | go-list-deps |
| 58 | TestConcurrentRoundTrip | §21.11 NF4 / E8 (orig #24, -race) | concurrency | Concurrent RoundTrip → no race | -race |
| 59 | TestTimes1RaceFix | §23.14 | concurrency | match-AND-increment-AND-respond single critical section | -race |
| 60 | TestTransportClose | §23.16 | lifecycle | `Transport.Close()` flushes recorded requests + closes any held resources | assert |
| 61 | TestTransportCloseIdempotent | §23.16 | lifecycle | Close twice → no error; errs.Join joined | assert |
| 62 | TestMultipartBodyMatch | §21.11 E15 | unit | Body matchers work on raw multipart bytes | assert |
| 63 | TestScriptErrorTypedAndUnwrap | §21.11 F28 | unit | ScriptError implements error + Unwrap | assert.ErrorAs |
| 64 | FuzzFixture | §23.10 / orig #25 | fuzz | Malformed JSON, oversize, control chars in fixture loader | testing.F |
| 65 | TestImportSurfaceFluentInternal | §21.11 NF9 | audit | `fluent` used internally for stub-list iteration | go-list |
| 66 | BenchmarkRoundTripFirstStubMatches | §23.13 / NF1 | bench | ≤ 5 µs/op | testing.B |
| 67 | BenchmarkRoundTripScanThirty | §21.11 NF1 | bench | 30-stub scan bounded | testing.B |
| 68 | BenchmarkBodyJSONMatch | §21.11 NF1 | bench | smart-equal cost bounded | testing.B |
| 69 | BenchmarkResponseJSON | §21.11 NF1 | bench | response build cost | testing.B |
| 70 | ExampleTransport_OnRequest | §21.11 examples | doc-test | Headline example | godoc |

### Coverage target

- **Line coverage:** 95% minimum — httpmock is small, behavior-rich, and trivially testable.
- **Public API coverage:** 100%.

### Edge cases not in spec

- **Empty stub** (no matchers) — does it match every request? Spec implies yes; add `TestEmptyStubMatchesAll`.
- **Stub with conflicting matchers** (Path AND PathPrefix on same stub) — spec only says PathPrefix vs Regex are exclusive. Add `TestPathAndPathPrefixCoexist` (recommend: AND-match).
- **Request with no Body** under BodyContains matcher — should match `BodyContains("")` only. Add `TestEmptyBodyMatcherSemantics`.
- **Sequence with one element + cycle** — single-element infinite cycle. Add `TestSequenceSingleElementCycles`.
- **Stub registered after first RoundTrip** — should the post-RoundTrip stub apply to subsequent requests? Recommend: yes (documented). Add `TestLateRegistrationApplies`.
- **`http.Client` `Timeout` interaction** — if client sets timeout, does Transport respect it? Recommend: yes (transport doesn't override). Add `TestClientTimeoutHonored`.

### Special concerns

- **§NF3 "never makes a real network call"** is verified two ways: (1) package-import audit (test 57); (2) integration test that runs in a `iptables`-blocked or DNS-disabled fixture (best-effort, build-tagged Linux).
- **Falcon §1.13 disciplines** for `LoadFixtures` carry through (size cap 16 MiB, DisallowUnknownFields, depth, UTF-8) — tests 48–56.

---

## Package: `httpc/`

### Test files

- `httpc/client_test.go` — New, Default, Client options
- `httpc/methods_test.go` — Get/Post/Put/Patch/Delete/Head/Do (all generic)
- `httpc/body_test.go` — JSONBody[T] (per §23.17) / MultipartBody / RawBody / StreamBody / FormBody / WithRequestHeaders
- `httpc/retry_test.go` — every retry option, MaxAttempts, MaxElapsed, RetryOn, RetryIf
- `httpc/dryrun_test.go` — WithDryRun, WithPlanSink, WithDryRunErrors, IsDryRun
- `httpc/response_test.go` — Response wrapper, Body, Elapsed, Drain
- `httpc/errors_test.go` — StatusError, BodyParseError (per §23.15 rename), ErrDryRun, ErrMaxAttempts, ErrMaxElapsed
- `httpc/limits_test.go` — WithMaxResponseBytes / WithUnboundedResponse (§23.7)
- `httpc/redaction_test.go` — header redaction in plan, body omission in errors (§23.11)
- `httpc/concurrency_test.go` — concurrent Get/Post (-race)
- `httpc/lifecycle_test.go` — Client.Close (§23.16)
- `httpc/property_test.go` — property tests
- `httpc/fuzz_test.go` — `FuzzResponseBody`
- `httpc/bench_test.go` — benchmarks
- `httpc/example_test.go` — godoc examples
- `httpc/testdata/...` — golden plans, attacker-shaped JSON, oversized bodies

### Test matrix

| # | Name | Spec ref | Type | Description | Helpers |
|---|---|---|---|---|---|
| 1 | TestNewDefaultClient | §21.13 F1 | unit | New() returns valid client | assert |
| 2 | TestDefaultPackageVar | §21.13 F2 | unit | `httpc.Default` is non-nil | assert |
| 3 | TestWithTransport | §21.13 F3 | unit | Custom transport used (compose with httpmock) | httpmock |
| 4 | TestWithTimeout | §21.13 F3 | unit | Per-request default timeout applied | httpmock, fixture.NewClock |
| 5 | TestWithBaseURL | §21.13 F3 | unit | Relative URL prepended with base | httpmock |
| 6 | TestWithBaseURLJoinSafety | §23.9 row 16 | unit | After-join URL >8 KiB → typed error | assert |
| 7 | TestWithBaseURLRejectsUserinfo | §23.9 row 16 | unit | URL with `user:pass@` rejected | assert |
| 8 | TestWithHeaders | §21.13 F3 | unit | Headers sent on every request | httpmock |
| 9 | TestWithRetry | §21.13 F3/F15 | unit | Default retry policy applied client-wide | httpmock |
| 10 | TestWithLogger | §21.13 F3 | unit | Logger receives request lifecycle events | mock(slog handler) |
| 11 | TestGetTyped | §21.13 F4 (orig #1) | unit | `Get[User](ctx, url)` round-trip | httpmock, assert |
| 12 | TestGetTypedTypeInference | §23.17 | unit | Compile-only: `httpc.Get[User]` is the only T mention | compile-only |
| 13 | TestGetByteSliceSpecialCase | §21.13 E3 (orig #4) | unit | `Get[[]byte]` returns raw body | httpmock |
| 14 | TestHead | §21.13 F5 | unit | HEAD returns response with no body | httpmock |
| 15 | TestPostTyped | §21.13 F6 | unit | `Post[T]` with JSONBody | httpmock |
| 16 | TestPutTyped | §21.13 F7 | unit | Put round-trip | httpmock |
| 17 | TestPatchTyped | §21.13 F7 | unit | Patch round-trip | httpmock |
| 18 | TestDeleteTyped | §21.13 F7 | unit | Delete round-trip | httpmock |
| 19 | TestDoEscapeHatch | §21.13 F8 | unit | `Do(ctx, req)` for raw round-trip; honors retry/dry-run/config | httpmock |
| 20 | TestStatusErrorOnNon2xx | §21.13 E1 / orig #2 | unit | `Get[T]` 500 → StatusError; ErrorAs works | httpmock, assert.ErrorAs |
| 21 | TestStatusErrorContains | §21.13 F31 | unit | StatusError fields populated (Status, Body, Cause) | assert |
| 22 | TestStatusErrorErrorOmitsBody | §23.11 | redaction | `StatusError.Error()` does not include body bytes; field accessible | assert |
| 23 | TestBodyParseErrorOnBadJSON | §21.13 E2 / §23.15 (orig #3) | unit | Bad JSON → `*BodyParseError` (renamed from ParseError) | assert.ErrorAs |
| 24 | TestBodyParseErrorBodySnippet | §21.13 F32 | unit | Body field carries first 1 KiB | assert |
| 25 | TestBodyParseErrorErrorOmitsBody | §23.11 | redaction | `BodyParseError.Error()` does not include body | assert |
| 26 | TestResponseWrapperFields | §21.13 F9 | unit | Response.Body, Elapsed, http.Response present | assert |
| 27 | TestResponseDrain | §21.13 F9 | unit | Drain() closes unread body | assert |
| 28 | TestJSONBodyTyped | §23.17 (was F10 untyped) | unit | `JSONBody[T any](gen func() T)` typed; compile-time arg type | compile-only |
| 29 | TestJSONBodyClosurePerAttempt | §21.13 NF3 (orig #5) | unit | Closure invoked exactly N times for N attempts | httpmock |
| 30 | TestMultipartBody | §21.13 F11 | unit | Multipart body construction works | httpmock |
| 31 | TestMultipartBodyClosurePerAttempt | §21.13 F11 (orig #6) | unit | Multipart closure called per attempt | httpmock |
| 32 | TestRawBody | §21.13 F12 | unit | Bytes + content type | httpmock |
| 33 | TestStreamBody | §21.13 F13 | unit | Fresh ReadCloser per attempt | httpmock |
| 34 | TestStreamBodyCloserClosed | §21.13 E11 | unit | Previous attempt's ReadCloser closed before retry | mock(io.Closer) |
| 35 | TestFormBody | §21.13 F14 | unit | `application/x-www-form-urlencoded` | httpmock |
| 36 | TestBodyClosureReturnsError | §21.13 E6 | unit | Closure returns error → request never sent; error returned | httpmock |
| 37 | TestWithRequestHeaders | §21.13 F-req-opt | unit | Per-request headers added | httpmock |
| 38 | TestMaxAttemptsDefaultOne | §21.13 F16 | unit | Default no retry | httpmock |
| 39 | TestMaxAttemptsRetries | §21.13 F16 (orig #7) | unit | n-1 retries on 503; final returns `errors.Is(ErrMaxAttempts)` | httpmock, assert.ErrorIs |
| 40 | TestExponentialBackoff | §21.13 F17 (orig #8) | unit | Sleeps grow as base * 2^attempt | fixture.NewClock |
| 41 | TestLinearBackoff | §21.13 F18 | unit | Fixed delay | fixture.NewClock |
| 42 | TestJittered | §21.13 F19 (orig #9) | property | ±25% jitter; mean over 100 runs ≈ base | rapid, fixture.NewClock |
| 43 | TestRetryOnDefault | §21.13 F20 (orig #10) | unit | Default retry on 500/502/503/504/429 | httpmock |
| 44 | TestRetryOnCustom | §21.13 F20 | unit | `RetryOn(503)` only retries 503 | httpmock |
| 45 | TestRetryIf | §21.13 F21 (orig #11) | unit | Custom predicate triggers retry | httpmock |
| 46 | TestRetryIfAndRetryOnCombine | §21.13 F21 | unit | Either condition triggers retry | httpmock |
| 47 | TestMaxElapsedFires | §21.13 F22 / E5 (orig #12) | unit | Total time exceeds → ErrMaxElapsed wrapped | fixture.NewClock |
| 48 | TestMaxElapsedDuringBackoff | §21.13 F22 | unit | Backoff itself exceeds → ErrMaxElapsed | fixture.NewClock |
| 49 | TestRetryClosureCallCountExact | §21.13 NF3 / property | property | For random 1≤N≤10, body closure invoked exactly N times | rapid |
| 50 | TestRetryRespectsCtxCancel | §21.13 NF3 | unit | Ctx cancel mid-backoff → ErrMaxElapsed wrapped or ctx err | fixture |
| 51 | TestDryRunCapturesPlan | §21.13 F23/F24 (orig #13) | unit | `WithPlanSink` receives plan; no network call | mock(plan sink) |
| 52 | TestDryRunReturnsZeroT | §21.13 E9 | unit | Default mode: zero T, nil response, nil error | assert |
| 53 | TestDryRunErrorsMode | §21.13 E8 / orig #14 | unit | `WithDryRunErrors()` → ErrDryRun returned | assert.ErrorIs |
| 54 | TestIsDryRun | §21.13 F27 / orig #15 | unit | Reflects ctx state | assert |
| 55 | TestRequestPlanShape | §21.13 F26 | unit | Plan carries Request, Body bytes, retry, timeout | assert |
| 56 | TestRequestPlanWithBaseURLApplied | §21.13 F26 | unit | Plan reflects base URL join | assert |
| 57 | TestRequestPlanRedactsAuthorization | §23.11 | redaction | Plan render scrubs `Authorization` header | assert |
| 58 | TestRequestPlanRedactsCookie | §23.11 | redaction | `Cookie` redacted | assert |
| 59 | TestRequestPlanRedactsSetCookie | §23.11 | redaction | `Set-Cookie` redacted | assert |
| 60 | TestRequestPlanRedactsXApiKey | §23.11 | redaction | `X-Api-Key` redacted | assert |
| 61 | TestRequestPlanRedactsXAuthToken | §23.11 | redaction | `X-Auth-Token` redacted | assert |
| 62 | TestRequestPlanRedactsProxyAuthorization | §23.11 | redaction | `Proxy-Authorization` redacted | assert |
| 63 | TestRequestPlanRedactsByRegex | §23.11 | redaction | Custom header `X-Auth-Foo` matches `(?i)auth` regex → redacted | assert |
| 64 | TestWithPlanIncludeSecrets | §23.11 | redaction | Opt-in re-includes raw values | assert |
| 65 | TestMaxResponseBytesDefault32MiB | §23.7 | unit | 33 MiB response → ErrBodyTooLarge | httpmock (oversize stream) |
| 66 | TestMaxResponseBytesCustom | §23.7 | unit | `WithMaxResponseBytes(1MiB)` enforced | httpmock |
| 67 | TestWithUnboundedResponse | §23.7 | unit | Opt-out allows >32 MiB | httpmock |
| 68 | TestJSONDepthCap32 | §23.7 / §23.9 row 14 | unit | JSON depth >32 → BodyParseError | httpmock |
| 69 | TestUTF8ValidationOnJSONContentType | §23.7 | unit | Malformed UTF-8 in JSON → BodyParseError | httpmock |
| 70 | TestUTF8ValidationOnTextStar | §23.7 | unit | Malformed UTF-8 in text/* → BodyParseError | httpmock |
| 71 | TestGzipDecompressionPreCap | §23.7 | unit | gzip compressed body cap pre-decompression | httpmock |
| 72 | TestGzipDecompressionPostCap | §23.7 | unit | post-decompression cap (zip-bomb defense) | httpmock |
| 73 | TestResponseHeaderTotalCap8KiB | §23.9 row 15 | unit | Total response header bytes >8 KiB → typed error | httpmock |
| 74 | TestResponseHeaderNULRejected | §23.9 row 15 | unit | Header value with NUL → typed error | httpmock |
| 75 | TestResponseHeaderCRLFInjectionRejected | §23.9 row 15 | unit | CRLF in header value → typed error | httpmock |
| 76 | TestErrDryRunSentinel | §21.13 F28 | unit | `errors.Is(err, httpc.ErrDryRun)` works | assert |
| 77 | TestErrMaxAttemptsSentinel | §21.13 F29 | unit | sentinel comparable | assert |
| 78 | TestErrMaxElapsedSentinel | §21.13 F30 | unit | sentinel comparable | assert |
| 79 | TestComposeWithHttpmock | §21.13 (orig #16) | integration | `New(WithTransport(httpmock.New()))` works end-to-end | httpmock |
| 80 | TestConcurrentGet | §21.13 NF5 / E12 (orig #17, -race) | concurrency | 100 goroutines Get simultaneously; client safe | -race |
| 81 | TestConcurrentPostBodyClosures | §21.13 NF3/NF5 | concurrency | Concurrent Posts; closures not invoked concurrently *for same request* | -race, atomic counter |
| 82 | TestClientCloseIdempotent | §23.16 | lifecycle | Client.Close() twice; errs.Join | assert |
| 83 | TestClientCloseClosesOwnedTransport | §23.16 | lifecycle | When transport built internally, Close closes it; not closed when WithTransport supplied externally-owned | mock(io.Closer) |
| 84 | TestImportSurfaceNoThirdParty | §21.13 NF8 | audit | Stdlib + framework kernel only | go-list |
| 85 | TestTLSVerificationDefault_Linux | §21.13 X-platform | integration | TLS verification on Linux | build-tag |
| 86 | TestTLSVerificationDefault_Darwin | §21.13 X-platform | integration | TLS verification on macOS | build-tag |
| 87 | TestTLSVerificationDefault_Windows | §21.13 X-platform | integration | TLS verification on Windows | build-tag |
| 88 | PropertyGetThenPostEquivalentServerState | §21.13 / property (parent-task spec) | property | `httpc.Get[T]` then `Post[T]` with same body returns equivalent server-state via httpmock | rapid + httpmock |
| 89 | FuzzResponseBody | §23.10 | fuzz | Deep nesting, huge strings, malformed UTF-8 attacker JSON | testing.F |
| 90 | TestStaticDryRunHelperOnClient | §21.13 F23 | unit | `client.Get(ctx, ...)` honors ctx-level dry-run | mock(plan sink) |
| 91 | TestSlogLogValuerInRedactionContext | §23.11 | redaction | Headers wrapped in log.Redact render `[REDACTED]` regardless of plan-sink config | log.Redact |
| 92 | BenchmarkGetTyped_With_Httpmock | §23.13 | bench | ≤ 50 µs/op qualified "with httpmock transport" | httpmock |
| 93 | BenchmarkPostJSONBody | §21.13 NF1 | bench | post w/ JSONBody | httpmock |
| 94 | BenchmarkRetry5xx | §21.13 NF1 | bench | retry overhead bounded | httpmock, fixture.NewClock |
| 95 | BenchmarkDryRun | §21.13 NF1 | bench | Dry-run path overhead | mock |
| 96 | ExampleGet | §21.13 examples | doc-test | Headline typed Get | godoc |
| 97 | ExamplePost_JSONBody | §21.13 examples | doc-test | Closure-body retry-safe | godoc |
| 98 | ExampleWithDryRun | §21.13 examples | doc-test | Dry-run flag propagation | godoc |

### Coverage target

- **Line coverage:** 92% minimum.
- **Public API coverage:** 100% — every Get/Head/Post/Put/Patch/Delete/Do, every body builder, every retry option, every dry-run helper, every error type, every Client option.

### Edge cases not in spec

- **`Get[T]` where T is a pointer type** (`*User`) — spec implies value-type. Add `TestGetPointerT`.
- **`Get[T]` where T is `any`** — should it special-case to map[string]any? Spec is silent. Add `TestGetAnyT` (recommend: unmarshal as `any`, common JSON shape).
- **Retry on transport-level error (not status)** — RetryIf can detect, but RetryOn is status-only. Add `TestRetryOnTransportError`.
- **`MaxAttempts(0)` and `MaxAttempts(1)`** — both should mean "no retry". Add `TestMaxAttemptsZero`.
- **Negative MaxAttempts** — should panic. Add `TestMaxAttemptsNegative`.
- **`WithBaseURL` plus absolute URL in call** — absolute wins. Add `TestAbsoluteURLOverridesBase`.
- **Dry-run + retry** — does dry-run skip the retry loop entirely? Recommend: yes, plan emitted once. Add `TestDryRunSkipsRetryLoop`.
- **Multipart body builder with closure that returns error** — error surfaced; not sent. Add `TestMultipartBodyClosureError`.
- **Streaming response (chunked) under MaxResponseBytes** — partial read up to cap, then ErrBodyTooLarge. Add `TestStreamingResponseCapped`.
- **Empty body on 200** — `Get[User]` on empty body → BodyParseError? or zero-T-no-error? Recommend: BodyParseError (explicit). Add `TestEmptyBody200ParseError`.

### Special concerns

- **§NF1 perf gate** is "with httpmock transport" qualified per §23.13; benchmarks must run httpmock-only and document the qualifier in the benchstat output.
- **§23.7 response cap** is the most security-sensitive surface; tests 65–75 collectively form the Falcon-mandated cap suite.
- **§23.11 redaction** is multi-layered: header allowlist + regex + log.Redact; each redaction test verifies one layer.
- **`internal/safejson`** routes the JSON decoding (per §23.10); tests 68–70 + 89 verify it's the active code path (not raw `json.Decoder`).

---

## Cross-leaf integration tests

These live under `tests/integration/leaves/` (or per-package `*_integration_test.go` files build-tagged `integration`). They exercise leaf-to-leaf composition that the runtime explicitly supports.

| # | Name | Spec ref | Description | Helpers |
|---|---|---|---|---|
| I1 | TestHttpcWithHttpmockComposition | §21.13 / §21.11 | `httpc.New(httpc.WithTransport(httpmock.New()))` end-to-end with stubs + retries + dry-run | httpc, httpmock, assert |
| I2 | TestCliWithMockedHandlerDeps | §21.9 / §21.10 | A cli command's handler depends on a `Repo` interface; test wires `mock.Of[Repo]` and runs `App.Run` | cli, mock, assert |
| I3 | TestCliWithHttpcAndHttpmock | §21.9 / §21.11 / §21.13 | A cli handler makes httpc calls; httpmock provides responses; full Run lifecycle | cli, httpc, httpmock, assert |
| I4 | TestCliDryRunPropagation | §21.9 / §21.13 F23 | cli flag `--dry-run=true` flips ctx dry-run; httpc calls inside become plan-only | cli, httpc, mock(plan sink) |
| I5 | TestCodegenedCliBuildsAndRuns | §21.9 / §23.8 | Run cligen on a testdata module, `go build`, exec, verify output | fixture.RunGo, cli/gen |
| I6 | TestCodegenedMockBuildsAndRuns | §21.10 F19-22 | Run cligen on `+glacier:mock` interface; build; use generated `RepoMock` in a test | fixture.RunGo, cli/gen, mock |
| I7 | TestRuntimeAndCodegenMocksCoexist | §21.10 / orig #23 | Same test uses `mock.Of[A]` runtime mock and codegen `BMock` for B | mock, assert |
| I8 | TestCli_Mock_Httpmock_Httpc_FullStack | "5-package composition" (Otter A1) | A handler under cli takes a Repo (mock), calls httpc (with WithTransport(httpmock)) | cli, mock, httpmock, httpc, assert |
| I9 | TestE2E_DogfoodedGlacier_Argv_Conf_Httpc | E2E smoke (when SDK lands; v0 placeholder) | The `glacier` binary parses argv, loads conf, makes httpc requests with traces. Build-tagged `glacier_sdk` since SDK is post-v0 (per §23.18). | cli, conf, httpc, obs, mock, httpmock |
| I10 | TestImportGraphRespectsLeafTier | §6 architecture | Static check: cli does not import mock; mock does not import httpmock; httpmock does not import httpc; httpc does not import httpmock (composition is in user code only) | go-list-deps |
| I11 | TestCloseAuditAcrossLeaves | §23.16 | Construct App + Mock + Transport + Client; close each in arbitrary order; all idempotent and errs.Join'd | cli, mock, httpmock, httpc, assert |
| I12 | TestPlanRedactionSurvivesCliWiring | §23.11 / §21.9 | cli renders dry-run plan to user (via Magpie-watched output); secrets redacted by default | cli, httpc, fixture.CaptureStdout |
| I13 | TestConcurrentE2EUnderRace | -race | I3 scenario run concurrently from N=10 goroutines; no race | -race |

---

## Test helper dogfooding inventory

The leaves are the **primary validation that the helpers are ergonomic**. Required imports per leaf's test files:

| Leaf | Helpers used (mandatory) | Notes |
|---|---|---|
| `cli` tests | `assert`, `assert/require`, `fixture` (CaptureStdout/CaptureStderr/SetEnv/NewClock/WatchGoroutines/RunGo), `mock` (slog handler interface, deps in handler tests), `httpmock` (when handler makes HTTP calls in integration tests) | All five leaf-tier helpers used; no raw `t.Errorf` / `t.Fatal` in tests; all `t.Setenv` routed through `fixture.SetEnv` |
| `cli/gen` tests | `assert`, `assert/require`, `fixture` (NewFS, GoldenFile, RunGo), `mock` (logger, file-system shim where applicable) | No raw filesystem ops outside fixture |
| `mock` tests | `assert`, `assert/require`, `httpmock` (only in `mock_codegen_integration_test` where mock interacts with HTTP-shaped interfaces, **and** for the meta-test `TestMockUsesAssertHelpersInTests`) | mock testing assert + httpmock per the assignment's "mock tests will use assert and httpmock" rule |
| `httpmock` tests | `assert`, `assert/require`, `fixture` (NewFS for fixture loading testdata), `mock` (where httpmock collaborates with mocked io.Closer / slog handler / etc.) | LoadFixtures tests use fixture.NewFS to construct on-disk testdata isolated per test |
| `httpc` tests | `assert`, `assert/require`, `httpmock` (transport for **every** non-trivial httpc test), `fixture` (NewClock for retry timing, CaptureStdout for log capture, WatchGoroutines for goroutine-leak guards), `mock` (plan sinks, slog handlers, io.Closer for Drain tests) | Every retry test uses `fixture.NewClock` (no real `time.Sleep`); every transport test uses `httpmock` (never real `http.DefaultTransport`) |

**Audit test:** `tests/dogfood/leaves_dogfood_test.go` runs `go list -deps -test ./cli/... ./mock/... ./httpmock/... ./httpc/...` and asserts that every leaf's `_test.go` package depends on `assert` (and that no leaf's `_test.go` package depends on raw test-helper-bypass patterns like a custom assertion lib).

---

## End-to-end smoke tests

These are the highest-confidence "release gate" checks; they are subset of integration tests above, run on CI as a separate `e2e-smoke` job.

| # | Name | Description |
|---|---|---|
| E1 | TestSmoke_AllLeavesCompile | `go build ./cli/... ./mock/... ./httpmock/... ./httpc/...` succeeds on Linux/macOS/Windows |
| E2 | TestSmoke_AllLeavesPassRace | `go test -race ./cli/... ./mock/... ./httpmock/... ./httpc/...` passes on Linux |
| E3 | TestSmoke_FivePackageComposition | I8 above; the headline composition test |
| E4 | TestSmoke_CodegenDriftClean | `glaciergen --check ./...` exit 0 on clean tree (per §23.12) |
| E5 | TestSmoke_FuzzShortRun | All fuzz targets run for 30s under `-fuzz=Fuzz` and find no crash |
| E6 | TestSmoke_BenchAllTargetsHit | All §23.13-recalibrated targets hit on the OS-baseline runner; benchstat <5% regression vs main |
| E7 | TestSmoke_PublicAPISignaturesUnchanged | Use `golang.org/x/exp/apidiff` (or hand-rolled) to detect public-API drift vs spec-locked surface |
| E8 | TestSmoke_NoUnsafeAcrossLeaves | All four leaf packages contain zero `unsafe` imports |
| E9 | TestSmoke_GenericsCompileChecks | Compile-only file `tests/compile/generics_check.go` that exercises every generic surface (Mock[T], Matcher[T], JSONBody[T], BodyJSON[T], Get[T], Post[T], etc.) — verifies §23.17 fixes hold at the type level |

---

## Sign-off conditions

Lynx will not sign off on `cli/`, `mock/`, `httpmock/`, or `httpc/` reaching `accepted` until **all** of the following hold:

1. **Every test row above is implemented and green** under `go test -race ./...` on Linux, macOS, and Windows CI runners (cross-platform tests honor build tags; signal-handling tests only assert OS-relevant matrix rows).
2. **Coverage thresholds met:** cli ≥ 90% line + 100% public API; cli/gen ≥ 92%; mock ≥ 92%; httpmock ≥ 95%; httpc ≥ 92%; **all four leaves: 100% public-symbol coverage**.
3. **Benchmarks meet §23.13-recalibrated targets** with the per-OS baseline (per §23.12 benchstat-flake mitigation): Run-simple ≤ 5 µs (cli); MockCall ≤ 6 allocs/op (mock); RoundTripFirstMatch ≤ 5 µs (httpmock); GetTyped ≤ 50 µs/op qualified "with httpmock" (httpc).
4. **Fuzz gates green:** `cli/gen:FuzzMarkerParse`, `cli:FuzzArgvParse`, `httpmock:FuzzFixture`, `httpc:FuzzResponseBody` each accumulate ≥ 1 hour CI fuzz time without crash before tag.
5. **§23.8 codegen safety** verified: tests 94–98 (strconv.Quote invariant, no `fmt.Sprintf`-quoted strings in emitter, output-path canonicalization) pass; `--check` drift gate (test 112) green on PR-protect job.
6. **§23.9 untrusted-input boundary tests** all green: HTTP body cap, header cap, marker payload, output path, mock method-name regex (test rows 6–10 (cli), 65–75 (httpc), 9–10 (mock), 50–53 (httpmock)).
7. **§23.11 info-disclosure defaults** all green: redaction tests 57–64 (httpc), 22, 25 (httpc errors), test 56 (mock), test I12 (cli E2E redaction).
8. **§23.13 perf targets** documented and met (benchmark rows 63–65 cli, 64–68 mock, 66–69 httpmock, 92–95 httpc).
9. **§23.14 race lock-ins** verified: tests 47 (mock Times(1)), 59 (httpmock Times(1)).
10. **§23.16 Close audit** complete: cli App.Close (tests 55–56), httpmock Transport.Close (60–61), mock.Mock[T].Close (53–55), httpc.Client.Close (82–83), and the cross-leaf Close-audit test I11.
11. **§23.17 generics fixes** verified at compile time: tests 22 (mock Matcher[T]), 17 (httpmock BodyJSON[T]), 12 + 28 (httpc Get[T] + JSONBody[T]) and the compile-only smoke E9.
12. **Dogfooding inventory** complete: every leaf test file imports `assert`; every retry test uses `fixture.NewClock`; every transport test uses `httpmock`; every interface dependency in cli handler tests uses `mock`. Audit test in `tests/dogfood/` green.
13. **Spec traceability:** every test name comment cites a §-number or decision ID; CI job `spec-traceability` greps for tests missing the citation comment and fails on any.
14. **Cross-platform CI matrix** green for the OS-relevant rows: cli signal handling (tests 28–30), httpc TLS (85–87), codegen output path UNC (test 98 on Windows).
15. **End-to-end smoke job** green (E1–E9 above).
16. **No leaf test bypasses test helpers** (audit: no `t.Setenv`, no raw `t.TempDir` outside fixture, no raw `time.Sleep` in retry tests, no raw `http.DefaultTransport` in httpc tests).

---

### Appendix A — test count summary

| Package | Unit | Codegen | Property | Bench | Fuzz | Concurrency / lifecycle / audit / X-plat | Total in matrix |
|---|---|---|---|---|---|---|---|
| cli (runtime) | 56 | 0 | 0 | 3 | 1 | 7 (2 race, 1 lifecycle, 1 audit, 3 signal/X-plat) | 67 |
| cli/gen | 0 | 64 (rows 68–131) | 2 (114, 115) | 1 (131) | 1 (125) | 2 (123, 124 concurrency) | 64 |
| mock | 49 | 1 (59) | 3 (60, 61, 62) | 5 | 0 | 11 (3 race, 3 lifecycle, 2 audit, 3 misc) | 69 |
| httpmock | 51 | 0 | 1 (31) | 4 | 1 | 13 (2 race, 2 lifecycle, 4 audit, 5 X-plat/security) | 70 |
| httpc | 73 | 0 | 2 (49, 88) | 4 | 1 | 18 (2 race, 2 lifecycle, 1 audit, 3 X-plat, 10 security/cap/redaction) | 98 |
| **Cross-leaf** | — | — | — | — | — | 13 | 13 |
| **E2E smoke** | — | — | — | — | — | 9 | 9 |
| **TOTAL** | | | | | | | **~390** |

Comfortably above the assignment's "230+ across the four leaves" floor; the inflation is largely §23.7/§23.9/§23.11 security cap + redaction tests on httpc and the codegen safety surface in cli/gen, both of which are non-negotiable per Falcon's lock-ins.

### Appendix B — files this review will *not* allow to ship without addition

- `cli/gen/safety_test.go` — must contain tests 94–99 verbatim; cli/gen will not be accepted without strconv.Quote invariant proven.
- `httpc/redaction_test.go` — must contain tests 57–64 verbatim; httpc will not be accepted with default plan-sink leaking auth headers.
- `httpc/limits_test.go` — must contain tests 65–75; the 32 MiB cap is a Falcon §1 invariant.
- `httpc/lifecycle_test.go` and the parallel `mock/lifecycle_test.go`, `httpmock/lifecycle_test.go`, `cli/lifecycle_test.go` — §23.16 Close audit is gating.
- `tests/integration/leaves/composition_test.go` — must contain I8 (the five-package composition); without it, the dogfooding claim is unverified.