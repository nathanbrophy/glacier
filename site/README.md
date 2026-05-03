# Glacier site

Source for `https://nathanbrophy.github.io/glacier/`. Authored per `specs/0031-public-site.md` at the repo root.

## Local development

```sh
cd site
npm install
npm run dev          # http://localhost:5173/glacier/
npm run build        # static output to .vitepress/dist/
npm run preview      # serve the built artifact
```

## One-time setup: vendored fonts

The site references three SIL OFL-licensed font families (Inter, Space
Grotesk, JetBrains Mono). They are NOT committed to the repo; download
them once with:

```sh
bash scripts/fetch-fonts.sh
```

Until you run it, the site falls back to the system `system-ui` /
`ui-monospace` stack and renders cleanly. Re-running is idempotent.

## Static checks

The full check sweep mirrors what CI runs on every PR:

```sh
bash scripts/check-site.sh
```

Individual checks live alongside it (`scripts/check-*.sh`); each one
exits non-zero on failure with a precise pointer at the offending file.
