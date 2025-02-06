package main

import (
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

type vfioDevice struct {
	pciName    string
	deviceId   string
	vendorId   string
	iommuGroup string
}

type vfioGroup struct {
	resourceName string
	iommuGroups  []string
	pciAddresses []string
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if strings.EqualFold(v, str) {
			return true
		}
	}

	return false
}

func scanDevices() []vfioDevice {
	var names []string
	var devices []vfioDevice

	path := "/sys/bus/pci/drivers/vfio-pci"
	bdf := regexp.MustCompile(`^[a-f0-9]{4}:[a-f0-9]{2}:[a-f0-9]{2}.[0-9]$`)
	iommu := regexp.MustCompile(`\/(\d+)$`)

	list, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range list {
		mode := file.Mode()
		name := file.Name()
		if (mode & os.ModeSymlink) == os.ModeSymlink {
			if bdf.MatchString(name) {
				names = append(names, name)
			}
		}
	}

	for _, name := range names {
		fullpath := path + "/" + name

		content, err := ioutil.ReadFile(fullpath + "/vendor")
		if err != nil {
			log.Print(err)
			continue
		}
		vendor := strings.TrimSpace(string(content))
		vendor = vendor[2:]

		content, err = ioutil.ReadFile(fullpath + "/device")
		if err != nil {
			log.Print(err)
			continue
		}
		device := strings.TrimSpace(string(content))
		device = device[2:]

		dest, err := os.Readlink(fullpath + "/iommu_group")
		if err != nil {
			log.Print(err)
			continue
		}

		match := iommu.FindStringSubmatch(dest)
		if len(match) == 0 {
			log.Print("Failed to get IOMMU group")
			continue
		}
		dest = match[1]

		if _, err := os.Stat("/dev/vfio/" + dest); os.IsNotExist(err) {
			log.Print(err)
			continue
		}

		log.Print("Found PCI device " + name)
		log.Print("Vendor " + vendor)
		log.Print("Device " + device)
		log.Print("IOMMU Group " + dest)

		devices = append(devices, vfioDevice{
			pciName:    name,
			vendorId:   vendor,
			deviceId:   device,
			iommuGroup: dest,
		})
	}

	return devices
}

func groupDevices(devices []vfioDevice, config []vfioConfig) []vfioGroup {
	var groups []vfioGroup

	for _, group := range config {
		var iommuGroups []string
		var pciAddresses []string

		for _, device := range devices {
			if strings.EqualFold(device.vendorId, group.Vendor) == false {
				continue
			}

			if contains(group.Device, device.deviceId) == false {
				continue
			}

			iommuGroups = append(iommuGroups, device.iommuGroup)
			pciAddresses = append(pciAddresses, device.pciName)
		}

		if len(iommuGroups) == 0 {
			continue
		}

		groups = append(groups, vfioGroup{
			resourceName: group.Name,
			iommuGroups:  iommuGroups,
			pciAddresses: pciAddresses,
		})

		log.Print("Creating Resource " + group.Name)
		log.Print("IOMMU Groups " + strings.Join(iommuGroups, " "))
	}

	return groups
}
