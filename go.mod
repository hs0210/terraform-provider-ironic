module github.com/openshift-metal3/terraform-provider-ironic

go 1.16

require (
	github.com/gophercloud/gophercloud v0.23.0
	github.com/gophercloud/utils v0.0.0-20210909165623-d7085207ff6d
	github.com/hashicorp/go-retryablehttp v0.7.0
	github.com/hashicorp/go-version v1.3.0
	github.com/hashicorp/terraform-plugin-sdk v1.17.2
	github.com/metal3-io/baremetal-operator v0.0.0-20211203102512-3572353e42e5
	github.com/metal3-io/baremetal-operator/apis v0.0.0
	github.com/metal3-io/baremetal-operator/pkg/hardwareutils v0.0.0
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
)

replace (
	github.com/gophercloud/gophercloud => github.com/gophercloud/gophercloud v0.17.0
	github.com/metal3-io/baremetal-operator/apis => github.com/metal3-io/baremetal-operator/apis v0.0.0-20211203102512-3572353e42e5
	github.com/metal3-io/baremetal-operator/pkg/hardwareutils => github.com/metal3-io/baremetal-operator/pkg/hardwareutils v0.0.0-20211203102512-3572353e42e5
)
