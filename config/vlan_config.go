package config

import (
	"fmt"
	lib "golibext"
	cmd "opennos-mgmt/config/command"
	"opennos-mgmt/gnmi/modeldata/oc"

	log "github.com/golang/glog"
	"github.com/jinzhu/copier"
	"github.com/r3labs/diff"
)

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
	if len(change.Path) != cmd.TrunkVlanEthPathItemsCountC {
		return false
	}

	if (change.Path[cmd.VlanEthIntfPathItemIdxC] == cmd.VlanEthIntfPathItemC) && (change.Path[cmd.VlanEthEthernetPathItemIdxC] == cmd.VlanEthEthernetPathItemC) && (change.Path[cmd.VlanEthSwVlanPathItemIdxC] == cmd.VlanEthSwVlanPathItemC) && (change.Path[cmd.VlanEthTrunkVlanPathItemIdxC] == cmd.VlanEthTrunkVlanPathItemC) {
		return true
	}

	return false
}

func doFindsetVlanModeEthIntfChange(changelog *DiffChangelogMgmtT) (*DiffChangeMgmtT, bool) {
	return nil, false
}

func doFindDeleteAccessVlanEthIntfChange(changelog *DiffChangelogMgmtT) (*DiffChangeMgmtT, bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type != diff.CREATE {
				if isChangedAccessVlan(ch.Change) {
					return ch, true
				}
			}
		}
	}

	return nil, false
}

func doFindSetNativeVlanEthIntfChange(changelog *DiffChangelogMgmtT) (*DiffChangeMgmtT, bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type == diff.CREATE {
				if isChangedNativeVlan(ch.Change) {
					return ch, true
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
					return ch, true
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
					return ch, true
				}
			}
		}
	}

	return nil, false
}

func findsetVlanModeEthIntfChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	return doFindSetNativeVlanEthIntfChange(changelog)
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

func findDeleteTrunkVlanEthIntfChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	return doFindDeleteTrunkVlanEthIntfChange(changelog)
}

func extractVlanIdFromInterface(vlanId interface{}) (lib.VidT, error) {
	var vid lib.VidT
	switch v := vlanId.(type) {
	case *oc.Interface_Ethernet_SwitchedVlan_TrunkVlans_Union_Uint16:
		vid = lib.VidT(v.Uint16)
	case oc.Interface_Ethernet_SwitchedVlan_TrunkVlans_Union_Uint16:
		vid = lib.VidT(v.Uint16)
	case *uint16:
		vid = lib.VidT(*v)
	case uint16:
		vid = lib.VidT(v)
	default:
		return 0, fmt.Errorf("Cannot convert %v to any of [uint16, Interface_Ethernet_SwitchedVlan_TrunkVlans_Union], unsupported union type, got: %T", v, v)
	}

	return vid, nil
}

func (this *ConfigMngrT) validateDeleteAccessVlanEthIntfChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ifname := changeItem.Change.Path[cmd.VlanEthIfnamePathItemIdxC]
	if !this.isEthIntfAvailable(ifname) {
		return fmt.Errorf("Ethernet interface %s is not available", ifname)
	}

	vid, err := extractVlanIdFromInterface(changeItem.Change.From)
	if err != nil {
		return err
	}

	vlanMode, err := this.transConfigLookupTbl.getVlanModeEthIntf(ifname)
	if vlanMode != oc.OpenconfigVlan_VlanModeType_ACCESS {
		return fmt.Errorf("Deletion of access VLAN %d from Ethernet interface %s is disallowed if VLAN interface mode is not access. Current mode: %v", vid, ifname, vlanMode)
	}

	log.Infof("Requested delete access VLAN %d from Ethernet interface %s", vid, ifname)
	deleteAccessVlanEthIntfCmd := cmd.NewDeleteAccessVlanEthIntfCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForDeleteAccessVlanFromEthIntf(ifname, vid); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from interface %s:\n%s",
			deleteAccessVlanEthIntfCmd.GetName(), ifname, err)
	}

	if this.transHasBeenStarted {
		if err = this.appendCmdToTransactionByIfname(ifname, deleteAccessVlanEthIntfCmd, deleteEthIntfFromAccessVlanC); err != nil {
			return err
		}

		if err := this.transConfigLookupTbl.deleteAccessVlanEthIntf(ifname, vid); err != nil {
			return err
		}

		// Update type carries info about old and new access VLAN ID. Let's create new change item
		// in order to process new native VLAN it by SetAccessVlanEthIntfCmd
		if (changeItem.Change.Type == diff.UPDATE) && (changeItem.Change.To != nil) {
			var newChange diff.Change
			copier.Copy(newChange, changeItem.Change)
			newChange.Type = diff.CREATE
			newChange.From = nil
			changelog.Changes = append(changelog.Changes, NewDiffChangeMgmtT(&newChange))
		}

		changeItem.MarkAsProcessed()
	}

	return nil
}

func (this *ConfigMngrT) validateSetNativeVlanEthIntfChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ifname := changeItem.Change.Path[cmd.VlanEthIfnamePathItemIdxC]
	if !this.isEthIntfAvailable(ifname) {
		return fmt.Errorf("Ethernet interface %s is not available", ifname)
	}

	vid, err := extractVlanIdFromInterface(changeItem.Change.To)
	if err != nil {
		return err
	}
	vlanModeChange, exists := findsetVlanModeEthIntfChange(changelog)
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
		if err := this.appendCmdToTransactionByIfname(ifname, setNativeVlanEthIntfCmd, setNativeVlanForEthIntfC); err != nil {
			return err
		}

		this.transConfigLookupTbl.setNativeVlanEthIntf(ifname, vid)
		changeItem.MarkAsProcessed()
	}

	return nil
}

func (this *ConfigMngrT) validateDeleteNativeVlanEthIntfChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ifname := changeItem.Change.Path[cmd.VlanEthIfnamePathItemIdxC]
	if !this.isEthIntfAvailable(ifname) {
		return fmt.Errorf("Ethernet interface %s is not available", ifname)
	}

	vid, err := extractVlanIdFromInterface(changeItem.Change.From)
	if err != nil {
		return err
	}

	vlanMode, err := this.transConfigLookupTbl.getVlanModeEthIntf(ifname)
	if vlanMode != oc.OpenconfigVlan_VlanModeType_TRUNK {
		return fmt.Errorf("Deletion of native VLAN %d from Ethernet interface %s is disallowed if VLAN interface mode is not trunk. Current mode: %v", vid, ifname, vlanMode)
	}

	log.Infof("Requested delete native VLAN %d from Ethernet interface %s", vid, ifname)
	deleteNativeVlanEthIntfCmd := cmd.NewDeleteNativeVlanEthIntfCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForDeleteNativeVlanFromEthIntf(ifname, vid); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from interface %s:\n%s",
			deleteNativeVlanEthIntfCmd.GetName(), ifname, err)
	}

	if this.transHasBeenStarted {
		if err = this.appendCmdToTransactionByIfname(ifname, deleteNativeVlanEthIntfCmd, deleteEthIntfFromNativeVlanC); err != nil {
			return err
		}

		if err := this.transConfigLookupTbl.deleteNativeVlanEthIntf(ifname, vid); err != nil {
			return err
		}

		// Update type carries info about old and new native VLAN ID. Let's create new change item
		// in order to process new native VLAN it by SetNativeVlanEthIntfCmd
		if (changeItem.Change.Type == diff.UPDATE) && (changeItem.Change.To != nil) {
			var newChange diff.Change
			copier.Copy(newChange, changeItem.Change)
			newChange.Type = diff.CREATE
			newChange.From = nil
			changelog.Changes = append(changelog.Changes, NewDiffChangeMgmtT(&newChange))
		}

		changeItem.MarkAsProcessed()
	}

	return nil
}

func (this *ConfigMngrT) validateDeleteTrunkVlanEthIntfChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ifname := changeItem.Change.Path[cmd.VlanEthIfnamePathItemIdxC]
	if !this.isEthIntfAvailable(ifname) {
		return fmt.Errorf("Ethernet interface %s is not available", ifname)
	}

	vid, err := extractVlanIdFromInterface(changeItem.Change.From)
	if err != nil {
		return err
	}

	vlanMode, err := this.transConfigLookupTbl.getVlanModeEthIntf(ifname)
	if vlanMode != oc.OpenconfigVlan_VlanModeType_TRUNK {
		return fmt.Errorf("Deletion of trunk VLAN %d from Ethernet interface %s is disallowed if VLAN interface mode is not trunk. Current mode: %v", vid, ifname, vlanMode)
	}

	log.Infof("Requested delete trunk VLAN %d from Ethernet interface %s", vid, ifname)
	deleteTrunkVlanEthIntfCmd := cmd.NewDeleteTrunkVlanEthIntfCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForDeleteTrunkVlanFromEthIntf(ifname, vid); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from interface %s:\n%s",
			deleteTrunkVlanEthIntfCmd.GetName(), ifname, err)
	}

	if this.transHasBeenStarted {
		if err = this.appendCmdToTransactionByIfname(ifname, deleteTrunkVlanEthIntfCmd, deleteEthIntfFromTrunkVlanC); err != nil {
			return err
		}

		if err := this.transConfigLookupTbl.deleteTrunkVlanEthIntf(ifname, vid); err != nil {
			return err
		}

		// Update type carries info about old and new trunk VLAN ID. Let's create new change item
		// in order to process new native VLAN it by SetTrunkVlanEthIntfCmd
		if (changeItem.Change.Type == diff.UPDATE) && (changeItem.Change.To != nil) {
			var newChange diff.Change
			copier.Copy(newChange, changeItem.Change)
			newChange.Type = diff.CREATE
			newChange.From = nil
			// Let's drop "Uint16"/"String"
			newChange.Path = newChange.Path[:len(newChange.Path)-1]
			changelog.Changes = append(changelog.Changes, NewDiffChangeMgmtT(&newChange))
		}

		changeItem.MarkAsProcessed()
	}

	return nil
}

func (this *ConfigMngrT) processDeleteAccessVlanEthIntfFromChangelog(changelog *DiffChangelogMgmtT) (int, error) {
	var count int = 0
	for {
		// Repeat till there is not any change related to delete native VLAN from Ethernet interface
		if change, exists := findDeleteAccessVlanEthIntfChange(changelog); exists {
			if err := this.validateDeleteAccessVlanEthIntfChange(change, changelog); err != nil {
				return 0, err
			}

			count++
		} else {
			break
		}
	}

	return count, nil
}

func (this *ConfigMngrT) processSetNativeVlanEthIntfFromChangelog(changelog *DiffChangelogMgmtT) (int, error) {
	var count int = 0
	for {
		// Repeat till there is not any change related to set native VLAN for Ethernet interface
		if change, exists := findSetNativeVlanEthIntfChange(changelog); exists {
			if err := this.validateSetNativeVlanEthIntfChange(change, changelog); err != nil {
				return 0, err
			}

			count++
		} else {
			break
		}
	}

	return count, nil
}

func (this *ConfigMngrT) processDeleteNativeVlanEthIntfFromChangelog(changelog *DiffChangelogMgmtT) (int, error) {
	var count int = 0
	for {
		// Repeat till there is not any change related to delete native VLAN from Ethernet interface
		if change, exists := findDeleteNativeVlanEthIntfChange(changelog); exists {
			if err := this.validateDeleteNativeVlanEthIntfChange(change, changelog); err != nil {
				return 0, err
			}

			count++
		} else {
			break
		}
	}

	return count, nil
}

func (this *ConfigMngrT) processDeleteTrunkVlanEthIntfFromChangelog(changelog *DiffChangelogMgmtT) (int, error) {
	var count int = 0
	for {
		// Repeat till there is not any change related to delete trunk VLANs from Ethernet interface
		if change, exists := findDeleteTrunkVlanEthIntfChange(changelog); exists {
			if err := this.validateDeleteTrunkVlanEthIntfChange(change, changelog); err != nil {
				return 0, err
			}

			count++
		} else {
			break
		}
	}

	return count, nil
}
