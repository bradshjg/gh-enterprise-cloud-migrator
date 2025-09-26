# gh-enterprise-to-cloud-migrator

Web UI for managing GitHub Enterprise Server to GitHub Enterprise Cloud migrations

## How it works

After supplying Personal Access Tokens (PATs) for the source and destination, select repos to migrate.

* A script will be generated and run to migrate the selected repos, and migration output will be displayed.
* Tokens are stored at rest client-side in encrypted cookies and only kept in memory server-side for the duration of a migration run.

## Demo (includes narration)

https://github.com/user-attachments/assets/c4a1e61c-d433-4260-82ce-ef876beb66a2

## Deploying

In addition to the `ghes-to-ghec` binary that starts the webserver, you'll need:

* `gh` and `gh gei` available on your `PATH`
* `pwsh` (PowerShell) available on your `PATH`
* environment variable configuration (see `.env.example`)

See the included `Dockerfile` as a starting point

## Acknowledgements

The [GitHub Enterprise Importer](https://github.com/github/gh-gei) does a ton of the heavy-lifting here!
