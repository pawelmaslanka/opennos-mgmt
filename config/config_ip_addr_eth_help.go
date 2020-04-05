package config

import (
	"errors"
	"fmt"
	lib "golibext"
	cmd "opennos-mgmt/config/command"

	log "github.com/golang/glog"
	"github.com/r3labs/diff"
)

func (this *ConfigMngrT) getIpv4AddrEthSubintfIpFromChangelog(ifname string, changelog *DiffChangelogMgmtT, goingToBeDeleted bool) (string, error) {
	var err error = nil
	var ip string
	for _, ch := range changelog.Changes {
		if this.IsChangedIpv4AddrEthSubintfIp(ch.Change) {
			log.Infof("Found changing IPv4 address request too:\n%+v", ch.Change)
			if ch.Change.Path[cmd.Ipv4AddrEthIfnamePathItemIdxC] == ifname {
				ip = fmt.Sprintf("%s", *ch.Change.To.(*string))
				break
			}
		}
	}

	if !lib.IsValidIpv4AddrIp(ip) {
		err = fmt.Errorf("IPv4 address (%s) is invalid", ip)
	}

	return ip, err
}

func (this *ConfigMngrT) getIpv4AddrEthSubintfPrfxLenFromChangelog(ifname string, changelog *DiffChangelogMgmtT) (uint8, error) {
	var err error = nil
	var prfxLen uint8
	for _, ch := range changelog.Changes {
		if this.IsChangedIpv4AddrEthSubintfPrfxLen(ch.Change) {
			log.Infof("Found changing IPv4 prefix len request too:\n%+v", ch.Change)
			if ch.Change.Path[cmd.Ipv4AddrEthIfnamePathItemIdxC] == ifname {
				prfxLen = *ch.Change.To.(*uint8)
				break
			}
		}
	}

	if !lib.IsValidIpv4AddrPrfxLen(prfxLen) {
		err = fmt.Errorf("IPv4 prefix length (%d) is invalid", prfxLen)
	}

	return uint8(prfxLen), err
}

func (this *ConfigMngrT) findSetIpv4AddrEthSubintfIpChangeFromChangelog(ifname string, changelog *DiffChangelogMgmtT) (*DiffChangeMgmtT, error) {
	var ip *DiffChangeMgmtT = nil
	for _, ch := range changelog.Changes {
		if ch.Change.Type != diff.DELETE {
			if this.IsChangedIpv4AddrEthSubintfIp(ch.Change) {
				log.Infof("Found changing IPv4 request too:\n%+v", ch.Change)
				if ch.Change.Path[cmd.Ipv4AddrEthIfnamePathItemIdxC] == ifname {
					ip = ch
					break
				}
			}
		}
	}

	if ip == nil {
		return nil, errors.New("Not found IPv4 address dependency")
	}

	if !lib.IsValidIpv4AddrIp(ip.Change.To.(string)) {
		return nil, fmt.Errorf("IPv4 address (%d) is invalid", ip.Change.To.(string))
	}

	return ip, nil
}

func (this *ConfigMngrT) findDeleteIpv4AddrEthSubintfIpChangeFromChangelog(ifname string, changelog *DiffChangelogMgmtT) (*DiffChangeMgmtT, error) {
	for _, ch := range changelog.Changes {
		if ch.Change.Type == diff.DELETE {
			if this.IsChangedIpv4AddrEthSubintfIp(ch.Change) {
				log.Infof("Found change IPv4 address request too:\n%+v", ch.Change)
				if ch.Change.Path[cmd.Ipv4AddrEthIfnamePathItemIdxC] == ifname {
					return ch, nil
				}
			}
		}
	}

	return nil, errors.New("Not found IPv4 address dependency")
}

func (this *ConfigMngrT) findSetIpv4AddrEthSubintfPrfxLenChangeFromChangelog(ifname string, changelog *DiffChangelogMgmtT) (*DiffChangeMgmtT, error) {
	var prfxLenChange *DiffChangeMgmtT = nil
	for _, ch := range changelog.Changes {
		if ch.Change.Type != diff.DELETE {
			if this.IsChangedIpv4AddrEthSubintfPrfxLen(ch.Change) {
				log.Infof("Found changing IPv4 prefix len request too:\n%+v", ch.Change)
				if ch.Change.Path[cmd.Ipv4AddrEthIfnamePathItemIdxC] == ifname {
					prfxLenChange = ch
					break
				}
			}
		}
	}

	if prfxLenChange == nil {
		return nil, errors.New("Not found IPv4 address prefix length")
	}

	if !lib.IsValidIpv4AddrPrfxLen(*prfxLenChange.Change.To.(*uint8)) {
		return nil, fmt.Errorf("IPv4 prefix length (%d) is invalid", *prfxLenChange.Change.To.(*uint8))
	}

	return prfxLenChange, nil
}

func (this *ConfigMngrT) findDeleteIpv4AddrEthSubintfPrfxLenChangeFromChangelog(ifname string, changelog *DiffChangelogMgmtT) (*DiffChangeMgmtT, error) {
	for _, ch := range changelog.Changes {
		if ch.Change.Type == diff.DELETE {
			if this.IsChangedIpv4AddrEthSubintfPrfxLen(ch.Change) {
				log.Infof("Found changing IPv4 prefix len request too:\n%+v", ch.Change)
				if ch.Change.Path[cmd.Ipv4AddrEthIfnamePathItemIdxC] == ifname {
					return ch, nil
				}
			}
		}
	}

	return nil, errors.New("Not found IPv4 address prefix length")
}

func (this *ConfigMngrT) FindSetIpv4AddrEthSubintfIp(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	findDeleteOperation := false
	return this.doFindIpv4AddrEthSubintfIp(changelog, findDeleteOperation)
}

func (this *ConfigMngrT) FindDeleteIpv4AddrEthSubintfIp(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	findDeleteOperation := true
	return this.doFindIpv4AddrEthSubintfIp(changelog, findDeleteOperation)
}

func (this *ConfigMngrT) doFindIpv4AddrEthSubintfIp(changelog *DiffChangelogMgmtT, findDeleteOperation bool) (change *DiffChangeMgmtT, exists bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			isDeleteOperation := ch.Change.Type == diff.DELETE
			if isDeleteOperation == findDeleteOperation {
				if len(ch.Change.Path) == cmd.Ipv4AddrEthPathItemsCountC {
					if (ch.Change.Path[cmd.Ipv4AddrEthIntfPathItemIdxC] == cmd.Ipv4AddrEthIntfPathItemC) && (ch.Change.Path[cmd.Ipv4AddrEthSubintfPathItemIdxC] == cmd.Ipv4AddrEthSubintfPathItemC) && (ch.Change.Path[cmd.Ipv4AddrEthSubintfIpv4PathItemIdxC] == cmd.Ipv4AddrEthSubintfIpv4PathItemC) && (ch.Change.Path[cmd.Ipv4AddrEthSubintfIpv4AddrPathItemIdxC] == cmd.Ipv4AddrEthSubintfIpv4AddrPathItemC) && (ch.Change.Path[cmd.Ipv4AddrEthSubintfIpv4AddrPartIpPathItemIdxC] == cmd.Ipv4AddrEthSubintfIpv4AddrPartIpPathItemC) {
						return ch, true
					}
				}
			}
		}
	}

	return nil, false
}

func (this *ConfigMngrT) IsChangedIpv4AddrEthSubintfIp(change *diff.Change) bool {
	if len(change.Path) < cmd.Ipv4AddrEthPathItemsCountC {
		return false
	}

	if (change.Path[cmd.Ipv4AddrEthIntfPathItemIdxC] != cmd.Ipv4AddrEthIntfPathItemC) || (change.Path[cmd.Ipv4AddrEthSubintfPathItemIdxC] != cmd.Ipv4AddrEthSubintfPathItemC) || (change.Path[cmd.Ipv4AddrEthSubintfIpv4PathItemIdxC] != cmd.Ipv4AddrEthSubintfIpv4PathItemC) || (change.Path[cmd.Ipv4AddrEthSubintfIpv4AddrPathItemIdxC] != cmd.Ipv4AddrEthSubintfIpv4AddrPathItemC) || (change.Path[cmd.Ipv4AddrEthSubintfIpv4AddrPartIpPathItemIdxC] != cmd.Ipv4AddrEthSubintfIpv4AddrPartIpPathItemC) {
		return false
	}

	return true
}

func (this *ConfigMngrT) IsChangedIpv4AddrEthSubintfPrfxLen(change *diff.Change) bool {
	if len(change.Path) < cmd.Ipv4AddrEthPathItemsCountC {
		return false
	}

	if (change.Path[cmd.Ipv4AddrEthIntfPathItemIdxC] != cmd.Ipv4AddrEthIntfPathItemC) || (change.Path[cmd.Ipv4AddrEthSubintfPathItemIdxC] != cmd.Ipv4AddrEthSubintfPathItemC) || (change.Path[cmd.Ipv4AddrEthSubintfIpv4PathItemIdxC] != cmd.Ipv4AddrEthSubintfIpv4PathItemC) || (change.Path[cmd.Ipv4AddrEthSubintfIpv4AddrPathItemIdxC] != cmd.Ipv4AddrEthSubintfIpv4AddrPathItemC) || (change.Path[cmd.Ipv4AddrEthSubintfIpv4AddrPartPrfxLenPathItemIdxC] != cmd.Ipv4AddrEthSubintfIpv4AddrPartPrfxLenPathItemC) {
		return false
	}

	return true
}

func (this *ConfigMngrT) appendSetIpv4AddrEthIntfCmdToTransaction(ifname string, cmdSet *cmd.SetIpv4AddrEthIntfCmdT) error {
	setIpv4AddrForEthIntfCmds := this.cmdByIfname[setIpv4AddrForEthIntfC]
	for _, setIpv4AddrForEthIntfCmd := range setIpv4AddrForEthIntfCmds {
		if setIpv4AddrForEthIntfCmd.Equals(cmdSet) {
			return fmt.Errorf("Command %q already exists in transaction", cmdSet.GetName())
		}
	}

	log.Infof("Append command %q to transaction", cmdSet.GetName())

	setIpv4AddrForEthIntfCmds[ifname] = cmdSet
	this.addCmdToListTrans(cmdSet)
	return nil
}

func (this *ConfigMngrT) appendDeleteIpv4AddrEthIntfCmdToTransaction(ifname string, cmdDelete *cmd.DeleteIpv4AddrEthIntfCmdT) error {
	deleteIpv4AddrFromEthIntfCmds := this.cmdByIfname[deleteIpv4AddrFromEthIntfC]
	for _, deleteIpv4AddrFromEthIntfCmd := range deleteIpv4AddrFromEthIntfCmds {
		if deleteIpv4AddrFromEthIntfCmd.Equals(cmdDelete) {
			return fmt.Errorf("Command %q already exists in transaction", cmdDelete.GetName())
		}
	}

	log.Infof("Append command %q to transaction", cmdDelete.GetName())

	deleteIpv4AddrFromEthIntfCmds[ifname] = cmdDelete
	this.addCmdToListTrans(cmdDelete)
	return nil
}

func (this *ConfigMngrT) validateSetIpv4AddrEthIntf(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ifname := changeItem.Change.Path[cmd.Ipv4AddrEthIfnamePathItemIdxC]
	if !this.isEthIntfAvailable(ifname) {
		if !this.isEthIntfGoingToBeAvailableAfterPortBreakout(ifname) {
			return fmt.Errorf("Ethernet interface %s is unrecognized", ifname)
		}
	}

	var ipChangeItem *DiffChangeMgmtT
	var prfxLenChangeItem *DiffChangeMgmtT
	var err error
	// Check if there is change of IP
	if changeItem.Change.Path[cmd.Ipv4AddrEthSubintfIpv4AddrPartIpPathItemIdxC] == cmd.Ipv4AddrEthSubintfIpv4AddrPartIpPathItemC {
		prfxLenChangeItem, err = this.findSetIpv4AddrEthSubintfPrfxLenChangeFromChangelog(ifname, changelog)
		if err != nil {
			return err
		}

		ipChangeItem = changeItem
	} else if changeItem.Change.Path[cmd.Ipv4AddrEthSubintfIpv4AddrPartPrfxLenPathItemIdxC] == cmd.Ipv4AddrEthSubintfIpv4AddrPartPrfxLenPathItemC {
		ipChangeItem, err = this.findSetIpv4AddrEthSubintfIpChangeFromChangelog(ifname, changelog)
		if err != nil {
			return err
		}

		prfxLenChangeItem = changeItem
	} else {
		return fmt.Errorf("Unable to get IPv4 address change")
	}

	if changeItem.Change.Type == diff.UPDATE {
		return errors.New("Unexpected UPDATE request for change IPv4 address")
	}

	cidr := fmt.Sprintf("%s/%d", *ipChangeItem.Change.To.(*string), *prfxLenChangeItem.Change.To.(*uint8))
	log.Infof("Requested set IPv4 address %s for Ethernet interface %s", cidr, ifname)
	setIpv4AddrEthIntfCmd := cmd.NewSetIpv4AddrEthIntfCmdT(ipChangeItem.Change, prfxLenChangeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForSetIpv4AddrForEthIntf(ifname, cidr); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from IPv4 address %s:\n%s",
			setIpv4AddrEthIntfCmd.GetName(), cidr, err)
	}

	if this.transHasBeenStarted {
		if err = this.appendSetIpv4AddrEthIntfCmdToTransaction(ifname, setIpv4AddrEthIntfCmd); err != nil {
			return err
		}

		if err := this.transConfigLookupTbl.addIpv4AddrEthIntf(ifname, cidr); err != nil {
			return err
		}

		ipChangeItem.MarkAsProcessed()
		prfxLenChangeItem.MarkAsProcessed()
	}

	return nil
}

func (this *ConfigMngrT) validateDeleteIpv4AddrEthIntf(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ifname := changeItem.Change.Path[cmd.Ipv4AddrEthIfnamePathItemIdxC]
	if !this.isEthIntfAvailable(ifname) {
		return fmt.Errorf("Ethernet interface %s is unrecognized", ifname)
	}

	var ipChangeItem *DiffChangeMgmtT
	var prfxLenChangeItem *DiffChangeMgmtT
	var err error

	// Check if there is changing of IP
	if changeItem.Change.Path[cmd.Ipv4AddrEthSubintfIpv4AddrPartIpPathItemIdxC] == cmd.Ipv4AddrEthSubintfIpv4AddrPartIpPathItemC {
		prfxLenChangeItem, err = this.findDeleteIpv4AddrEthSubintfPrfxLenChangeFromChangelog(ifname, changelog)
		if err != nil {
			return err
		}

		ipChangeItem = changeItem
	} else if changeItem.Change.Path[cmd.Ipv4AddrEthSubintfIpv4AddrPartPrfxLenPathItemIdxC] == cmd.Ipv4AddrEthSubintfIpv4AddrPartPrfxLenPathItemC {
		ipChangeItem, err = this.findDeleteIpv4AddrEthSubintfIpChangeFromChangelog(ifname, changelog)
		if err != nil {
			return err
		}

		prfxLenChangeItem = changeItem
	} else {
		return fmt.Errorf("Unable to get IPv4 address change")
	}

	cidr := fmt.Sprintf("%s/%d", *ipChangeItem.Change.From.(*string), *prfxLenChangeItem.Change.From.(*uint8))
	log.Infof("Requested delete IPv4 address %s from Ethernet interface %s", cidr, ifname)
	deleteIpv4AddrEthIntfCmd := cmd.NewDeleteIpv4AddrEthIntfCmdT(ipChangeItem.Change, prfxLenChangeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForDeleteIpv4AddrFromEthIntf(ifname, cidr); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from interface %s:\n%s",
			deleteIpv4AddrEthIntfCmd.GetName(), ifname, err)
	}

	if this.transHasBeenStarted {
		if err = this.appendDeleteIpv4AddrEthIntfCmdToTransaction(ifname, deleteIpv4AddrEthIntfCmd); err != nil {
			return err
		}

		if err := this.transConfigLookupTbl.deleteIpv4AddrEthIntf(ifname, cidr); err != nil {
			return err
		}

		ipChangeItem.MarkAsProcessed()
		prfxLenChangeItem.MarkAsProcessed()
	}

	return nil
}

func (this *ConfigMngrT) processSetIpv4AddrEthIntfFromChangelog(changelog *DiffChangelogMgmtT) (int, error) {
	var count int = 0
	for {
		// Repeat till there is not any change related to delete IPv4 address from Ethernet interface
		if change, exists := this.FindSetIpv4AddrEthSubintfIp(changelog); exists {
			if err := this.validateSetIpv4AddrEthIntf(change, changelog); err != nil {
				return 0, err
			}

			count++
		} else {
			break
		}
	}

	return count, nil
}

func (this *ConfigMngrT) processDeleteIpv4AddrEthIntfFromChangelog(changelog *DiffChangelogMgmtT) (int, error) {
	var count int = 0
	for {
		// Repeat till there is not any change related to delete IPv4 address from Ethernet interface
		if change, exists := this.FindDeleteIpv4AddrEthSubintfIp(changelog); exists {
			if err := this.validateDeleteIpv4AddrEthIntf(change, changelog); err != nil {
				return 0, err
			}

			count++
		} else {
			break
		}
	}

	return count, nil
}
