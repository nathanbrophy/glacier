<!--
  Banner ASCII is the canonical art from specs/0001-brand-identity.md.
  Source of truth: assets/logo/banner.txt. Do not edit here in isolation —
  any change is a brand-identity spec change.
-->

```
                          ███╗   ███╗ ██████╗ ███╗   ██╗ ██████╗  ██████╗  ██████╗ ███████╗███████╗
       .-""""-.            ████╗ ████║██╔═══██╗████╗  ██║██╔════╝ ██╔═══██╗██╔═══██╗██╔════╝██╔════╝
      / -    - \           ██╔████╔██║██║   ██║██╔██╗ ██║██║  ███╗██║   ██║██║   ██║███████╗█████╗
     ( =o    o= )          ██║╚██╔╝██║██║   ██║██║╚██╗██║██║   ██║██║   ██║██║   ██║╚════██║██╔══╝
      \    --  /__         ██║ ╚═╝ ██║╚██████╔╝██║ ╚████║╚██████╔╝╚██████╔╝╚██████╔╝███████║███████╗
       \_______/  \        ╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝  ╚═════╝  ╚═════╝ ╚══════╝╚══════╝
                                          Less plumbing. More Go.
```

Mongoose is a Go SDK that handles the plumbing so you can focus on what's yours. Like the animal it's named for, mongoose is small, alert, and fearless about the messy parts: argument parsing, configuration layering, lifecycle and signal handling, mock-driven testing, and HTTP transport faking. You write the logic. Mongoose keeps the den safe.

## Status

Mongoose is in early design. The repo currently holds the development lifecycle and the brand identity. Code lands as component specs are accepted.

- [`specs/`](specs/) — the source of truth. Every change is a spec first.
- [`specs/0000-spec-process.md`](specs/0000-spec-process.md) — how mongoose is built.
- [`specs/0001-brand-identity.md`](specs/0001-brand-identity.md) — what mongoose looks and feels like.
- [`CLAUDE.md`](CLAUDE.md) — the rules.

## The Promise

When you use mongoose, you should be able to say each of these truthfully:

1. *"I'm only writing what's mine."*
2. *"I trust the defaults."*
3. *"The error tells me what to do next."*
4. *"Tests are easy because the framework helps."*

Every component spec is reviewed against these four statements. If a design doesn't deliver them, the design is wrong.

## License

License will be selected when the first code spec lands.
