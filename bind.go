// Copyright (c) file authors. All rights reserved.
//=============================================================================
//
// Author(s):
//   freesky-edward<freesky.edward@gmail.com>
//
// =============================================================================

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/jaypipes/ghw"
)

const (
	DRIVER_UNBIND_PATTEN   string = "/sys/bus/pci/drivers/%s/unbind"
	DRIVER_OVERRIDE_PATTEN string = "/sys/bus/pci/device/%s/driver_override"
	DRIVER_PROBE_PATH      string = "/sys/bus/pci/drivers_probe"
	DRIVER_NEW_ID_PATH     string = "/sys/bus/pci/drivers/vfio-pci/new_id"

	VFIO_PCI             string = "vfio_pci"
	DEVDRV_DEVICE_DRIVER        = "devdrv_device_driver"
)

// Check whether the system has load the vfio_pci driver
func isVfioPCIDriverReady() bool {
	cmd := exec.Command("lsmod")
	output, err := cmd.Output()
	if err != nil {
		log.Print(err)
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, VFIO_PCI) {
			return true
		}
	}
	return false
}

// Run the modprobe to load the vfio_pci driver
func loadVfioPCIDriver() {
	cmd := exec.Command("modprobe", VFIO_PCI)
	err := cmd.Run()
	if err != nil {
		log.Print(err)
	}
}

// Unbind the orignal driver of the device by write the device address to <driver>/unbind file.
// e.g. if attempt to unbind the BDF with address value: "0000:c0:00.0" from "vfio_pci" driver.
// The following command will be executed:
//
//	echo "0000:c0:00.0" > /sys/bus/pci/drivers/vfio-pci/unbind
func unbindPCIDevice(addr string, driver string) error {
	path := fmt.Sprintf(DRIVER_UNBIND_PATTEN, driver)
	err := writeValueToFile(addr, path)
	if err != nil {
		return err
	}
	return nil
}

// Bind the device to driver by write the driver name to  device/<device-id>/driver_override
// e.g. if attempt to bind the "0000:c1:00.0" to "devdrv_device_driver".
// The following command will be executed:
//
//	echo devdrv_device_driver > /sys/bus/pci/devices/0000:c0:00.0/driver_override
func driverOverride(addr string, driver string) error {
	path := fmt.Sprintf(DRIVER_OVERRIDE_PATTEN, addr)
	err := writeValueToFile(driver, path)
	if err != nil {
		return err
	}
	return nil
}

// Write the device to /sys/bus/pci/drivers_probe
func driverProbe(addr string) error {
	if err := writeValueToFile(addr, DRIVER_PROBE_PATH); err != nil {
		return err
	}
	return nil
}

// Write the device's "vendorid deviceId" into "/sys/bus/pci/drivers/vfio-pci/new_id""
func addDeviceToVfioPCI(vendorId int, deviceId int) error {
	var id string = fmt.Sprintf("%x %x", vendorId, deviceId)
	if err := writeValueToFile(id, DRIVER_NEW_ID_PATH); err != nil {
		return err
	}
	return nil
}

// Write a value to a file
func writeValueToFile(value string, filePath string) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY, 0400)
	if err != nil {
		return err
	}
	_, err = file.WriteString(value)
	if err != nil {
		return err
	}
	file.Close()
	return nil
}

// Bind the device to vfio-pci driver.
// which will execute the following steps:
//  1. check whether the device exist.
//  2. check whether the system spport iommu
//  3. unbind the device from bdf
//  4. override the driver to vfio
//  5. driver probe
//
// the error will return if only of the above step get failed.
func bindDeviceToVfioPCI(address string) error {
	var devicePath string = fmt.Sprintf("/sys/bus/pci/devices/%s", address)
	var iommuPath string = fmt.Sprintf("/sys/bus/pci/devices/%s/iommu/", address)
	if _, err := os.Stat(devicePath); os.IsNotExist(err) {
		return fmt.Errorf("Device is not exist: %v", err)
	}

	if _, err := os.Stat(iommuPath); os.IsNotExist(err) {
		return fmt.Errorf("Error: Check your hardware or linux cmdline parameters. Use intel_iommu=on or iommu=pt iommu=1")
	}

	if err := unbindPCIDevice(address, DEVDRV_DEVICE_DRIVER); err != nil {
		return fmt.Errorf("unbind device from devdrv_device_driver error: %v", err)
	}

	if err := driverOverride(address, "vfio-pci"); err != nil {
		return fmt.Errorf("driver_override error: %v", err)
	}

	if err := driverProbe(address); err != nil {
		return fmt.Errorf("driver_probe error: %v", err)
	}
	return nil
}

// Find all the devices by the device's venderid and deviceTypeID.
// e.g. if the node has the NPU 910b. its vendor id "19e5", deviceType "d802"
// the result will return the dbdf number of all the devices like:
// ["0000:00.00.0","0000:01:00.0","0000:c0:00.0","0000:c1:00.0"]
func findDevices(vendorId string, deviceId string) ([]string, error) {

	ret := []string{}
	pci, err := ghw.PCI()
	if err != nil {
		log.Print(err)
		return ret, err
	}
	for _, device := range pci.Devices {
		vendor := device.Vendor
		deviceType := device.Product

		if vendor.ID == vendorId && deviceType.ID == deviceId {
			fmt.Printf("%s %s has device %s\n", vendorId, deviceId, device.Address)
			ret = append(ret, device.Address)
		}
	}

	return ret, nil
}
