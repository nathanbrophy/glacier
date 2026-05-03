// 14-package manifest for the public site. See specs/0002-framework-shape.md
// for the canonical three-tier DAG and specs/0031-public-site.md §Schema for
// the PackageManifest invariant (length === 14; tier counts {kernel: 5, mid: 5, leaf: 4}).

import type { PackageManifest } from '../types'

export const packages: PackageManifest = [
  // Tier 0: kernel
  {
    name: 'option',
    slug: 'option',
    tier: 'kernel',
    specId: '0003',
    teaser: 'Functional options. Apply, Validate, Required, Option[T]: every Glacier package configurable at construction speaks this protocol.',
  },
  {
    name: 'errs',
    slug: 'errs',
    tier: 'kernel',
    specId: '0004',
    teaser: 'Error stories without ceremony. Wrap, Join, Chain, Sentinel, IsAny, Retryable, Coded: tree-walking, classification, library and CLI registers.',
  },
  {
    name: 'log',
    slug: 'log',
    tier: 'kernel',
    specId: '0005',
    teaser: 'Structured logging on log/slog with Trace and Notice levels, brand-palette TTY color, context attribute attachment, and a Redact helper.',
  },
  {
    name: 'assert',
    slug: 'assert',
    tier: 'kernel',
    specId: '0006',
    teaser: 'Test assertions and runtime invariants. Equal[T] with smart deep-compare, Must for init-time panics, require/ for halt-on-failure.',
  },
  {
    name: 'term',
    slug: 'term',
    tier: 'kernel',
    specId: '0016',
    teaser: 'Terminal as first-class output. Capability detection, 24-bit ANSI styling, glyph registry, beauty-writer layout, prompts, animation.',
  },

  // Tier 1: mid
  {
    name: 'concur',
    slug: 'concur',
    tier: 'mid',
    specId: '0007',
    teaser: 'Concurrency primitives that play with context: Mutex, RWMutex, Group with panic recovery, Semaphore, Pool[T], Once[T], WaitGroup.',
  },
  {
    name: 'fluent',
    slug: 'fluent',
    tier: 'mid',
    specId: '0008',
    teaser: 'Lazy iter.Seq pipeline operators: Map, Filter, Take, Window, GroupBy, joins, set ops, aggregations. Generics-first; zero deps.',
  },
  {
    name: 'conf',
    slug: 'conf',
    tier: 'mid',
    specId: '0009',
    teaser: 'Layered configuration with atomic snapshots. Defaults, JSON file, env, flags, overrides: Register[T] returns a typed accessor.',
  },
  {
    name: 'fixture',
    slug: 'fixture',
    tier: 'mid',
    specId: '0010',
    teaser: 'Test resources: golden files, typed snapshots, deterministic fake clocks, in-memory filesystems, leak guards for goroutines, FDs, env vars.',
  },
  {
    name: 'obs',
    slug: 'obs',
    tier: 'mid',
    specId: '0017',
    teaser: 'Opt-in OpenTelemetry: MeterProvider and TracerProvider via OTLP gRPC, instrumentation hooks for httpc / cli / conf, zero overhead when off.',
  },

  // Tier 2: leaves
  {
    name: 'cli',
    slug: 'cli',
    tier: 'leaf',
    specId: '0011',
    teaser: 'Build CLIs from a struct and a Run method. Markers + glaciergen codegen emit flag parsing, help text, routing. Banner via go:embed.',
  },
  {
    name: 'mock',
    slug: 'mock',
    tier: 'leaf',
    specId: '0012',
    teaser: 'Interface mocks: mock.Of[T] reflect-based, or +glacier:mock codegen for typed wrappers. Fluent expectation builder, automatic Verify on cleanup.',
  },
  {
    name: 'httpmock',
    slug: 'httpmock',
    tier: 'leaf',
    specId: '0013',
    teaser: 'A programmable http.RoundTripper for tests. Stub builder, generic JSON[T] responses, strict-by-default; testdata fixtures land in JSON.',
  },
  {
    name: 'httpc',
    slug: 'httpc',
    tier: 'leaf',
    specId: '0015',
    teaser: 'Typed HTTP client: Get[T], Post[T], Put[T] auto-unmarshal. Retry with backoff, closure-body retry-safe payloads, dry-run via context.',
  },
] as const
