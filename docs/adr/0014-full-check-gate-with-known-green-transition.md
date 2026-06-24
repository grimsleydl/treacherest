# Full check gate with known-green transition

Status: accepted

Treacherest's `just check` should represent the desired full verification gate, even if it fails while pre-existing baseline tests are red. During the migration, a temporary `just check-known-green` command may provide a reliable passing subset for day-to-day work and agent verification. This keeps the main gate honest while avoiding a false claim that the current partial green bar is release-complete.
