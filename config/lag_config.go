package config

import (
	"fmt"
	cmd "opennos-mgmt/config/command"
	"opennos-mgmt/gnmi/modeldata/oc"
	"opennos-mgmt/utils"

	log "github.com/golang/glog"
	"github.com/jinzhu/copier"
	"github.com/r3labs/diff"
)

const (
	idSetAggIntfMemberNameFmt    = "sm-%s"
	idDeleteAggIntfMemberNameFmt = "sm-%s"
)

func isChangedAggIntfMember(change *diff.Change) bool {
	if len(change.Path) != cmd.AggIntfMemberPathItemsCountC {
		return false
	}

	if (change.Path[cmd.AggIntfInterfacePathItemIdxC] == cmd.AggIntfInterfacePathItemC) && (change.Path[cmd.AggIntfMemberEthernetPathItemIdxC] == cmd.AggIntfMemberEthernetPathItemC) && (change.Path[cmd.AggIntfMemberAggIdPathItemIdxC] == cmd.AggIntfMemberAggIdPathItemC) {
		return true
	}

	return false
}

func isChangedAggIntf(change *diff.Change) bool {
	if len(change.Path) != cmd.AggIntfPathItemsCountC {
		return false
	}

	if (change.Path[cmd.AggIntfInterfacePathItemIdxC] == cmd.AggIntfInterfacePathItemC) && (change.Path[cmd.AggIntfNamePathItemIdxC] == cmd.AggIntfNamePathItemC) {
		return true
	}

	return false
}

func (this *ConfigMngrT) validateSetAggIntfMemberChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ifname := changeItem.Change.Path[cmd.AggIntfIfnamePathItemIdxC]
	if !this.isEthIntfAvailable(ifname) {
		return fmt.Errorf("Ethernet interface %s is not available", ifname)
	}

	aggIfname, err := utils.ConvertGoInterfaceIntoString(changeItem.Change.To)
	if err != nil {
		return err
	}

	log.Infof("Requested add Ethernet interface %s as LAG member %s", ifname, aggIfname)
	setAggIntfMemberCmd := cmd.NewSetAggIntfMemberCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForSetAggIntfMember(aggIfname, ifname); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from interface %s:\n%s",
			setAggIntfMemberCmd.GetName(), ifname, err)
	}

	if this.transHasBeenStarted {
		id := fmt.Sprintf(idSetAggIntfMemberNameFmt, aggIfname)
		if err := this.appendCmdToTransaction(id, setAggIntfMemberCmd, setAggIntfMemberC); err != nil {
			return err
		}
	}

	if err := this.transConfigLookupTbl.setAggIntfMember(aggIfname, ifname); err != nil {
		return err
	}

	changeItem.MarkAsProcessed()

	return nil
}

func (this *ConfigMngrT) validateDeleteAggIntfMemberChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ifname := changeItem.Change.Path[cmd.AggIntfIfnamePathItemIdxC]
	if !this.isEthIntfAvailable(ifname) {
		return fmt.Errorf("Ethernet interface %s is not available", ifname)
	}

	aggIfname, err := utils.ConvertGoInterfaceIntoString(changeItem.Change.From)
	if err != nil {
		return err
	}

	var newChange diff.Change
	needsCreateNewChange := (changeItem.Change.Type == diff.UPDATE) && (changeItem.Change.To != nil)
	if needsCreateNewChange {
		// Update type carries info about old and new LAG. Let's create new change item
		// in order to process new LAG by SetAggIntfMemberCmd
		newAggIfname, err := utils.ConvertGoInterfaceIntoString(changeItem.Change.To)
		if err != nil {
			return err
		}
		if len(newAggIfname) > 0 {
			copier.Copy(&newChange, changeItem.Change)
			newChange.Type = diff.CREATE
			newChange.From = nil
			// Update current change
			changeItem.Change.Type = diff.DELETE
			changeItem.Change.To = nil
		} else {
			needsCreateNewChange = false
		}
	}

	log.Infof("Requested remove Ethernet interface %s from LAG member %s", ifname, aggIfname)
	deleteAggIntfMemberCmd := cmd.NewDeleteAggIntfMemberCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForDeleteAggIntfMember(aggIfname, ifname); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from interface %s:\n%s",
			deleteAggIntfMemberCmd.GetName(), ifname, err)
	}

	if this.transHasBeenStarted {
		id := fmt.Sprintf(idDeleteAggIntfMemberNameFmt, aggIfname)
		if err := this.appendCmdToTransaction(id, deleteAggIntfMemberCmd, deleteAggIntfMemberC); err != nil {
			return err
		}
	}

	if err := this.transConfigLookupTbl.deleteAggIntfMember(aggIfname, ifname); err != nil {
		return err
	}

	if needsCreateNewChange {
		changelog.Changes = append(changelog.Changes, NewDiffChangeMgmtT(&newChange))
	}

	changeItem.MarkAsProcessed()

	return nil
}

func (this *ConfigMngrT) validateSetAggIntfChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	aggIfname, err := utils.ConvertGoInterfaceIntoString(changeItem.Change.To)
	if err != nil {
		return err
	}

	log.Infof("Requested set LAG interface %s", aggIfname)
	setAggIntfCmd := cmd.NewSetAggIntfCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForSetAggIntf(aggIfname); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from LAG interface %s:\n%s",
			setAggIntfCmd.GetName(), aggIfname, err)
	}

	if this.transHasBeenStarted {
		if err := this.appendCmdToTransaction(aggIfname, setAggIntfCmd, setAggIntfC); err != nil {
			return err
		}
	}

	if err := this.transConfigLookupTbl.setAggIntf(aggIfname); err != nil {
		return err
	}

	changeItem.MarkAsProcessed()

	return nil
}

func (this *ConfigMngrT) validateDeleteAggIntfChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	aggIfname, err := utils.ConvertGoInterfaceIntoString(changeItem.Change.From)
	if err != nil {
		return err
	}

	log.Infof("Requested delete LAG interface %s", aggIfname)
	deleteAggIntfCmd := cmd.NewDeleteAggIntfCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForDeleteAggIntf(aggIfname); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from LAG interface %s:\n%s",
			deleteAggIntfCmd.GetName(), aggIfname, err)
	}

	if this.transHasBeenStarted {
		if err := this.appendCmdToTransaction(aggIfname, deleteAggIntfCmd, deleteAggIntfC); err != nil {
			return err
		}
	}

	if err := this.transConfigLookupTbl.deleteAggIntf(aggIfname); err != nil {
		return err
	}

	changeItem.MarkAsProcessed()

	return nil
}

func findSetAggIntfMemberChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type != diff.DELETE {
				if isChangedAggIntfMember(ch.Change) {
					if ch.Change.To != nil {
						return ch, true
					}
				}
			}
		}
	}

	return nil, false
}

func findDeleteAggIntfMemberChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type != diff.CREATE {
				if isChangedAggIntfMember(ch.Change) {
					if ch.Change.From != nil {
						return ch, true
					}
				}
			}
		}
	}

	return nil, false
}

func findSetAggIntfChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type != diff.DELETE {
				if isChangedAggIntf(ch.Change) {
					if ch.Change.To != nil {
						return ch, true
					}
				}
			}
		}
	}

	return nil, false
}

func findDeleteAggIntfChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type != diff.CREATE {
				if isChangedAggIntf(ch.Change) {
					if ch.Change.From != nil {
						return ch, true
					}
				}
			}
		}
	}

	return nil, false
}

func (this *ConfigMngrT) processSetAggIntfMemberFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to set LAG interface member
		if change, exists := findSetAggIntfMemberChange(changelog); exists {
			if err := this.validateSetAggIntfMemberChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (this *ConfigMngrT) processDeleteAggIntfMemberFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to delete LAG interface member
		if change, exists := findDeleteAggIntfMemberChange(changelog); exists {
			if err := this.validateDeleteAggIntfMemberChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (this *ConfigMngrT) processSetAggIntfFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to set LAG interface
		if change, exists := findSetAggIntfChange(changelog); exists {
			if err := this.validateSetAggIntfChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (this *ConfigMngrT) processDeleteAggIntfFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to delete LAG interface
		if change, exists := findDeleteAggIntfChange(changelog); exists {
			if err := this.validateDeleteAggIntfChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (this *ConfigMngrT) setAggIntf(device *oc.Device) error {
	var err error
	for _, aggIfname := range this.configLookupTbl.aggIfnameByIdx {
		var change diff.Change
		change.Type = diff.CREATE
		change.From = nil
		change.To = aggIfname
		change.Path = make([]string, cmd.AggIntfPathItemsCountC)
		change.Path[cmd.AggIntfInterfacePathItemIdxC] = cmd.AggIntfInterfacePathItemC
		change.Path[cmd.AggIntfIfnamePathItemIdxC] = aggIfname
		change.Path[cmd.AggIntfNamePathItemIdxC] = cmd.AggIntfNamePathItemC

		command := cmd.NewSetAggIntfCmdT(&change, this.ethSwitchMgmtClient)
		if err = this.appendCmdToTransaction(aggIfname, command, setAggIntfC); err != nil {
			return err
		}
	}

	return nil
}

func (this *ConfigMngrT) setAggIntfMember(device *oc.Device) error {
	var err error
	for aggIdx, aggIfname := range this.configLookupTbl.aggIfnameByIdx {
		for _, ethIdx := range this.configLookupTbl.ethByAgg[aggIdx].IdxTs() {
			var change diff.Change
			change.Type = diff.CREATE
			change.From = nil
			change.To = aggIfname
			change.Path = make([]string, cmd.AggIntfMemberPathItemsCountC)
			change.Path[cmd.AggIntfInterfacePathItemIdxC] = cmd.AggIntfInterfacePathItemC
			change.Path[cmd.AggIntfIfnamePathItemIdxC] = this.configLookupTbl.ethIfnameByIdx[ethIdx]
			change.Path[cmd.AggIntfMemberEthernetPathItemIdxC] = cmd.AggIntfMemberEthernetPathItemC
			change.Path[cmd.AggIntfMemberAggIdPathItemIdxC] = cmd.AggIntfMemberAggIdPathItemC

			command := cmd.NewSetAggIntfMemberCmdT(&change, this.ethSwitchMgmtClient)
			id := fmt.Sprintf(idSetAggIntfMemberNameFmt, aggIfname)
			if err = this.appendCmdToTransaction(id, command, setAggIntfMemberC); err != nil {
				return err
			}
		}
	}

	return nil
}
