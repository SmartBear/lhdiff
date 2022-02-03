# Releasing

- Update changelog. See [.github/RELEASING](https://github.com/cucumber/.github/blob/main/RELEASING.md) for details.
- Create a tag: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`
- Run `export GITHUB_TOKEN=...` - see https://github.com/settings/tokens
- Run `goreleaser release --rm-dist`

