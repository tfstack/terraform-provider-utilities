package provider

// import (
// 	"context"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"os"
// 	"path/filepath"
// 	"regexp"

// 	"archive/tar"

// 	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
// 	"github.com/hashicorp/terraform-plugin-framework/attr"
// 	"github.com/hashicorp/terraform-plugin-framework/diag"
// 	"github.com/hashicorp/terraform-plugin-framework/path"
// 	"github.com/hashicorp/terraform-plugin-framework/resource"
// 	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
// 	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
// 	"github.com/hashicorp/terraform-plugin-framework/types"
// )

// type resourceUtilitiesExtractTar struct{}

// type ExtractTar struct {
// 	Source             types.String `tfsdk:"source"`
// 	Url                types.String `tfsdk:"url"`
// 	Destination        types.String `tfsdk:"destination"`
// 	CreatedFiles       types.List   `tfsdk:"created_files"`
// 	FileHash           types.String `tfsdk:"file_hash"`
// 	DestinationCreated types.Bool   `tfsdk:"destination_created"`
// }

// // NewResourceUtilitiesExtractTar creates a new instance of the resource.
// func NewResourceUtilitiesExtractTar() resource.Resource {
// 	return &resourceUtilitiesExtractTar{}
// }

// // Metadata sets the resource type name.
// func (r *resourceUtilitiesExtractTar) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
// 	resp.TypeName = "utilities_extract_tar"
// }

// func (r *resourceUtilitiesExtractTar) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
// 	var data ExtractTar
// 	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}

// 	var source string
// 	var url string

// 	// If source is set, get its value
// 	if !data.Source.IsNull() {
// 		source = data.Source.ValueString()
// 	}

// 	// If url is set, get its value
// 	if !data.Url.IsNull() {
// 		url = data.Url.ValueString()
// 	}

// 	destination := data.Destination.ValueString()

// 	// Calculate the hash of the source file if the source is provided
// 	var fileHash string
// 	var err error
// 	if source != "" {
// 		fileHash, err = calculateFileHash(source)
// 		if err != nil {
// 			resp.Diagnostics.AddError(
// 				"File Hash Calculation Failed",
// 				fmt.Sprintf("Error calculating hash for file '%s': %v", source, err),
// 			)
// 			return
// 		}
// 	}

// 	// Extract the Tar file based on source or URL
// 	var createdFiles []string
// 	var destinationCreated bool
// 	if source != "" {
// 		createdFiles, destinationCreated, err = r.validateAndExtractTar(ctx, source, destination, &resp.Diagnostics)
// 	} else if url != "" {
// 		createdFiles, destinationCreated, err = r.validateAndExtractTarFromURL(ctx, url, destination, &resp.Diagnostics)
// 	}

// 	if err != nil {
// 		resp.Diagnostics.AddError(
// 			"Tar Extraction Failed",
// 			fmt.Sprintf("Error extracting Tar file from '%s' to '%s': %v", source, destination, err),
// 		)
// 		return
// 	}

// 	// Convert created files to Terraform state format
// 	var createdFileList []types.String
// 	for _, file := range createdFiles {
// 		createdFileList = append(createdFileList, types.StringValue(file))
// 	}

// 	var attrList []attr.Value
// 	for _, createdFile := range createdFileList {
// 		attrList = append(attrList, createdFile)
// 	}

// 	// Update state with hash and created files
// 	stateData := ExtractTar{
// 		Source:             data.Source,
// 		Url:                data.Url,
// 		Destination:        data.Destination,
// 		CreatedFiles:       types.ListValueMust(types.StringType, attrList),
// 		FileHash:           types.StringValue(fileHash),
// 		DestinationCreated: types.BoolValue(destinationCreated),
// 	}

// 	resp.State.Set(ctx, &stateData)
// }

// func (r *resourceUtilitiesExtractTar) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
// 	var data ExtractTar

// 	// Load the current state
// 	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}

// 	// Convert CreatedFiles from basetypes.ListValue to a Go slice
// 	var createdFiles []string
// 	resp.Diagnostics.Append(data.CreatedFiles.ElementsAs(ctx, &createdFiles, false)...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}

// 	// Delete all created files in reverse order to ensure directories are deleted last
// 	for i := len(createdFiles) - 1; i >= 0; i-- {
// 		file := createdFiles[i]

// 		// Check if file exists and attempt to delete
// 		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
// 			resp.Diagnostics.AddWarning(
// 				"File Deletion Failed",
// 				fmt.Sprintf("Could not delete file '%s': %v", file, err),
// 			)
// 		}
// 	}

// 	// Clear the state to remove the resource from Terraform's state
// 	resp.State.RemoveResource(ctx)
// }

// func (r *resourceUtilitiesExtractTar) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
// 	var data ExtractTar

// 	// Load the current state
// 	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}

// 	// Skip file hash validation if URL is set
// 	hash := ""

// 	if data.Url.IsNull() {
// 		// Check for hash validation if FileHash is set and URL is not provided
// 		if !data.FileHash.IsNull() {
// 			var err error
// 			hash, err = calculateFileHash(data.Source.ValueString())
// 			if err != nil {
// 				resp.Diagnostics.AddError(
// 					"File Hash Calculation Failed",
// 					fmt.Sprintf("Unable to calculate hash for source file: %s", err),
// 				)
// 				return
// 			}

// 			// Detect drift: if hash mismatch, update the FileHash field
// 			if hash != data.FileHash.ValueString() {
// 				data.FileHash = types.StringValue(hash)
// 				resp.Diagnostics.AddWarning(
// 					"File Hash Mismatch Detected",
// 					"The hash of the Tar file has changed, marking the resource for update.",
// 				)
// 			}
// 		}
// 	}

// 	// Update the state at the end
// 	resp.State.Set(ctx, &data)
// }

// func (r *resourceUtilitiesExtractTar) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
// 	resp.Schema = schema.Schema{
// 		MarkdownDescription: "Extracts a Tar archive to a specified directory.",
// 		Attributes: map[string]schema.Attribute{
// 			"source": schema.StringAttribute{
// 				Optional:    true,
// 				Description: "The path to the source Tar file to be extracted.",
// 				Validators: []validator.String{
// 					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("url")),
// 				},
// 			},
// 			"url": schema.StringAttribute{
// 				Optional:    true,
// 				Description: "The URL to the source Tar file to be extracted. This URL must point to a valid Tar file.",
// 				Validators: []validator.String{
// 					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("source")),
// 					stringvalidator.RegexMatches(regexp.MustCompile(`^https?://.+\.tar$`), "must be a valid HTTP(S) URL ending with '.tar'"),
// 				},
// 			},
// 			"destination": schema.StringAttribute{
// 				Required:    true,
// 				Description: "The destination directory where the Tar file will be extracted.",
// 			},
// 			"created_files": schema.ListAttribute{
// 				ElementType: types.StringType,
// 				Computed:    true,
// 				Description: "A list of paths to the files created during Tar extraction.",
// 			},
// 			"file_hash": schema.StringAttribute{
// 				Computed:    true,
// 				Description: "The hash of the source Tar file, used for integrity verification.",
// 			},
// 			"destination_created": schema.BoolAttribute{
// 				Computed:    true,
// 				Description: "Indicates whether the destination directory was created by the resource.",
// 			},
// 		},
// 	}
// }

// func (r *resourceUtilitiesExtractTar) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
// 	var planData ExtractTar
// 	var stateData ExtractTar

// 	// Get plan and state
// 	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
// 	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}

// 	source := planData.Source.ValueString()
// 	destination := planData.Destination.ValueString()

// 	// If the source is a URL, skip file hash calculation
// 	newFileHash := ""
// 	if !planData.Url.IsNull() {
// 		// Calculate the new hash of the source file
// 		var err error
// 		newFileHash, err = calculateFileHash(source)
// 		if err != nil {
// 			resp.Diagnostics.AddError(
// 				"File Hash Calculation Failed",
// 				fmt.Sprintf("Error calculating hash for file '%s': %v", source, err),
// 			)
// 			return
// 		}
// 	}

// 	// Compare new hash with the existing state hash if hash is being used (not a URL)
// 	if stateData.FileHash.ValueString() == newFileHash {
// 		// No change, update state without re-extraction
// 		resp.State.Set(ctx, &stateData)
// 		return
// 	}

// 	// Extract the Tar file since the hash has changed
// 	createdFiles, destinationCreated, err := r.validateAndExtractTar(ctx, source, destination, &resp.Diagnostics)
// 	if err != nil {
// 		resp.Diagnostics.AddError(
// 			"Tar Extraction Failed",
// 			fmt.Sprintf("Error extracting Tar file from '%s' to '%s': %v", source, destination, err),
// 		)
// 		return
// 	}

// 	// Convert created files to Terraform state format
// 	var createdFileList []types.String
// 	for _, file := range createdFiles {
// 		createdFileList = append(createdFileList, types.StringValue(file))
// 	}

// 	var attrList []attr.Value
// 	for _, createdFile := range createdFileList {
// 		attrList = append(attrList, createdFile)
// 	}

// 	// Update state with new hash and created files
// 	updatedStateData := ExtractTar{
// 		Source:             planData.Source,
// 		Url:                planData.Url,
// 		Destination:        planData.Destination,
// 		CreatedFiles:       types.ListValueMust(types.StringType, attrList),
// 		FileHash:           types.StringValue(newFileHash),
// 		DestinationCreated: types.BoolValue(destinationCreated),
// 	}

// 	resp.State.Set(ctx, &updatedStateData)
// }

// func (r *resourceUtilitiesExtractTar) validateAndExtractTar(ctx context.Context, source, destination string, diagnostics *diag.Diagnostics) ([]string, bool, error) {
// 	createdFiles := []string{}
// 	destinationCreated := false

// 	// Check if the context is already canceled
// 	select {
// 	case <-ctx.Done():
// 		diagnostics.AddError("Operation Canceled", "Context was canceled before validation started.")
// 		return nil, false, ctx.Err()
// 	default:
// 		// Continue execution
// 	}

// 	// Validate the source Tar file
// 	if _, err := os.Stat(source); os.IsNotExist(err) {
// 		diagnostics.AddError(
// 			"Source File Not Found",
// 			fmt.Sprintf("The source file '%s' does not exist.", source),
// 		)
// 		return nil, false, fmt.Errorf("source file '%s' does not exist", source)
// 	}

// 	// Check if the destination directory exists
// 	if _, err := os.Stat(destination); os.IsNotExist(err) {
// 		select {
// 		case <-ctx.Done():
// 			diagnostics.AddError("Operation Canceled", "Context was canceled before creating the destination directory.")
// 			return nil, false, ctx.Err()
// 		default:
// 			// Attempt to create the directory
// 			if err := os.MkdirAll(destination, 0755); err != nil {
// 				diagnostics.AddError(
// 					"Destination Directory Creation Failed",
// 					fmt.Sprintf("Failed to create destination directory '%s': %v", destination, err),
// 				)
// 				return nil, false, fmt.Errorf("failed to create directory '%s': %w", destination, err)
// 			}
// 			destinationCreated = true // Mark directory as created
// 		}
// 	}

// 	// Extract the Tar file
// 	select {
// 	case <-ctx.Done():
// 		diagnostics.AddError("Operation Canceled", "Context was canceled before Tar extraction.")
// 		return nil, false, ctx.Err()
// 	default:
// 		if err := extractTarFile(ctx, source, destination, diagnostics, &createdFiles); err != nil {
// 			diagnostics.AddError(
// 				"Tar Extraction Failed",
// 				fmt.Sprintf("Error extracting Tar file from '%s' to '%s': %v", source, destination, err),
// 			)
// 			return nil, false, err
// 		}
// 	}

// 	return createdFiles, destinationCreated, nil
// }

// // extractTarFile extracts the contents of a Tar file to a specified destination.
// func extractTarFile(ctx context.Context, source, destination string, diagnostics *diag.Diagnostics, createdFiles *[]string) error {
// 	// Open the source Tar file
// 	file, err := os.Open(source)
// 	if err != nil {
// 		diagnostics.AddError(
// 			"Tar File Open Error",
// 			fmt.Sprintf("Unable to open the Tar file '%s': %v", source, err),
// 		)
// 		return fmt.Errorf("unable to open Tar file '%s': %w", source, err)
// 	}
// 	defer file.Close()

// 	// Create a new tar reader
// 	tarReader := tar.NewReader(file)

// 	// Process each file in the Tar archive
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			diagnostics.AddError("Operation Canceled", "Context was canceled during Tar extraction.")
// 			return ctx.Err()
// 		default:
// 			// Continue processing
// 		}

// 		header, err := tarReader.Next()
// 		if err == io.EOF {
// 			break // End of archive
// 		}
// 		if err != nil {
// 			diagnostics.AddError(
// 				"Tar Read Error",
// 				fmt.Sprintf("Error reading Tar file '%s': %v", source, err),
// 			)
// 			return fmt.Errorf("error reading Tar file '%s': %w", source, err)
// 		}

// 		targetPath := filepath.Join(destination, header.Name)

// 		switch header.Typeflag {
// 		case tar.TypeDir:
// 			// Create directory
// 			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
// 				diagnostics.AddError(
// 					"Directory Creation Error",
// 					fmt.Sprintf("Unable to create directory '%s': %v", targetPath, err),
// 				)
// 				return fmt.Errorf("unable to create directory '%s': %w", targetPath, err)
// 			}
// 		case tar.TypeReg:
// 			// Extract regular file
// 			outFile, err := os.Create(targetPath)
// 			if err != nil {
// 				diagnostics.AddError(
// 					"File Creation Error",
// 					fmt.Sprintf("Unable to create file '%s': %v", targetPath, err),
// 				)
// 				return fmt.Errorf("unable to create file '%s': %w", targetPath, err)
// 			}

// 			// Copy file contents
// 			if _, err := io.Copy(outFile, tarReader); err != nil {
// 				outFile.Close()
// 				diagnostics.AddError(
// 					"File Copy Error",
// 					fmt.Sprintf("Failed to copy data to file '%s': %v", targetPath, err),
// 				)
// 				return fmt.Errorf("failed to copy data to file '%s': %w", targetPath, err)
// 			}
// 			outFile.Close()

// 			// Update file permissions
// 			if err := os.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
// 				diagnostics.AddWarning(
// 					"File Permission Warning",
// 					fmt.Sprintf("Failed to set permissions for file '%s': %v", targetPath, err),
// 				)
// 			}

// 			// Add to the created files list
// 			*createdFiles = append(*createdFiles, targetPath)
// 		default:
// 			diagnostics.AddWarning(
// 				"Unsupported Tar Entry",
// 				fmt.Sprintf("Skipping unsupported entry '%s' in Tar file.", header.Name),
// 			)
// 		}
// 	}

// 	return nil
// }

// // extractTarEntry extracts an individual entry from a Tar file.
// func extractTarEntry(ctx context.Context, header *tar.Header, tarReader *tar.Reader, destination string, diagnostics *diag.Diagnostics, createdFiles *[]string) error {
// 	// Get the file's destination path
// 	destPath := filepath.Join(destination, header.Name)

// 	// Check for cancellation or timeout before proceeding
// 	select {
// 	case <-ctx.Done():
// 		return fmt.Errorf("operation canceled: %w", ctx.Err())
// 	default:
// 	}

// 	// Debug: print the directory and file path
// 	fmt.Printf("Debug: Destination path: %s\n", destPath)

// 	// Check if it's a directory
// 	if header.Typeflag == tar.TypeDir {
// 		// Debug: print directory path
// 		fmt.Printf("Debug: Directory path to be created: %s\n", destPath)

// 		// Only create the directory if it doesn't exist
// 		if err := os.MkdirAll(destPath, 0755); err != nil {
// 			diagnostics.AddError(
// 				"Directory Creation Failed",
// 				fmt.Sprintf("Failed to create directory '%s': %v", destPath, err),
// 			)
// 			return err
// 		}

// 		// Append to createdFiles list if it's not the root destination
// 		if destPath != destination {
// 			*createdFiles = append(*createdFiles, destPath)
// 		}
// 		return nil
// 	}

// 	// Ensure that the directory for this file exists
// 	dirPath := filepath.Dir(destPath)
// 	if err := os.MkdirAll(dirPath, 0755); err != nil {
// 		diagnostics.AddError(
// 			"Directory Creation Failed",
// 			fmt.Sprintf("Failed to create directory '%s': %v", dirPath, err),
// 		)
// 		return err
// 	}

// 	// Create the destination file
// 	destFile, err := os.Create(destPath)
// 	if err != nil {
// 		diagnostics.AddError(
// 			"File Create Failed",
// 			fmt.Sprintf("Failed to create file '%s': %v", destPath, err),
// 		)
// 		return err
// 	}
// 	defer destFile.Close()

// 	// Copy the contents from the tarReader to the destination file
// 	if _, err := io.Copy(destFile, tarReader); err != nil {
// 		diagnostics.AddError(
// 			"File Copy Failed",
// 			fmt.Sprintf("Failed to copy contents to '%s': %v", destPath, err),
// 		)
// 		return err
// 	}

// 	// Debug: print the created file
// 	fmt.Printf("Debug: Created file: %s\n", destPath)

// 	// Append to the createdFiles list
// 	*createdFiles = append(*createdFiles, destPath)

// 	return nil
// }

// // func calculateFileHash(filepath string) (string, error) {
// // 	file, err := os.Open(filepath)
// // 	if err != nil {
// // 		return "", fmt.Errorf("unable to open file: %w", err)
// // 	}
// // 	defer file.Close()

// // 	hash := sha256.New()
// // 	if _, err := io.Copy(hash, file); err != nil {
// // 		return "", fmt.Errorf("unable to calculate hash: %w", err)
// // 	}

// // 	return fmt.Sprintf("%x", hash.Sum(nil)), nil
// // }

// func (r *resourceUtilitiesExtractTar) validateAndExtractTarFromURL(ctx context.Context, url, destination string, diagnostics *diag.Diagnostics) ([]string, bool, error) {
// 	createdFiles := []string{}
// 	destinationCreated := false

// 	// Check if the context is already canceled
// 	select {
// 	case <-ctx.Done():
// 		diagnostics.AddError("Operation Canceled", "Context was canceled before validation started.")
// 		return nil, false, ctx.Err()
// 	default:
// 		// Continue execution
// 	}

// 	// Download the Tar file from the URL
// 	tmpFile, err := downloadTarFile(ctx, url, diagnostics)
// 	if err != nil {
// 		diagnostics.AddError(
// 			"Tar Download Failed",
// 			fmt.Sprintf("Error downloading the Tar file from '%s': %v", url, err),
// 		)
// 		return nil, false, err
// 	}
// 	defer os.Remove(tmpFile) // Ensure the temporary file is removed after extraction

// 	// Check if the destination directory exists
// 	if _, err := os.Stat(destination); os.IsNotExist(err) {
// 		select {
// 		case <-ctx.Done():
// 			diagnostics.AddError("Operation Canceled", "Context was canceled before creating the destination directory.")
// 			return nil, false, ctx.Err()
// 		default:
// 			// Attempt to create the directory
// 			if err := os.MkdirAll(destination, 0755); err != nil {
// 				diagnostics.AddError(
// 					"Destination Directory Creation Failed",
// 					fmt.Sprintf("Failed to create destination directory '%s': %v", destination, err),
// 				)
// 				return nil, false, fmt.Errorf("failed to create directory '%s': %w", destination, err)
// 			}
// 			destinationCreated = true // Mark directory as created
// 		}
// 	}

// 	// Extract the Tar file
// 	select {
// 	case <-ctx.Done():
// 		diagnostics.AddError("Operation Canceled", "Context was canceled before Tar extraction.")
// 		return nil, false, ctx.Err()
// 	default:
// 		if err := extractTarFile(ctx, tmpFile, destination, diagnostics, &createdFiles); err != nil {
// 			diagnostics.AddError(
// 				"Tar Extraction Failed",
// 				fmt.Sprintf("Error extracting Tar file from '%s' to '%s': %v", tmpFile, destination, err),
// 			)
// 			return nil, false, err
// 		}
// 	}

// 	return createdFiles, destinationCreated, nil
// }

// // Helper function to download the Tar file.
// func downloadTarFile(ctx context.Context, url string, diagnostics *diag.Diagnostics) (string, error) {
// 	// Create a temporary file to store the Tar content
// 	tmpFile, err := os.CreateTemp("", "downloaded-tar-*.tar")
// 	if err != nil {
// 		diagnostics.AddError("Temp File Creation Failed", fmt.Sprintf("Error creating temporary file: %v", err))
// 		return "", err
// 	}

// 	// Create an HTTP client with timeout from the context
// 	client := &http.Client{}

// 	// Create a request with context
// 	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
// 	if err != nil {
// 		diagnostics.AddError("Request Creation Failed", fmt.Sprintf("Error creating request for URL '%s': %v", url, err))
// 		return "", err
// 	}

// 	// Download the Tar file
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		diagnostics.AddError("HTTP Request Failed", fmt.Sprintf("Error making HTTP request to '%s': %v", url, err))
// 		return "", err
// 	}
// 	defer resp.Body.Close()

// 	// Check if the response status is OK
// 	if resp.StatusCode != http.StatusOK {
// 		diagnostics.AddError("HTTP Request Failed", fmt.Sprintf("Failed to download file, HTTP status: %s", resp.Status))
// 		return "", fmt.Errorf("failed to download file, HTTP status: %s", resp.Status)
// 	}

// 	// Copy the content to the temporary file
// 	_, err = io.Copy(tmpFile, resp.Body)
// 	if err != nil {
// 		diagnostics.AddError("File Write Failed", fmt.Sprintf("Error writing the downloaded file to temporary file: %v", err))
// 		return "", err
// 	}

// 	// Return the path to the temporary file
// 	return tmpFile.Name(), nil
// }
