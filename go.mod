module github.com/openshift-metal3/terraform-provider-ironic

go 1.13

require (
	github.com/apparentlymart/go-dump v0.0.0-20190214190832-042adf3cf4a0 // indirect
	github.com/aws/aws-sdk-go v1.25.3 // indirect
	github.com/gophercloud/gophercloud v0.18.0
	github.com/gophercloud/utils v0.0.0-20210530213738-7c693d7efe47
	github.com/hashicorp/go-retryablehttp v0.6.4
	github.com/hashicorp/go-version v1.2.0
	github.com/hashicorp/hil v0.0.0-20190212132231-97b3a9cdfa93 // indirect
	github.com/hashicorp/terraform-plugin-sdk v1.0.0
	github.com/mattn/go-colorable v0.1.1 // indirect
	github.com/mattn/go-isatty v0.0.7 // indirect
	github.com/metal3-io/baremetal-operator/apis v0.0.0
	github.com/metal3-io/baremetal-operator/pkg/ironic v0.0.0-20211019071605-57305319bec4
	github.com/ulikunitz/xz v0.5.6 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
)

replace github.com/metal3-io/baremetal-operator/apis => github.com/metal3-io/baremetal-operator/apis v0.0.0-20211019071605-57305319bec4
