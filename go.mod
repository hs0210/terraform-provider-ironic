module github.com/openshift-metal3/terraform-provider-ironic

go 1.16

require (
	github.com/gophercloud/gophercloud v0.23.0
	github.com/gophercloud/utils v0.0.0-20210909165623-d7085207ff6d
	github.com/hashicorp/go-retryablehttp v0.7.0
	github.com/hashicorp/go-version v1.3.0
	github.com/hashicorp/terraform-plugin-sdk v1.17.2
	github.com/metal3-io/baremetal-operator v0.0.0-20220128094204-28771f489634
	github.com/metal3-io/baremetal-operator/apis v0.0.0
)

replace (
	github.com/metal3-io/baremetal-operator => github.com/hs0210/baremetal-operator v0.5.1 // Use OpenShift fork
	github.com/metal3-io/baremetal-operator/apis => github.com/openshift/baremetal-operator/apis v0.0.0-20220128094204-28771f489634 // Use OpenShift fork
	github.com/metal3-io/baremetal-operator/pkg/hardwareutils => github.com/openshift/baremetal-operator/pkg/hardwareutils v0.0.0-20220128094204-28771f489634 // Use OpenShift fork
)
