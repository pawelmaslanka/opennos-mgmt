package config

import (
	"fmt"
	cmd "opennos-mgmt/config/command"
	"opennos-mgmt/gnmi/modeldata/oc"

	log "github.com/golang/glog"
	"github.com/jinzhu/copier"
	"github.com/r3labs/diff"
)

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
		if change.From != nil {
			return true
		}
	}

	return false
}

func isChangedNativeVlan(change *diff.Change) bool {
	if len(change.Path) != cmd.NativeVlanEthPathItemsCountC {
		return false
	}

	if (change.Path[cmd.VlanEthIntfPathItemIdxC] == cmd.VlanEthIntfPathItemC) && (change.Path[cmd.VlanEthEthernetPathItemIdxC] == cmd.VlanEthEthernetPathItemC) && (change.Path[cmd.VlanEthSwVlanPathItemIdxC] == cmd.VlanEthSwVlanPathItemC) && (change.Path[cmd.VlanEthNativeVlanPathItemIdxC] == cmd.VlanEthNativeVlanPathItemC) {
		if change.From != nil {
			return true
		}
	}

	return false
}

func isChangedTrunkVlan(change *diff.Change) bool {
	if (len(change.Path) != cmd.TrunkVlanEthPathItemsCountC) && (len(change.Path) != cmd.TrunkVlanEthPathItemsCountIfUpdateC) {
		return false
	}

	if (change.Path[cmd.VlanEthIntfPathItemIdxC] == cmd.VlanEthIntfPathItemC) && (change.Path[cmd.VlanEthEthernetPathItemIdxC] == cmd.VlanEthEthernetPathItemC) && (change.Path[cmd.VlanEthSwVlanPathItemIdxC] == cmd.VlanEthSwVlanPathItemC) && (change.Path[cmd.VlanEthTrunkVlanPathItemIdxC] == cmd.VlanEthTrunkVlanPathItemC) {
		if change.From != nil {
			return true
		}
	}

	return false
}

func doFindSetVlanModeEthIntfChange(changelog *DiffChangelogMgmtT) (*DiffChangeMgmtT, bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type != diff.DELETE {
				if isChangedVlanMode(ch.Change) {
					return ch, true
				}
			}
		}
	}

	return nil, false
}

func doFindSetAccessVlanEthIntfChange(changelog *DiffChangelogMgmtT) (*DiffChangeMgmtT, bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type == diff.CREATE {
				if isChangedAccessVlan(ch.Change) {
					return ch, true
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

func doFindSetTrunkVlanEthIntfChange(changelog *DiffChangelogMgmtT) (*DiffChangeMgmtT, bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type == diff.CREATE {
				if isChangedTrunkVlan(ch.Change) {
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
		if err := this.appendCmdToTransaction(ifname, setVlanModeEthIntfCmd, setVlanModeForEthIntfC); err != nil {
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

	vid, err := convertInterfaceIntoVlanId(changeItem.Change.To)
	if err != nil {
		return err
	}

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
		if err := this.appendCmdToTransaction(ifname, setAccessVlanEthIntfCmd, setAccessVlanForEthIntfC); err != nil {
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

	vid, err := convertInterfaceIntoVlanId(changeItem.Change.From)
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
		if err = this.appendCmdToTransaction(ifname, deleteAccessVlanEthIntfCmd, deleteEthIntfFromAccessVlanC); err != nil {
			return err
		}
	}

	if err := this.transConfigLookupTbl.deleteAccessVlanEthIntf(ifname, vid); err != nil {
		return err
	}

	// Update type carries info about old and new access VLAN ID. Let's create new change item
	// in order to process new native VLAN it by SetAccessVlanEthIntfCmd
	if (changeItem.Change.Type == diff.UPDATE) && (changeItem.Change.To != nil) {
		var newChange diff.Change
		copier.Copy(&newChange, changeItem.Change)
		newChange.Type = diff.CREATE
		newChange.From = nil
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

	vid, err := convertInterfaceIntoVlanId(changeItem.Change.To)
	if err != nil {
		return err
	}

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
		if err := this.appendCmdToTransaction(ifname, setNativeVlanEthIntfCmd, setNativeVlanForEthIntfC); err != nil {
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

	vid, err := convertInterfaceIntoVlanId(changeItem.Change.From)
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
		if err = this.appendCmdToTransaction(ifname, deleteNativeVlanEthIntfCmd, deleteEthIntfFromNativeVlanC); err != nil {
			return err
		}
	}

	if err := this.transConfigLookupTbl.deleteNativeVlanEthIntf(ifname, vid); err != nil {
		return err
	}

	// Update type carries info about old and new native VLAN ID. Let's create new change item
	// in order to process new native VLAN it by SetNativeVlanEthIntfCmd
	if (changeItem.Change.Type == diff.UPDATE) && (changeItem.Change.To != nil) {
		var newChange diff.Change
		copier.Copy(&newChange, changeItem.Change)
		newChange.Type = diff.CREATE
		newChange.From = nil
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

	vid, err := convertInterfaceIntoVlanId(changeItem.Change.To)
	if err != nil {
		return err
	}

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
		if err = this.appendCmdToTransaction(ifname, setTrunkVlanEthIntfCmd, setTrunkVlanForEthIntfC); err != nil {
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

	vid, err := convertInterfaceIntoVlanId(changeItem.Change.From)
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
		if err = this.appendCmdToTransaction(ifname, deleteTrunkVlanEthIntfCmd, deleteEthIntfFromTrunkVlanC); err != nil {
			return err
		}
	}

	if err := this.transConfigLookupTbl.deleteTrunkVlanEthIntf(ifname, vid); err != nil {
		return err
	}

	// Update type carries info about old and new trunk VLAN ID. Let's create new change item
	// in order to process new native VLAN it by SetTrunkVlanEthIntfCmd
	if (changeItem.Change.Type == diff.UPDATE) && (changeItem.Change.To != nil) {
		var newChange diff.Change
		copier.Copy(&newChange, changeItem.Change)
		newChange.Type = diff.CREATE
		newChange.From = nil
		// Let's drop "Uint16"/"String"
		newChange.Path = newChange.Path[:len(newChange.Path)-1]
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
