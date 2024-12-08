## Download Utilities

curl -L -o /tmp/repo.zip https://github.com/tfstack/terraform-provider-scaffolding-framework/archive/refs/heads/main.zip

## Setup Dev Environment

go mod edit -module terraform-provider-utilities
go mod tidy

## Prepare Terraform for Local Provider Install

cat << EOF > ~/.terraformrc
provider_installation {
dev_overrides {
"hashicorp.com/tfstack/utilities" = "/go/bin"
}
direct {}
}
EOF

## Verify the Initial Provider

go run main.go

## Locally Install Provider and Verify with Terraform

go install .

## Test

cd examples/provider-install-verification/
terraform plan
terraform -chdir=./examples/resources/utilities_extract_zip plan
go test -v ./internal/provider/...
go test -v ./internal/provider/provider_test.go

### Run the Tests for the Package:

go test -v ./internal/provider

### Run a Specific Test File:

go test -v ./internal/provider/provider_test.go

### Run with Coverage:

go test -v -cover ./internal/provider

### Running Tests with Specific Test Names:

go test -v -run TestProviderConfigure ./internal/provider

### Running Acceptance Tests:

go test -v ./internal/provider

### Running with Debug Logging:

TF_LOG=DEBUG go test -v ./internal/provider

### Generate documentation

tfplugindocs validate

known error, you can manually remediate but comes back when `tfplugindocs generate` gets executed

```bash
Error executing command: validation errors found: 
docs/index.md: error checking file frontmatter: YAML frontmatter should not contain subcategory
```

tfplugindocs generate
