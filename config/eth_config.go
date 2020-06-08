package config

import (
	"fmt"
	cmd "opennos-mgmt/config/command"
	"opennos-mgmt/gnmi/modeldata/oc"
	"opennos-mgmt/utils"
	"strings"

	log "github.com/golang/glog"
	"github.com/r3labs/diff"
)

func isCreatedEthIntf(change *diff.Change) bool {
	if len(change.Path) != cmd.EthIntfPathItemsCountC {
		return false
	}

	if (change.Path[cmd.EthIntfInterfacePathItemIdxC] == cmd.EthIntfInterfacePathItemC) && strings.Contains(change.Path[cmd.EthIntfIfnamePathItemIdxC], "eth") && (change.Path[cmd.EthIntfEthernetPathItemIdxC] == cmd.EthIntfEthernetPathItemC) {
		return true
	}

	return false
}

func isChangedEthIntf(change *diff.Change) bool {
	if len(change.Path) != cmd.EthIntfPathItemsCountC {
		return false
	}

	if (change.Path[cmd.EthIntfInterfacePathItemIdxC] == cmd.EthIntfInterfacePathItemC) && strings.Contains(change.Path[cmd.EthIntfIfnamePathItemIdxC], "eth") && (change.Path[cmd.EthIntfNamePathItemIdxC] == cmd.EthIntfNamePathItemC) {
		return true
	}

	return false
}

func (this *ConfigMngrT) validateSetEthIntfChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ethIfname, err := utils.ConvertGoInterfaceIntoString(changeItem.Change.To)
	if err != nil {
		return err
	}

	log.Infof("Requested set Ethernet interface %s", ethIfname)
	setEthIntfCmd := cmd.NewSetEthIntfCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForSetEthIntf(ethIfname); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from Ethernet interface %s:\n%s",
			setEthIntfCmd.GetName(), ethIfname, err)
	}

	if this.transHasBeenStarted {
		if err := this.appendCmdToTransaction(ethIfname, setEthIntfCmd, setEthIntfC, false); err != nil {
			return err
		}
	}

	if err := this.transConfigLookupTbl.setEthIntf(ethIfname); err != nil {
		return err
	}

	changeItem.MarkAsProcessed()

	return nil
}

func (this *ConfigMngrT) validateDeleteEthIntfChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ethIfname, err := utils.ConvertGoInterfaceIntoString(changeItem.Change.From)
	if err != nil {
		return err
	}

	log.Infof("Requested delete Ethernet interface %s", ethIfname)
	deleteEthIntfCmd := cmd.NewDeleteEthIntfCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForDeleteEthIntf(ethIfname); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from Ethernet interface %s:\n%s",
			deleteEthIntfCmd.GetName(), ethIfname, err)
	}

	if this.transHasBeenStarted {
		if err := this.appendCmdToTransaction(ethIfname, deleteEthIntfCmd, deleteEthIntfC, false); err != nil {
			return err
		}
	}

	if err := this.transConfigLookupTbl.deleteEthIntf(ethIfname); err != nil {
		return err
	}

	changeItem.MarkAsProcessed()

	return nil
}

func findSetEthIntfChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type != diff.DELETE {
				if isChangedEthIntf(ch.Change) {
					if ch.Change.To != nil {
						return ch, true
					}
				}
			}
		}
	}

	return nil, false
}

func findDeleteEthIntfChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type != diff.CREATE {
				if isChangedEthIntf(ch.Change) {
					if ch.Change.From != nil {
						return ch, true
					}
				}
			}
		}
	}

	return nil, false
}

func (this *ConfigMngrT) processSetEthIntfFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to set Ethernet interface
		if change, exists := findSetEthIntfChange(changelog); exists {
			if err := this.validateSetEthIntfChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (this *ConfigMngrT) processDeleteEthIntfFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to delete Ethernet interface
		if change, exists := findDeleteEthIntfChange(changelog); exists {
			if err := this.validateDeleteEthIntfChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (this *ConfigMngrT) setEthIntf(device *oc.Device) error {
	var err error
	for _, ethIfname := range this.configLookupTbl.ethIfnameByIdx {
		var change diff.Change
		change.Type = diff.CREATE
		change.From = nil
		change.To = ethIfname
		change.Path = make([]string, cmd.EthIntfPathItemsCountC)
		change.Path[cmd.EthIntfInterfacePathItemIdxC] = cmd.EthIntfInterfacePathItemC
		change.Path[cmd.EthIntfIfnamePathItemIdxC] = ethIfname
		change.Path[cmd.EthIntfNamePathItemIdxC] = cmd.EthIntfNamePathItemC

		command := cmd.NewSetEthIntfCmdT(&change, this.ethSwitchMgmtClient)
		if err = this.appendCmdToTransaction(ethIfname, command, setEthIntfC, true); err != nil {
			return err
		}
	}

	return nil
}
