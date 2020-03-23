package config

import (
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
	ipv4ByIntf map[lib.IdxT]*lib.StringSet
	// L3 LAG can have assigned many IPv4 addresses
	ipv4ByLag map[lib.IdxT]*lib.StringSet
	// L3 VLAN can have assigned many IPv4 addresses
	ipv4ByVlan map[lib.VidT]*lib.StringSet
	intfByIpv4 map[string]lib.IdxT
	lagByIpv4  map[string]lib.IdxT
	vlanByIpv4 map[string]lib.VidT
	ipv6ByIntf map[lib.IdxT]*lib.StringSet
	ipv6ByLag  map[lib.IdxT]*lib.StringSet
	ipv6ByVlan map[lib.VidT]*lib.StringSet
	// L3 interface can have assigned many IPv6 addresses
	intfByIpv6 map[string]lib.IdxT
	// L3 LAG can have assigned many IPv6 addresses
	lagByIpv6 map[string]lib.IdxT
	// L3 VLAN interface can have assigned many IPv6 addresses
	vlanByIpv6 map[string]lib.VidT
	lagByIntf  map[lib.IdxT]lib.IdxT
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
		ipv4ByIntf:         make(map[lib.IdxT]*lib.StringSet),
		ipv4ByLag:          make(map[lib.IdxT]*lib.StringSet),
		ipv4ByVlan:         make(map[lib.VidT]*lib.StringSet),
		intfByIpv4:         make(map[string]lib.IdxT),
		lagByIpv4:          make(map[string]lib.IdxT),
		vlanByIpv4:         make(map[string]lib.VidT),
		ipv6ByIntf:         make(map[lib.IdxT]*lib.StringSet),
		ipv6ByLag:          make(map[lib.IdxT]*lib.StringSet),
		ipv6ByVlan:         make(map[lib.VidT]*lib.StringSet),
		intfByIpv6:         make(map[string]lib.IdxT),
		lagByIpv6:          make(map[string]lib.IdxT),
		vlanByIpv6:         make(map[string]lib.VidT),
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
		ipAddrs, exists := t.ipv4ByIntf[idx]
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
		ipAddrs, exists := t.ipv6ByIntf[idx]
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

func (t *configLookupTablesT) saveIpv4AddressForInterface(ifname string, ip string) error {
	intfIdx := t.idxByIntfName[ifname]
	if _, exists := t.intfByIpv4[ip]; exists {
		return fmt.Errorf("Failed to assign IPv4 address %s to interface %s because it is already in use",
			ip, ifname)
	}

	t.intfByIpv4[ip] = intfIdx
	if _, exists := t.ipv4ByIntf[t.idxByIntfName[ifname]]; !exists {
		t.ipv4ByIntf[t.idxByIntfName[ifname]] = lib.NewStringSet()
	}
	// TODO: Check if IP is valid
	t.ipv4ByIntf[t.idxByIntfName[ifname]].Add(ip)
	log.Infof("Saved IPv4 %s for interface %s", ip, ifname)
	return nil
}

func (t *configLookupTablesT) saveIpv6AddressForInterface(ifname string, ip string) error {
	intfIdx := t.idxByIntfName[ifname]
	if _, exists := t.intfByIpv6[ip]; exists {
		return fmt.Errorf("Failed to assign IPv6 address %s to interface %s because it is already in use",
			ip, ifname)
	}

	t.intfByIpv6[ip] = intfIdx
	if _, exists := t.ipv6ByIntf[intfIdx]; !exists {
		t.ipv6ByIntf[intfIdx] = lib.NewStringSet()
	}
	// TODO: Check if IP is valid
	t.ipv6ByIntf[intfIdx].Add(ip)
	log.Infof("Saved IPv6 %s for interface %s", ip, ifname)
	return nil
}

func (t *configLookupTablesT) saveIpv4AddressForLag(lagName string, ip string) error {
	lagIdx := t.idxByLagName[lagName]
	if _, exists := t.lagByIpv4[ip]; exists {
		return fmt.Errorf("Failed to assign IPv4 address %s to LAG %s because it is already in use",
			ip, lagName)
	}

	t.lagByIpv4[ip] = lagIdx
	if _, exists := t.ipv4ByLag[lagIdx]; !exists {
		t.ipv4ByLag[lagIdx] = lib.NewStringSet()
	}
	// TODO: Check if IP is valid
	t.ipv4ByLag[lagIdx].Add(ip)
	log.Infof("Saved IPv4 %s for LAG %s", ip, lagName)
	return nil
}

func (t *configLookupTablesT) saveIpv6AddressForLag(lagName string, ip string) error {
	lagIdx := t.idxByLagName[lagName]
	if _, exists := t.lagByIpv6[ip]; exists {
		return fmt.Errorf("Failed to assign IPv6 address %s to interface %s because it is already in use",
			ip, lagName)
	}
	t.lagByIpv6[ip] = lagIdx

	if _, exists := t.ipv6ByLag[lagIdx]; !exists {
		t.ipv6ByLag[lagIdx] = lib.NewStringSet()
	}
	// TODO: Check if IP is valid
	t.ipv6ByLag[lagIdx].Add(ip)
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
	ipv4 := subIntf.GetIpv4()
	if ipv4 != nil {
		for _, addr := range ipv4.Address {
			ip := fmt.Sprintf("%s/%d", addr.GetIp(), addr.GetPrefixLength())
			if err := t.saveIpv4AddressForInterface(ifname, ip); err != nil {
				return err
			}
		}
	}

	ipv6 := subIntf.GetIpv6()
	if ipv6 != nil {
		for _, addr := range ipv6.Address {
			ip := fmt.Sprintf("%s/%d", addr.GetIp(), addr.GetPrefixLength())
			if err := t.saveIpv6AddressForInterface(ifname, ip); err != nil {
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
