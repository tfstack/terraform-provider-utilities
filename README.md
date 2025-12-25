# Terraform Provider utilities (Terraform Plugin Framework)

A Terraform provider for various utility functions and tools, including file extraction, directory management, HTTP requests, path operations, and cryptographic hashing.

## Features

- **File Extraction**: Extract ZIP and TarGz archives from local files or URLs
- **Directory Management**: Create and manage local directories with permissions and ownership
- **HTTP Requests**: Make HTTP requests and retrieve responses
- **Path Operations**: Check path existence, get file permissions, and query ownership
- **Cryptographic Hashing**: Generate bcrypt hashes for passwords and secrets

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22

## Installation

### Using Terraform Registry

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    utilities = {
      source  = "hashicorp.com/tfstack/utilities"
      version = "~> 0.1"
    }
  }
}

provider "utilities" {}
```

### Building from Source

1. Clone the repository:

   ```bash
   git clone https://github.com/tfstack/terraform-provider-utilities.git
   cd terraform-provider-utilities
   ```

2. Build the provider:

   ```bash
   go install
   ```

3. Install the provider to your local Terraform plugins directory:

   ```bash
   mkdir -p ~/.terraform.d/plugins/registry.terraform.io/tfstack/utilities/0.1.0/linux_amd64
   cp $GOPATH/bin/terraform-provider-utilities ~/.terraform.d/plugins/registry.terraform.io/tfstack/utilities/0.1.0/linux_amd64/
   ```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Quick Start

Here's a simple example to get you started:

```hcl
terraform {
  required_providers {
    utilities = {
      source  = "hashicorp.com/tfstack/utilities"
      version = "~> 0.1"
    }
  }
}

provider "utilities" {}

# Generate a bcrypt hash
data "utilities_bcrypt_hash" "password" {
  plaintext = "my-secret-password"
  cost      = 10
}

# Create a local directory
resource "utilities_local_directory" "example" {
  path        = "/tmp/example"
  permissions = "0755"
}

# Extract a ZIP file
resource "utilities_extract_zip" "example" {
  url         = "https://example.com/archive.zip"
  destination = "/tmp/extracted"
}

output "bcrypt_hash" {
  value = data.utilities_bcrypt_hash.password.hash
}
```

For more examples, see the [`examples/`](examples/) directory.

## Data Sources

- [`utilities_bcrypt_hash`](docs/data-sources/bcrypt_hash.md) - Generates a bcrypt hash from plaintext
- [`utilities_local_directory`](docs/data-sources/local_directory.md) - Retrieves information about a local directory

## Resources

- [`utilities_extract_tar_gz`](docs/resources/extract_tar_gz.md) - Extracts a TarGz archive to a specified directory
- [`utilities_extract_zip`](docs/resources/extract_zip.md) - Extracts a ZIP archive to a specified directory
- [`utilities_local_directory`](docs/resources/local_directory.md) - Creates and manages a local directory

## Functions

- [`utilities_http_request`](docs/functions/http_request.md) - Makes an HTTP request and returns the response
- [`utilities_path_exists`](docs/functions/path_exists.md) - Checks if a file or directory path exists
- [`utilities_path_owner`](docs/functions/path_owner.md) - Gets the owner of a file or directory
- [`utilities_path_permission`](docs/functions/path_permission.md) - Gets the permissions of a file or directory

## Local Testing (Development Container)

When developing in the devcontainer, you can test the provider locally using the following steps:

### 1. Build the Provider

Build the provider binary:

```bash
make build
# or
go build -o terraform-provider-utilities -buildvcs=false
```

### 2. Install Provider Locally

Install the provider to Terraform's local plugin directory so Terraform can find it:

#### Using Make (Recommended)

```bash
make install-local
```

#### Manual installation

```bash
# Create the plugin directory structure
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/tfstack/utilities/0.1.0/linux_amd64

# Copy the built binary
cp terraform-provider-utilities ~/.terraform.d/plugins/registry.terraform.io/tfstack/utilities/0.1.0/linux_amd64/
```

**Note:** The version number (`0.1.0`) should match the version in your Terraform configuration's `required_providers` block.

### 3. Setup .terraformrc for Local Development

**Important:** When using `dev_overrides`, you must skip `terraform init` and use `terraform plan` or `terraform apply` directly.

Setup the `.terraformrc` file:

```bash
make setup-terraformrc
```

This will:

- Build and install the provider locally
- Configure `~/.terraformrc` with `dev_overrides` to use the local provider

### 4. Initialize Examples (Automated)

#### Skip initialization (Recommended with dev_overrides)

When using `dev_overrides`, skip `terraform init` entirely:

```bash
cd examples/data-sources/utilities_bcrypt_hash
# Skip terraform init - go directly to plan/apply
terraform plan
```

**Note:** The warning about provider development overrides is expected and can be ignored. It's just informing you that you're using a local development version of the provider.

### 5. Test with Example Configuration

After setup, navigate to any example directory and test the provider:

```bash
cd examples/data-sources/utilities_bcrypt_hash

# Skip terraform init - use plan/apply directly
terraform plan

# Apply to test the provider
terraform apply
```

**Important:** When using `dev_overrides`, **skip `terraform init`** and use `terraform plan` or `terraform apply` directly. The `terraform init` command will try to query the registry and may fail with a 429 error.

### 6. Run Unit Tests

Run the unit tests:

```bash
make test
# or
go test -v ./...
```

### 7. Run Test Coverage

Generate a test coverage report:

#### Using Make for Coverage (Recommended)

```bash
make test-coverage
```

This will:

- Run tests with coverage
- Display coverage summary in the terminal
- Generate an HTML coverage report (`coverage.html`)

#### Manual commands

```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage report in terminal
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# View coverage for specific package
go test -cover ./internal/provider/
```

**Coverage Options:**

- `-coverprofile=coverage.out` - Generate coverage profile file
- `-covermode=count` - Show how many times each statement was executed (default: `set`)
- `-covermode=atomic` - Same as count but thread-safe (useful for parallel tests)
- `-coverpkg=./...` - Include coverage for all packages, not just tested ones

**Example output:**

```text
github.com/tfstack/terraform-provider-utilities/internal/provider/data_source_bcrypt_hash.go:Metadata    100.0%
github.com/tfstack/terraform-provider-utilities/internal/provider/data_source_bcrypt_hash.go:Schema      100.0%
...
total:                                                                    (statements)    85.5%
```

### 8. Run Acceptance Tests

Acceptance tests make real operations (file system operations, network requests, etc.). Set the `TF_ACC` environment variable to enable them:

```bash
export TF_ACC=1
make testacc
# or
TF_ACC=1 go test -v ./...
```

**Warning:** Acceptance tests create and destroy real resources. Use appropriate caution when running them.

### 9. Run End-to-End Tests

Run the end-to-end validation test:

```bash
make test-e2e
```

This will test the bcrypt hash data source with a full Terraform plan/apply cycle.

### 10. Quick Setup Scripts

Helper scripts are available to automate common tasks:

**Install Provider Locally:**

```bash
make install-local
```

**Setup .terraformrc:**

```bash
make setup-terraformrc
```

**Initialize All Examples:**

```bash
make init-examples
```

**Initialize Specific Example:**

```bash
make init-example EXAMPLE=examples/data-sources/utilities_bcrypt_hash
```

### Troubleshooting

- **Provider not found:** Ensure the version in your Terraform config matches the directory version (`0.1.0`)
- **Permission denied:** Make sure the plugin directory is writable: `chmod -R 755 ~/.terraform.d/plugins/`
- **Provider version mismatch:** Update the version in your Terraform config or rename the plugin directory to match
- **Build errors:** Ensure you have Go 1.22+ installed and all dependencies are downloaded: `go mod download`
- **Terraform init fails with 429 error:** This is expected when using `dev_overrides`. **Skip `terraform init`** and use `terraform plan` or `terraform apply` directly. The warning about provider development overrides is normal and can be ignored.
- **Terraform init fails (other errors):** Make sure the provider is installed locally using `make setup-terraformrc`

## Examples

Comprehensive examples are available in the [`examples/`](examples/) directory:

- **Data Sources**: See [`examples/data-sources/`](examples/data-sources/) for examples of querying and generating data
- **Resources**: See [`examples/resources/`](examples/resources/) for examples of managing resources
- **Functions**: See [`examples/functions/`](examples/functions/) for examples of using provider functions

Each example includes a `data-source.tf`, `resource.tf`, or function usage file with working Terraform configuration.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate` or `make docs`.

In order to run the full suite of Acceptance tests, run `make testacc`.

_Note:_ Acceptance tests create real resources and perform real operations. Use appropriate caution when running them.

```shell
make testacc
```

## Documentation

Full documentation for all data sources, resources, and functions is available in the [`docs/`](docs/) directory:

- [Data Sources Documentation](docs/data-sources/)
- [Resources Documentation](docs/resources/)
- [Functions Documentation](docs/functions/)

Documentation is automatically generated using `make docs` or `go generate`.
