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
	idxByIntfName      map[string]lib.IdxT
	intfNameByIdx      map[lib.IdxT]string
	idxOfLastAddedLag  lib.IdxT
	idxByLagName       map[string]lib.IdxT
	lagNameByIdx       map[lib.IdxT]string
	idxOfLastAddedVlan lib.IdxT
	idxByVlanName      map[string]lib.IdxT
	vlanNameByIdx      map[lib.IdxT]string

	// L3 interface can have assigned many IPv4 addresses
	ipv4AddrByIntf map[lib.IdxT]*lib.StringSet
	// L3 LAG can have assigned many IPv4 addresses
	ipv4AddrByLag map[lib.IdxT]*lib.StringSet
	// L3 VLAN can have assigned many IPv4 addresses
	ipv4AddrByVlan map[lib.VidT]*lib.StringSet
	intfByIpv4Addr map[string]lib.IdxT
	lagByIpv4Addr  map[string]lib.IdxT
	vlanByIpv4Addr map[string]lib.VidT
	ipv6AddrByIntf map[lib.IdxT]*lib.StringSet
	ipv6AddrByLag  map[lib.IdxT]*lib.StringSet
	ipv6AddrByVlan map[lib.VidT]*lib.StringSet
	// L3 interface can have assigned many IPv6 addresses
	intfByIpv6Addr map[string]lib.IdxT
	// L3 LAG can have assigned many IPv6 addresses
	lagByIpv6Addr map[string]lib.IdxT
	// L3 VLAN interface can have assigned many IPv6 addresses
	vlanByIpv6Addr map[string]lib.VidT
	lagByIntf      map[lib.IdxT]lib.IdxT
	// LAG can have many interface members
	intfByLag        map[lib.IdxT]*lib.IdxTSet
	stpByIntf        *lib.IdxTSet
	vlanAccessByIntf map[lib.IdxT]lib.VidT
	// There can be many ports in specific VLAN ID for access mode
	intfByVlanAccess map[lib.VidT]*lib.IdxTSet
	vlanAccessByLag  map[lib.IdxT]lib.VidT
	// There can be many LAGs in specific VLAN ID for access mode
	lagByVlanAccess  map[lib.VidT]*lib.IdxTSet
	vlanNativeByIntf map[lib.IdxT]lib.VidT
	// There can be many ports in specific VLAN ID for native tag
	intfByVlanNative map[lib.VidT]*lib.IdxTSet
	vlanNativeByLag  map[lib.IdxT]lib.VidT
	// There can be many LAGs in specific VLAN ID for native tag
	lagByVlanNative map[lib.VidT]*lib.IdxTSet
	vlanTrunkByIntf map[lib.IdxT]*lib.VidTSet
	// There can be many ports in VLAN trunk
	intfByVlanTrunk map[lib.VidT]*lib.IdxTSet
	vlanTrunkByLag  map[lib.IdxT]*lib.VidTSet
	// There can be many LAGs in VLAN trunk
	lagByVlanTrunk map[lib.VidT]*lib.IdxTSet
}

func newConfigLookupTables() *configLookupTablesT {
	return &configLookupTablesT{
		idxOfLastAddedIntf: 0,
		idxByIntfName:      make(map[string]lib.IdxT, maxPortsC),
		intfNameByIdx:      make(map[lib.IdxT]string, maxPortsC),
		idxOfLastAddedLag:  0,
		idxByLagName:       make(map[string]lib.IdxT),
		lagNameByIdx:       make(map[lib.IdxT]string),
		idxOfLastAddedVlan: 0,
		idxByVlanName:      make(map[string]lib.IdxT),
		vlanNameByIdx:      make(map[lib.IdxT]string),
		ipv4AddrByIntf:     make(map[lib.IdxT]*lib.StringSet),
		ipv4AddrByLag:      make(map[lib.IdxT]*lib.StringSet),
		ipv4AddrByVlan:     make(map[lib.VidT]*lib.StringSet),
		intfByIpv4Addr:     make(map[string]lib.IdxT),
		lagByIpv4Addr:      make(map[string]lib.IdxT),
		vlanByIpv4Addr:     make(map[string]lib.VidT),
		ipv6AddrByIntf:     make(map[lib.IdxT]*lib.StringSet),
		ipv6AddrByLag:      make(map[lib.IdxT]*lib.StringSet),
		ipv6AddrByVlan:     make(map[lib.VidT]*lib.StringSet),
		intfByIpv6Addr:     make(map[string]lib.IdxT),
		lagByIpv6Addr:      make(map[string]lib.IdxT),
		vlanByIpv6Addr:     make(map[string]lib.VidT),
		lagByIntf:          make(map[lib.IdxT]lib.IdxT),
		intfByLag:          make(map[lib.IdxT]*lib.IdxTSet),
		stpByIntf:          lib.NewIdxTSet(),
		vlanAccessByIntf:   make(map[lib.IdxT]lib.VidT),
		vlanAccessByLag:    make(map[lib.IdxT]lib.VidT),
		intfByVlanAccess:   make(map[lib.VidT]*lib.IdxTSet),
		lagByVlanAccess:    make(map[lib.VidT]*lib.IdxTSet),
		vlanNativeByIntf:   make(map[lib.IdxT]lib.VidT),
		vlanNativeByLag:    make(map[lib.IdxT]lib.VidT),
		intfByVlanNative:   make(map[lib.VidT]*lib.IdxTSet),
		lagByVlanNative:    make(map[lib.VidT]*lib.IdxTSet),
		vlanTrunkByIntf:    make(map[lib.IdxT]*lib.VidTSet),
		vlanTrunkByLag:     make(map[lib.IdxT]*lib.VidTSet),
		intfByVlanTrunk:    make(map[lib.VidT]*lib.IdxTSet),
		lagByVlanTrunk:     make(map[lib.VidT]*lib.IdxTSet),
	}
}

func (this *configLookupTablesT) checkDependenciesForDeleteOrRemoveEthIntfFromLagIntf(ifname string, lagName string) error {
	intfIdx, exists := this.idxByIntfName[ifname]
	if !exists {
		return fmt.Errorf("Ethernet interface %s does not exists", ifname)
	}

	expectedLagIdx, exists := this.idxByLagName[lagName]
	if !exists {
		return fmt.Errorf("LAG %s does not exists", lagName)
	}

	if this.lagByIntf[intfIdx] != expectedLagIdx {
		return fmt.Errorf("Ethernet interface %s does not exists in LAG %s", ifname, lagName)
	}

	return nil
}

func (this *configLookupTablesT) checkDependenciesForDeleteOrRemoveLagIntf(lagName string) error {
	var err error
	strBuilder := strings.Builder{}
	lagIdx, exists := this.idxByLagName[lagName]
	if !exists {
		return fmt.Errorf("LAG %s does not exists", lagName)
	}

	if intfs, exists := this.intfByLag[lagIdx]; exists {
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
	intfIdx, exists := this.intfByIpv4Addr[cidr4]
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
	intfIdx := this.idxByIntfName[ifname]
	allIpv4Addr, exists := this.ipv4AddrByIntf[intfIdx]
	if !exists {
		if _, err = strBuilder.WriteString(fmt.Sprintf("There is not any IPv4 address on Ethernet interface %s", ifname)); err != nil {
			return err
		}
	} else if !allIpv4Addr.Has(cidr4) {
		if _, err = strBuilder.WriteString(fmt.Sprintf("There is not IPv4 address %s on Ethernet interface %s", cidr4, ifname)); err != nil {
			return err
		}
	}

	foundIpIntfIdx, exists := this.intfByIpv4Addr[cidr4]
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

func (this *configLookupTablesT) checkDependenciesForDeletePortBreakout(ifname string) error {
	var err error
	strBuilder := strings.Builder{}
	intfIdx := this.idxByIntfName[ifname]
	if allIpv4Addr, exists := this.ipv4AddrByIntf[intfIdx]; exists {
		for _, ip4 := range allIpv4Addr.Strings() {
			if _, err = strBuilder.WriteString("IPv4: " + ip4 + "\n"); err != nil {
				return err
			}
		}
	}

	if allIpv6Addr, exists := this.ipv6AddrByIntf[intfIdx]; exists {
		for _, ip6 := range allIpv6Addr.Strings() {
			if _, err = strBuilder.WriteString("IPv6: " + ip6 + "\n"); err != nil {
				return err
			}
		}
	}

	if vid, exists := this.vlanAccessByIntf[intfIdx]; exists {
		if _, err = strBuilder.WriteString(fmt.Sprintf("Access VLAN: %d\n", vid)); err != nil {
			return err
		}
	}

	if vid, exists := this.vlanNativeByIntf[intfIdx]; exists {
		if _, err = strBuilder.WriteString(fmt.Sprintf("Native VLAN: %d\n", vid)); err != nil {
			return err
		}
	}

	if trunkVlans, exists := this.vlanTrunkByIntf[intfIdx]; exists {
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

	if lagIdx, exists := this.lagByIntf[intfIdx]; exists {
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
		if _, exists := table.idxByIntfName[ifname]; !exists {
			table.idxByIntfName[ifname] = table.idxOfLastAddedIntf
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

func (table *configLookupTablesT) setNativeVlanOnPort(ifname string, vid lib.VidT) {
	// TODO: Add asserts for checking if interface exists in map
	table.vlanNativeByIntf[table.idxByIntfName[ifname]] = vid
	if _, exists := table.intfByVlanNative[vid]; !exists {
		table.intfByVlanNative[vid] = lib.NewIdxTSet()
	}

	table.intfByVlanNative[vid].Add(table.idxByIntfName[ifname])
	log.Infof("Set native VLAN %d on interface %s", vid, ifname)
}

func (table *configLookupTablesT) setAccessVlanOnPort(ifname string, vid lib.VidT) {
	// TODO: Add asserts for checking if LAG exists in map
	table.vlanAccessByIntf[table.idxByIntfName[ifname]] = vid
	if _, exists := table.intfByVlanAccess[vid]; !exists {
		table.intfByVlanAccess[vid] = lib.NewIdxTSet()
	}

	table.intfByVlanAccess[vid].Add(table.idxByIntfName[ifname])
	log.Infof("Set access VLAN %d on port %s", vid, ifname)
}

func (table *configLookupTablesT) setTrunkVlansOnPort(ifname string, vids []lib.VidT) {
	// TODO: Add asserts for checking if interface exists in map
	for _, vid := range vids {
		if _, exists := table.vlanTrunkByIntf[table.idxByIntfName[ifname]]; !exists {
			table.vlanTrunkByIntf[table.idxByIntfName[ifname]] = lib.NewVidTSet()
		}

		table.vlanTrunkByIntf[table.idxByIntfName[ifname]].Add(vid)
		if _, exists := table.intfByVlanTrunk[vid]; !exists {
			table.intfByVlanTrunk[vid] = lib.NewIdxTSet()
		}

		table.intfByVlanTrunk[vid].Add(table.idxByIntfName[ifname])
		log.Infof("Set trunk VLAN %d on interface %s", vid, ifname)
	}
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
	intfIdx := t.idxByIntfName[ifname]
	if _, exists := t.intfByIpv4Addr[ip]; exists {
		return fmt.Errorf("Failed to assign IPv4 address %s to interface %s because it is already in use",
			ip, ifname)
	}

	t.intfByIpv4Addr[ip] = intfIdx
	if _, exists := t.ipv4AddrByIntf[t.idxByIntfName[ifname]]; !exists {
		t.ipv4AddrByIntf[t.idxByIntfName[ifname]] = lib.NewStringSet()
	}
	// TODO: Check if IP is valid
	t.ipv4AddrByIntf[t.idxByIntfName[ifname]].Add(ip)
	log.Infof("Saved IPv4 %s for interface %s", ip, ifname)
	return nil
}

func (this *configLookupTablesT) deleteIpv4AddrEthIntf(ifname string, ip string) error {
	if _, exists := this.intfByIpv4Addr[ip]; !exists {
		return fmt.Errorf("Failed to delete IPv4 address %s from Ethernet interface %s because interface does not exist",
			ip, ifname)
	}

	delete(this.intfByIpv4Addr, ip)
	intfIdx := this.idxByIntfName[ifname]
	this.ipv4AddrByIntf[intfIdx].Delete(ip)
	log.Infof("Deleted IPv4 %s from Ethernet interface %s", ip, ifname)
	return nil
}

func (t *configLookupTablesT) saveIpv6AddrAddressForInterface(ifname string, ip string) error {
	intfIdx := t.idxByIntfName[ifname]
	if _, exists := t.intfByIpv6Addr[ip]; exists {
		return fmt.Errorf("Failed to assign IPv6 address %s to interface %s because it is already in use",
			ip, ifname)
	}

	t.intfByIpv6Addr[ip] = intfIdx
	if _, exists := t.ipv6AddrByIntf[intfIdx]; !exists {
		t.ipv6AddrByIntf[intfIdx] = lib.NewStringSet()
	}
	// TODO: Check if IP is valid
	t.ipv6AddrByIntf[intfIdx].Add(ip)
	log.Infof("Saved IPv6 %s for interface %s", ip, ifname)
	return nil
}

func (this *configLookupTablesT) deleteIpv6AddrEthIntf(ifname string, ip string) error {
	if _, exists := this.intfByIpv6Addr[ip]; !exists {
		return fmt.Errorf("Failed to delete IPv6 address %s from Ethernet interface %s because interface does not exist",
			ip, ifname)
	}

	delete(this.intfByIpv6Addr, ip)
	intfIdx := this.idxByIntfName[ifname]
	this.ipv6AddrByIntf[intfIdx].Delete(ip)
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

	intfIdx := t.idxByIntfName[ifname]
	if lag, exists := t.lagByIntf[intfIdx]; exists {
		if lag == lagIdx {
			return fmt.Errorf("Interface %s exists in another LAG %s", ifname, t.lagNameByIdx[lag])
		}
	}
	t.lagByIntf[intfIdx] = lagIdx

	if _, exists = t.intfByLag[lagIdx]; !exists {
		t.intfByLag[lagIdx] = lib.NewIdxTSet()
	}
	t.intfByLag[lagIdx].Add(intfIdx)

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
			t.setAccessVlanOnPort(ifname, vid)
			log.Infof("Set access VLAN %d for interface %s", vid, ifname)
		} else {
			return fmt.Errorf("Failed to parse VLAN on interface %s in access mode", ifname)
		}
	} else if intfMode == oc.OpenconfigVlan_VlanModeType_TRUNK {
		nativeVid := lib.VidT(swVlan.GetNativeVlan())
		if nativeVid != 0 {
			t.setNativeVlanOnPort(ifname, nativeVid)
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
	}

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
			log.Infof("Set native VLAN %d for LAG %s", lagName, nativeVid)
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
	}

	return nil
}

func (this *configLookupTablesT) makeCopy() *configLookupTablesT {
	copy := newConfigLookupTables()

	copy.idxOfLastAddedIntf = this.idxOfLastAddedIntf
	copy.idxByIntfName = make(map[string]lib.IdxT, maxPortsC)
	for k, v := range this.idxByIntfName {
		copy.idxByIntfName[k] = v
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

	copy.ipv4AddrByIntf = make(map[lib.IdxT]*lib.StringSet, len(this.ipv4AddrByIntf))
	for k, v := range this.ipv4AddrByIntf {
		copy.ipv4AddrByIntf[k] = v.MakeCopy()
	}
	copy.ipv4AddrByLag = make(map[lib.IdxT]*lib.StringSet, len(this.ipv4AddrByLag))
	for k, v := range this.ipv4AddrByLag {
		copy.ipv4AddrByLag[k] = v.MakeCopy()
	}
	copy.ipv4AddrByVlan = make(map[lib.VidT]*lib.StringSet, len(this.ipv4AddrByVlan))
	for k, v := range this.ipv4AddrByVlan {
		copy.ipv4AddrByVlan[k] = v.MakeCopy()
	}
	copy.intfByIpv4Addr = make(map[string]lib.IdxT, len(this.intfByIpv4Addr))
	for k, v := range this.intfByIpv4Addr {
		copy.intfByIpv4Addr[k] = v
	}
	copy.lagByIpv4Addr = make(map[string]lib.IdxT, len(this.lagByIpv4Addr))
	for k, v := range this.lagByIpv4Addr {
		copy.lagByIpv4Addr[k] = v
	}
	copy.vlanByIpv4Addr = make(map[string]lib.VidT, len(this.vlanByIpv4Addr))
	for k, v := range this.vlanByIpv4Addr {
		copy.vlanByIpv4Addr[k] = v
	}

	copy.ipv6AddrByIntf = make(map[lib.IdxT]*lib.StringSet, len(this.ipv6AddrByIntf))
	for k, v := range this.ipv6AddrByIntf {
		copy.ipv6AddrByIntf[k] = v.MakeCopy()
	}
	copy.ipv6AddrByLag = make(map[lib.IdxT]*lib.StringSet, len(this.ipv6AddrByLag))
	for k, v := range this.ipv6AddrByLag {
		copy.ipv6AddrByLag[k] = v.MakeCopy()
	}
	copy.ipv6AddrByVlan = make(map[lib.VidT]*lib.StringSet, len(this.ipv6AddrByVlan))
	for k, v := range this.ipv6AddrByVlan {
		copy.ipv6AddrByVlan[k] = v.MakeCopy()
	}
	copy.intfByIpv6Addr = make(map[string]lib.IdxT, len(this.intfByIpv6Addr))
	for k, v := range this.intfByIpv6Addr {
		copy.intfByIpv6Addr[k] = v
	}
	copy.lagByIpv6Addr = make(map[string]lib.IdxT, len(this.lagByIpv6Addr))
	for k, v := range this.lagByIpv6Addr {
		copy.lagByIpv6Addr[k] = v
	}
	copy.vlanByIpv6Addr = make(map[string]lib.VidT, len(this.vlanByIpv6Addr))
	for k, v := range this.vlanByIpv6Addr {
		copy.vlanByIpv6Addr[k] = v
	}

	copy.lagByIntf = make(map[lib.IdxT]lib.IdxT, len(this.lagByIntf))
	for k, v := range this.lagByIntf {
		copy.lagByIntf[k] = v
	}
	copy.intfByLag = make(map[lib.IdxT]*lib.IdxTSet, len(this.intfByLag))
	for k, v := range this.intfByLag {
		copy.intfByLag[k] = v.MakeCopy()
	}

	copy.stpByIntf = this.stpByIntf.MakeCopy()

	copy.vlanAccessByIntf = make(map[lib.IdxT]lib.VidT, len(this.vlanAccessByIntf))
	for k, v := range this.vlanAccessByIntf {
		copy.vlanAccessByIntf[k] = v
	}
	copy.vlanAccessByLag = make(map[lib.IdxT]lib.VidT, len(this.vlanAccessByLag))
	for k, v := range this.vlanAccessByLag {
		copy.vlanAccessByLag[k] = v
	}
	copy.intfByVlanAccess = make(map[lib.VidT]*lib.IdxTSet, len(this.intfByVlanAccess))
	for k, v := range this.intfByVlanAccess {
		copy.intfByVlanAccess[k] = v.MakeCopy()
	}
	copy.lagByVlanAccess = make(map[lib.VidT]*lib.IdxTSet, len(this.lagByVlanAccess))
	for k, v := range this.lagByVlanAccess {
		copy.lagByVlanAccess[k] = v.MakeCopy()
	}

	copy.vlanNativeByIntf = make(map[lib.IdxT]lib.VidT, len(this.vlanNativeByIntf))
	for k, v := range this.vlanNativeByIntf {
		copy.vlanNativeByIntf[k] = v
	}
	copy.vlanNativeByLag = make(map[lib.IdxT]lib.VidT, len(this.vlanNativeByLag))
	for k, v := range this.vlanNativeByLag {
		copy.vlanNativeByLag[k] = v
	}
	copy.intfByVlanNative = make(map[lib.VidT]*lib.IdxTSet, len(this.intfByVlanNative))
	for k, v := range this.intfByVlanNative {
		copy.intfByVlanNative[k] = v.MakeCopy()
	}
	copy.lagByVlanNative = make(map[lib.VidT]*lib.IdxTSet, len(this.lagByVlanNative))
	for k, v := range this.lagByVlanNative {
		copy.lagByVlanNative[k] = v.MakeCopy()
	}

	copy.vlanTrunkByIntf = make(map[lib.IdxT]*lib.VidTSet, len(this.vlanTrunkByIntf))
	for k, v := range this.vlanTrunkByIntf {
		copy.vlanTrunkByIntf[k] = v.MakeCopy()
	}
	copy.vlanTrunkByLag = make(map[lib.IdxT]*lib.VidTSet, len(this.vlanTrunkByLag))
	for k, v := range this.vlanTrunkByLag {
		copy.vlanTrunkByLag[k] = v.MakeCopy()
	}
	copy.intfByVlanTrunk = make(map[lib.VidT]*lib.IdxTSet, len(this.intfByVlanTrunk))
	for k, v := range this.intfByVlanTrunk {
		copy.intfByVlanTrunk[k] = v.MakeCopy()
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
	for ifname, _ := range t.idxByIntfName {
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
	log.Infoln("Print list of LAG interfaces:")
	for _, lagName := range lags {
		log.Infoln(lagName)
	}

	log.Infoln("========================================")
	log.Infoln("Print membership of interfaces in access VLAN:")
	for _, ifname := range intfs {
		if accessVid, exists := t.vlanAccessByIntf[t.idxByIntfName[ifname]]; exists {
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
		idx := t.idxByIntfName[ifname]
		vids, exists := t.vlanTrunkByIntf[idx]
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
		log.Infof("There are %d VLANs on interface %s:", len(vlans), ifname)
		for _, vid := range vlans {
			log.Infoln(vid)
		}

		if nativeVid, exists := t.vlanNativeByIntf[idx]; exists {
			log.Infof("Native VLAN on interface %s: %d", ifname, nativeVid)
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
		idx := t.idxByIntfName[ifname]
		ipAddrs, exists := t.ipv4AddrByIntf[idx]
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
		idx := t.idxByIntfName[ifname]
		ipAddrs, exists := t.ipv6AddrByIntf[idx]
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
