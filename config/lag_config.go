package config

import (
	"fmt"
	cmd "opennos-mgmt/config/command"

	log "github.com/golang/glog"
	"github.com/jinzhu/copier"
	"github.com/r3labs/diff"
)

func isChangedLagIntfMember(change *diff.Change) bool {
	if len(change.Path) != cmd.LagIntfMemberPathItemsCountC {
		return false
	}

	if (change.Path[cmd.LagIntfInterfacePathItemIdxC] == cmd.LagIntfInterfacePathItemC) && (change.Path[cmd.LagIntfMemberEthernetPathItemIdxC] == cmd.LagIntfMemberEthernetPathItemC) && (change.Path[cmd.LagIntfMemberAggIdPathItemIdxC] == cmd.LagIntfMemberAggIdPathItemC) {
		return true
	}

	return false
}

func isChangedLagIntf(change *diff.Change) bool {
	if len(change.Path) != cmd.LagIntfPathItemsCountC {
		return false
	}

	if (change.Path[cmd.LagIntfInterfacePathItemIdxC] == cmd.LagIntfInterfacePathItemC) && (change.Path[cmd.LagIntfNamePathItemIdxC] == cmd.LagIntfNamePathItemC) {
		return true
	}

	return false
}

func (this *ConfigMngrT) validateSetLagIntfMemberChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ifname := changeItem.Change.Path[cmd.LagIntfIfnamePathItemIdxC]
	if !this.isEthIntfAvailable(ifname) {
		return fmt.Errorf("Ethernet interface %s is not available", ifname)
	}

	lagName, err := convertInterfaceIntoString(changeItem.Change.To)
	if err != nil {
		return err
	}

	log.Infof("Requested add Ethernet interface %s as LAG member %s", ifname, lagName)
	setLagIntfMemberCmd := cmd.NewSetLagIntfMemberCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForSetLagIntfMember(lagName, ifname); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from interface %s:\n%s",
			setLagIntfMemberCmd.GetName(), ifname, err)
	}

	if this.transHasBeenStarted {
		if err := this.appendCmdToTransaction(ifname, setLagIntfMemberCmd, setLagIntfMemberC); err != nil {
			return err
		}
	}

	if err := this.transConfigLookupTbl.setLagIntfMember(lagName, ifname); err != nil {
		return err
	}

	changeItem.MarkAsProcessed()

	return nil
}

func (this *ConfigMngrT) validateDeleteLagIntfMemberChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ifname := changeItem.Change.Path[cmd.LagIntfIfnamePathItemIdxC]
	if !this.isEthIntfAvailable(ifname) {
		return fmt.Errorf("Ethernet interface %s is not available", ifname)
	}

	lagName, err := convertInterfaceIntoString(changeItem.Change.From)
	if err != nil {
		return err
	}
	log.Infof("Requested remove Ethernet interface %s from LAG member %s", ifname, lagName)
	deleteLagIntfMemberCmd := cmd.NewDeleteLagIntfMemberCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForDeleteLagIntfMember(lagName, ifname); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from interface %s:\n%s",
			deleteLagIntfMemberCmd.GetName(), ifname, err)
	}

	if this.transHasBeenStarted {
		if err := this.appendCmdToTransaction(ifname, deleteLagIntfMemberCmd, deleteLagIntfMemberC); err != nil {
			return err
		}
	}

	if err := this.transConfigLookupTbl.deleteLagIntfMember(lagName, ifname); err != nil {
		return err
	}

	// Update type carries info about old and new LAG. Let's create new change item
	// in order to process new LAG by SetLagIntfMemberCmd
	if (changeItem.Change.Type == diff.UPDATE) && (changeItem.Change.To != nil) {
		newLagName, err := convertInterfaceIntoString(changeItem.Change.To)
		if err != nil {
			return err
		}
		if len(newLagName) > 0 {
			var newChange diff.Change
			copier.Copy(&newChange, changeItem.Change)
			newChange.Type = diff.CREATE
			newChange.From = nil
			changelog.Changes = append(changelog.Changes, NewDiffChangeMgmtT(&newChange))
		}
	}

	changeItem.MarkAsProcessed()

	return nil
}

func (this *ConfigMngrT) validateSetLagIntfChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	lagName, err := convertInterfaceIntoString(changeItem.Change.To)
	if err != nil {
		return err
	}

	log.Infof("Requested set LAG interface %s", lagName)
	setLagIntfCmd := cmd.NewSetLagIntfCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForSetLagIntf(lagName); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from LAG interface %s:\n%s",
			setLagIntfCmd.GetName(), lagName, err)
	}

	if this.transHasBeenStarted {
		if err := this.appendCmdToTransaction(lagName, setLagIntfCmd, setLagIntfC); err != nil {
			return err
		}
	}

	if err := this.transConfigLookupTbl.setLagIntf(lagName); err != nil {
		return err
	}

	changeItem.MarkAsProcessed()

	return nil
}

func (this *ConfigMngrT) validateDeleteLagIntfChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	lagName, err := convertInterfaceIntoString(changeItem.Change.From)
	if err != nil {
		return err
	}

	log.Infof("Requested delete LAG interface %s", lagName)
	deleteLagIntfCmd := cmd.NewDeleteLagIntfCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForDeleteLagIntf(lagName); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from LAG interface %s:\n%s",
			deleteLagIntfCmd.GetName(), lagName, err)
	}

	if this.transHasBeenStarted {
		if err := this.appendCmdToTransaction(lagName, deleteLagIntfCmd, deleteLagIntfC); err != nil {
			return err
		}
	}

	if err := this.transConfigLookupTbl.deleteLagIntf(lagName); err != nil {
		return err
	}

	changeItem.MarkAsProcessed()

	return nil
}

func findSetLagIntfMemberChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type != diff.DELETE {
				if isChangedLagIntfMember(ch.Change) {
					if ch.Change.To != nil {
						return ch, true
					}
				}
			}
		}
	}

	return nil, false
}

func findDeleteLagIntfMemberChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type == diff.UPDATE {
				if isChangedLagIntfMember(ch.Change) {
					if ch.Change.From != nil {
						return ch, true
					}
				}
			}
		}
	}

	return nil, false
}

func findSetLagIntfChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type == diff.CREATE {
				if isChangedLagIntf(ch.Change) {
					if ch.Change.To != nil {
						return ch, true
					}
				}
			}
		}
	}

	return nil, false
}

func findDeleteLagIntfChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type == diff.DELETE {
				if isChangedLagIntf(ch.Change) {
					if ch.Change.From != nil {
						return ch, true
					}
				}
			}
		}
	}

	return nil, false
}

func (this *ConfigMngrT) processSetLagIntfMemberFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to set LAG interface member
		if change, exists := findSetLagIntfMemberChange(changelog); exists {
			if err := this.validateSetLagIntfMemberChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (this *ConfigMngrT) processDeleteLagIntfMemberFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to delete LAG interface member
		if change, exists := findDeleteLagIntfMemberChange(changelog); exists {
			if err := this.validateDeleteLagIntfMemberChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (this *ConfigMngrT) processSetLagIntfFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to set LAG interface
		if change, exists := findSetLagIntfChange(changelog); exists {
			if err := this.validateSetLagIntfChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (this *ConfigMngrT) processDeleteLagIntfFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to delete LAG interface
		if change, exists := findDeleteLagIntfChange(changelog); exists {
			if err := this.validateDeleteLagIntfChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}
