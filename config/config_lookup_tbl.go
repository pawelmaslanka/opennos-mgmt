package config

import (
	"errors"
	"fmt"
	lib "golibext"
	"opennos-mgmt/gnmi/modeldata/oc"
	"strings"

	log "github.com/golang/glog"

	"sort"
)

type configLookupTablesT struct {
	idxOfLastAddedIntf lib.IdxT
	idxByEthName       map[string]lib.IdxT
	intfNameByIdx      map[lib.IdxT]string
	idxOfLastAddedLag  lib.IdxT
	idxByLagName       map[string]lib.IdxT
	lagNameByIdx       map[lib.IdxT]string
	idxOfLastAddedVlan lib.IdxT
	idxByVlanName      map[string]lib.IdxT
	vlanNameByIdx      map[lib.IdxT]string

	// L3 interface can have assigned many IPv4 addresses
	ipv4AddrByEth map[lib.IdxT]*lib.StringSet
	// L3 LAG can have assigned many IPv4 addresses
	ipv4AddrByLag map[lib.IdxT]*lib.StringSet
	// L3 VLAN can have assigned many IPv4 addresses
	ipv4AddrByVlan map[lib.VidT]*lib.StringSet
	ethByIpv4Addr  map[string]lib.IdxT
	lagByIpv4Addr  map[string]lib.IdxT
	vlanByIpv4Addr map[string]lib.VidT
	ipv6AddrByEth  map[lib.IdxT]*lib.StringSet
	ipv6AddrByLag  map[lib.IdxT]*lib.StringSet
	ipv6AddrByVlan map[lib.VidT]*lib.StringSet
	// L3 interface can have assigned many IPv6 addresses
	ethByIpv6Addr map[string]lib.IdxT
	// L3 LAG can have assigned many IPv6 addresses
	lagByIpv6Addr map[string]lib.IdxT
	// L3 VLAN interface can have assigned many IPv6 addresses
	vlanByIpv6Addr map[string]lib.VidT
	lagByEth       map[lib.IdxT]lib.IdxT
	// LAG can have many interface members
	ethByLag        map[lib.IdxT]*lib.IdxTSet
	stpByEth        *lib.IdxTSet
	vlanModeByEth   map[lib.IdxT]oc.E_OpenconfigVlan_VlanModeType
	vlanModeByLag   map[lib.IdxT]oc.E_OpenconfigVlan_VlanModeType
	vlanAccessByEth map[lib.IdxT]lib.VidT
	// There can be many ports in specific VLAN ID for access mode
	ethByVlanAccess map[lib.VidT]*lib.IdxTSet
	vlanAccessByLag map[lib.IdxT]lib.VidT
	// There can be many LAGs in specific VLAN ID for access mode
	lagByVlanAccess map[lib.VidT]*lib.IdxTSet
	vlanNativeByEth map[lib.IdxT]lib.VidT
	// There can be many ports in specific VLAN ID for native tag
	ethByVlanNative map[lib.VidT]*lib.IdxTSet
	vlanNativeByLag map[lib.IdxT]lib.VidT
	// There can be many LAGs in specific VLAN ID for native tag
	lagByVlanNative map[lib.VidT]*lib.IdxTSet
	vlanTrunkByEth  map[lib.IdxT]*lib.VidTSet
	// There can be many ports in VLAN trunk
	ethByVlanTrunk map[lib.VidT]*lib.IdxTSet
	vlanTrunkByLag map[lib.IdxT]*lib.VidTSet
	// There can be many LAGs in VLAN trunk
	lagByVlanTrunk map[lib.VidT]*lib.IdxTSet
}

func newConfigLookupTables() *configLookupTablesT {
	return &configLookupTablesT{
		idxOfLastAddedIntf: 0,
		idxByEthName:       make(map[string]lib.IdxT, maxPortsC),
		intfNameByIdx:      make(map[lib.IdxT]string, maxPortsC),
		idxOfLastAddedLag:  0,
		idxByLagName:       make(map[string]lib.IdxT),
		lagNameByIdx:       make(map[lib.IdxT]string),
		idxOfLastAddedVlan: 0,
		idxByVlanName:      make(map[string]lib.IdxT),
		vlanNameByIdx:      make(map[lib.IdxT]string),
		ipv4AddrByEth:      make(map[lib.IdxT]*lib.StringSet),
		ipv4AddrByLag:      make(map[lib.IdxT]*lib.StringSet),
		ipv4AddrByVlan:     make(map[lib.VidT]*lib.StringSet),
		ethByIpv4Addr:      make(map[string]lib.IdxT),
		lagByIpv4Addr:      make(map[string]lib.IdxT),
		vlanByIpv4Addr:     make(map[string]lib.VidT),
		ipv6AddrByEth:      make(map[lib.IdxT]*lib.StringSet),
		ipv6AddrByLag:      make(map[lib.IdxT]*lib.StringSet),
		ipv6AddrByVlan:     make(map[lib.VidT]*lib.StringSet),
		ethByIpv6Addr:      make(map[string]lib.IdxT),
		lagByIpv6Addr:      make(map[string]lib.IdxT),
		vlanByIpv6Addr:     make(map[string]lib.VidT),
		lagByEth:           make(map[lib.IdxT]lib.IdxT),
		ethByLag:           make(map[lib.IdxT]*lib.IdxTSet),
		stpByEth:           lib.NewIdxTSet(),
		vlanModeByEth:      make(map[lib.IdxT]oc.E_OpenconfigVlan_VlanModeType),
		vlanModeByLag:      make(map[lib.IdxT]oc.E_OpenconfigVlan_VlanModeType),
		vlanAccessByEth:    make(map[lib.IdxT]lib.VidT),
		vlanAccessByLag:    make(map[lib.IdxT]lib.VidT),
		ethByVlanAccess:    make(map[lib.VidT]*lib.IdxTSet),
		lagByVlanAccess:    make(map[lib.VidT]*lib.IdxTSet),
		vlanNativeByEth:    make(map[lib.IdxT]lib.VidT),
		vlanNativeByLag:    make(map[lib.IdxT]lib.VidT),
		ethByVlanNative:    make(map[lib.VidT]*lib.IdxTSet),
		lagByVlanNative:    make(map[lib.VidT]*lib.IdxTSet),
		vlanTrunkByEth:     make(map[lib.IdxT]*lib.VidTSet),
		vlanTrunkByLag:     make(map[lib.IdxT]*lib.VidTSet),
		ethByVlanTrunk:     make(map[lib.VidT]*lib.IdxTSet),
		lagByVlanTrunk:     make(map[lib.VidT]*lib.IdxTSet),
	}
}

func (this *configLookupTablesT) checkDependenciesForSetLagIntfMember(lagName string, ifname string) error {
	intfIdx, exists := this.idxByEthName[ifname]
	if !exists {
		return fmt.Errorf("Ethernet interface %s does not exists", ifname)
	}

	var err error
	strBuilder := strings.Builder{}
	if _, exists := this.idxByLagName[lagName]; exists {
		configuredLagIdx, exists := this.lagByEth[intfIdx]
		if exists {
			msg := fmt.Sprintf("Ethernet interface %s is already member of LAG %s", ifname, this.lagNameByIdx[configuredLagIdx])
			if _, err = strBuilder.WriteString(msg); err != nil {
				return err
			}
		}
	} else {
		msg := fmt.Sprintf("LAG interface %s does not exist", lagName)
		if _, err = strBuilder.WriteString(msg); err != nil {
			return err
		}
	}

	if strBuilder.Len() == 0 {
		return nil
	}

	return errors.New(strBuilder.String())
}

func (this *configLookupTablesT) checkDependenciesForDeleteLagIntfMember(lagName string, ifname string) error {
	intfIdx, exists := this.idxByEthName[ifname]
	if !exists {
		return fmt.Errorf("Ethernet interface %s does not exist", ifname)
	}

	var err error
	strBuilder := strings.Builder{}
	lagIdx, exists := this.idxByLagName[lagName]
	if !exists {
		msg := fmt.Sprintf("LAG %s does not exist", lagName)
		if _, err = strBuilder.WriteString(msg); err != nil {
			return err
		}
	}

	if this.lagByEth[intfIdx] != lagIdx {
		msg := fmt.Sprintf("Ethernet interface %s does not exists in LAG %s", ifname, lagName)
		if _, err = strBuilder.WriteString(msg); err != nil {
			return err
		}
	}

	if strBuilder.Len() == 0 {
		return nil
	}

	return errors.New(strBuilder.String())
}

func (this *configLookupTablesT) setLagIntfMember(lagName string, ifname string) error {
	intfIdx, exists := this.idxByEthName[ifname]
	if !exists {
		return fmt.Errorf("Ethernet interface %s does not exists", ifname)
	}

	lagIdx, exists := this.idxByLagName[lagName]
	if !exists {
		if err := this.addNewInterfaceIfItDoesNotExist(lagName); err != nil {
			return err
		}
	}

	if !this.ethByLag[lagIdx].Has(intfIdx) {
		this.ethByLag[lagIdx].Add(intfIdx)
		this.lagByEth[intfIdx] = lagIdx
	}

	return nil
}

func (this *configLookupTablesT) deleteLagIntfMember(lagName string, ifname string) error {
	intfIdx, exists := this.idxByEthName[ifname]
	if !exists {
		return fmt.Errorf("Ethernet interface %s does not exist", ifname)
	}

	lagIdx, exists := this.idxByLagName[lagName]
	if !exists {
		return fmt.Errorf("LAG %s does not exist", lagName)
	}

	if this.lagByEth[intfIdx] != lagIdx {
		return fmt.Errorf("Ethernet interface %s in not member of LAG %s", ifname, lagName)
	}

	delete(this.lagByEth, intfIdx)
	this.ethByLag[lagIdx].Delete(intfIdx)

	return nil
}

func (this *configLookupTablesT) checkDependenciesForDeleteOrRemoveLagIntf(lagName string) error {
	var err error
	strBuilder := strings.Builder{}
	lagIdx, exists := this.idxByLagName[lagName]
	if !exists {
		return fmt.Errorf("LAG %s does not exists", lagName)
	}

	if intfs, exists := this.ethByLag[lagIdx]; exists {
		if intfs.Size() > 0 {
			if _, err = strBuilder.WriteString("LAG members:"); err != nil {
				return err
			}

			for _, intfIdx := range intfs.IdxTs() {
				if _, err = strBuilder.WriteString(fmt.Sprintf(" %s", this.intfNameByIdx[intfIdx])); err != nil {
					return err
				}
			}

			if _, err = strBuilder.WriteString("\n"); err != nil {
				return err
			}
		}
	}

	if strBuilder.Len() == 0 {
		return nil
	}

	return errors.New(strBuilder.String())
}

func (this *configLookupTablesT) checkDependenciesForSetIpv4AddrForEthIntf(ifname string, cidr4 string) error {
	var err error
	strBuilder := strings.Builder{}
	intfIdx, exists := this.ethByIpv4Addr[cidr4]
	if exists {
		msg := fmt.Sprintf("IPv4 address %s is configured on Ethernet interface %s",
			cidr4, this.intfNameByIdx[intfIdx])
		if _, err = strBuilder.WriteString(msg); err != nil {
			return err
		}
	}

	if strBuilder.Len() == 0 {
		return nil
	}

	return errors.New(strBuilder.String())
}

func (this *configLookupTablesT) checkDependenciesForDeleteIpv4AddrFromEthIntf(ifname string, cidr4 string) error {
	var err error
	strBuilder := strings.Builder{}
	intfIdx := this.idxByEthName[ifname]
	allIpv4Addr, exists := this.ipv4AddrByEth[intfIdx]
	if !exists {
		if _, err = strBuilder.WriteString(fmt.Sprintf("There is not any IPv4 address on Ethernet interface %s", ifname)); err != nil {
			return err
		}
	} else if !allIpv4Addr.Has(cidr4) {
		if _, err = strBuilder.WriteString(fmt.Sprintf("There is not IPv4 address %s on Ethernet interface %s", cidr4, ifname)); err != nil {
			return err
		}
	}

	foundIpIntfIdx, exists := this.ethByIpv4Addr[cidr4]
	if exists && foundIpIntfIdx != intfIdx {
		if _, err = strBuilder.WriteString(fmt.Sprintf("IPv4 address %s is on Ethernet interface %s", cidr4, this.intfNameByIdx[intfIdx])); err != nil {
			return err
		}
	}

	if strBuilder.Len() == 0 {
		return nil
	}

	return errors.New(strBuilder.String())
}

func (this *configLookupTablesT) checkDependenciesForSetVlanModeForEthIntf(ifname string, setVlanMode oc.E_OpenconfigVlan_VlanModeType) error {
	var err error
	strBuilder := strings.Builder{}
	intfIdx := this.idxByEthName[ifname]

	if vlanMode, exists := this.vlanModeByEth[intfIdx]; exists {
		if vlanMode == setVlanMode {
			msg := fmt.Sprintf("VLAN mode (%d) is already configured on Ethernet interface %s", setVlanMode, ifname)
			if _, err = strBuilder.WriteString(msg); err != nil {
				return err
			}
		}
	}

	if strBuilder.Len() == 0 {
		return nil
	}

	return errors.New(strBuilder.String())
}

func (this *configLookupTablesT) checkDependenciesForSetAccessVlanForEthIntf(ifname string, setVid lib.VidT) error {
	var err error
	strBuilder := strings.Builder{}
	intfIdx := this.idxByEthName[ifname]

	if accessVid, exists := this.vlanAccessByEth[intfIdx]; exists {
		if accessVid == setVid {
			msg := fmt.Sprintf("Access VLAN %d is already configured on Ethernet interface %s", setVid, ifname)
			if _, err = strBuilder.WriteString(msg); err != nil {
				return err
			}
		} else {
			msg := fmt.Sprintf("There is other native VLAN %d configured on Ethernet interface %s", accessVid, ifname)
			if _, err = strBuilder.WriteString(msg); err != nil {
				return err
			}
		}
	}

	if strBuilder.Len() == 0 {
		return nil
	}

	return errors.New(strBuilder.String())
}

func (this *configLookupTablesT) checkDependenciesForDeleteAccessVlanFromEthIntf(ifname string, deleteVid lib.VidT) error {
	var err error
	strBuilder := strings.Builder{}
	intfIdx := this.idxByEthName[ifname]

	accessVid, exists := this.vlanAccessByEth[intfIdx]
	if !exists {
		msg := fmt.Sprintf("Access VLAN %d is not configured on Ethernet interface %s", deleteVid, ifname)
		if _, err = strBuilder.WriteString(msg); err != nil {
			return err
		}
	} else if accessVid != deleteVid {
		msg := fmt.Sprintf("Currently access VLAN %d is configured on Ethernet interface %s", accessVid, ifname)
		if _, err = strBuilder.WriteString(msg); err != nil {
			return err
		}
	}

	vlanMode, err := this.getVlanModeEthIntf(ifname)
	if err != nil {
		msg := fmt.Sprintf("%s", err)
		if _, err = strBuilder.WriteString(msg); err != nil {
			return err
		}
	}

	if vlanMode != oc.OpenconfigVlan_VlanModeType_ACCESS {
		msg := fmt.Sprintf("There is not set access VLAN mode on interface %s. Current mode: %v", ifname, vlanMode)
		if _, err = strBuilder.WriteString(msg); err != nil {
			return err
		}
	}

	if strBuilder.Len() == 0 {
		return nil
	}

	return errors.New(strBuilder.String())
}

func (this *configLookupTablesT) checkDependenciesForSetNativeVlanForEthIntf(ifname string, setVid lib.VidT) error {
	var err error
	strBuilder := strings.Builder{}
	intfIdx := this.idxByEthName[ifname]

	if nativeVid, exists := this.vlanNativeByEth[intfIdx]; exists {
		if nativeVid == setVid {
			msg := fmt.Sprintf("Native VLAN %d is already configured on Ethernet interface %s", setVid, ifname)
			if _, err = strBuilder.WriteString(msg); err != nil {
				return err
			}
		} else {
			msg := fmt.Sprintf("There is other native VLAN %d configured on Ethernet interface %s", nativeVid, ifname)
			if _, err = strBuilder.WriteString(msg); err != nil {
				return err
			}
		}
	}

	if strBuilder.Len() == 0 {
		return nil
	}

	return errors.New(strBuilder.String())
}

func (this *configLookupTablesT) checkDependenciesForDeleteNativeVlanFromEthIntf(ifname string, deleteVid lib.VidT) error {
	var err error
	strBuilder := strings.Builder{}
	intfIdx := this.idxByEthName[ifname]

	nativeVid, exists := this.vlanNativeByEth[intfIdx]
	if !exists {
		msg := fmt.Sprintf("Native VLAN %d is not configured on Ethernet interface %s", deleteVid, ifname)
		if _, err = strBuilder.WriteString(msg); err != nil {
			return err
		}
	} else if nativeVid != deleteVid {
		msg := fmt.Sprintf("Currently native VLAN %d is configured on Ethernet interface %s", nativeVid, ifname)
		if _, err = strBuilder.WriteString(msg); err != nil {
			return err
		}
	}

	vlanMode, err := this.getVlanModeEthIntf(ifname)
	if err != nil {
		msg := fmt.Sprintf("%s", err)
		if _, err = strBuilder.WriteString(msg); err != nil {
			return err
		}
	}

	if vlanMode != oc.OpenconfigVlan_VlanModeType_TRUNK {
		msg := fmt.Sprintf("There is not set trunk VLAN mode on interface %s. Current mode: %v", ifname, vlanMode)
		if _, err = strBuilder.WriteString(msg); err != nil {
			return err
		}
	}

	if strBuilder.Len() == 0 {
		return nil
	}

	return errors.New(strBuilder.String())
}

func (this *configLookupTablesT) checkDependenciesForSetTrunkVlanForEthIntf(ifname string, setVid lib.VidT) error {
	var err error
	strBuilder := strings.Builder{}
	intfIdx := this.idxByEthName[ifname]

	if trunkVids, exists := this.vlanTrunkByEth[intfIdx]; exists {
		if trunkVids.Has(setVid) {
			msg := fmt.Sprintf("Trunk VLAN %d is already configured on Ethernet interface %s", setVid, ifname)
			if _, err = strBuilder.WriteString(msg); err != nil {
				return err
			}
		}
	}

	if strBuilder.Len() == 0 {
		return nil
	}

	return errors.New(strBuilder.String())
}

func (this *configLookupTablesT) checkDependenciesForDeleteTrunkVlanFromEthIntf(ifname string, deleteVid lib.VidT) error {
	var err error
	strBuilder := strings.Builder{}
	intfIdx := this.idxByEthName[ifname]
	if _, exists := this.vlanTrunkByEth[intfIdx]; !exists {
		msg := fmt.Sprintf("Trunk VLAN %d is not configured on Ethernet interface %s", deleteVid, ifname)
		if _, err = strBuilder.WriteString(msg); err != nil {
			return err
		}
	}

	vlanMode, err := this.getVlanModeEthIntf(ifname)
	if err != nil {
		msg := fmt.Sprintf("%s", err)
		if _, err = strBuilder.WriteString(msg); err != nil {
			return err
		}
	}

	if vlanMode != oc.OpenconfigVlan_VlanModeType_TRUNK {
		msg := fmt.Sprintf("There is not set trunk VLAN mode on interface %s. Current mode: %v", ifname, vlanMode)
		if _, err = strBuilder.WriteString(msg); err != nil {
			return err
		}
	}

	if strBuilder.Len() == 0 {
		return nil
	}

	return errors.New(strBuilder.String())
}

func (this *configLookupTablesT) checkDependenciesForDeletePortBreakout(ifname string) error {
	var err error
	strBuilder := strings.Builder{}
	intfIdx := this.idxByEthName[ifname]
	if allIpv4Addr, exists := this.ipv4AddrByEth[intfIdx]; exists {
		for _, ip4 := range allIpv4Addr.Strings() {
			if _, err = strBuilder.WriteString("IPv4: " + ip4 + "\n"); err != nil {
				return err
			}
		}
	}

	if allIpv6Addr, exists := this.ipv6AddrByEth[intfIdx]; exists {
		for _, ip6 := range allIpv6Addr.Strings() {
			if _, err = strBuilder.WriteString("IPv6: " + ip6 + "\n"); err != nil {
				return err
			}
		}
	}

	if vid, exists := this.vlanAccessByEth[intfIdx]; exists {
		if _, err = strBuilder.WriteString(fmt.Sprintf("Access VLAN: %d\n", vid)); err != nil {
			return err
		}
	}

	if vid, exists := this.vlanNativeByEth[intfIdx]; exists {
		if _, err = strBuilder.WriteString(fmt.Sprintf("Native VLAN: %d\n", vid)); err != nil {
			return err
		}
	}

	if trunkVlans, exists := this.vlanTrunkByEth[intfIdx]; exists {
		if trunkVlans.Size() > 0 {
			vlans := trunkVlans.VidTs()
			if _, err = strBuilder.WriteString("Trunk VLANs:"); err != nil {
				return err
			}

			for _, vid := range vlans {
				if _, err = strBuilder.WriteString(fmt.Sprintf(" %d", vid)); err != nil {
					return err
				}
			}

			if _, err = strBuilder.WriteString("\n"); err != nil {
				return err
			}
		}
	}

	if lagIdx, exists := this.lagByEth[intfIdx]; exists {
		if _, err = strBuilder.WriteString(fmt.Sprintf("LAG: %s\n", this.lagNameByIdx[lagIdx])); err != nil {
			return err
		}
	}

	if strBuilder.Len() == 0 {
		return nil
	}

	return errors.New(strBuilder.String())
}

func (this *configLookupTablesT) checkLagDependenciesDuringAdd(ifname string, lagName string) error {
	return nil
}

func (table *configLookupTablesT) addNewInterfaceIfItDoesNotExist(ifname string) error {
	if strings.Contains(ifname, "ae") {
		if _, exists := table.idxByLagName[ifname]; !exists {
			table.idxByLagName[ifname] = table.idxOfLastAddedLag
			table.lagNameByIdx[table.idxOfLastAddedLag] = ifname
			table.idxOfLastAddedLag++
			log.Infof("Saved LAG %s", ifname)
		}
	} else if strings.Contains(ifname, "eth") {
		if _, exists := table.idxByEthName[ifname]; !exists {
			table.idxByEthName[ifname] = table.idxOfLastAddedIntf
			table.intfNameByIdx[table.idxOfLastAddedIntf] = ifname
			table.idxOfLastAddedIntf++
			log.Infof("Saved interface %s", ifname)
		}
	} else {
		err := fmt.Errorf("Unrecognized type of interface %s", ifname)
		return err
	}

	return nil
}

func (table *configLookupTablesT) setVlanModeEthIntf(ifname string, vlanMode oc.E_OpenconfigVlan_VlanModeType) error {
	intfIdx, exists := table.idxByEthName[ifname]
	if !exists {
		return fmt.Errorf("Ethernet interface %s does not exist", ifname)
	}
	table.vlanModeByEth[intfIdx] = vlanMode

	return nil
}

func (this *configLookupTablesT) deleteAccessVlanEthIntf(ifname string, vidDelete lib.VidT) error {
	intfIdx, exists := this.idxByEthName[ifname]
	if !exists {
		return fmt.Errorf("Ethernet interface %s does not exist", ifname)
	}

	vid, exists := this.vlanAccessByEth[intfIdx]
	if vid != vidDelete {
		return fmt.Errorf("There is configured other access VLAN %d", vid, ifname)
	}

	delete(this.vlanAccessByEth, intfIdx)
	this.ethByVlanAccess[vidDelete].Delete(intfIdx)
	log.Infof("Deleted access VLAN %d from Ethernet interface %s", vid, ifname)
	return nil
}

func (table *configLookupTablesT) setNativeVlanEthIntf(ifname string, vid lib.VidT) {
	// TODO: Add asserts for checking if interface exists in map
	table.vlanNativeByEth[table.idxByEthName[ifname]] = vid
	if _, exists := table.ethByVlanNative[vid]; !exists {
		table.ethByVlanNative[vid] = lib.NewIdxTSet()
	}

	table.ethByVlanNative[vid].Add(table.idxByEthName[ifname])
	log.Infof("Set native VLAN %d on interface %s", vid, ifname)
}

func (this *configLookupTablesT) deleteNativeVlanEthIntf(ifname string, vidDelete lib.VidT) error {
	intfIdx, exists := this.idxByEthName[ifname]
	if !exists {
		return fmt.Errorf("Ethernet interface %s does not exist", ifname)
	}

	vid, exists := this.vlanNativeByEth[intfIdx]
	if vid != vidDelete {
		return fmt.Errorf("There is configured other native VLAN %d", vid, ifname)
	}

	delete(this.vlanNativeByEth, intfIdx)
	this.ethByVlanNative[vidDelete].Delete(intfIdx)
	log.Infof("Deleted native VLAN %d from Ethernet interface %s", vid, ifname)
	return nil
}

func (table *configLookupTablesT) setTrunkVlanEthIntf(ifname string, vid lib.VidT) error {
	ethIdx, exists := table.idxByEthName[ifname]
	if !exists {
		return fmt.Errorf("Not found index of EThernet interface %s", ifname)
	}

	if _, exists := table.vlanTrunkByEth[table.idxByEthName[ifname]]; !exists {
		table.vlanTrunkByEth[ethIdx] = lib.NewVidTSet()
	}
	table.vlanTrunkByEth[ethIdx].Add(vid)

	if _, exists := table.ethByVlanTrunk[vid]; !exists {
		table.ethByVlanTrunk[vid] = lib.NewIdxTSet()
	}

	table.ethByVlanTrunk[vid].Add(ethIdx)
	log.Infof("Set trunk VLAN %d on Ethernet interface %s", vid, ifname)
	return nil
}

func (this *configLookupTablesT) deleteTrunkVlanEthIntf(ifname string, vidDelete lib.VidT) error {
	intfIdx, exists := this.idxByEthName[ifname]
	if !exists {
		return fmt.Errorf("Ethernet interface %s does not exist", ifname)
	}

	vlans, exists := this.vlanTrunkByEth[intfIdx]
	if !vlans.Has(vidDelete) {
		return fmt.Errorf("There is not configured trunk VLAN %d", vidDelete)
	}

	delete(this.vlanTrunkByEth, intfIdx)
	this.ethByVlanTrunk[vidDelete].Delete(intfIdx)
	log.Infof("Deleted trunk VLAN %d from Ethernet interface %s", vidDelete, ifname)
	return nil
}

func (table *configLookupTablesT) setAccessVlanEthIntf(ifname string, vid lib.VidT) {
	// TODO: Add asserts for checking if LAG exists in map
	table.vlanAccessByEth[table.idxByEthName[ifname]] = vid
	if _, exists := table.ethByVlanAccess[vid]; !exists {
		table.ethByVlanAccess[vid] = lib.NewIdxTSet()
	}

	table.ethByVlanAccess[vid].Add(table.idxByEthName[ifname])
	log.Infof("Set access VLAN %d on port %s", vid, ifname)
}

func (table *configLookupTablesT) setTrunkVlansOnPort(ifname string, vids []lib.VidT) {
	// TODO: Add asserts for checking if interface exists in map
	for _, vid := range vids {
		if _, exists := table.vlanTrunkByEth[table.idxByEthName[ifname]]; !exists {
			table.vlanTrunkByEth[table.idxByEthName[ifname]] = lib.NewVidTSet()
		}

		table.vlanTrunkByEth[table.idxByEthName[ifname]].Add(vid)
		if _, exists := table.ethByVlanTrunk[vid]; !exists {
			table.ethByVlanTrunk[vid] = lib.NewIdxTSet()
		}

		table.ethByVlanTrunk[vid].Add(table.idxByEthName[ifname])
		log.Infof("Set trunk VLAN %d on interface %s", vid, ifname)
	}
}

func (table *configLookupTablesT) SetVlanModeLagIntf(lagName string, vlanMode oc.E_OpenconfigVlan_VlanModeType) {
	// TODO: Add asserts for checking if LAG interface exists in map
	table.vlanModeByLag[table.idxByLagName[lagName]] = vlanMode
}

func (table *configLookupTablesT) getVlanModeEthIntf(ifname string) (oc.E_OpenconfigVlan_VlanModeType, error) {
	idxIntf, exists := table.idxByEthName[ifname]
	if !exists {
		return oc.OpenconfigVlan_VlanModeType_UNSET, fmt.Errorf("Ethernet interface %s does not exist", ifname)
	}

	vlanMode, exists := table.vlanModeByEth[idxIntf]
	if !exists {
		return oc.OpenconfigVlan_VlanModeType_UNSET, fmt.Errorf("There is not set VLAN mode on Ethernet interface %s", ifname)
	}

	return vlanMode, nil
}

func (table *configLookupTablesT) getVlanModeLagIntf(lagName string) (oc.E_OpenconfigVlan_VlanModeType, error) {
	idxLag, exists := table.idxByLagName[lagName]
	if !exists {
		return oc.OpenconfigVlan_VlanModeType_UNSET, fmt.Errorf("LAG interface %s does not exist", lagName)
	}

	vlanMode, exists := table.vlanModeByLag[idxLag]
	if !exists {
		return oc.OpenconfigVlan_VlanModeType_UNSET, fmt.Errorf("There is not set VLAN mode on LAG interface %s", lagName)
	}

	return vlanMode, nil
}

func (table *configLookupTablesT) setAccessVlanOnLag(lagName string, vid lib.VidT) {
	// TODO: Add asserts for checking if LAG exists in map
	table.vlanAccessByLag[table.idxByLagName[lagName]] = vid
	if _, exists := table.lagByVlanAccess[vid]; !exists {
		table.lagByVlanAccess[vid] = lib.NewIdxTSet()
	}

	table.lagByVlanAccess[vid].Add(table.idxByLagName[lagName])
	log.Infof("Set access VLAN %d on LAG %s", vid, lagName)
}

func (table *configLookupTablesT) setNativeVlanOnLag(lagName string, vid lib.VidT) {
	// TODO: Add asserts for checking if LAG exists in map
	table.vlanNativeByLag[table.idxByLagName[lagName]] = vid
	if _, exists := table.lagByVlanNative[vid]; !exists {
		table.lagByVlanNative[vid] = lib.NewIdxTSet()
	}

	table.lagByVlanNative[vid].Add(table.idxByLagName[lagName])
	log.Infof("Set native VLAN %d on LAG %s", vid, lagName)
}

func (table *configLookupTablesT) setTrunkVlansOnLag(lagName string, vids []lib.VidT) {
	// TODO: Add asserts for checking if LAG exists in map
	for _, vid := range vids {
		if _, exists := table.vlanTrunkByLag[table.idxByLagName[lagName]]; !exists {
			table.vlanTrunkByLag[table.idxByLagName[lagName]] = lib.NewVidTSet()
		}

		table.vlanTrunkByLag[table.idxByLagName[lagName]].Add(vid)
		if _, exists := table.lagByVlanTrunk[vid]; !exists {
			table.lagByVlanTrunk[vid] = lib.NewIdxTSet()
		}

		table.lagByVlanTrunk[vid].Add(table.idxByLagName[lagName])
		log.Infof("Set trunk VLAN %d on LAG %s", vid, lagName)
	}
}

func (t *configLookupTablesT) addIpv4AddrEthIntf(ifname string, ip string) error {
	intfIdx := t.idxByEthName[ifname]
	if _, exists := t.ethByIpv4Addr[ip]; exists {
		return fmt.Errorf("Failed to assign IPv4 address %s to interface %s because it is already in use",
			ip, ifname)
	}

	t.ethByIpv4Addr[ip] = intfIdx
	if _, exists := t.ipv4AddrByEth[t.idxByEthName[ifname]]; !exists {
		t.ipv4AddrByEth[t.idxByEthName[ifname]] = lib.NewStringSet()
	}
	// TODO: Check if IP is valid
	t.ipv4AddrByEth[t.idxByEthName[ifname]].Add(ip)
	log.Infof("Saved IPv4 %s for interface %s", ip, ifname)
	return nil
}

func (this *configLookupTablesT) deleteIpv4AddrEthIntf(ifname string, ip string) error {
	if _, exists := this.ethByIpv4Addr[ip]; !exists {
		return fmt.Errorf("Failed to delete IPv4 address %s from Ethernet interface %s because interface does not exist",
			ip, ifname)
	}

	delete(this.ethByIpv4Addr, ip)
	intfIdx := this.idxByEthName[ifname]
	this.ipv4AddrByEth[intfIdx].Delete(ip)
	log.Infof("Deleted IPv4 %s from Ethernet interface %s", ip, ifname)
	return nil
}

func (t *configLookupTablesT) saveIpv6AddrAddressForInterface(ifname string, ip string) error {
	intfIdx := t.idxByEthName[ifname]
	if _, exists := t.ethByIpv6Addr[ip]; exists {
		return fmt.Errorf("Failed to assign IPv6 address %s to interface %s because it is already in use",
			ip, ifname)
	}

	t.ethByIpv6Addr[ip] = intfIdx
	if _, exists := t.ipv6AddrByEth[intfIdx]; !exists {
		t.ipv6AddrByEth[intfIdx] = lib.NewStringSet()
	}
	// TODO: Check if IP is valid
	t.ipv6AddrByEth[intfIdx].Add(ip)
	log.Infof("Saved IPv6 %s for interface %s", ip, ifname)
	return nil
}

func (this *configLookupTablesT) deleteIpv6AddrEthIntf(ifname string, ip string) error {
	if _, exists := this.ethByIpv6Addr[ip]; !exists {
		return fmt.Errorf("Failed to delete IPv6 address %s from Ethernet interface %s because interface does not exist",
			ip, ifname)
	}

	delete(this.ethByIpv6Addr, ip)
	intfIdx := this.idxByEthName[ifname]
	this.ipv6AddrByEth[intfIdx].Delete(ip)
	log.Infof("Deleted IPv6 %s from Ethernet interface %s", ip, ifname)
	return nil
}

func (t *configLookupTablesT) saveIpv4AddrAddressForLag(lagName string, ip string) error {
	lagIdx := t.idxByLagName[lagName]
	if _, exists := t.lagByIpv4Addr[ip]; exists {
		return fmt.Errorf("Failed to assign IPv4 address %s to LAG %s because it is already in use",
			ip, lagName)
	}

	t.lagByIpv4Addr[ip] = lagIdx
	if _, exists := t.ipv4AddrByLag[lagIdx]; !exists {
		t.ipv4AddrByLag[lagIdx] = lib.NewStringSet()
	}
	// TODO: Check if IP is valid
	t.ipv4AddrByLag[lagIdx].Add(ip)
	log.Infof("Saved IPv4 %s for LAG %s", ip, lagName)
	return nil
}

func (t *configLookupTablesT) saveIpv6AddrAddressForLag(lagName string, ip string) error {
	lagIdx := t.idxByLagName[lagName]
	if _, exists := t.lagByIpv6Addr[ip]; exists {
		return fmt.Errorf("Failed to assign IPv6 address %s to interface %s because it is already in use",
			ip, lagName)
	}
	t.lagByIpv6Addr[ip] = lagIdx

	if _, exists := t.ipv6AddrByLag[lagIdx]; !exists {
		t.ipv6AddrByLag[lagIdx] = lib.NewStringSet()
	}
	// TODO: Check if IP is valid
	t.ipv6AddrByLag[lagIdx].Add(ip)
	log.Infof("Saved IPv6 %s for interface %s", ip, lagName)
	return nil
}

func (t *configLookupTablesT) parseInterfaceAsLagMember(ifname string, eth *oc.Interface_Ethernet) error {
	lagName := eth.GetAggregateId()
	if len(lagName) == 0 {
		return nil
	}
	lagIdx, exists := t.idxByLagName[lagName]
	if !exists {
		return fmt.Errorf("Invalid LAG %s on interface %s: LAG not exists", lagName, ifname)
	}

	intfIdx := t.idxByEthName[ifname]
	if lag, exists := t.lagByEth[intfIdx]; exists {
		if lag == lagIdx {
			return fmt.Errorf("Interface %s exists in another LAG %s", ifname, t.lagNameByIdx[lag])
		}
	}
	t.lagByEth[intfIdx] = lagIdx

	if _, exists = t.ethByLag[lagIdx]; !exists {
		t.ethByLag[lagIdx] = lib.NewIdxTSet()
	}
	t.ethByLag[lagIdx].Add(intfIdx)

	log.Infof("Added interface %s as member of LAG %s", ifname, lagName)
	return nil
}

func (t *configLookupTablesT) parseSubinterface(ifname string, subIntf *oc.Interface_Subinterface) error {
	ipv4 := subIntf.GetIpv4Addr()
	if ipv4 != nil {
		for _, addr := range ipv4.Address {
			ip := fmt.Sprintf("%s/%d", addr.GetIp(), addr.GetPrefixLength())
			if err := t.addIpv4AddrEthIntf(ifname, ip); err != nil {
				return err
			}
		}
	}

	ipv6 := subIntf.GetIpv6Addr()
	if ipv6 != nil {
		for _, addr := range ipv6.Address {
			ip := fmt.Sprintf("%s/%d", addr.GetIp(), addr.GetPrefixLength())
			if err := t.saveIpv6AddrAddressForInterface(ifname, ip); err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *configLookupTablesT) parseVlanForIntf(ifname string, swVlan *oc.Interface_Ethernet_SwitchedVlan) error {
	intfMode := swVlan.GetInterfaceMode()
	if intfMode == oc.OpenconfigVlan_VlanModeType_ACCESS {
		vid := lib.VidT(swVlan.GetAccessVlan())
		if vid != 0 {
			t.setAccessVlanEthIntf(ifname, vid)
			log.Infof("Set access VLAN %d for interface %s", vid, ifname)
		} else {
			return fmt.Errorf("Failed to parse VLAN on interface %s in access mode", ifname)
		}
	} else if intfMode == oc.OpenconfigVlan_VlanModeType_TRUNK {
		nativeVid := lib.VidT(swVlan.GetNativeVlan())
		if nativeVid != 0 {
			t.setNativeVlanEthIntf(ifname, nativeVid)
			log.Infof("Set native VLAN %d for interface %s", nativeVid, ifname)
		}

		trunkVlans := swVlan.GetTrunkVlans()
		if trunkVlans != nil {
			vlans := make([]lib.VidT, 0)
			for _, v := range trunkVlans {
				switch t := v.(type) {
				case *oc.Interface_Ethernet_SwitchedVlan_TrunkVlans_Union_String:
					var lower, upper lib.VidT
					n, err := fmt.Sscanf(t.String, "%d..%d", &lower, &upper)
					if n != 2 || err != nil {
						return fmt.Errorf("Failed to parse lower and upper bound of trunk VLAN rane: %s", err)
					}

					if lower >= maxVlansC || upper >= maxVlansC {
						return fmt.Errorf("Out of range lowwer and upper bound of trunk VLANs (%d, %d)", lower, upper)
					}

					for ; lower <= upper; lower++ {
						vlans = append(vlans, lower)
					}
				case *oc.Interface_Ethernet_SwitchedVlan_TrunkVlans_Union_Uint16:
					vlans = append(vlans, lib.VidT(t.Uint16))
				default:
					return fmt.Errorf("Cannot convert %v to Interface_Ethernet_SwitchedVlan_TrunkVlans_Union, unknown union type, got: %T, want any of [string, uint16]", v, v)
				}
			}

			t.setTrunkVlansOnPort(ifname, vlans)
		}

		if nativeVid == 0 && trunkVlans == nil {
			return fmt.Errorf("Failed to parse VLANs on interface %s in trunk mode", ifname)
		}
	} else {
		intfMode = oc.OpenconfigVlan_VlanModeType_UNSET
	}

	t.setVlanModeEthIntf(ifname, intfMode)

	return nil
}

func (t *configLookupTablesT) parseVlanForLagIntf(lagName string, swVlan *oc.Interface_Aggregation_SwitchedVlan) error {
	intfMode := swVlan.GetInterfaceMode()
	if intfMode == oc.OpenconfigVlan_VlanModeType_ACCESS {
		vid := lib.VidT(swVlan.GetAccessVlan())
		if vid != 0 {
			t.setAccessVlanOnLag(lagName, vid)
			log.Infof("Set access VLAN %d for LAG %s", vid, lagName)
		} else {
			return fmt.Errorf("Failed to parse VLAN on interface %s in access mode", lagName)
		}
	} else if intfMode == oc.OpenconfigVlan_VlanModeType_TRUNK {
		nativeVid := lib.VidT(swVlan.GetNativeVlan())
		if nativeVid != 0 {
			t.setNativeVlanOnLag(lagName, nativeVid)
			log.Infof("Set native VLAN %d for LAG %s", nativeVid, lagName)
		}

		trunkVlans := swVlan.GetTrunkVlans()
		if trunkVlans != nil {
			vlans := make([]lib.VidT, 0)
			for _, v := range trunkVlans {
				switch t := v.(type) {
				case *oc.Interface_Aggregation_SwitchedVlan_TrunkVlans_Union_String:
					var lower, upper lib.VidT
					n, err := fmt.Sscanf(t.String, "%d..%d", &lower, &upper)
					if n != 2 || err != nil {
						return fmt.Errorf("Failed to parse lower and upper bound of trunk VLAN rane: %s", err)
					}

					if lower >= maxVlansC || upper >= maxVlansC {
						return fmt.Errorf("Out of range lowwer and upper bound of trunk VLANs (%d, %d)", lower, upper)
					}

					for ; lower <= upper; lower++ {
						vlans = append(vlans, lower)
					}
				case *oc.Interface_Aggregation_SwitchedVlan_TrunkVlans_Union_Uint16:
					vlans = append(vlans, lib.VidT(t.Uint16))
				default:
					return fmt.Errorf("Cannot convert %v to Interface_Aggregation_SwitchedVlan_TrunkVlans_Union, unknown union type, got: %T, want any of [string, uint16]", v, v)
				}
			}

			t.setTrunkVlansOnLag(lagName, vlans)
		}

		if nativeVid == 0 && trunkVlans == nil {
			return fmt.Errorf("Failed to parse VLANs on interface %s in trunk mode", lagName)
		}
	} else {
		intfMode = oc.OpenconfigVlan_VlanModeType_UNSET
	}

	t.SetVlanModeLagIntf(lagName, intfMode)

	return nil
}

func (this *configLookupTablesT) makeCopy() *configLookupTablesT {
	copy := newConfigLookupTables()

	copy.idxOfLastAddedIntf = this.idxOfLastAddedIntf
	copy.idxByEthName = make(map[string]lib.IdxT, maxPortsC)
	for k, v := range this.idxByEthName {
		copy.idxByEthName[k] = v
	}
	copy.intfNameByIdx = make(map[lib.IdxT]string, maxPortsC)
	for k, v := range this.intfNameByIdx {
		copy.intfNameByIdx[k] = v
	}

	copy.idxOfLastAddedLag = this.idxOfLastAddedLag
	copy.idxByLagName = make(map[string]lib.IdxT, len(this.idxByLagName))
	for k, v := range this.idxByLagName {
		copy.idxByLagName[k] = v
	}
	copy.lagNameByIdx = make(map[lib.IdxT]string, len(this.lagNameByIdx))
	for k, v := range this.lagNameByIdx {
		copy.lagNameByIdx[k] = v
	}

	copy.idxOfLastAddedVlan = this.idxOfLastAddedVlan
	copy.idxByVlanName = make(map[string]lib.IdxT, len(this.idxByVlanName))
	for k, v := range this.idxByVlanName {
		copy.idxByVlanName[k] = v
	}
	copy.vlanNameByIdx = make(map[lib.IdxT]string, len(this.vlanNameByIdx))
	for k, v := range this.vlanNameByIdx {
		copy.vlanNameByIdx[k] = v
	}

	copy.ipv4AddrByEth = make(map[lib.IdxT]*lib.StringSet, len(this.ipv4AddrByEth))
	for k, v := range this.ipv4AddrByEth {
		copy.ipv4AddrByEth[k] = v.MakeCopy()
	}
	copy.ipv4AddrByLag = make(map[lib.IdxT]*lib.StringSet, len(this.ipv4AddrByLag))
	for k, v := range this.ipv4AddrByLag {
		copy.ipv4AddrByLag[k] = v.MakeCopy()
	}
	copy.ipv4AddrByVlan = make(map[lib.VidT]*lib.StringSet, len(this.ipv4AddrByVlan))
	for k, v := range this.ipv4AddrByVlan {
		copy.ipv4AddrByVlan[k] = v.MakeCopy()
	}
	copy.ethByIpv4Addr = make(map[string]lib.IdxT, len(this.ethByIpv4Addr))
	for k, v := range this.ethByIpv4Addr {
		copy.ethByIpv4Addr[k] = v
	}
	copy.lagByIpv4Addr = make(map[string]lib.IdxT, len(this.lagByIpv4Addr))
	for k, v := range this.lagByIpv4Addr {
		copy.lagByIpv4Addr[k] = v
	}
	copy.vlanByIpv4Addr = make(map[string]lib.VidT, len(this.vlanByIpv4Addr))
	for k, v := range this.vlanByIpv4Addr {
		copy.vlanByIpv4Addr[k] = v
	}

	copy.ipv6AddrByEth = make(map[lib.IdxT]*lib.StringSet, len(this.ipv6AddrByEth))
	for k, v := range this.ipv6AddrByEth {
		copy.ipv6AddrByEth[k] = v.MakeCopy()
	}
	copy.ipv6AddrByLag = make(map[lib.IdxT]*lib.StringSet, len(this.ipv6AddrByLag))
	for k, v := range this.ipv6AddrByLag {
		copy.ipv6AddrByLag[k] = v.MakeCopy()
	}
	copy.ipv6AddrByVlan = make(map[lib.VidT]*lib.StringSet, len(this.ipv6AddrByVlan))
	for k, v := range this.ipv6AddrByVlan {
		copy.ipv6AddrByVlan[k] = v.MakeCopy()
	}
	copy.ethByIpv6Addr = make(map[string]lib.IdxT, len(this.ethByIpv6Addr))
	for k, v := range this.ethByIpv6Addr {
		copy.ethByIpv6Addr[k] = v
	}
	copy.lagByIpv6Addr = make(map[string]lib.IdxT, len(this.lagByIpv6Addr))
	for k, v := range this.lagByIpv6Addr {
		copy.lagByIpv6Addr[k] = v
	}
	copy.vlanByIpv6Addr = make(map[string]lib.VidT, len(this.vlanByIpv6Addr))
	for k, v := range this.vlanByIpv6Addr {
		copy.vlanByIpv6Addr[k] = v
	}

	copy.lagByEth = make(map[lib.IdxT]lib.IdxT, len(this.lagByEth))
	for k, v := range this.lagByEth {
		copy.lagByEth[k] = v
	}
	copy.ethByLag = make(map[lib.IdxT]*lib.IdxTSet, len(this.ethByLag))
	for k, v := range this.ethByLag {
		copy.ethByLag[k] = v.MakeCopy()
	}

	copy.stpByEth = this.stpByEth.MakeCopy()

	copy.vlanModeByEth = make(map[lib.IdxT]oc.E_OpenconfigVlan_VlanModeType, len(this.vlanModeByEth))
	for k, v := range this.vlanModeByEth {
		copy.vlanModeByEth[k] = v
	}
	copy.vlanModeByLag = make(map[lib.IdxT]oc.E_OpenconfigVlan_VlanModeType, len(this.vlanModeByEth))
	for k, v := range this.vlanModeByLag {
		copy.vlanModeByLag[k] = v
	}

	copy.vlanAccessByEth = make(map[lib.IdxT]lib.VidT, len(this.vlanAccessByEth))
	for k, v := range this.vlanAccessByEth {
		copy.vlanAccessByEth[k] = v
	}
	copy.vlanAccessByLag = make(map[lib.IdxT]lib.VidT, len(this.vlanAccessByLag))
	for k, v := range this.vlanAccessByLag {
		copy.vlanAccessByLag[k] = v
	}
	copy.ethByVlanAccess = make(map[lib.VidT]*lib.IdxTSet, len(this.ethByVlanAccess))
	for k, v := range this.ethByVlanAccess {
		copy.ethByVlanAccess[k] = v.MakeCopy()
	}
	copy.lagByVlanAccess = make(map[lib.VidT]*lib.IdxTSet, len(this.lagByVlanAccess))
	for k, v := range this.lagByVlanAccess {
		copy.lagByVlanAccess[k] = v.MakeCopy()
	}

	copy.vlanNativeByEth = make(map[lib.IdxT]lib.VidT, len(this.vlanNativeByEth))
	for k, v := range this.vlanNativeByEth {
		copy.vlanNativeByEth[k] = v
	}
	copy.vlanNativeByLag = make(map[lib.IdxT]lib.VidT, len(this.vlanNativeByLag))
	for k, v := range this.vlanNativeByLag {
		copy.vlanNativeByLag[k] = v
	}
	copy.ethByVlanNative = make(map[lib.VidT]*lib.IdxTSet, len(this.ethByVlanNative))
	for k, v := range this.ethByVlanNative {
		copy.ethByVlanNative[k] = v.MakeCopy()
	}
	copy.lagByVlanNative = make(map[lib.VidT]*lib.IdxTSet, len(this.lagByVlanNative))
	for k, v := range this.lagByVlanNative {
		copy.lagByVlanNative[k] = v.MakeCopy()
	}

	copy.vlanTrunkByEth = make(map[lib.IdxT]*lib.VidTSet, len(this.vlanTrunkByEth))
	for k, v := range this.vlanTrunkByEth {
		copy.vlanTrunkByEth[k] = v.MakeCopy()
	}
	copy.vlanTrunkByLag = make(map[lib.IdxT]*lib.VidTSet, len(this.vlanTrunkByLag))
	for k, v := range this.vlanTrunkByLag {
		copy.vlanTrunkByLag[k] = v.MakeCopy()
	}
	copy.ethByVlanTrunk = make(map[lib.VidT]*lib.IdxTSet, len(this.ethByVlanTrunk))
	for k, v := range this.ethByVlanTrunk {
		copy.ethByVlanTrunk[k] = v.MakeCopy()
	}
	copy.lagByVlanTrunk = make(map[lib.VidT]*lib.IdxTSet, len(this.lagByVlanTrunk))
	for k, v := range this.lagByVlanTrunk {
		copy.lagByVlanTrunk[k] = v.MakeCopy()
	}

	return copy
}

func (t *configLookupTablesT) dump() {
	log.Infoln("Dump internal state of config lookup tables")
	log.Infoln("========================================")
	intfs := make([]string, 0)
	for ifname, _ := range t.idxByEthName {
		intfs = append(intfs, ifname)
	}
	sort.Strings(intfs)
	log.Infof("There are %d LAG interfaces", len(intfs))
	log.Infoln("Print list of interfaces:")
	for _, ifname := range intfs {
		log.Infoln(ifname)
	}

	log.Infoln("========================================")
	lags := make([]string, 0)
	for lagName, _ := range t.idxByLagName {
		lags = append(lags, lagName)
	}
	sort.Strings(lags)
	log.Infof("There are %d LAG interfaces", len(lags))
	log.Infoln("Print list of LAG interfaces and their members:")
	for _, lagName := range lags {
		log.Infof("%s", lagName)
		for _, ethIntf := range t.ethByLag[t.idxByLagName[lagName]].IdxTs() {
			log.Infof("  %s", t.intfNameByIdx[ethIntf])
		}
	}

	log.Infoln("========================================")
	log.Infoln("Print VLAN mode of Ethernet interface:")
	for _, ifname := range intfs {
		if vlanMode, exists := t.vlanModeByEth[t.idxByEthName[ifname]]; exists {
			log.Infof("VLAN mode on interface %s: %d", ifname, vlanMode)
		} else {
			log.Infof("There isn't set VLAN mode on interface %s", ifname)
			log.Infoln("----------------------------------------")
			continue
		}
	}

	log.Infoln("Print membership of interfaces in access VLAN:")
	for _, ifname := range intfs {
		if accessVid, exists := t.vlanAccessByEth[t.idxByEthName[ifname]]; exists {
			log.Infof("Access VLAN on interface %s: %d", ifname, accessVid)
		} else {
			log.Infof("There isn't access VLAN on interface %s", ifname)
			log.Infoln("----------------------------------------")
			continue
		}
	}

	log.Infoln("========================================")
	log.Infoln("Print membership of interfaces in trunk VLANs:")
	for _, ifname := range intfs {
		idx := t.idxByEthName[ifname]
		vids, exists := t.vlanTrunkByEth[idx]
		if !exists {
			log.Infof("There aren't any trunk VLANs on interface %s", ifname)
			log.Infoln("----------------------------------------")
			continue
		}

		vlans := make([]int, 0)
		for _, vid := range vids.VidTs() {
			vlans = append(vlans, int(vid))
		}
		sort.Ints(vlans)
		log.Infof("There are %d trunk VLANs on interface %s:", len(vlans), ifname)
		for _, vid := range vlans {
			log.Infoln(vid)
		}

		if nativeVid, exists := t.vlanNativeByEth[idx]; exists {
			log.Infof("Native VLAN on interface %s: %d", ifname, nativeVid)
		}
	}

	log.Infoln("========================================")
	log.Infoln("Print VLAN mode of LAG interface:")
	for _, lagName := range lags {
		if vlanMode, exists := t.vlanModeByLag[t.idxByLagName[lagName]]; exists {
			log.Infof("VLAN mode on LAG interface %s: %d", lagName, vlanMode)
		} else {
			log.Infof("There isn't set VLAN mode on LAG interface %s", lagName)
			log.Infoln("----------------------------------------")
			continue
		}
	}

	log.Infoln("========================================")
	log.Infoln("Print membership of LAG in access VLAN:")
	for _, lagName := range lags {
		if accessVid, exists := t.vlanAccessByLag[t.idxByLagName[lagName]]; exists {
			log.Infof("Access VLAN on LAG %s: %d", lagName, accessVid)
		} else {
			log.Infof("There isn't access VLAN on LAG %s", lagName)
			log.Infoln("----------------------------------------")
			continue
		}
	}

	log.Infoln("========================================")
	log.Infoln("Print membership of LAG interfaces in trunk VLANs:")
	for _, lagName := range lags {
		idx := t.idxByLagName[lagName]
		vids, exists := t.vlanTrunkByLag[idx]
		if !exists {
			log.Infof("There aren't any trunk VLANs on LAG %s", lagName)
			log.Infoln("----------------------------------------")
			continue
		}

		vlans := make([]int, 0)
		for _, vid := range vids.VidTs() {
			vlans = append(vlans, int(vid))
		}
		sort.Ints(vlans)
		log.Infof("There are %d VLANs on LAG %s:", len(vlans), lagName)
		for _, vid := range vlans {
			log.Infoln(vid)
		}

		if nativeVid, exists := t.vlanNativeByLag[idx]; exists {
			log.Infof("Native VLAN on LAG %s: %d", lagName, nativeVid)
		}
	}

	log.Infoln("========================================")
	log.Infoln("Print IPv4 addresses on interfaces:")
	for _, ifname := range intfs {
		idx := t.idxByEthName[ifname]
		ipAddrs, exists := t.ipv4AddrByEth[idx]
		if !exists {
			log.Infof("There aren't any IPv4 addresses on interface %s", ifname)
			log.Infoln("----------------------------------------")
			continue
		}

		addrs := make([]string, 0)
		for _, ip := range ipAddrs.Strings() {
			addrs = append(addrs, ip)
		}
		sort.Strings(addrs)
		log.Infof("There are %d IPv4 addresses on interface %s:", len(addrs), ifname)
		for _, ip := range addrs {
			log.Infoln(ip)
		}
	}

	log.Infoln("========================================")
	log.Infoln("Print IPv6 addresses on interfaces:")
	for _, ifname := range intfs {
		idx := t.idxByEthName[ifname]
		ipAddrs, exists := t.ipv6AddrByEth[idx]
		if !exists {
			log.Infof("There aren't any IPv6 addresses on interface %s", ifname)
			log.Infoln("----------------------------------------")
			continue
		}

		addrs := make([]string, 0)
		for _, ip := range ipAddrs.Strings() {
			addrs = append(addrs, ip)
		}
		sort.Strings(addrs)
		log.Infof("There are %d IPv6 addresses on interface %s:", len(addrs), ifname)
		for _, ip := range addrs {
			log.Infoln(ip)
		}
	}
}
