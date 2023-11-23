package provider

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	// EnvTfAccExternalTimeoutTest is the name of the environment variable used
	// to enable the 20 minute timeout test. The environment variable can be
	// set to any value to enable the test.
	EnvTfAccExternalTimeoutTest = "TF_ACC_EXTERNAL_TIMEOUT_TEST"
)

const testResourceConfig_basic = `
resource "toolbox_external" "test" {
  program = ["%s", "cheese"]

  query = {
    value = "pizza"
  }
}

output "query_value" {
  value = "${toolbox_external.test.result["query_value"]}"
}

output "argument" {
  value = "${toolbox_external.test.result["argument"]}"
}
`

func TestResource_basic(t *testing.T) {
	programPath, err := buildGoTestProgram()
	if err != nil {
		t.Fatal(err)
		return
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testResourceConfig_basic, programPath),
				Check: func(s *terraform.State) error {
					_, ok := s.RootModule().Resources["toolbox_external.test"]
					if !ok {
						return fmt.Errorf("missing data resource")
					}

					outputs := s.RootModule().Outputs

					if outputs["argument"] == nil {
						return fmt.Errorf("missing 'argument' output")
					}
					if outputs["query_value"] == nil {
						return fmt.Errorf("missing 'query_value' output")
					}

					if outputs["argument"].Value != "cheese" {
						return fmt.Errorf(
							"'argument' output is %q; want 'cheese'",
							outputs["argument"].Value,
						)
					}
					if outputs["query_value"].Value != "pizza" {
						return fmt.Errorf(
							"'query_value' output is %q; want 'pizza'",
							outputs["query_value"].Value,
						)
					}

					return nil
				},
			},
		},
	})
}

const testResourceConfig_error = `
resource "toolbox_external" "test" {
  program = ["%s"]

  query = {
    fail = "true"
  }
}
`

func TestResource_error(t *testing.T) {
	programPath, err := buildGoTestProgram()
	if err != nil {
		t.Fatal(err)
		return
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      fmt.Sprintf(testResourceConfig_error, programPath),
				ExpectError: regexp.MustCompile("I was asked to fail"),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-external/issues/110
func TestResource_Program_OnlyEmptyString(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "toolbox_external" "test" {
						program = [
							"", # e.g. a variable that became empty
						]
				
						query = {
							value = "valuetest"
						}
					}
				`,
				ExpectError: regexp.MustCompile(`External Program Missing`),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-external/issues/110
func TestResource_Program_PathAndEmptyString(t *testing.T) {
	programPath, err := buildGoTestProgram()
	if err != nil {
		t.Fatal(err)
		return
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "toolbox_external" "test" {
						program = [
							%[1]q,
							"", # e.g. a variable that became empty
						]
				
						query = {
							value = "valuetest"
						}
					}
				`, programPath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("toolbox_external.test", "result.query_value", "valuetest"),
				),
			},
		},
	})
}

func TestResource_Program_EmptyStringAndNullValues(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "toolbox_external" "test" {
						program = [
							null, "", # e.g. a variable that became empty
						]
				
						query = {
							value = "valuetest"
						}
					}
				`,
				ExpectError: regexp.MustCompile(`External Program Missing`),
			},
		},
	})
}

func TestResource_Query_NullAndEmptyValue(t *testing.T) {
	programPath, err := buildGoTestProgram()
	if err != nil {
		t.Fatal(err)
		return
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "toolbox_external" "test" {
						program = [%[1]q]
				
						query = {
							value = null,
							value2 = ""
						}
					}
				`, programPath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("toolbox_external.test", "result.value", ""),
					resource.TestCheckResourceAttr("toolbox_external.test", "result.value2", ""),
				),
			},
		},
	})
}

/*
Each resources "query" will have a "stage" attribute that will be set to one of the following values:
- create
- read
- update
- destroy
*/
func TestResource_Example(t *testing.T) {
	programPath, err := buildGoTestProgram()
	if err != nil {
		t.Fatal(err)
		return
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
						resource "toolbox_external" "example" {
							create = true
							delete = true
							update = true
							read = true
							program = [%[1]q]

							recreate = {}
							query = {
								value = null,
								value2 = ""
							}
						}
					`, programPath),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("toolbox_external.example", plancheck.ResourceActionCreate),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("toolbox_external.example", plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("toolbox_external.example", plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("toolbox_external.example", "stage"),
					resource.TestCheckResourceAttr("toolbox_external.example", "stage", "create"),
				),
			},
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
				Config: fmt.Sprintf(`
						resource "toolbox_external" "example" {
							create = true
							delete = true
							update = true
							read = true
							program = [%[1]q]

							recreate = {}
							query = {
								value = null,
								value2 = ""
							}
						}
					`, programPath),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("toolbox_external.example", plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("toolbox_external.example", plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("toolbox_external.example", "stage"),
					resource.TestCheckResourceAttr("toolbox_external.example", "stage", "read"),
				),
			},
			{
				Config: fmt.Sprintf(`
						resource "toolbox_external" "example" {
							create = true
							delete = true
							update = true
							read = true
							program = [%[1]q]

							recreate = {}
							query = {
								value = null,
								value2 = ""
							}
						}
					`, programPath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("toolbox_external.example", "stage"),
					resource.TestCheckResourceAttr("toolbox_external.example", "stage", "read"),
					resource.TestCheckResourceAttr("toolbox_external.example", "result.value2", ""),
				),
			},
			{
				Config: fmt.Sprintf(`
						resource "toolbox_external" "example" {
							create = true
							delete = true
							update = true
							read = true
							program = [%[1]q]
							recreate = {foo: "bar"}
							query = {
								value = null,
								value2 = ""
							}
						}
					`, programPath),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("toolbox_external.example", plancheck.ResourceActionDestroyBeforeCreate),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("toolbox_external.example", plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("toolbox_external.example", plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("toolbox_external.example", "stage"),
					resource.TestCheckResourceAttr("toolbox_external.example", "stage", "create"),
					resource.TestCheckResourceAttr("toolbox_external.example", "result.value", ""),
				),
			},
			{
				Config: fmt.Sprintf(`
						resource "toolbox_external" "example" {
							create = true
							delete = true
							update = true
							read = true
							program = [%[1]q]
							recreate = {foo: "bar"}
							query = {
								value = null,
								value2 = "test"
							}
						}
					`, programPath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("toolbox_external.example", "stage"),
					resource.TestCheckResourceAttr("toolbox_external.example", "stage", "update"),
					resource.TestCheckResourceAttr("toolbox_external.example", "result.value2", "test"),
				),
			},
			{
				Destroy: true,
				Config: fmt.Sprintf(`
						resource "toolbox_external" "example" {
							create = true
							delete = true
							update = true
							read = true
							program = [%[1]q]
							recreate = {foo: "bar"}
							query = {
								value = null,
								value2 = "test"
							}
						}
					`, programPath),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("toolbox_external.example", plancheck.ResourceActionDestroy),
					},
					PostApplyPreRefresh:  []plancheck.PlanCheck{},
					PostApplyPostRefresh: []plancheck.PlanCheck{},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("toolbox_external.example", "stage"),
					resource.TestCheckResourceAttr("toolbox_external.example", "stage", "read"),
				),
			},
		},
	})
}

// Simple test to see all the stages in action with bash
func TestResource_Bash_File(t *testing.T) {
	bashacc := os.Getenv("TF_ACC_BASH")
	bashlog := os.Getenv("TF_ACC_BASH_LOGFILE")
	if bashlog == "" {
		bashlog = "/tmp/resource_external.bash.log"
	}
	if bashacc == "" {
		t.Skipf("Skipping this test since the %s environment variable is not set to any value. "+
			"This test writes to log file %s set with %s environment variable, so it is disabled by default.",
			"TF_ACC_BASH", bashlog, "TF_ACC_BASH_LOGFILE",
		)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
						resource "toolbox_external" "example" {
							create = true
							delete = true
							update = true
							read = true
							program = [
								"bash",
								"-c",
								<<-EOT
								# Read input from stdin
								read INPUT
								printf "$INPUT\n" > %s
								printf '{"value":"two","test":{"inner":"value"}}'
								EOT
							]
							recreate = {foo: "bar"}
							query = {
								value = "test",
								value2 = null
							}
						}
					`, bashlog),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("toolbox_external.example", plancheck.ResourceActionCreate),
					},
					PostApplyPreRefresh:  []plancheck.PlanCheck{},
					PostApplyPostRefresh: []plancheck.PlanCheck{},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("toolbox_external.example", "stage"),
					resource.TestCheckResourceAttr("toolbox_external.example", "stage", "create"),
				),
			},
			{
				Config: fmt.Sprintf(`
						resource "toolbox_external" "example" {
							create = true
							delete = true
							update = true
							read = true
							program = [
								"bash",
								"-c",
								<<-EOT
								# Read input from stdin
								read INPUT
								printf "$INPUT\n" >> %s
								printf '{"value":"two"}'
								EOT
							]
							recreate = {foo: "bar"}
							query = {
								value = null,
								value2 = "test"
								value3 = "other"
							}
						}
						`, bashlog),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("toolbox_external.example", plancheck.ResourceActionUpdate),
					},
					PostApplyPreRefresh:  []plancheck.PlanCheck{},
					PostApplyPostRefresh: []plancheck.PlanCheck{},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("toolbox_external.example", "stage"),
					resource.TestCheckResourceAttr("toolbox_external.example", "stage", "update"),
				),
			},
			{
				Destroy: true,
				Config: fmt.Sprintf(`
						resource "toolbox_external" "example" {
							create = true
							delete = true
							update = true
							read = true
							program = [
								"bash",
								"-c",
								<<-EOT
								# Read input from stdin
								read INPUT
								printf "$INPUT\n" >> %s
								printf '{"value":"delete"}'
								EOT
							]
							recreate = {foo: "bar"}
							query = {
								value = null,
								value2 = "test"
								value3 = "never_seen_during_destroy"
							}
						}
						`, bashlog),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("toolbox_external.example", plancheck.ResourceActionDestroy),
					},
					PostApplyPreRefresh:  []plancheck.PlanCheck{},
					PostApplyPostRefresh: []plancheck.PlanCheck{},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("toolbox_external.example", "stage"),
					resource.TestCheckResourceAttr("toolbox_external.example", "stage", "read"),
				),
			},
		},
	})
}

func TestResource_upgrade(t *testing.T) {
	programPath, err := buildGoTestProgram()
	if err != nil {
		t.Fatal(err)
		return
	}

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion(),
				Config:            fmt.Sprintf(testResourceConfig_basic, programPath),
				Check: func(s *terraform.State) error {
					_, ok := s.RootModule().Resources["toolbox_external.test"]
					if !ok {
						return fmt.Errorf("missing resource")
					}

					outputs := s.RootModule().Outputs

					if outputs["argument"] == nil {
						return fmt.Errorf("missing 'argument' output")
					}
					if outputs["query_value"] == nil {
						return fmt.Errorf("missing 'query_value' output")
					}

					if outputs["argument"].Value != "cheese" {
						return fmt.Errorf(
							"'argument' output is %q; want 'cheese'",
							outputs["argument"].Value,
						)
					}
					if outputs["query_value"].Value != "pizza" {
						return fmt.Errorf(
							"'query_value' output is %q; want 'pizza'",
							outputs["query_value"].Value,
						)
					}

					return nil
				},
			},
			{
				ProtoV6ProviderFactories: protoV6ProviderFactories(),
				Config:                   fmt.Sprintf(testResourceConfig_basic, programPath),
				PlanOnly:                 true,
			},
			{
				ProtoV6ProviderFactories: protoV6ProviderFactories(),
				Config:                   fmt.Sprintf(testResourceConfig_basic, programPath),
				Check: func(s *terraform.State) error {
					_, ok := s.RootModule().Resources["toolbox_external.test"]
					if !ok {
						return fmt.Errorf("missing resource")
					}

					outputs := s.RootModule().Outputs

					if outputs["argument"] == nil {
						return fmt.Errorf("missing 'argument' output")
					}
					if outputs["query_value"] == nil {
						return fmt.Errorf("missing 'query_value' output")
					}

					if outputs["argument"].Value != "cheese" {
						return fmt.Errorf(
							"'argument' output is %q; want 'cheese'",
							outputs["argument"].Value,
						)
					}
					if outputs["query_value"].Value != "pizza" {
						return fmt.Errorf(
							"'query_value' output is %q; want 'pizza'",
							outputs["query_value"].Value,
						)
					}

					return nil
				},
			},
		},
	})
}

func buildGoTestProgram() (string, error) {
	// We have a simple Go program that we use as a stub for testing.
	cmd := exec.Command(
		"go", "install",
		"github.com/bryan-bar/terraform-provider-toolbox/internal/provider/test-programs/tf-acc-external-resource",
	)
	err := cmd.Run()

	if err != nil {
		return "", fmt.Errorf("failed to build test stub program: %s", err)
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(os.Getenv("HOME") + "/go")
	}

	programPath := path.Join(
		filepath.SplitList(gopath)[0], "bin", "tf-acc-external-resource",
	)
	return programPath, nil
}

// Reference: https://github.com/hashicorp/terraform-provider-external/issues/145
func TestResource_20MinuteTimeout(t *testing.T) {
	if os.Getenv(EnvTfAccExternalTimeoutTest) == "" {
		t.Skipf("Skipping this test since the %s environment variable is not set to any value. "+
			"This test requires 20 minutes to run, so it is disabled by default.",
			EnvTfAccExternalTimeoutTest,
		)
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "toolbox_external" "test" {
						program = ["sleep", "1205"] # over 20 minutes
					}
				`,
				// Not External Program Execution Failed / State: signal: killed
				ExpectError: regexp.MustCompile(`Unexpected External Program Results`),
			},
		},
	})
}
