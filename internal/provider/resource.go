package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type externalResource struct{}
type externalResourceModelV0 struct {
	Program    types.List   `tfsdk:"program"`
	Create     types.Bool   `tfsdk:"create"`
	Read       types.Bool   `tfsdk:"read"`
	Update     types.Bool   `tfsdk:"update"`
	Destroy    types.Bool   `tfsdk:"destroy"`
	WorkingDir types.String `tfsdk:"working_dir"`
	Query      types.Map    `tfsdk:"query"`
	Result     types.Map    `tfsdk:"result"`
	ID         types.String `tfsdk:"id"`
}

var _ resource.Resource = (*externalResource)(nil)

func NewExternalResource() resource.Resource {
	return &externalResource{}
}

func (e *externalResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_external"
}

func (e *externalResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `external` resource allows an external program implementing a specific protocol " +
			"(defined below) to act as a resource, exposing arbitrary data for use elsewhere in the Terraform " +
			"configuration.\n" +
			"As of now, this resource will be re-created if any value is changed." +
			"Similar to null_resource combined with trigger but the output can be saved into our state.\n" +
			"\n" +
			"**Warning** This mechanism is provided as an \"escape hatch\" for exceptional situations where a " +
			"first-class Terraform provider is not more appropriate. Its capabilities are limited in comparison " +
			"to a true resource, and implementing a resource via an external program is likely to hurt the " +
			"portability of your Terraform configuration by creating dependencies on external programs and " +
			"libraries that may not be available (or may need to be used differently) on different operating " +
			"systems.\n" +
			"\n" +
			"**Warning** Terraform Enterprise does not guarantee availability of any particular language runtimes " +
			"or external programs beyond standard shell utilities, so it is not recommended to use this resource " +
			"within configurations that are applied within Terraform Enterprise.",

		Attributes: map[string]schema.Attribute{
			"program": schema.ListAttribute{
				Description: "A list of strings, whose first element is the program to run and whose " +
					"subsequent elements are optional command line arguments to the program. Terraform does " +
					"not execute the program through a shell, so it is not necessary to escape shell " +
					"metacharacters nor add quotes around arguments containing spaces.",
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplaceIfConfigured(),
				},
			},

			"create": schema.BoolAttribute{
				Description: "Run on create: enabled by default",
				Optional:    true,
				Default:     booldefault.StaticBool(true),
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
			},

			"read": schema.BoolAttribute{
				Description: "Run on read: disabled by default",
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
			},

			"update": schema.BoolAttribute{
				Description: "Run on update: disabled by default",
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
			},

			"destroy": schema.BoolAttribute{
				Description: "Run on destroy: disabled by default",
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
			},

			"working_dir": schema.StringAttribute{
				Description: "Working directory of the program. If not supplied, the program will run " +
					"in the current directory.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},

			"query": schema.MapAttribute{
				Description: "A map of string values to pass to the external program as the query " +
					"arguments. If not supplied, the program will receive an empty object as its input.",
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplaceIfConfigured(),
				},
			},

			"result": schema.MapAttribute{
				Description: "A map of string values returned from the external program.",
				ElementType: types.StringType,
				Computed:    true,
			},

			"id": schema.StringAttribute{
				Description: "The id of the resource. This will always be set to `-`",
				Computed:    true,
			},
		},
	}
}

func (e *externalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "Creating resource")
	var config externalResourceModelV0

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &config)...)
	config.ID = types.StringValue("-")

	result, errors := run_external(ctx, config, config.Create.ValueBool(), "create")

	if errors != nil {
		resp.Diagnostics.Append(errors...)
		return
	}

	config.Result = result

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

func (e *externalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Trace(ctx, "Updating resource")
	var config externalResourceModelV0
	
	// Read Terraform state
	resp.Diagnostics.Append(req.Plan.Get(ctx, &config)...)
	
	// Set Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (e *externalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "Reading resource")
	var config externalResourceModelV0

	// Read Terraform state
	resp.Diagnostics.Append(req.State.Get(ctx, &config)...)

	// Set Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (e *externalResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "Deleting resource")
	var config externalResourceModelV0
	// Read Terraform plan
	resp.Diagnostics.Append(req.State.Get(ctx, &config)...)
	config.ID = types.StringValue("-")

	diags := req.State.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, errors := run_external(ctx, config, config.Destroy.ValueBool(), "destroy")

	if errors != nil {
		resp.Diagnostics.Append(errors...)
		return
	}
	config.Result = result
	resp.Diagnostics.Append(diags...)
}

func run_external(ctx context.Context, config externalResourceModelV0, execute bool, stage string) (types.Map, diag.Diagnostics) {

	var diag diag.Diagnostics

	//initMap := make(map[string]string)
	initMap := map[string]string{}
	emptyMap, _ := types.MapValueFrom(ctx, types.StringType, initMap)

	// Setup program variable
	// Grab program list and filter out empty/null values
	var program []types.String
	diag = config.Program.ElementsAs(ctx, &program, false)
	if diag.HasError() {
		return emptyMap, diag
	}
	filteredProgram := make([]string, 0, len(program))
	for _, programArgRaw := range program {
		if programArgRaw.IsNull() || programArgRaw.ValueString() == "" {
			continue
		}

		filteredProgram = append(filteredProgram, programArgRaw.ValueString())
	}
	if len(filteredProgram) == 0 {
		diag.AddAttributeError(path.Root("program"),
			"External Program Missing",
			"The resource was configured without a program to execute. Verify the configuration contains at least one non-empty value.",
		)
		return emptyMap, diag
	}
	// first element is assumed to be an executable command, possibly found
	// using the PATH environment variable.
	_, err := exec.LookPath(filteredProgram[0])

	if err != nil {
		diag.AddAttributeError(
			path.Root("program"),
			"External Program Lookup Failed",
			"The resource received an unexpected error while attempting to parse the query. "+
				`The resource received an unexpected error while attempting to find the program.

The program must be accessible according to the platform where Terraform is running.

If the expected program should be automatically found on the platform where Terraform is running, ensure that the program is in an expected directory. On Unix-based platforms, these directories are typically searched based on the '$PATH' environment variable. On Windows-based platforms, these directories are typically searched based on the '%PATH%' environment variable.

If the expected program is relative to the Terraform configuration, it is recommended that the program name includes the interpolated value of 'path.module' before the program name to ensure that it is compatible with varying module usage. For example: "${path.module}/my-program"

The program must also be executable according to the platform where Terraform is running. On Unix-based platforms, the file on the filesystem must have the executable bit set. On Windows-based platforms, no action is typically necessary.
`+
				fmt.Sprintf("\nPlatform: %s", runtime.GOOS)+
				fmt.Sprintf("\nProgram: %s", program[0])+
				fmt.Sprintf("\nError: %s", err),
		)
		return emptyMap, diag
	}

	// Setup query variable
	// Grab query mapping and setup "stage" key
	var query map[string]types.String
	diag = config.Query.ElementsAs(ctx, &query, false)
	if diag.HasError() {
		return emptyMap, diag
	}
	if query == nil {
		query = make(map[string]types.String)
	}
	query["stage"] = types.StringValue(stage)
	filteredQuery := make(map[string]string)
	for key, value := range query {
		filteredQuery[key] = value.ValueString()
	}

	queryJson, err := json.Marshal(filteredQuery)
	if err != nil {
		diag.AddAttributeError(
			path.Root("query"),
			"Query Handling Failed",
			"The resource received an unexpected error while attempting to parse the query. "+
				"This is always a bug in the external provider code and should be reported to the provider developers."+
				fmt.Sprintf("\n\nError: %s", err),
		)
		return emptyMap, diag
	}

	// Setup working directory
	workingDir := config.WorkingDir.ValueString()

	// Check whether to execute the program
	// Depending on stage, we need to return the previous result
	if !execute {
		return emptyMap, nil
	}

	cmd := exec.CommandContext(ctx, filteredProgram[0], filteredProgram[1:]...)
	cmd.Dir = workingDir
	cmd.Stdin = bytes.NewReader(queryJson)

	tflog.Trace(ctx, "Executing external program", map[string]interface{}{"program": cmd.String()})

	resultJson, err := cmd.Output()

	tflog.Trace(ctx, "Executed external program", map[string]interface{}{"program": cmd.String(), "output": string(resultJson)})

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.Stderr != nil && len(exitErr.Stderr) > 0 {
				diag.AddAttributeError(
					path.Root("program"),
					"External Program Execution Failed",
					"The resource received an unexpected error while attempting to execute the program."+
						fmt.Sprintf("\n\nProgram: %s", cmd.Path)+
						fmt.Sprintf("\nError Message: %s", string(exitErr.Stderr))+
						fmt.Sprintf("\nState: %s", err),
				)
				return emptyMap, diag
			}

			diag.AddAttributeError(
				path.Root("program"),
				"External Program Execution Failed",
				"The resource received an unexpected error while attempting to execute the program.\n\n"+
					"The program was executed, however it returned no additional error messaging."+
					fmt.Sprintf("\n\nProgram: %s", cmd.Path)+
					fmt.Sprintf("\nState: %s", err),
			)
			return emptyMap, diag
		}

		diag.AddAttributeError(
			path.Root("program"),
			"External Program Execution Failed",
			"The resource received an unexpected error while attempting to execute the program."+
				fmt.Sprintf("\n\nProgram: %s", cmd.Path)+
				fmt.Sprintf("\nError: %s", err),
		)
		return emptyMap, diag
	}

	result := map[string]string{}
	err = json.Unmarshal(resultJson, &result)
	if err != nil {
		diag.AddAttributeError(
			path.Root("program"),
			"Unexpected External Program Results",
			`The resource received unexpected results after executing the program.

Program output must be a JSON encoded map of string keys and string values.

If the error is unclear, the output can be viewed by enabling Terraform's logging at TRACE level. Terraform documentation on logging: https://www.terraform.io/internals/debugging
`+
				fmt.Sprintf("\nProgram: %s", cmd.Path)+
				fmt.Sprintf("\nResult Error: %s", err),
		)
		return emptyMap, diag
	}

	from_result, from_diag := types.MapValueFrom(ctx, types.StringType, result)
	if err != nil {
		diag.Append(from_diag...)
		return emptyMap, diag
	}

	return from_result, nil
}
