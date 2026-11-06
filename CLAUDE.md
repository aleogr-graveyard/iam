# Working agreements

- Talk to the owner in **Portuguese**. Code, identifiers, comments and commit
  messages are in English; versioned documentation is in Portuguese until the
  first release (docs/brief.md, 4.28).
- **Guiding the owner through manual steps: one command at a time.** Give a
  single command, wait for the owner to run it and show the output, and only
  then give the next one. Never hand over a batch of commands, even when they
  are independent. This holds in every session and every repository.
- This repository is **public**. No secret may ever be committed.
- Terraform under `deploy/` is applied only by CI after a merge to `main`,
  except for the documented bootstrap (deploy/gcp/README.md). A session never
  runs `apply`.
