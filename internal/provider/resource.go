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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type externalResource struct{}

var _ resource.Resource = (*externalResource)(nil)

func NewExternalResource() resource.Resource {
	return &externalResource{}
}

func (e *externalResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_external"
}

func (e *externalResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `external` data source allows an external program implementing a specific protocol " +
			"(defined below) to act as a data source, exposing arbitrary data for use elsewhere in the Terraform " +
			"configuration.\n" +
			"\n" +
			"**Warning** This mechanism is provided as an \"escape hatch\" for exceptional situations where a " +
			"first-class Terraform provider is not more appropriate. Its capabilities are limited in comparison " +
			"to a true data source, and implementing a data source via an external program is likely to hurt the " +
			"portability of your Terraform configuration by creating dependencies on external programs and " +
			"libraries that may not be available (or may need to be used differently) on different operating " +
			"systems.\n" +
			"\n" +
			"**Warning** Terraform Enterprise does not guarantee availability of any particular language runtimes " +
			"or external programs beyond standard shell utilities, so it is not recommended to use this data source " +
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
			},

			"working_dir": schema.StringAttribute{
				Description: "Working directory of the program. If not supplied, the program will run " +
					"in the current directory.",
				Optional: true,
			},

			"query": schema.MapAttribute{
				Description: "A map of string values to pass to the external program as the query " +
					"arguments. If not supplied, the program will receive an empty object as its input.",
				ElementType: types.StringType,
				Optional:    true,
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

			"enable": schema.BoolAttribute{
				Description: "Allow user to skip running of external programs without needing to remove it from state",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					EnableDefaultValue(),
				},
			},
		},
	}
}

type externalResourceModelV0 struct {
	Program    types.List   `tfsdk:"program"`
	WorkingDir types.String `tfsdk:"working_dir"`
	Query      types.Map    `tfsdk:"query"`
	Result     types.Map    `tfsdk:"result"`
	ID         types.String `tfsdk:"id"`
	Enable     types.Bool   `tfsdk:"enable"`
}

func (e *externalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "Creating resource")
	var config externalResourceModelV0

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.ID = types.StringValue("-")

	var program []types.String

	diags = config.Program.ElementsAs(ctx, &program, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var query map[string]types.String

	diags = config.Query.ElementsAs(ctx, &query, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	workingDir := config.WorkingDir.ValueString()

	result, errors := run_external(ctx, program, query, workingDir)

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

	config.ID = types.StringValue("-")

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var program []types.String

	diags = config.Program.ElementsAs(ctx, &program, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var query map[string]types.String

	diags = config.Query.ElementsAs(ctx, &query, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	workingDir := config.WorkingDir.ValueString()

	result, errors := run_external(ctx, program, query, workingDir)

	if errors != nil {
		resp.Diagnostics.Append(errors...)
		return
	}
	config.Result = result

	diags = resp.State.Set(ctx, config)
	resp.Diagnostics.Append(diags...)
}

func (e *externalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}
func (e *externalResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

type enableDefaultValuePlanModifier struct {
	Default types.Bool
}

func EnableDefaultValue() planmodifier.Bool {
	return &enableDefaultValuePlanModifier{}
}

var _ planmodifier.Bool = (*enableDefaultValuePlanModifier)(nil)

func (r *enableDefaultValuePlanModifier) Description(ctx context.Context) string {
	return fmt.Sprintf("If value is not set, defaults to %s", r.Default)
}

func (r *enableDefaultValuePlanModifier) MarkdownDescription(ctx context.Context) string {
	return r.Description(ctx)
}

func (r *enableDefaultValuePlanModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {

	if req.Plan.Raw.IsNull() {
		resp.Diagnostics.AddWarning(
			"Removes resource from state",
			"Call to program skipped during this stage",
		)
		return
	}

	var enable types.Bool
	diags := req.Plan.GetAttribute(ctx, path.Root("enable"), &enable)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if enable.IsNull() {
		resp.PlanValue = types.BoolValue(true)
		return
	}

	resp.PlanValue = enable
}

func run_external(ctx context.Context, program []types.String, query map[string]types.String, workingDir string) (types.Map, diag.Diagnostics) {

	var diag diag.Diagnostics
	//initMap := make(map[string]string)
	initMap := map[string]string{}
	emptyMap, _ := types.MapValueFrom(ctx, types.StringType, initMap)
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
			"The data source was configured without a program to execute. Verify the configuration contains at least one non-empty value.",
		)
		return emptyMap, diag
	}

	filteredQuery := make(map[string]string)
	for key, value := range query {
		if value.IsNull() || value.ValueString() == "" {
			continue
		}

		filteredQuery[key] = value.ValueString()
	}

	queryJson, err := json.Marshal(filteredQuery)
	if err != nil {
		diag.AddAttributeError(
			path.Root("query"),
			"Query Handling Failed",
			"The data source received an unexpected error while attempting to parse the query. "+
				"This is always a bug in the external provider code and should be reported to the provider developers."+
				fmt.Sprintf("\n\nError: %s", err),
		)
		return emptyMap, diag
	}

	// first element is assumed to be an executable command, possibly found
	// using the PATH environment variable.
	_, err = exec.LookPath(filteredProgram[0])

	if err != nil {
		diag.AddAttributeError(
			path.Root("program"),
			"External Program Lookup Failed",
			"The data source received an unexpected error while attempting to parse the query. "+
				`The data source received an unexpected error while attempting to find the program.

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
					"The data source received an unexpected error while attempting to execute the program."+
						fmt.Sprintf("\n\nProgram: %s", cmd.Path)+
						fmt.Sprintf("\nError Message: %s", string(exitErr.Stderr))+
						fmt.Sprintf("\nState: %s", err),
				)
				return emptyMap, diag
			}

			diag.AddAttributeError(
				path.Root("program"),
				"External Program Execution Failed",
				"The data source received an unexpected error while attempting to execute the program.\n\n"+
					"The program was executed, however it returned no additional error messaging."+
					fmt.Sprintf("\n\nProgram: %s", cmd.Path)+
					fmt.Sprintf("\nState: %s", err),
			)
			return emptyMap, diag
		}

		diag.AddAttributeError(
			path.Root("program"),
			"External Program Execution Failed",
			"The data source received an unexpected error while attempting to execute the program."+
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
			`The data source received unexpected results after executing the program.

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
