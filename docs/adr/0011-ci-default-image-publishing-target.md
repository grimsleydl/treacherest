# CI-default image publishing target

Status: accepted

Treacherest's target release model is for CI to build, verify, tag, and publish container images, with local image pushing retained as a manual override until CI/CD exists. The current repo does not yet have CI/CD, so migration work should preserve local `just image-push` behavior while shaping commands and metadata so the same contract can move into CI later. Release identity should use immutable tags such as the short git SHA and semantic release tags rather than relying on `latest`. Dirty local working-tree images may use `latest` as a convenience tag because no clean commit identity exactly describes their contents.
