# Neptune 🌊

A Terraform pull request automation tool inspired by Atlantis! 🔱

## Features

- Run Terraform plans on pull requests (bring the same experience as Atlantis)
- Apply approved Terraform changes
- Easy to use and integrate with your workflow
- Type-safe CLI with auto-completion support

## Dependencies

- The Infrastructure as Code Repository must have a `.neptune.yaml` file with the following structure:

```yaml
repository:
  object_storage: gs://object_storage_url
  branch: master
  plan_requirements:
    - undiverged
  apply_requirements:
    - approved
    - mergeable
    - undiverged
  allowed_workflow: default

workflows:
  default:
    plan:
      steps:
        - run: echo "Custom command"
        - run: terramate run --parallel $(nproc --all) --changed -- terragrunt init -upgrade
        - run: terramate run --changed -- terragrunt plan
    apply:
      depends_on:
        - plan
      steps:
        - run: echo "Custom command"
        - run: terramate run --changed -- terragrunt apply -auto-approve
```

- The repository must have Terramate orchestrating the Terraform stacks and be able to use the `--changed` flag to run the terramate commands.

## Installation

### Go (recommended)

```bash
# Build from source
go build -o neptune .

# Or install into $GOPATH/bin
go install .
```

Binaries for Linux, macOS, and Windows are published to the GitHub Releases page when you push a version tag (e.g. `v0.2.0`) via GoReleaser.

### Python (legacy)

```bash
# From the repository root
pip install -e .

# Enable shell completion (optional)
neptune --install-completion
```

### Using with GitHub Actions

Set the same environment variables in your workflow (`GITHUB_REPOSITORY`, `GITHUB_PULL_REQUEST_BRANCH`, `GITHUB_PULL_REQUEST_NUMBER`, `GITHUB_PULL_REQUEST_COMMENT_ID`, `GITHUB_RUN_ID`, `GITHUB_TOKEN`), then run `neptune command plan` or `neptune command apply` as needed.

## Usage

```bash
# Show help
neptune --help

# Print version
neptune version

# Run a workflow phase (plan or apply)
neptune command plan
neptune command apply

# Unlock all stacks (requires --all)
neptune unlock --all
```

## Development

### Go

```bash
make build      # build binary
make test-all   # run tests
make check-fmt  # check formatting
make lint       # run golangci-lint (optional)
```

### Python (legacy)

This tool is built with:
- Typer - For modern, type-safe CLI interface
- Rich - For beautiful terminal output (included with Typer[all])
- PyYAML - For configuration handling
- Requests - For API interactions

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

3. Create a .env file with the following variables:
   ```bash
   GITHUB_REPOSITORY=https://github.com/example/infrastructure           # Update this with the repository url that you are testing
   GITHUB_PULL_REQUEST_BRANCH=feat/add-pull-request-automation-framework # Update this with the pull request branch that you are testing
   GITHUB_PULL_REQUEST_NUMBER=281                                        # Update this with the pull request number that you are testing
   GITHUB_PULL_REQUEST_COMMENT_ID=2945607606                             # Update this with the pull request comment id that you are testing
   GITHUB_RUN_ID=1234567890                                              # Update this with the GitHub Actions run id that you are emulating
   ```

4. Export the .env file:
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

6. Copy the .neptune.example.yaml and customize it to your needs:
   ```bash
   cp .neptune.example.yaml .neptune.yaml
   ```

7. Run the tool:
   ```bash
   python3 neptune/cli.py --help
   ```
