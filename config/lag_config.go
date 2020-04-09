package config

import (
	"fmt"
	cmd "opennos-mgmt/config/command"

	log "github.com/golang/glog"
	"github.com/r3labs/diff"
)

func isChangedLagIntfMember(change *diff.Change) bool {
	if len(change.Path) != cmd.LagIntfMemberPathItemsCountC {
		return false
	}

	if (change.Path[cmd.LagIntfMemberIntfPathItemIdxC] == cmd.LagIntfMemberIntfPathItemC) && (change.Path[cmd.LagIntfMemberEthernetPathItemIdxC] == cmd.LagIntfMemberEthernetPathItemC) && (change.Path[cmd.LagIntfMemberAggIdPathItemIdxC] == cmd.LagIntfMemberAggIdPathItemC) {
		return true
	}

	return false
}

func (this *ConfigMngrT) validateSetLagIntfMemberChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ifname := changeItem.Change.Path[cmd.LagIntfMemberIfnamePathItemIdxC]
	if !this.isEthIntfAvailable(ifname) {
		return fmt.Errorf("Ethernet interface %s is not available", ifname)
	}

	lagName := changeItem.Change.To.(string)
	log.Infof("Requested add Ethernet interface %s as LAG member %s", ifname, lagName)
	setLagIntfMemberCmd := cmd.NewSetLagIntfMemberCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForSetLagIntfMember(lagName, ifname); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from interface %s:\n%s",
			setLagIntfMemberCmd.GetName(), ifname, err)
	}

	if this.transHasBeenStarted {
		if err := this.appendCmdToTransactionByIfname(ifname, setLagIntfMemberCmd, setLagIntfMemberC); err != nil {
			return err
		}

		if err := this.transConfigLookupTbl.setLagIntfMember(lagName, ifname); err != nil {
			return err
		}

		changeItem.MarkAsProcessed()
	}

	return nil
}

func findSetLagIntfMemberChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type != diff.DELETE {
				if isChangedLagIntfMember(ch.Change) {
					return ch, true
				}
			}
		}
	}

	return nil, false
}

func (this *ConfigMngrT) processSetLagIntfMemberFromChangelog(changelog *DiffChangelogMgmtT) (int, error) {
	var count int = 0
	for {
		// Repeat till there is not any change related to set LAG interface member
		if change, exists := findSetLagIntfMemberChange(changelog); exists {
			if err := this.validateSetLagIntfMemberChange(change, changelog); err != nil {
				return 0, err
			}

			count++
		} else {
			break
		}
	}

	return count, nil
}
