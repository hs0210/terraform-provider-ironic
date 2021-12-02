package ironic

import (
	"fmt"
	"reflect"

	"github.com/gophercloud/gophercloud/openstack/baremetal/v1/nodes"
	metal3v1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
)

const (
	noRAIDInterface       string = "no-raid"
	softwareRAIDInterface string = "agent"
)

// buildTargetRAIDCfg build RAID logical disks, this method doesn't set the root volume
func buildTargetRAIDCfg(raid *metal3v1alpha1.RAIDConfig) (logicalDisks []nodes.LogicalDisk, err error) {
	// Deal possible panic
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("panic in build RAID settings: %v", r)
		}
	}()

	if raid == nil {
		return
	}

	// build logicalDisks
	if len(raid.HardwareRAIDVolumes) != 0 {
		logicalDisks, err = buildTargetHardwareRAIDCfg(raid.HardwareRAIDVolumes)
	} else if len(raid.SoftwareRAIDVolumes) != 0 {
		logicalDisks, err = buildTargetSoftwareRAIDCfg(raid.SoftwareRAIDVolumes)
	}

	return
}

// A private method to build hardware RAID disks
func buildTargetHardwareRAIDCfg(volumes []metal3v1alpha1.HardwareRAIDVolume) (logicalDisks []nodes.LogicalDisk, err error) {
	var (
		logicalDisk    nodes.LogicalDisk
		nameCheckFlags map[string]int = make(map[string]int)
	)

	if len(volumes) == 0 {
		return
	}

	for index, volume := range volumes {
		// Check volume's name
		if volume.Name != "" {
			i, exist := nameCheckFlags[volume.Name]
			if exist {
				return nil, fmt.Errorf("the names(%s) of volume[%d] and volume[%d] are repeated", volume.Name, index, i)
			}
			nameCheckFlags[volume.Name] = index
		}
		// Build logicalDisk
		logicalDisk = nodes.LogicalDisk{
			SizeGB:     volume.SizeGibibytes,
			RAIDLevel:  nodes.RAIDLevel(volume.Level),
			VolumeName: volume.Name,
		}
		if volume.Rotational != nil {
			if *volume.Rotational {
				logicalDisk.DiskType = nodes.HDD
			} else {
				logicalDisk.DiskType = nodes.SSD
			}
		}
		if volume.NumberOfPhysicalDisks != nil {
			logicalDisk.NumberOfPhysicalDisks = *volume.NumberOfPhysicalDisks
		}
		// Add to logicalDisks
		logicalDisks = append(logicalDisks, logicalDisk)
	}

	return
}

// A private method to build software RAID disks
func buildTargetSoftwareRAIDCfg(volumes []metal3v1alpha1.SoftwareRAIDVolume) (logicalDisks []nodes.LogicalDisk, err error) {
	var (
		logicalDisk nodes.LogicalDisk
	)

	if len(volumes) == 0 {
		return
	}

	if nodes.RAIDLevel(volumes[0].Level) != nodes.RAID1 {
		return nil, fmt.Errorf("the level in first volume of software raid must be RAID1")
	}

	for _, volume := range volumes {
		// Build logicalDisk
		logicalDisk = nodes.LogicalDisk{
			SizeGB:     volume.SizeGibibytes,
			RAIDLevel:  nodes.RAIDLevel(volume.Level),
			Controller: "software",
		}
		// Build physical disks hint
		for i := range volume.PhysicalDisks {
			logicalDisk.PhysicalDisks = append(logicalDisk.PhysicalDisks, makeHintMap(&volume.PhysicalDisks[i]))
		}
		// Add to logicalDisks
		logicalDisks = append(logicalDisks, logicalDisk)
	}

	return
}

// makeHintMap converts a RootDeviceHints instance into a string map
// suitable to pass to ironic.
func makeHintMap(source *metal3v1alpha1.RootDeviceHints) map[string]string {
	hints := map[string]string{}

	if source == nil {
		return hints
	}

	if source.DeviceName != "" {
		hints["name"] = fmt.Sprintf("s== %s", source.DeviceName)
	}
	if source.HCTL != "" {
		hints["hctl"] = fmt.Sprintf("s== %s", source.HCTL)
	}
	if source.Model != "" {
		hints["model"] = fmt.Sprintf("<in> %s", source.Model)
	}
	if source.Vendor != "" {
		hints["vendor"] = fmt.Sprintf("<in> %s", source.Vendor)
	}
	if source.SerialNumber != "" {
		hints["serial"] = fmt.Sprintf("s== %s", source.SerialNumber)
	}
	if source.MinSizeGigabytes != 0 {
		hints["size"] = fmt.Sprintf(">= %d", source.MinSizeGigabytes)
	}
	if source.WWN != "" {
		hints["wwn"] = fmt.Sprintf("s== %s", source.WWN)
	}
	if source.WWNWithExtension != "" {
		hints["wwn_with_extension"] = fmt.Sprintf("s== %s", source.WWNWithExtension)
	}
	if source.WWNVendorExtension != "" {
		hints["wwn_vendor_extension"] = fmt.Sprintf("s== %s", source.WWNVendorExtension)
	}
	switch {
	case source.Rotational == nil:
	case *source.Rotational == true:
		hints["rotational"] = "true"
	case *source.Rotational == false:
		hints["rotational"] = "false"
	}

	return hints
}

// buildRAIDCleanSteps build the clean steps for RAID configuration from BaremetalHost spec
func buildRAIDCleanSteps(raidInterface string, target *metal3v1alpha1.RAIDConfig, actual *metal3v1alpha1.RAIDConfig) (cleanSteps []nodes.CleanStep, err error) {
	err = checkRAIDConfigure(raidInterface, target)
	if err != nil {
		return nil, err
	}

	// No RAID
	if raidInterface == noRAIDInterface {
		return
	}

	// Software RAID
	if raidInterface == softwareRAIDInterface {
		// Ignore HardwareRAIDVolumes
		if target != nil {
			target.HardwareRAIDVolumes = nil
		}
		if actual != nil {
			actual.HardwareRAIDVolumes = nil
		}
		if reflect.DeepEqual(target, actual) {
			return
		}

		cleanSteps = append(
			cleanSteps,
			[]nodes.CleanStep{
				{
					Interface: "raid",
					Step:      "delete_configuration",
				},
				{
					Interface: "deploy",
					Step:      "erase_devices_metadata",
				},
			}...,
		)

		// If software raid configuration is empty, only need to clear old configuration
		if target == nil || len(target.SoftwareRAIDVolumes) == 0 {
			return
		}

		cleanSteps = append(
			cleanSteps,
			nodes.CleanStep{
				Interface: "raid",
				Step:      "create_configuration",
			},
		)
		return
	}

	// Hardware RAID
	// If hardware RAID configuration is nil,
	// keep old hardware RAID configuration
	if target == nil || target.HardwareRAIDVolumes == nil {
		return
	}

	// Ignore SoftwareRAIDVolumes
	target.SoftwareRAIDVolumes = nil
	if actual != nil {
		actual.SoftwareRAIDVolumes = nil
	}
	if reflect.DeepEqual(target, actual) {
		return
	}

	// Add ‘delete_configuration’ before ‘create_configuration’ to make sure
	// that only the desired logical disks exist in the system after manual cleaning.
	cleanSteps = append(
		cleanSteps,
		nodes.CleanStep{
			Interface: "raid",
			Step:      "delete_configuration",
		},
	)

	// If hardware raid configuration is empty, only need to clear old configuration
	if len(target.HardwareRAIDVolumes) == 0 {
		return
	}

	// ‘create_configuration’ doesn’t remove existing disks. It is recommended
	// that only the desired logical disks exist in the system after manual cleaning.
	cleanSteps = append(
		cleanSteps,
		nodes.CleanStep{
			Interface: "raid",
			Step:      "create_configuration",
		},
	)
	return
}

func checkRAIDConfigure(raidInterface string, raid *metal3v1alpha1.RAIDConfig) error {
	switch raidInterface {
	case noRAIDInterface:
		if raid != nil && (len(raid.HardwareRAIDVolumes) != 0 || len(raid.SoftwareRAIDVolumes) != 0) {
			return fmt.Errorf("raid settings are defined, but the node's driver %s does not support RAID", raidInterface)
		}
	case softwareRAIDInterface:
		if raid != nil && len(raid.HardwareRAIDVolumes) != 0 {
			return fmt.Errorf("node's driver %s does not support hardware RAID", raidInterface)
		}
	default:
		if raid != nil && len(raid.HardwareRAIDVolumes) == 0 && len(raid.SoftwareRAIDVolumes) != 0 {
			return fmt.Errorf("node's driver %s does not support software RAID", raidInterface)
		}
	}
	return nil
}
