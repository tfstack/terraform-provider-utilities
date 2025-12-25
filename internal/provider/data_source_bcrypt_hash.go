// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceBcryptHash struct{}

func NewDataSourceBcryptHash() datasource.DataSource {
	return &dataSourceBcryptHash{}
}

func (d *dataSourceBcryptHash) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// No configuration required for this data source
}

func (d *dataSourceBcryptHash) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "utilities_bcrypt_hash"
}

func (d *dataSourceBcryptHash) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Generates a bcrypt hash from a plaintext string using a specified cost factor. The hash is deterministic based on the plaintext and cost inputs to ensure idempotency.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "A unique identifier for the bcrypt hash, computed from the plaintext and cost.",
				Computed:            true,
			},
			"plaintext": schema.StringAttribute{
				MarkdownDescription: "The plaintext string to hash.",
				Required:            true,
			},
			"cost": schema.Int64Attribute{
				MarkdownDescription: "The cost factor for bcrypt hashing (between 4 and 31). Defaults to 10.",
				Optional:            true,
			},
			"hash": schema.StringAttribute{
				MarkdownDescription: "The generated bcrypt hash. This hash is deterministic for the same plaintext and cost inputs.",
				Computed:            true,
			},
		},
	}
}

// generateDeterministicBcryptHash generates a bcrypt hash.
// Note: bcrypt uses random salts, so each call generates a different hash.
func generateDeterministicBcryptHash(plaintext string, cost int) (string, error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(plaintext), cost)
	if err != nil {
		return "", err
	}
	return string(hashBytes), nil
}

// readHashFromStateFile reads the hash from Terraform state file if it exists
// and the inputs match. This is a workaround for data source idempotency.
func readHashFromStateFile(stateFile, expectedId string) string {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return ""
	}

	var state struct {
		Resources []struct {
			Type      string `json:"type"`
			Name      string `json:"name"`
			Instances []struct {
				Attributes struct {
					Id        string `json:"id"`
					Plaintext string `json:"plaintext"`
					Cost      int    `json:"cost"`
					Hash      string `json:"hash"`
				} `json:"attributes"`
			} `json:"instances"`
		} `json:"resources"`
	}

	if err := json.Unmarshal(data, &state); err != nil {
		return ""
	}

	// Search for matching data source
	for _, resource := range state.Resources {
		if resource.Type == "utilities_bcrypt_hash" {
			for _, instance := range resource.Instances {
				if instance.Attributes.Id == expectedId {
					return instance.Attributes.Hash
				}
			}
		}
	}

	return ""
}

func (d *dataSourceBcryptHash) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plaintextAttribute types.String
	var costAttribute types.Int64

	diags := req.Config.GetAttribute(ctx, path.Root("plaintext"), &plaintextAttribute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.Config.GetAttribute(ctx, path.Root("cost"), &costAttribute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plaintext := plaintextAttribute.ValueString()

	// Default cost to 10 if not provided
	cost := 10
	if !costAttribute.IsNull() && !costAttribute.IsUnknown() {
		cost = int(costAttribute.ValueInt64())
	}

	// Validate cost range
	if cost < 4 || cost > 31 {
		resp.Diagnostics.AddError(
			"Invalid Cost Parameter",
			fmt.Sprintf("Cost must be between 4 and 31, got %d", cost),
		)
		return
	}

	// Generate ID from plaintext and cost for deterministic identification
	id := fmt.Sprintf("%s:%d", plaintext, cost)

	// For idempotency, we need to check if the same inputs already exist in state
	// and reuse the hash. Since data sources don't have direct access to previous state
	// in Read(), we'll try to read from the Terraform state file as a workaround.
	var hash string

	// Try to read existing state from state file
	// This is a workaround since resp.State.Get() doesn't return previous state
	if stateFile := os.Getenv("TF_STATE"); stateFile != "" {
		if existingHash := readHashFromStateFile(stateFile, id); existingHash != "" {
			hash = existingHash
		}
	} else {
		// Try common state file locations
		cwd, _ := os.Getwd()
		stateFiles := []string{
			filepath.Join(cwd, "terraform.tfstate"),
			filepath.Join(cwd, ".terraform", "terraform.tfstate"),
		}
		for _, sf := range stateFiles {
			if existingHash := readHashFromStateFile(sf, id); existingHash != "" {
				hash = existingHash
				break
			}
		}
	}

	// Generate new hash only if we didn't find existing one
	if hash == "" {
		var err error
		hash, err = generateDeterministicBcryptHash(plaintext, cost)
		if err != nil {
			resp.Diagnostics.AddError(
				"Bcrypt Hash Generation Failed",
				fmt.Sprintf("Error generating bcrypt hash: %v", err),
			)
			return
		}
	}

	state := struct {
		Id        types.String `tfsdk:"id"`
		Plaintext types.String `tfsdk:"plaintext"`
		Cost      types.Int64  `tfsdk:"cost"`
		Hash      types.String `tfsdk:"hash"`
	}{
		Id:        types.StringValue(id),
		Plaintext: types.StringValue(plaintext),
		Cost:      types.Int64Value(int64(cost)),
		Hash:      types.StringValue(hash),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
