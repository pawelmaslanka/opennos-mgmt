package config

import (
	"fmt"
	lib "golibext"
	cmd "opennos-mgmt/config/command"
	"opennos-mgmt/gnmi/modeldata/oc"
	"opennos-mgmt/utils"

	log "github.com/golang/glog"
	"github.com/jinzhu/copier"
	"github.com/r3labs/diff"
)

const (
	idSetVlanNameFmt          = "sv-%d"
	idDeleteVlanNameFmt       = "dv-%d"
	idSetAccessVlanNameFmt    = "sav-%d"
	idDeleteAccessVlanNameFmt = "dav-%d"
	idSetNativeVlanNameFmt    = "snv-%d"
	idDeleteNativeVlanNameFmt = "dnv-%d"
	idSetTrunkVlanNameFmt     = "stv-%d"
	idDeleteTrunkVlanNameFmt  = "stv-%d"
)

func extractVlanRelatedParametersFromEthIntf(ifname string, ethIntf *oc.Interface_Ethernet, isDelete bool) ([]diff.Change, error) {
	changes := make([]diff.Change, 0)
	swVlan := ethIntf.GetSwitchedVlan()
	if swVlan == nil {
		return changes, nil
	}

	fmt.Printf("Dump Switched Vlan:\n%v\n", *swVlan)

	if vlanMode := swVlan.GetInterfaceMode(); vlanMode != oc.OpenconfigVlan_VlanModeType_UNSET {
		changes = append(changes, *createVlanModeDiffChange(ifname, vlanMode, isDelete))
	}

	if accessVlan := swVlan.GetAccessVlan(); accessVlan != 0 { // 0 is incorrect VLAN ID
		fmt.Printf("Requested modify ACCESS VLAN %d\n", accessVlan)
		changes = append(changes, *createAccessVlanDiffChange(ifname, accessVlan, isDelete))
	}

	if nativeVlan := swVlan.GetNativeVlan(); nativeVlan != 0 { // 0 is incorrect VLAN ID
		fmt.Printf("Requested modify NATIVE VLAN %d\n", nativeVlan)
		changes = append(changes, *createNativeVlanDiffChange(ifname, nativeVlan, isDelete))
	}

	if trunkVlans := swVlan.GetTrunkVlans(); len(trunkVlans) > 0 {
		fmt.Printf("Requested modify TRUNK VLANs %v\n", trunkVlans)
		newChanges, err := createTrunkVlansDiffChange(ifname, trunkVlans, isDelete)
		if err != nil {
			return nil, err
		}

		changes = append(changes, newChanges...)
	}

	return changes, nil
}

func createEmptyDiffChangeWithVlanAndNilPath(vid uint16, isDelete bool) diff.Change {
	var ch diff.Change
	if isDelete {
		ch.Type = diff.DELETE
		ch.From = vid
		ch.To = nil
	} else {
		ch.Type = diff.CREATE
		ch.From = nil
		ch.To = vid
	}

	return ch
}

func createVlanModeDiffChange(ifname string, vlanMode oc.E_OpenconfigVlan_VlanModeType, isDelete bool) *diff.Change {
	var ch diff.Change
	if isDelete {
		ch.Type = diff.DELETE
		ch.From = vlanMode
		ch.To = nil
	} else {
		ch.Type = diff.CREATE
		ch.From = nil
		ch.To = vlanMode
	}

	ch.Path = make([]string, cmd.VlanModeEthPathItemsCountC)
	ch.Path[cmd.VlanEthIntfPathItemIdxC] = cmd.VlanEthIntfPathItemC
	ch.Path[cmd.VlanEthIfnamePathItemIdxC] = ifname
	ch.Path[cmd.VlanEthEthernetPathItemIdxC] = cmd.VlanEthEthernetPathItemC
	ch.Path[cmd.VlanEthSwVlanPathItemIdxC] = cmd.VlanEthSwVlanPathItemC
	ch.Path[cmd.VlanEthVlanModePathItemIdxC] = cmd.VlanEthVlanModePathItemC

	return &ch
}

func createAccessVlanDiffChange(ifname string, vid uint16, isDelete bool) *diff.Change {
	ch := createEmptyDiffChangeWithVlanAndNilPath(vid, isDelete)
	ch.Path = make([]string, cmd.AccessVlanEthPathItemsCountC)
	ch.Path[cmd.VlanEthIntfPathItemIdxC] = cmd.VlanEthIntfPathItemC
	ch.Path[cmd.VlanEthIfnamePathItemIdxC] = ifname
	ch.Path[cmd.VlanEthEthernetPathItemIdxC] = cmd.VlanEthEthernetPathItemC
	ch.Path[cmd.VlanEthSwVlanPathItemIdxC] = cmd.VlanEthSwVlanPathItemC
	ch.Path[cmd.VlanEthAccessVlanPathItemIdxC] = cmd.VlanEthAccessVlanPathItemC

	return &ch
}

func createNativeVlanDiffChange(ifname string, vid uint16, isDelete bool) *diff.Change {
	ch := createEmptyDiffChangeWithVlanAndNilPath(vid, isDelete)
	ch.Path = make([]string, cmd.NativeVlanEthPathItemsCountC)
	ch.Path[cmd.VlanEthIntfPathItemIdxC] = cmd.VlanEthIntfPathItemC
	ch.Path[cmd.VlanEthIfnamePathItemIdxC] = ifname
	ch.Path[cmd.VlanEthEthernetPathItemIdxC] = cmd.VlanEthEthernetPathItemC
	ch.Path[cmd.VlanEthSwVlanPathItemIdxC] = cmd.VlanEthSwVlanPathItemC
	ch.Path[cmd.VlanEthNativeVlanPathItemIdxC] = cmd.VlanEthNativeVlanPathItemC

	return &ch
}

func createTrunkVlansDiffChange(ifname string, vids []oc.Interface_Ethernet_SwitchedVlan_TrunkVlans_Union, isDelete bool) ([]diff.Change, error) {
	vlans := make([]uint16, 0)
	for _, vid := range vids {
		switch v := vid.(type) {
		case *oc.Interface_Ethernet_SwitchedVlan_TrunkVlans_Union_String:
			var lower, upper uint16
			n, err := fmt.Sscanf(v.String, "%d..%d", &lower, &upper)
			if n != 2 || err != nil {
				return nil, fmt.Errorf("Failed to parse lower and upper bound of trunk VLAN rane: %s", err)
			}

			if lower >= maxVlansC || upper >= maxVlansC {
				return nil, fmt.Errorf("Out of range lowwer and upper bound of trunk VLANs (%d, %d)", lower, upper)
			}

			for ; lower <= upper; lower++ {
				vlans = append(vlans, lower)
			}
		case *oc.Interface_Ethernet_SwitchedVlan_TrunkVlans_Union_Uint16:
			vlans = append(vlans, v.Uint16)
		default:
			return nil, fmt.Errorf("Cannot convert %v to Interface_Ethernet_SwitchedVlan_TrunkVlans_Union, unknown union type, got: %T, want any of [string, uint16]", v, v)
		}
	}

	changes := make([]diff.Change, len(vlans))
	for i, vid := range vlans {
		ch := createEmptyDiffChangeWithVlanAndNilPath(vid, isDelete)
		ch.Path = make([]string, cmd.TrunkVlanEthPathItemsCountC)
		ch.Path[cmd.VlanEthIntfPathItemIdxC] = cmd.VlanEthIntfPathItemC
		ch.Path[cmd.VlanEthIfnamePathItemIdxC] = ifname
		ch.Path[cmd.VlanEthEthernetPathItemIdxC] = cmd.VlanEthEthernetPathItemC
		ch.Path[cmd.VlanEthSwVlanPathItemIdxC] = cmd.VlanEthSwVlanPathItemC
		ch.Path[cmd.VlanEthTrunkVlanPathItemIdxC] = cmd.VlanEthTrunkVlanPathItemC
		ch.Path[cmd.TrunkVlanEthIdxPathItemIdxC] = fmt.Sprintf("%d", i)

		changes = append(changes, ch)
	}

	return changes, nil
}

func isChangedVlanMode(change *diff.Change) bool {
	if len(change.Path) != cmd.VlanModeEthPathItemsCountC {
		return false
	}

	if (change.Path[cmd.VlanEthIntfPathItemIdxC] == cmd.VlanEthIntfPathItemC) && (change.Path[cmd.VlanEthEthernetPathItemIdxC] == cmd.VlanEthEthernetPathItemC) && (change.Path[cmd.VlanEthSwVlanPathItemIdxC] == cmd.VlanEthSwVlanPathItemC) && (change.Path[cmd.VlanEthVlanModePathItemIdxC] == cmd.VlanEthVlanModePathItemC) {
		return true
	}

	return false
}

func isChangedAccessVlan(change *diff.Change) bool {
	if len(change.Path) != cmd.AccessVlanEthPathItemsCountC {
		return false
	}

	if (change.Path[cmd.VlanEthIntfPathItemIdxC] == cmd.VlanEthIntfPathItemC) && (change.Path[cmd.VlanEthEthernetPathItemIdxC] == cmd.VlanEthEthernetPathItemC) && (change.Path[cmd.VlanEthSwVlanPathItemIdxC] == cmd.VlanEthSwVlanPathItemC) && (change.Path[cmd.VlanEthAccessVlanPathItemIdxC] == cmd.VlanEthAccessVlanPathItemC) {
		return true
	}

	return false
}

func isChangedNativeVlan(change *diff.Change) bool {
	if len(change.Path) != cmd.NativeVlanEthPathItemsCountC {
		return false
	}

	if (change.Path[cmd.VlanEthIntfPathItemIdxC] == cmd.VlanEthIntfPathItemC) && (change.Path[cmd.VlanEthEthernetPathItemIdxC] == cmd.VlanEthEthernetPathItemC) && (change.Path[cmd.VlanEthSwVlanPathItemIdxC] == cmd.VlanEthSwVlanPathItemC) && (change.Path[cmd.VlanEthNativeVlanPathItemIdxC] == cmd.VlanEthNativeVlanPathItemC) {
		return true
	}

	return false
}

func isChangedTrunkVlan(change *diff.Change) bool {
	if (len(change.Path) != cmd.TrunkVlanEthPathItemsCountC) && (len(change.Path) != cmd.TrunkVlanEthPathItemsCountIfUpdateC) {
		return false
	}

	if (change.Path[cmd.VlanEthIntfPathItemIdxC] == cmd.VlanEthIntfPathItemC) && (change.Path[cmd.VlanEthEthernetPathItemIdxC] == cmd.VlanEthEthernetPathItemC) && (change.Path[cmd.VlanEthSwVlanPathItemIdxC] == cmd.VlanEthSwVlanPathItemC) && (change.Path[cmd.VlanEthTrunkVlanPathItemIdxC] == cmd.VlanEthTrunkVlanPathItemC) {
		return true
	}

	return false
}

func doFindSetVlanModeEthIntfChange(changelog *DiffChangelogMgmtT) (*DiffChangeMgmtT, bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type != diff.DELETE {
				if isChangedVlanMode(ch.Change) {
					if ch.Change.To != nil {
						return ch, true
					}
				}
			}
		}
	}

	return nil, false
}

func doFindSetAccessVlanEthIntfChange(changelog *DiffChangelogMgmtT) (*DiffChangeMgmtT, bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type != diff.DELETE {
				if isChangedAccessVlan(ch.Change) {
					if ch.Change.To != nil {
						return ch, true
					}
				}
			}
		}
	}

	return nil, false
}

func doFindDeleteAccessVlanEthIntfChange(changelog *DiffChangelogMgmtT) (*DiffChangeMgmtT, bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type != diff.CREATE {
				if isChangedAccessVlan(ch.Change) {
					if ch.Change.From != nil {
						return ch, true
					}
				}
			}
		}
	}

	return nil, false
}

func doFindSetNativeVlanEthIntfChange(changelog *DiffChangelogMgmtT) (*DiffChangeMgmtT, bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type != diff.DELETE {
				if isChangedNativeVlan(ch.Change) {
					if ch.Change.To != nil {
						return ch, true
					}
				}
			}
		}
	}

	return nil, false
}

func doFindDeleteNativeVlanEthIntfChange(changelog *DiffChangelogMgmtT) (*DiffChangeMgmtT, bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type != diff.CREATE {
				if isChangedNativeVlan(ch.Change) {
					if ch.Change.From != nil {
						return ch, true
					}
				}
			}
		}
	}

	return nil, false
}

func doFindSetTrunkVlanEthIntfChange(changelog *DiffChangelogMgmtT) (*DiffChangeMgmtT, bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type != diff.DELETE {
				if isChangedTrunkVlan(ch.Change) {
					if ch.Change.To != nil {
						return ch, true
					}
				}
			}
		}
	}

	return nil, false
}

func doFindDeleteTrunkVlanEthIntfChange(changelog *DiffChangelogMgmtT) (*DiffChangeMgmtT, bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type != diff.CREATE {
				if isChangedTrunkVlan(ch.Change) {
					if ch.Change.From != nil {
						return ch, true
					}
				}
			}
		}
	}

	return nil, false
}

func findSetVlanModeEthIntfChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	return doFindSetVlanModeEthIntfChange(changelog)
}

func findSetAccessVlanEthIntfChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	return doFindSetAccessVlanEthIntfChange(changelog)
}

func findDeleteAccessVlanEthIntfChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	return doFindDeleteAccessVlanEthIntfChange(changelog)
}

func findSetNativeVlanEthIntfChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	return doFindSetNativeVlanEthIntfChange(changelog)
}

func findDeleteNativeVlanEthIntfChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	return doFindDeleteNativeVlanEthIntfChange(changelog)
}

func findSetTrunkVlanEthIntfChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	return doFindSetTrunkVlanEthIntfChange(changelog)
}

func findDeleteTrunkVlanEthIntfChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	return doFindDeleteTrunkVlanEthIntfChange(changelog)
}

func (this *ConfigMngrT) validateSetVlanModeEthIntfChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ifname := changeItem.Change.Path[cmd.VlanEthIfnamePathItemIdxC]
	if !this.isEthIntfAvailable(ifname) {
		return fmt.Errorf("Ethernet interface %s is not available", ifname)
	}

	vlanMode := changeItem.Change.To.(oc.E_OpenconfigVlan_VlanModeType)
	log.Infof("Requested set VLAN mode (%d) for Ethernet interface %s", vlanMode, ifname)
	setVlanModeEthIntfCmd := cmd.NewSetVlanModeEthIntfCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForSetVlanModeForEthIntf(ifname, vlanMode); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from interface %s:\n%s",
			setVlanModeEthIntfCmd.GetName(), ifname, err)
	}

	if this.transHasBeenStarted {
		if err := this.appendCmdToTransaction(ifname, setVlanModeEthIntfCmd, setVlanModeForEthIntfC, false); err != nil {
			return err
		}
	}

	if err := this.transConfigLookupTbl.setVlanModeEthIntf(ifname, vlanMode); err != nil {
		return err
	}

	changeItem.MarkAsProcessed()

	return nil
}

func (this *ConfigMngrT) validateSetAccessVlanEthIntfChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ifname := changeItem.Change.Path[cmd.VlanEthIfnamePathItemIdxC]
	if !this.isEthIntfAvailable(ifname) {
		return fmt.Errorf("Ethernet interface %s is not available", ifname)
	}

	vid16, err := utils.ConvertGoInterfaceIntoUint16(changeItem.Change.To)
	if err != nil {
		return err
	}

	vid := lib.VidT(vid16)
	vlanModeChange, exists := findSetVlanModeEthIntfChange(changelog)
	if exists {
		reqVlanMode := oc.E_OpenconfigVlan_VlanModeType(*vlanModeChange.Change.To.(*uint8))
		if reqVlanMode != oc.OpenconfigVlan_VlanModeType_TRUNK {
			return fmt.Errorf("Set access VLAN %d for Ethernet interface %s is disallowed if VLAN interface mode is not going to be access.\nRequested mode: %v", vid, ifname, reqVlanMode)
		}
	} else {
		vlanMode, err := this.transConfigLookupTbl.getVlanModeEthIntf(ifname)
		if err != nil {
			return fmt.Errorf("Could not determine VLAN mode on Ethernet interface %s", ifname)
		}

		if vlanMode != oc.OpenconfigVlan_VlanModeType_ACCESS {
			return fmt.Errorf("Set access VLAN %d for Ethernet interface %s is disallowed if VLAN interface mode is not access. Current mode: %v", vid, ifname, vlanMode)
		}
	}

	log.Infof("Requested set access VLAN %d for Ethernet interface %s", vid, ifname)
	setAccessVlanEthIntfCmd := cmd.NewSetAccessVlanEthIntfCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForSetAccessVlanForEthIntf(ifname, vid); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from interface %s:\n%s",
			setAccessVlanEthIntfCmd.GetName(), ifname, err)
	}

	if this.transHasBeenStarted {
		if !this.transConfigLookupTbl.IsThereAnyMemberVlan(vid) {
			var newChange diff.Change
			newChange.Type = diff.CREATE
			newChange.From = nil
			newChange.To = vid
			setVlanCmd := cmd.NewSetVlanCmdT(&newChange, this.ethSwitchMgmtClient)
			id := fmt.Sprintf(idSetVlanNameFmt, vid)
			if err := this.appendCmdToTransaction(id, setVlanCmd, setVlanC, false); err != nil {
				return err
			}
		}

		id := fmt.Sprintf(idSetAccessVlanNameFmt, vid)
		if err := this.appendCmdToTransaction(id, setAccessVlanEthIntfCmd, setAccessVlanForEthIntfC, false); err != nil {
			return err
		}
	}

	this.transConfigLookupTbl.setAccessVlanEthIntf(ifname, vid)
	changeItem.MarkAsProcessed()

	return nil
}

func (this *ConfigMngrT) validateDeleteAccessVlanEthIntfChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ifname := changeItem.Change.Path[cmd.VlanEthIfnamePathItemIdxC]
	if !this.isEthIntfAvailable(ifname) {
		return fmt.Errorf("Ethernet interface %s is not available", ifname)
	}

	vid16, err := utils.ConvertGoInterfaceIntoUint16(changeItem.Change.From)
	if err != nil {
		return err
	}

	vid := lib.VidT(vid16)
	vlanMode, err := this.transConfigLookupTbl.getVlanModeEthIntf(ifname)
	if vlanMode != oc.OpenconfigVlan_VlanModeType_ACCESS {
		return fmt.Errorf("Deletion of access VLAN %d from Ethernet interface %s is disallowed if VLAN interface mode is not access. Current mode: %v", vid, ifname, vlanMode)
	}

	var newChange diff.Change
	needsCreateNewChange := (changeItem.Change.Type == diff.UPDATE) && (changeItem.Change.To != nil)
	if needsCreateNewChange {
		// Update type carries info about old and new access VLAN ID. Let's create new change item
		// in order to process new native VLAN it by SetAccessVlanEthIntfCmd
		copier.Copy(&newChange, changeItem.Change)
		newChange.Type = diff.CREATE
		newChange.From = nil
		// Update current change
		changeItem.Change.Type = diff.DELETE
		changeItem.Change.To = nil
	}

	log.Infof("Requested delete access VLAN %d from Ethernet interface %s", vid, ifname)
	deleteAccessVlanEthIntfCmd := cmd.NewDeleteAccessVlanEthIntfCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForDeleteAccessVlanFromEthIntf(ifname, vid); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from interface %s:\n%s",
			deleteAccessVlanEthIntfCmd.GetName(), ifname, err)
	}

	if this.transHasBeenStarted {
		id := fmt.Sprintf(idDeleteAccessVlanNameFmt, vid)
		if err = this.appendCmdToTransaction(id, deleteAccessVlanEthIntfCmd, deleteEthIntfFromAccessVlanC, false); err != nil {
			return err
		}
	}

	if err := this.transConfigLookupTbl.deleteAccessVlanEthIntf(ifname, vid); err != nil {
		return err
	}

	if !this.transConfigLookupTbl.IsThereAnyMemberVlan(vid) {
		var newChange diff.Change
		newChange.Type = diff.DELETE
		newChange.From = vid
		newChange.To = nil
		deleteVlanCmd := cmd.NewDeleteVlanCmdT(&newChange, this.ethSwitchMgmtClient)
		id := fmt.Sprintf(idDeleteVlanNameFmt, vid)
		if err := this.appendCmdToTransaction(id, deleteVlanCmd, deleteVlanC, false); err != nil {
			return err
		}
	}

	if needsCreateNewChange {
		changelog.Changes = append(changelog.Changes, NewDiffChangeMgmtT(&newChange))
	}

	changeItem.MarkAsProcessed()

	return nil
}

func (this *ConfigMngrT) validateSetNativeVlanEthIntfChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ifname := changeItem.Change.Path[cmd.VlanEthIfnamePathItemIdxC]
	if !this.isEthIntfAvailable(ifname) {
		return fmt.Errorf("Ethernet interface %s is not available", ifname)
	}

	vid16, err := utils.ConvertGoInterfaceIntoUint16(changeItem.Change.To)
	if err != nil {
		return err
	}

	vid := lib.VidT(vid16)
	vlanModeChange, exists := findSetVlanModeEthIntfChange(changelog)
	if exists {
		reqVlanMode := oc.E_OpenconfigVlan_VlanModeType(*vlanModeChange.Change.To.(*uint8))
		if reqVlanMode != oc.OpenconfigVlan_VlanModeType_TRUNK {
			return fmt.Errorf("Set native VLAN %d for Ethernet interface %s is disallowed if VLAN interface mode is not going to be trunk.\nRequested mode: %v", vid, ifname, reqVlanMode)
		}
	} else {
		vlanMode, err := this.transConfigLookupTbl.getVlanModeEthIntf(ifname)
		if err != nil {
			return fmt.Errorf("Could not determine VLAN mode on Ethernet interface %s", ifname)
		}

		if vlanMode != oc.OpenconfigVlan_VlanModeType_TRUNK {
			return fmt.Errorf("Set native VLAN %d for Ethernet interface %s is disallowed if VLAN interface mode is not trunk. Current mode: %v", vid, ifname, vlanMode)
		}
	}

	log.Infof("Requested set native VLAN %d for Ethernet interface %s", vid, ifname)
	setNativeVlanEthIntfCmd := cmd.NewSetNativeVlanEthIntfCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForSetNativeVlanForEthIntf(ifname, vid); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from interface %s:\n%s",
			setNativeVlanEthIntfCmd.GetName(), ifname, err)
	}

	if this.transHasBeenStarted {
		if !this.transConfigLookupTbl.IsThereAnyMemberVlan(vid) {
			var newChange diff.Change
			newChange.Type = diff.CREATE
			newChange.From = nil
			newChange.To = vid
			setVlanCmd := cmd.NewSetVlanCmdT(&newChange, this.ethSwitchMgmtClient)
			id := fmt.Sprintf(idSetVlanNameFmt, vid)
			if err := this.appendCmdToTransaction(id, setVlanCmd, setVlanC, false); err != nil {
				return err
			}
		}

		id := fmt.Sprintf(idSetNativeVlanNameFmt, vid)
		if err = this.appendCmdToTransaction(id, setNativeVlanEthIntfCmd, setNativeVlanForEthIntfC, false); err != nil {
			return err
		}
	}

	this.transConfigLookupTbl.setNativeVlanEthIntf(ifname, vid)
	changeItem.MarkAsProcessed()

	return nil
}

func (this *ConfigMngrT) validateDeleteNativeVlanEthIntfChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ifname := changeItem.Change.Path[cmd.VlanEthIfnamePathItemIdxC]
	if !this.isEthIntfAvailable(ifname) {
		return fmt.Errorf("Ethernet interface %s is not available", ifname)
	}

	vid16, err := utils.ConvertGoInterfaceIntoUint16(changeItem.Change.From)
	if err != nil {
		return err
	}

	vid := lib.VidT(vid16)
	vlanMode, err := this.transConfigLookupTbl.getVlanModeEthIntf(ifname)
	if vlanMode != oc.OpenconfigVlan_VlanModeType_TRUNK {
		return fmt.Errorf("Deletion of native VLAN %d from Ethernet interface %s is disallowed if VLAN interface mode is not trunk. Current mode: %v", vid, ifname, vlanMode)
	}

	var newChange diff.Change
	needsCreateNewChange := (changeItem.Change.Type == diff.UPDATE) && (changeItem.Change.To != nil)
	if needsCreateNewChange {
		// Update type carries info about old and new native VLAN ID. Let's create new change item
		// in order to process new native VLAN it by SetNativeVlanEthIntfCmd
		copier.Copy(&newChange, changeItem.Change)
		newChange.Type = diff.CREATE
		newChange.From = nil
		// Update current change
		changeItem.Change.Type = diff.DELETE
		changeItem.Change.To = nil
	}

	log.Infof("Requested delete native VLAN %d from Ethernet interface %s", vid, ifname)
	deleteNativeVlanEthIntfCmd := cmd.NewDeleteNativeVlanEthIntfCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForDeleteNativeVlanFromEthIntf(ifname, vid); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from interface %s:\n%s",
			deleteNativeVlanEthIntfCmd.GetName(), ifname, err)
	}

	if this.transHasBeenStarted {
		id := fmt.Sprintf(idDeleteNativeVlanNameFmt, vid)
		if err = this.appendCmdToTransaction(id, deleteNativeVlanEthIntfCmd, deleteEthIntfFromNativeVlanC, false); err != nil {
			return err
		}
	}

	if err := this.transConfigLookupTbl.deleteNativeVlanEthIntf(ifname, vid); err != nil {
		return err
	}

	if !this.transConfigLookupTbl.IsThereAnyMemberVlan(vid) {
		var newChange diff.Change
		newChange.Type = diff.DELETE
		newChange.From = vid
		newChange.To = nil
		deleteVlanCmd := cmd.NewDeleteVlanCmdT(&newChange, this.ethSwitchMgmtClient)
		id := fmt.Sprintf(idDeleteVlanNameFmt, vid)
		if err := this.appendCmdToTransaction(id, deleteVlanCmd, deleteVlanC, false); err != nil {
			return err
		}
	}

	if needsCreateNewChange {
		changelog.Changes = append(changelog.Changes, NewDiffChangeMgmtT(&newChange))
	}

	changeItem.MarkAsProcessed()

	return nil
}

func (this *ConfigMngrT) validateSetTrunkVlanEthIntfChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ifname := changeItem.Change.Path[cmd.VlanEthIfnamePathItemIdxC]
	if !this.isEthIntfAvailable(ifname) {
		return fmt.Errorf("Ethernet interface %s is not available", ifname)
	}

	vid16, err := utils.ConvertGoInterfaceIntoUint16(changeItem.Change.To)
	if err != nil {
		return err
	}

	vid := lib.VidT(vid16)
	vlanModeChange, exists := findSetVlanModeEthIntfChange(changelog)
	if exists {
		reqVlanMode := oc.E_OpenconfigVlan_VlanModeType(*vlanModeChange.Change.To.(*uint8))
		if reqVlanMode != oc.OpenconfigVlan_VlanModeType_TRUNK {
			return fmt.Errorf("Set trunk VLAN %d for Ethernet interface %s is disallowed if VLAN interface mode is not going to be trunk.\nRequested mode: %v", vid, ifname, reqVlanMode)
		}
	} else {
		vlanMode, err := this.transConfigLookupTbl.getVlanModeEthIntf(ifname)
		if err != nil {
			return fmt.Errorf("Could not determine VLAN mode on Ethernet interface %s", ifname)
		}

		if vlanMode != oc.OpenconfigVlan_VlanModeType_TRUNK {
			return fmt.Errorf("Set trunk VLAN %d for Ethernet interface %s is disallowed if VLAN interface mode is not trunk. Current mode: %v", vid, ifname, vlanMode)
		}
	}

	log.Infof("Requested set trunk VLAN %d from Ethernet interface %s", vid, ifname)
	setTrunkVlanEthIntfCmd := cmd.NewSetTrunkVlanEthIntfCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForSetTrunkVlanForEthIntf(ifname, vid); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from interface %s:\n%s",
			setTrunkVlanEthIntfCmd.GetName(), ifname, err)
	}

	if this.transHasBeenStarted {
		if !this.transConfigLookupTbl.IsThereAnyMemberVlan(vid) {
			var newChange diff.Change
			newChange.Type = diff.CREATE
			newChange.From = nil
			newChange.To = vid
			setVlanCmd := cmd.NewSetVlanCmdT(&newChange, this.ethSwitchMgmtClient)
			id := fmt.Sprintf(idSetVlanNameFmt, vid)
			if err := this.appendCmdToTransaction(id, setVlanCmd, setVlanC, false); err != nil {
				return err
			}
		}

		id := fmt.Sprintf(idSetTrunkVlanNameFmt, vid)
		if err = this.appendCmdToTransaction(id, setTrunkVlanEthIntfCmd, setTrunkVlanForEthIntfC, false); err != nil {
			return err
		}
	}

	if err := this.transConfigLookupTbl.setTrunkVlanEthIntf(ifname, vid); err != nil {
		return err
	}

	changeItem.MarkAsProcessed()

	return nil
}

func (this *ConfigMngrT) validateDeleteTrunkVlanEthIntfChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ifname := changeItem.Change.Path[cmd.VlanEthIfnamePathItemIdxC]
	if !this.isEthIntfAvailable(ifname) {
		return fmt.Errorf("Ethernet interface %s is not available", ifname)
	}

	vid16, err := utils.ConvertGoInterfaceIntoUint16(changeItem.Change.From)
	if err != nil {
		return err
	}

	vid := lib.VidT(vid16)
	vlanMode, err := this.transConfigLookupTbl.getVlanModeEthIntf(ifname)
	if vlanMode != oc.OpenconfigVlan_VlanModeType_TRUNK {
		return fmt.Errorf("Deletion of trunk VLAN %d from Ethernet interface %s is disallowed if VLAN interface mode is not trunk. Current mode: %v", vid, ifname, vlanMode)
	}

	var newChange diff.Change
	needsCreateNewChange := (changeItem.Change.Type == diff.UPDATE) && (changeItem.Change.To != nil)
	if needsCreateNewChange {
		// Update type carries info about old and new trunk VLAN ID. Let's create new change item
		// in order to process new native VLAN it by SetTrunkVlanEthIntfCmd
		copier.Copy(&newChange, changeItem.Change)
		newChange.Type = diff.CREATE
		newChange.From = nil
		// Let's drop "Uint16"/"String"
		newChange.Path = newChange.Path[:len(newChange.Path)-1]
		// Update current change
		changeItem.Change.Type = diff.DELETE
		changeItem.Change.To = nil
	}

	log.Infof("Requested delete trunk VLAN %d from Ethernet interface %s", vid, ifname)
	deleteTrunkVlanEthIntfCmd := cmd.NewDeleteTrunkVlanEthIntfCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForDeleteTrunkVlanFromEthIntf(ifname, vid); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from interface %s:\n%s",
			deleteTrunkVlanEthIntfCmd.GetName(), ifname, err)
	}

	if this.transHasBeenStarted {
		id := fmt.Sprintf(idDeleteTrunkVlanNameFmt, vid)
		if err = this.appendCmdToTransaction(id, deleteTrunkVlanEthIntfCmd, deleteEthIntfFromTrunkVlanC, false); err != nil {
			return err
		}
	}

	if err := this.transConfigLookupTbl.deleteTrunkVlanEthIntf(ifname, vid); err != nil {
		return err
	}

	if !this.transConfigLookupTbl.IsThereAnyMemberVlan(vid) {
		var newChange diff.Change
		newChange.Type = diff.DELETE
		newChange.From = vid
		newChange.To = nil
		deleteVlanCmd := cmd.NewDeleteVlanCmdT(&newChange, this.ethSwitchMgmtClient)
		id := fmt.Sprintf(idDeleteVlanNameFmt, vid)
		if err := this.appendCmdToTransaction(id, deleteVlanCmd, deleteVlanC, false); err != nil {
			return err
		}
	}

	if needsCreateNewChange {
		changelog.Changes = append(changelog.Changes, NewDiffChangeMgmtT(&newChange))
	}

	changeItem.MarkAsProcessed()

	return nil
}

func (this *ConfigMngrT) processSetVlanModeEthIntfFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to set native VLAN for Ethernet interface
		if change, exists := findSetVlanModeEthIntfChange(changelog); exists {
			if err := this.validateSetVlanModeEthIntfChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (this *ConfigMngrT) processSetAccessVlanEthIntfFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to set native VLAN for Ethernet interface
		if change, exists := findSetAccessVlanEthIntfChange(changelog); exists {
			if err := this.validateSetAccessVlanEthIntfChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (this *ConfigMngrT) processDeleteAccessVlanEthIntfFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to delete native VLAN from Ethernet interface
		if change, exists := findDeleteAccessVlanEthIntfChange(changelog); exists {
			if err := this.validateDeleteAccessVlanEthIntfChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (this *ConfigMngrT) processSetNativeVlanEthIntfFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to set native VLAN for Ethernet interface
		if change, exists := findSetNativeVlanEthIntfChange(changelog); exists {
			if err := this.validateSetNativeVlanEthIntfChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (this *ConfigMngrT) processDeleteNativeVlanEthIntfFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to delete native VLAN from Ethernet interface
		if change, exists := findDeleteNativeVlanEthIntfChange(changelog); exists {
			if err := this.validateDeleteNativeVlanEthIntfChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (this *ConfigMngrT) processSetTrunkVlanEthIntfFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to set trunk VLAN for Ethernet interface
		if change, exists := findSetTrunkVlanEthIntfChange(changelog); exists {
			if err := this.validateSetTrunkVlanEthIntfChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (this *ConfigMngrT) processDeleteTrunkVlanEthIntfFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to delete trunk VLANs from Ethernet interface
		if change, exists := findDeleteTrunkVlanEthIntfChange(changelog); exists {
			if err := this.validateDeleteTrunkVlanEthIntfChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (this *ConfigMngrT) setVlanEthIntf(device *oc.Device) error {
	createdVlans := lib.NewVidTSet()
	var err error
	for _, ethIfname := range this.configLookupTbl.ethIfnameByIdx {
		intf := device.GetInterface(ethIfname)
		if intf == nil {
			continue
		}

		eth := intf.GetEthernet()
		if eth == nil {
			continue
		}

		swVlan := eth.GetSwitchedVlan()
		if swVlan == nil {
			continue
		}

		mode := swVlan.GetInterfaceMode()
		if mode == oc.OpenconfigVlan_VlanModeType_UNSET {
			continue
		}

		modeChange := createVlanModeDiffChange(ethIfname, mode, false)
		vlanModeCmd := cmd.NewSetVlanModeEthIntfCmdT(modeChange, this.ethSwitchMgmtClient)
		if err = this.appendCmdToTransaction(ethIfname, vlanModeCmd, setVlanModeForEthIntfC, true); err != nil {
			return err
		}

		switch mode {
		case oc.OpenconfigVlan_VlanModeType_ACCESS:
			if accessVlan, exists := this.configLookupTbl.vlanAccessByEth[this.configLookupTbl.idxByEthIfname[ethIfname]]; exists {
				var accessChange diff.Change
				accessChange.Type = diff.CREATE
				accessChange.From = nil
				accessChange.To = accessVlan
				accessChange.Path = make([]string, cmd.AccessVlanEthPathItemsCountC)
				accessChange.Path[cmd.VlanEthIntfPathItemIdxC] = cmd.VlanEthIntfPathItemC
				accessChange.Path[cmd.VlanEthIfnamePathItemIdxC] = ethIfname
				accessChange.Path[cmd.VlanEthEthernetPathItemIdxC] = cmd.VlanEthEthernetPathItemC
				accessChange.Path[cmd.VlanEthSwVlanPathItemIdxC] = cmd.VlanEthSwVlanPathItemC
				accessChange.Path[cmd.VlanEthAccessVlanPathItemIdxC] = cmd.VlanEthAccessVlanPathItemC

				accessVlanCmd := cmd.NewSetAccessVlanEthIntfCmdT(&accessChange, this.ethSwitchMgmtClient)
				id := fmt.Sprintf(idSetAccessVlanNameFmt, accessVlan)
				if err = this.appendCmdToTransaction(id, accessVlanCmd, setAccessVlanForEthIntfC, true); err != nil {
					return err
				}

				if !createdVlans.Has(accessVlan) {
					fmt.Printf("Access VLAN %d does not exist. Creating...", accessVlan)
					if err := this.CreateVlanCmd(accessVlan); err != nil {
						return err
					}

					createdVlans.Add(accessVlan)
				}
			}

		case oc.OpenconfigVlan_VlanModeType_TRUNK:
			if nativeVlan, exists := this.configLookupTbl.vlanNativeByEth[this.configLookupTbl.idxByEthIfname[ethIfname]]; exists {
				var nativeChange diff.Change
				nativeChange.Type = diff.CREATE
				nativeChange.From = nil
				nativeChange.To = nativeVlan
				nativeChange.Path = make([]string, cmd.NativeVlanEthPathItemsCountC)
				nativeChange.Path[cmd.VlanEthIntfPathItemIdxC] = cmd.VlanEthIntfPathItemC
				nativeChange.Path[cmd.VlanEthIfnamePathItemIdxC] = ethIfname
				nativeChange.Path[cmd.VlanEthEthernetPathItemIdxC] = cmd.VlanEthEthernetPathItemC
				nativeChange.Path[cmd.VlanEthSwVlanPathItemIdxC] = cmd.VlanEthSwVlanPathItemC
				nativeChange.Path[cmd.VlanEthNativeVlanPathItemIdxC] = cmd.VlanEthNativeVlanPathItemC

				nativeVlanCmd := cmd.NewSetNativeVlanEthIntfCmdT(&nativeChange, this.ethSwitchMgmtClient)
				id := fmt.Sprintf(idSetNativeVlanNameFmt, nativeVlan)
				if err = this.appendCmdToTransaction(id, nativeVlanCmd, setNativeVlanForEthIntfC, true); err != nil {
					return err
				}

				if !createdVlans.Has(nativeVlan) {
					fmt.Printf("Native VLAN %d does not exist. Creating...", nativeVlan)
					if err := this.CreateVlanCmd(nativeVlan); err != nil {
						return err
					}

					createdVlans.Add(nativeVlan)
				}
			}

			for i, trunkVlan := range this.configLookupTbl.vlanTrunkByEth[this.configLookupTbl.idxByEthIfname[ethIfname]].VidTs() {
				var trunkChange diff.Change
				trunkChange.Type = diff.CREATE
				trunkChange.From = nil
				trunkChange.To = trunkVlan
				trunkChange.Path = make([]string, cmd.TrunkVlanEthPathItemsCountC)
				trunkChange.Path[cmd.VlanEthIntfPathItemIdxC] = cmd.VlanEthIntfPathItemC
				trunkChange.Path[cmd.VlanEthIfnamePathItemIdxC] = ethIfname
				trunkChange.Path[cmd.VlanEthEthernetPathItemIdxC] = cmd.VlanEthEthernetPathItemC
				trunkChange.Path[cmd.VlanEthSwVlanPathItemIdxC] = cmd.VlanEthSwVlanPathItemC
				trunkChange.Path[cmd.VlanEthTrunkVlanPathItemIdxC] = cmd.VlanEthTrunkVlanPathItemC
				trunkChange.Path[cmd.TrunkVlanEthIdxPathItemIdxC] = fmt.Sprintf("%d", i)

				trunkVlanCmd := cmd.NewSetTrunkVlanEthIntfCmdT(&trunkChange, this.ethSwitchMgmtClient)
				id := fmt.Sprintf(idSetTrunkVlanNameFmt, trunkVlan)
				if err = this.appendCmdToTransaction(id, trunkVlanCmd, setTrunkVlanForEthIntfC, true); err != nil {
					return err
				}

				if !createdVlans.Has(trunkVlan) {
					fmt.Printf("Trunk VLAN %d does not exist. Creating...", trunkVlan)
					if err := this.CreateVlanCmd(trunkVlan); err != nil {
						return err
					}

					createdVlans.Add(trunkVlan)
				}
			}
		}
	}

	return nil
}

func (this *ConfigMngrT) CreateVlanCmd(vid lib.VidT) error {
	var newChange diff.Change
	newChange.Type = diff.CREATE
	newChange.From = nil
	newChange.To = vid
	setVlanCmd := cmd.NewSetVlanCmdT(&newChange, this.ethSwitchMgmtClient)
	id := fmt.Sprintf(idSetVlanNameFmt, vid)
	return this.appendCmdToTransaction(id, setVlanCmd, setVlanC, false)
}
