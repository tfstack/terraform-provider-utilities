package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ provider.ProviderWithFunctions = (*utilitiesProvider)(nil)
)

type utilitiesProvider struct {
	version string
}

// New is a helper function to initialize the provider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &utilitiesProvider{
			version: version,
		}
	}
}

// Metadata defines provider name and version.
func (p *utilitiesProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "utilities"
	resp.Version = p.version
}

// Schema returns the provider schema, which is empty since no configuration is required.
func (p *utilitiesProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
The Utilities provider offers various utility functions and tools for use in Terraform configurations. This provider does not require configuration.
`,
	}
}

// Configure initializes the provider with no special configuration.
func (p *utilitiesProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// No configuration needed
}

// Resources returns an empty list since no resources are implemented.
func (p *utilitiesProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewResourceUtilitiesLocalDirectory,
	}
}

// DataSources lists the available data sources.
func (p *utilitiesProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDataSourceLocalDirectory,
	}
}

// Functions lists the available functions.
func (p *utilitiesProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewFunctionHttpRequest,
		NewFunctionPathExists,
		NewFunctionPathOwner,
		NewFunctionPathPermission,
	}
}
