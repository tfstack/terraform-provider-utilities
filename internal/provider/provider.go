package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type utilitiesProvider struct {
	version string
}

type utilitiesProviderModel struct {
	ApiToken types.String `tfsdk:"api_token"`
}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &utilitiesProvider{
			version: version,
		}
	}
}

func (p *utilitiesProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "utilities"
	resp.Version = p.version
}

func (p *utilitiesProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
The Utilities provider offers various utility functions and tools for use in Terraform configurations. This provider does not require authentication or credentials.
`,
		Attributes: map[string]schema.Attribute{}, // No attributes required
	}
}

func (p *utilitiesProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config utilitiesProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (p *utilitiesProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *utilitiesProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *utilitiesProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewFunctionHttpRequest,
		NewFunctionPathExists,
	}
}
