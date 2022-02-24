// +build acceptance

package ironic

import (
	"fmt"
	"testing"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/baremetal/v1/nodes"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccIronicNode(t *testing.T) {
	var node nodes.Node
	var targetRAIDConfig map[string]interface{}

	// raidConfig = make(map[string]interface{})
	targetRAIDConfig = make(map[string]interface{})
	// raidConfig["SoftwareRAIDVolume"] = map[string]string{"Level": "1"}
	targetRAIDConfig["SoftwareRAIDVolume"] = map[string]string{"Level": "1"}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccNodeDestroy,
		Steps: []resource.TestStep{

			// Create a node and check that it exists
			{
				Config: testAccNodeResource("", nil),
				Check: resource.ComposeTestCheckFunc(
					CheckNodeExists("ironic_node_v1.node-0", &node),
					resource.TestCheckResourceAttr("ironic_node_v1.node-0",
						"provision_state", "enroll",
					),
				),
			},

			// Ensure node is manageable
			{
				Config: testAccNodeResource("manage = true", nil),
				Check: resource.ComposeTestCheckFunc(
					CheckNodeExists("ironic_node_v1.node-0", &node),
					resource.TestCheckResourceAttr("ironic_node_v1.node-0",
						"provision_state", "manageable"),
				),
			},

			// Inspect the node
			{
				Config: testAccNodeResource("inspect = true", nil),
				Check: resource.ComposeTestCheckFunc(
					CheckNodeExists("ironic_node_v1.node-0", &node),
					resource.TestCheckResourceAttr("ironic_node_v1.node-0",
						"provision_state", "manageable"),
				),
			},

			// Clean the node
			{
				Config: testAccNodeResource(`
					"clean = true"
					raid_interface = agent`, targetRAIDConfig),
				Check: resource.ComposeTestCheckFunc(
					CheckNodeExists("ironic_node_v1.node-0", &node),
					resource.TestCheckResourceAttr("ironic_node_v1.node-0",
						"provision_state", "manageable"),
					resource.TestCheckResourceAttr("ironic_node_v1.node-0",
						"raid_config", "fake"),
				),
			},

			// Change the node's power state to 'power on', with a timeout
			{
				Config: testAccNodeResource(`
					target_power_state = "power on"
					power_state_timeout = 10
				`, nil),
				Check: resource.ComposeTestCheckFunc(
					CheckNodeExists("ironic_node_v1.node-0", &node),
					resource.TestCheckResourceAttr("ironic_node_v1.node-0",
						"power_state", "power on"),
				),
			},

			// Change the node's power state to 'power off'.
			{
				Config: testAccNodeResource("target_power_state = \"power off\"", nil),
				Check: resource.ComposeTestCheckFunc(
					CheckNodeExists("ironic_node_v1.node-0", &node),
					resource.TestCheckResourceAttr("ironic_node_v1.node-0",
						"power_state", "power off"),
				),
			},

			// Change the node's power state to 'rebooting', it probably
			// doesn't make a whole lot of sense for a terraform user to
			// declare a node's state as forever rebooting, as it'd reboot
			// every time, but we should check anyway that if they do say
			// rebooting, power_state goes to power_on and terraform exits
			// successfully.
			{
				Config: testAccNodeResource("target_power_state = \"rebooting\"", nil),
				Check: resource.ComposeTestCheckFunc(
					CheckNodeExists("ironic_node_v1.node-0", &node),
					resource.TestCheckResourceAttr("ironic_node_v1.node-0",
						"power_state", "power on"),
				),
			},
		},
	})
}

func CheckNodeExists(name string, node *nodes.Node) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client, err := testAccProvider.Meta().(*Clients).GetIronicClient()
		if err != nil {
			return err
		}

		rs, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no node ID is set")
		}

		result, err := nodes.Get(client, rs.Primary.ID).Extract()
		if err != nil {
			return fmt.Errorf("node (%s) not found: %s", rs.Primary.ID, err)
		}

		*node = *result

		return nil
	}
}

func testAccNodeDestroy(state *terraform.State) error {
	client, err := testAccProvider.Meta().(*Clients).GetIronicClient()
	if err != nil {
		return err
	}

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "ironic_node_v1" {
			continue
		}

		_, err := nodes.Get(client, rs.Primary.ID).Extract()
		if _, ok := err.(gophercloud.ErrDefault404); !ok {
			return fmt.Errorf("unexpected error: %s, expected 404", err)
		}
	}

	return nil
}

func testAccNodeResource(extraValue string, targetRAIDConfig map[string]interface{}) string {
	return fmt.Sprintf(`resource "ironic_node_v1" "node-0" {
			name = "node-0"
			driver = "fake-hardware"

			boot_interface = "pxe"
			deploy_interface = "fake"
			inspect_interface = "fake"
			management_interface = "fake"
			power_interface = "fake"
			resource_class = "baremetal"
			vendor_interface = "no-vendor"

			driver_info = {
				ipmi_port      = "6230"
				ipmi_username  = "admin"
				ipmi_password  = "admin"
			}

			%s
			%+v
		}`, extraValue, targetRAIDConfig)
}
