// Cross-reference table between task pages and package pages.
// See specs/0031-public-site.md §Architecture / Information architecture.

export interface TaskMeta {
  slug: string
  title: string
  packagesUsed: readonly string[]
}

export const tasks: readonly TaskMeta[] = [
  {
    slug: 'building-a-cli',
    title: 'Building a CLI',
    packagesUsed: ['cli', 'option', 'errs', 'log', 'term', 'conf'],
  },
  {
    slug: 'writing-tests',
    title: 'Writing tests',
    packagesUsed: ['assert', 'fixture', 'mock', 'errs'],
  },
  {
    slug: 'mocking-http',
    title: 'Mocking HTTP',
    packagesUsed: ['httpmock', 'httpc', 'fixture'],
  },
  {
    slug: 'loading-config',
    title: 'Loading config',
    packagesUsed: ['conf', 'option', 'errs', 'log'],
  },
  {
    slug: 'structured-logging',
    title: 'Structured logging',
    packagesUsed: ['log', 'errs', 'obs'],
  },
  {
    slug: 'observability',
    title: 'Observability',
    packagesUsed: ['obs', 'log', 'httpc'],
  },
  {
    slug: 'concurrency',
    title: 'Concurrency',
    packagesUsed: ['concur', 'errs', 'log'],
  },
] as const

// Inverse lookup: which task pages compose a given package?
export function tasksUsingPackage(name: string): readonly TaskMeta[] {
  return tasks.filter((t) => t.packagesUsed.includes(name))
}

// Lookup task metadata by slug.
export function taskBySlug(slug: string): TaskMeta | undefined {
  return tasks.find((t) => t.slug === slug)
}
