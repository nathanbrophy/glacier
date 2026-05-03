import { defineConfig } from 'vitepress'
import { withMermaid } from 'vitepress-plugin-mermaid'

// Public site config. See specs/0031-public-site.md §Architecture / Build & deploy.
export default withMermaid(defineConfig({
  title: 'Glacier',
  description: 'Less plumbing. More Go.',
  base: '/glacier/',
  cleanUrls: true,
  lang: 'en-US',
  appearance: 'force-dark',
  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/glacier/favicon.svg' }],
    ['link', { rel: 'mask-icon', href: '/glacier/favicon.svg', color: '#22D3EE' }],
    ['link', { rel: 'apple-touch-icon', href: '/glacier/favicon.svg' }],
    ['meta', { name: 'theme-color', content: '#0E1116' }],
    ['meta', { name: 'color-scheme', content: 'dark' }],
    ['meta', { property: 'og:title', content: 'Glacier' }],
    ['meta', { property: 'og:description', content: 'Less plumbing. More Go.' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
  ],
  themeConfig: {
    logo: { src: '/wordmark.svg', alt: 'GLACIER', width: 160, height: 32 },
    siteTitle: false,
    nav: [
      { text: 'Why', link: '/why' },
      { text: 'SDK', link: '/sdk/' },
      { text: 'Features', link: '/features' },
      { text: 'Examples', link: '/examples' },
      { text: 'Concepts', link: '/concepts' },
      { text: 'Docs', link: '/docs/' },
    ],
    socialLinks: [
      { icon: 'github', link: 'https://github.com/nathanbrophy/glacier' },
    ],
    footer: {
      message: 'Apache-2.0',
      copyright: 'github.com/nathanbrophy/glacier',
    },
    search: {
      provider: 'local',
    },
    sidebar: {
      '/sdk/': [
        {
          text: 'Glacier SDK',
          collapsed: false,
          items: [
            { text: 'Overview',      link: '/sdk/' },
            { text: 'Install',       link: '/sdk/install' },
            { text: 'Configuration', link: '/sdk/configuration' },
          ],
        },
        {
          text: 'Commands',
          collapsed: false,
          items: [
            { text: 'All commands',      link: '/sdk/commands/' },
            { text: 'glacier init',      link: '/sdk/commands/init' },
            { text: 'glacier new',       link: '/sdk/commands/new' },
            { text: 'glacier generate',  link: '/sdk/commands/generate' },
            { text: 'glacier lint',      link: '/sdk/commands/lint' },
            { text: 'glacier test',      link: '/sdk/commands/test' },
            { text: 'glacier explain',   link: '/sdk/commands/explain' },
            { text: 'glacier version',   link: '/sdk/commands/version' },
            { text: 'glacier completions', link: '/sdk/commands/completions' },
            { text: 'glacier vibe',      link: '/sdk/commands/vibe' },
          ],
        },
        {
          text: 'Codegen examples',
          collapsed: true,
          items: [
            { text: 'Overview',                          link: '/sdk/codegen/' },
            { text: 'cli: all markers',                  link: '/sdk/codegen/cli-all-markers' },
            { text: 'cli: nested commands',              link: '/sdk/codegen/cli-nested' },
            { text: 'cli: simple',                       link: '/sdk/codegen/cli-simple' },
            { text: 'mock: interface only',              link: '/sdk/codegen/mock-interface-only' },
            { text: 'mock: nested interfaces',           link: '/sdk/codegen/mock-nested-interfaces' },
            { text: 'mock: types from other package',    link: '/sdk/codegen/mock-with-types-from-other-package' },
            { text: 'httpmock: programmable router',     link: '/sdk/codegen/httpmock-programmable-router' },
            { text: 'httpmock: recording disabled',      link: '/sdk/codegen/httpmock-recording-disabled' },
            { text: 'httpmock: body closure',            link: '/sdk/codegen/httpmock-with-body-closure' },
          ],
        },
      ],
      '/docs/': [
        {
          text: 'Tasks',
          collapsed: false,
          items: [
            { text: 'Overview', link: '/docs/' },
            { text: 'Building a CLI',     link: '/docs/building-a-cli' },
            { text: 'Writing tests',      link: '/docs/writing-tests' },
            { text: 'Mocking HTTP',       link: '/docs/mocking-http' },
            { text: 'Loading config',     link: '/docs/loading-config' },
            { text: 'Structured logging', link: '/docs/structured-logging' },
            { text: 'Observability',      link: '/docs/observability' },
            { text: 'Concurrency',        link: '/docs/concurrency' },
          ],
        },
        {
          text: 'Packages: Kernel',
          collapsed: false,
          items: [
            { text: 'option', link: '/docs/packages/option' },
            { text: 'errs',   link: '/docs/packages/errs' },
            { text: 'log',    link: '/docs/packages/log' },
            { text: 'assert', link: '/docs/packages/assert' },
            { text: 'term',   link: '/docs/packages/term' },
          ],
        },
        {
          text: 'Packages: Mid',
          collapsed: false,
          items: [
            { text: 'concur',  link: '/docs/packages/concur' },
            { text: 'fluent',  link: '/docs/packages/fluent' },
            { text: 'conf',    link: '/docs/packages/conf' },
            { text: 'fixture', link: '/docs/packages/fixture' },
            { text: 'obs',     link: '/docs/packages/obs' },
          ],
        },
        {
          text: 'Packages: Leaves',
          collapsed: false,
          items: [
            { text: 'cli',      link: '/docs/packages/cli' },
            { text: 'mock',     link: '/docs/packages/mock' },
            { text: 'httpmock', link: '/docs/packages/httpmock' },
            { text: 'httpc',    link: '/docs/packages/httpc' },
            { text: 'cache',    link: '/docs/packages/cache' },
          ],
        },
      ],
    },
    outline: {
      level: [2, 3],
      label: 'On this page',
    },
  },
  vite: {
    server: {
      fs: {
        // Allow serving the wordmark.txt asset from the parent assets/ directory in dev.
        allow: ['..'],
      },
    },
  },
  mermaid: {
    // Mermaid theme tuned to the Glacier dark palette.
    theme: 'dark',
    themeVariables: {
      darkMode: true,
      background: '#0E1116',
      primaryColor: '#161B22',
      primaryTextColor: '#E6EDF3',
      primaryBorderColor: '#22D3EE',
      lineColor: '#30363D',
      secondaryColor: '#1F262E',
      tertiaryColor: '#1F262E',
      clusterBkg: '#161B22',
      clusterBorder: '#30363D',
      titleColor: '#E6EDF3',
    },
  },
}))
