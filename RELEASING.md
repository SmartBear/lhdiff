# Releasing

## Update changelog

See [.github/RELEASING](https://github.com/cucumber/.github/blob/main/RELEASING.md) for details. 

## Commit and create a tag

    git commit -am "Release v${next_release}"
    git tag -a "v${next_release}" -m "Release v${next_release}"
    git push --tags

## Publish executables

Get a [personal GitHub access token](https://github.com/settings/tokens).
You may have to authorize the token to push to this repo.

    export GITHUB_TOKEN=...
    goreleaser release --rm-dist

