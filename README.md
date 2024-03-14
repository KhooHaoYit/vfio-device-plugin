# Kubernetes VFIO Device Plugin
This plugin is used to present VFIO devices as allocatable resources in Kubernetes. While the plugin was written specifically to allow GPU/NPU passthrough for [Kata Containers](https://katacontainers.io), in practice the plugin is completely generic and it can handle any device controlled by the `vfio-pci` driver on the host machine. 



## Credits
This plugin is forked from [mrlhansen/vfio-device-plugin](https://github.com/mrlhansen/vfio-device-plugin). Add the following feature support:

1.Support to update vfio devices automatically when the vfio device bind.
2.Support to bind the vfio devices according to a config.

Contains the [vfio-pci bind script](https://github.com/andre-richter/vfio-pci-bind/blob/master/vfio-pci-bind.sh) which can bind device by manual.


## Configuration
In the `manifest` directory there is a YAML file for starting the plugin as a container in Kubernetes. This manifest also contains a configration file for the plugin, where you can specify what VFIO devices you want to register as a resource.

```yaml
- resourceName: ascend/910b
  vendorId: '19e5'
  deviceId: ['d802']
- resourceName: ascend/310
  vendorId: '19e5'
  deviceId: ['d301']
```

`resourceName` is the name of the resource in Kubernetes which will be used in resource request and it must be of the format `xxxx/yyyy`. 
`vendorId` is the value of the PCI device vendor. it can be get by `cat  /sys/bus/pci/drivers/vfio-pci/<DEVICE-ADDRESS>/vendor`.
`deviceId` is the identify value of one kind of PCI device. it can be get by `cat  /sys/bus/pci/drivers/vfio-pci/<DEVICE-ADDRESS>/device`.

Multiple device types with single vendor will be allowed as the `deviceId` is a list. 
