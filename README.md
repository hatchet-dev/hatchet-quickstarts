# hatchet-quickstarts

This repository owns the `hatchet quickstart` CLI templates and the generated
examples.

## Template layout

Templates live under `templates/`, are embedded into the module, and are
returned by `TemplatesFS()`. The Hatchet CLI templater reads fixed paths from
this tree, so the layout is a contract rather than an internal detail.

The default simple quickstart lives at the language root:

- `templates/go/` for Go.
- `templates/<language>/shared/` plus `templates/<language>/<package-manager>/`
  for Python and TypeScript, where the package-manager directory is overlaid on
  shared.

The main repo templater hardcodes these `templates/<language>/...` paths, so the
default templates must not move while released CLI versions depend on them.
Use-case templates therefore go under a separate subtree that mirrors the same
structure one level down:

- `templates/use-cases/<use-case>/go/`
- `templates/use-cases/<use-case>/<language>/shared/` plus
  `templates/use-cases/<use-case>/<language>/<package-manager>/`

This keeps the default quickstart untouched, avoids collisions with language
directory names, and lets a future `--use-case` path builder reuse the existing
overlay of shared and package-manager directories with a different root. No
use-case templates exist yet. This section records the agreed layout for when
the first one is added.
