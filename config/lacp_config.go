package config

import (
	"fmt"
	"opennos-mgmt/gnmi/modeldata/oc"
)

func (cfgMngr *ConfigMngrT) isLacpDeleted(ifname string) bool {
	device := (*cfgMngr.transCandidateConfig).(*oc.Device)
	lacpIntf, err := getLacpIntf(device, ifname)
	if lacpIntf == nil && err != nil {
		// OK, perhaps there were last LACP entry
		return true
	}

	return false
}

func getLacpIntf(device *oc.Device, ifname string) (*oc.Lacp_Interface, error) {
	lacp := device.GetLacp()
	if lacp == nil {
		return nil, fmt.Errorf("Failed to get LACP details from device config")
	}

	lacpIntf := lacp.GetInterface(ifname)
	if lacpIntf == nil {
		return nil, fmt.Errorf("Failed to get LACP interface %s details from device config", ifname)
	}

	return lacpIntf, nil
}
