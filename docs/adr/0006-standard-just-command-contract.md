# Standard just command contract

Status: accepted

Treacherest uses `just` as the human command surface, with a small standard command contract shared across projects and project-specific extras named separately. The shared commands are `dev`, `check`, `test`, `build`, `image`, `image-run`, `image-push`, and `release <tag>`; project-specific commands may exist when their names make the local concern explicit. This keeps daily muscle memory portable across projects without hiding project-specific build steps inside unrelated command names.
