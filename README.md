# gh-enterprise-to-cloud-migrator

Web UI for managing GitHub Enterprise to GitHub Enterprise Cloud migrations

## How it works

After authenticating to the source and target GitHub instances via OAuth, select repos to migrate.

* A script will be generated and run to migrate the selected repos

## Deploying

In addition to the `ghes-to-ghec` binary that starts the webserver, you'll need:

* `gh` and `gh gei` available on your `PATH`
* `pwsh` (PowerShell) available on your `PATH`
* environment variable configuration (see `.env.example`)

See the included `Dockerfile` as a starting point

## Acknowledgements

The [GitHub Enterprise Importer](https://github.com/github/gh-gei) does a ton of the heavy-lifting here!
