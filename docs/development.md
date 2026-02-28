# Development

## End-to-end tests

E2E tests run Neptune against MinIO via Docker Compose and do not require a real GitHub PR. See [e2e/README.md](../e2e/README.md) for prerequisites and how to run `./e2e/run.sh`. PR emulation is done with an isolated git repo (main + pr-1 branch with changed stacks) and `NEPTUNE_E2E=1` so GitHub checks are skipped.

## Go

```bash
make build      # build binary
make test-all   # run tests
make check-fmt  # check formatting
make lint       # run golangci-lint (optional)
```

Use the Go version from `go.mod`. See [AGENTS.md](../AGENTS.md) for code style, testing, and CI.

## Python (legacy)

This tool is built with:

- Typer – For modern, type-safe CLI interface
- Rich – For beautiful terminal output (included with Typer[all])
- PyYAML – For configuration handling
- Requests – For API interactions

To set up for development:

1. Create a virtual environment:

   ```bash
   python3 -m venv --clear --prompt "venv: dev" .venv
   source .venv/bin/activate
   ```

2. Install development dependencies:

   ```bash
   pip install -e ".[dev]"
   ```

3. Create a `.env` file with the following variables:

   ```bash
   GITHUB_REPOSITORY=https://github.com/example/infrastructure           # Update this with the repository url that you are testing
   GITHUB_PULL_REQUEST_BRANCH=feat/add-pull-request-automation-framework # Update this with the pull request branch that you are testing
   GITHUB_PULL_REQUEST_NUMBER=281                                        # Update this with the pull request number that you are testing
   GITHUB_RUN_ID=1234567890                                              # Update this with the GitHub Actions run id that you are emulating
   ```

4. Export the `.env` file:

   ```bash
   export $(cat .env | xargs)
   ```

5. Authenticate with GitHub and GCP:

   ```bash
   gh auth login
   export GITHUB_TOKEN=$(gh auth token)
   gcloud auth login
   gcloud auth application-default login
   ```

6. Copy the `.neptune.example.yaml` and customize it to your needs:

   ```bash
   cp .neptune.example.yaml .neptune.yaml
   ```

7. Run the tool:

   ```bash
   python3 neptune/cli.py --help
   ```
