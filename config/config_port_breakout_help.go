package config

import (
	"fmt"
	cmd "opennos-mgmt/config/command"
	"opennos-mgmt/gnmi/modeldata/oc"

	log "github.com/golang/glog"
	"github.com/r3labs/diff"
)

func (this *ConfigMngrT) getPortBreakoutChannelSpeedFromChangelog(ifname string, changelog *DiffChangelogMgmtT) (oc.E_OpenconfigIfEthernet_ETHERNET_SPEED, error) {
	var err error = nil
	channelSpeed := oc.OpenconfigIfEthernet_ETHERNET_SPEED_UNSET
	for _, changedItem := range changelog.Changes {
		if this.isChangedPortBreakoutChannelSpeed(changedItem.Change) {
			log.Infof("Found channel speed request too:\n%+v", *changedItem)
			if changedItem.Change.Path[cmd.PortBreakoutIfnamePathItemIdxC] == ifname {
				channelSpeed = changedItem.Change.To.(oc.E_OpenconfigIfEthernet_ETHERNET_SPEED)
				break
			}
		}
	}

	if channelSpeed == oc.OpenconfigIfEthernet_ETHERNET_SPEED_UNSET {
		err = fmt.Errorf("Could not found set channel speed request")
	}

	return channelSpeed, err
}

func (this *ConfigMngrT) getPortBreakoutChannelSpeedChangeItemFromChangelog(ifname string, changelog *DiffChangelogMgmtT) (*DiffChangeMgmtT, error) {
	var err error = nil
	var changeItem *DiffChangeMgmtT
	channelSpeed := oc.OpenconfigIfEthernet_ETHERNET_SPEED_UNSET
	for _, ch := range changelog.Changes {
		if this.isChangedPortBreakoutChannelSpeed(ch.Change) {
			log.Infof("Found channel speed request too:\n%+v", ch)
			if ch.Change.Path[cmd.PortBreakoutIfnamePathItemIdxC] == ifname {
				channelSpeed = ch.Change.To.(oc.E_OpenconfigIfEthernet_ETHERNET_SPEED)
				changeItem = ch
				break
			}
		}
	}

	if channelSpeed == oc.OpenconfigIfEthernet_ETHERNET_SPEED_UNSET {
		err = fmt.Errorf("Could not found set channel speed request")
	}

	return changeItem, err
}

func (this *ConfigMngrT) isChangedPortBreakoutChannelSpeed(change *diff.Change) bool {
	if len(change.Path) < cmd.PortBreakoutPathItemsCountC {
		return false
	}

	if (change.Path[cmd.PortBreakoutCompPathItemIdxC] != cmd.PortBreakoutCompPathItemC) || (change.Path[cmd.PortBreakoutPortPathItemIdxC] != cmd.PortBreakoutPortPathItemC) || (change.Path[cmd.PortBreakoutModePathItemIdxC] != cmd.PortBreakoutModePathItemC) || (change.Path[cmd.PortBreakoutChanSpeedPathItemIdxC] != cmd.PortBreakoutChanSpeedPathItemC) {
		return false
	}

	return true
}

func (this *ConfigMngrT) isChangedPortBreakoutNumChannels(change *diff.Change) bool {
	if len(change.Path) < cmd.PortBreakoutPathItemsCountC {
		return false
	}

	if (change.Path[cmd.PortBreakoutCompPathItemIdxC] != cmd.PortBreakoutCompPathItemC) || (change.Path[cmd.PortBreakoutPortPathItemIdxC] != cmd.PortBreakoutPortPathItemC) || (change.Path[cmd.PortBreakoutModePathItemIdxC] != cmd.PortBreakoutModePathItemC) || (change.Path[cmd.PortBreakoutNumChanPathItemIdxC] != cmd.PortBreakoutNumChanPathItemC) {
		return false
	}

	return true
}

func (this *ConfigMngrT) isValidPortBreakoutNumChannels(numChannels cmd.PortBreakoutModeT) bool {
	if numChannels == cmd.PortBreakoutModeNoneC || numChannels == cmd.PortBreakoutMode4xC {
		return true
	}

	return false
}

func (this *ConfigMngrT) isValidPortBreakoutChannelSpeed(numChannels cmd.PortBreakoutModeT,
	channelSpeed oc.E_OpenconfigIfEthernet_ETHERNET_SPEED) bool {
	log.Infof("Split (%d), speed (%d)", numChannels, channelSpeed)
	switch channelSpeed {
	case oc.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_10GB:
		if numChannels == cmd.PortBreakoutMode4xC {
			return true
		}
	case oc.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_100GB:
		fallthrough
	case oc.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_40GB:
		if numChannels == cmd.PortBreakoutModeNoneC {
			return true
		}
	}

	return false
}

func (this *ConfigMngrT) getPortBreakoutNumChannelsFromChangelog(ifname string, changelog *DiffChangelogMgmtT) (cmd.PortBreakoutModeT, error) {
	var err error = nil
	numChannels := cmd.PortBreakoutModeInvalidC
	for _, ch := range changelog.Changes {
		if this.isChangedPortBreakoutNumChannels(ch.Change) {
			log.Infof("Found changing number of channels request too:\n%+v", ch.Change)
			if ch.Change.Path[cmd.PortBreakoutIfnamePathItemIdxC] == ifname {
				numChannels = cmd.PortBreakoutModeT(ch.Change.To.(uint8))
				break
			}
		}
	}

	if !this.isValidPortBreakoutNumChannels(numChannels) {
		err = fmt.Errorf("Number of channels (%d) to breakout is invalid", numChannels)
	}

	return numChannels, err
}

func (this *ConfigMngrT) getPortBreakoutNumChannelsChangeItemFromChangelog(ifname string, changelog *DiffChangelogMgmtT) (*DiffChangeMgmtT, error) {
	var err error = nil
	var changeItem *DiffChangeMgmtT
	numChannels := cmd.PortBreakoutModeInvalidC
	for _, ch := range changelog.Changes {
		if this.isChangedPortBreakoutNumChannels(ch.Change) {
			log.Infof("Found changing number of channels request too:\n%+v", ch.Change)
			if ch.Change.Path[cmd.PortBreakoutIfnamePathItemIdxC] == ifname {
				numChannels = cmd.PortBreakoutModeT(ch.Change.To.(uint8))
				changeItem = ch
				break
			}
		}
	}

	if !this.isValidPortBreakoutNumChannels(numChannels) {
		err = fmt.Errorf("Number of channels (%d) to breakout is invalid", numChannels)
	}

	return changeItem, err
}

func (this *ConfigMngrT) validatePortBreakoutChannSpeedChanging(ch *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ifname := ch.Change.Path[cmd.PortBreakoutIfnamePathItemIdxC]
	log.Infof("Requested changing of channel speed on subports of port %s", ifname)
	device := this.runningConfig.(*oc.Device)
	numChannels := device.GetComponent(ifname).GetPort().GetBreakoutMode().GetNumChannels()
	mode := cmd.PortBreakoutModeT(numChannels)
	if mode == cmd.PortBreakoutModeNoneC {
		return fmt.Errorf("Unable change channel speed if port %s is not splitted", ifname)
	}

	chanSpeed := ch.Change.To.(oc.E_OpenconfigIfEthernet_ETHERNET_SPEED)
	if !this.isValidPortBreakoutChannelSpeed(mode, chanSpeed) {
		return fmt.Errorf("Requested channel speed (%d) on subports of port %s is invalid", chanSpeed, ifname)
	}

	if this.transHasBeenStarted {
		setPortBreakoutChanSpeedCmd := cmd.NewSetPortBreakoutChanSpeedCmdT(ch.Change, this.ethSwitchMgmtClient)
		if err := this.appendSetPortBreakoutChanSpeedCmdToTransaction(ifname, setPortBreakoutChanSpeedCmd); err != nil {
			return err
		}

		// TODO: Update	this.transConfigLookupTbl
		ch.MarkAsProcessed()
	}

	return nil
}

func (this *ConfigMngrT) appendSetPortBreakoutCmdToTransaction(ifname string, cmdToAdd *cmd.SetPortBreakoutCmdT) error {
	setPortBreakoutCmds := this.cmdByIfname[setPortBreakoutC]
	for _, setPortBreakoutCmd := range setPortBreakoutCmds {
		if setPortBreakoutCmd.Equals(cmdToAdd) {
			return fmt.Errorf("Command %q already exists in transaction", cmdToAdd.GetName())
		}
	}

	log.Infof("Append command %q to transaction", cmdToAdd.GetName())

	setPortBreakoutCmds[ifname] = cmdToAdd
	return nil
}

func (this *ConfigMngrT) appendSetPortBreakoutChanSpeedCmdToTransaction(ifname string, cmdToAdd *cmd.SetPortBreakoutChanSpeedCmdT) error {
	setPortBreakoutChanSpeedCmds := this.cmdByIfname[setPortBreakoutChanSpeedC]
	for _, setPortBreakoutChanSpeedCmd := range setPortBreakoutChanSpeedCmds {
		if setPortBreakoutChanSpeedCmd.Equals(cmdToAdd) {
			return fmt.Errorf("Command %q already exists in transaction", cmdToAdd.GetName())
		}
	}

	log.Infof("Append command %q to transaction", cmdToAdd.GetName())

	setPortBreakoutChanSpeedCmds[ifname] = cmdToAdd
	this.addCmdToListTrans(cmdToAdd)
	return nil
}

func (this *ConfigMngrT) isEthIntfGoingToBeAvailableAfterPortBreakout(ifname string) bool {
	if _, exists := this.transConfigLookupTbl.idxByIntfName[ifname]; exists {
		return true
	}

	return false
}

func (this *ConfigMngrT) ValidatePortBreakoutChange(changedItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ifname := changedItem.Change.Path[cmd.PortBreakoutIfnamePathItemIdxC]
	if !this.isEthIntfAvailable(ifname) {
		return fmt.Errorf("Port %s is unrecognized", ifname)
	}

	var numChannels cmd.PortBreakoutModeT = cmd.PortBreakoutModeInvalidC
	var channelSpeed oc.E_OpenconfigIfEthernet_ETHERNET_SPEED = oc.OpenconfigIfEthernet_ETHERNET_SPEED_UNSET
	var numChannelsChangeItem *DiffChangeMgmtT
	var channelSpeedChangeItem *DiffChangeMgmtT
	var err error

	if changedItem.Change.Path[cmd.PortBreakoutNumChanPathItemIdxC] == cmd.PortBreakoutNumChanPathItemC {
		channelSpeed, err = this.getPortBreakoutChannelSpeedFromChangelog(ifname, changelog)
		if err != nil {
			return err
		}

		numChannels = cmd.PortBreakoutModeT(changedItem.Change.To.(uint8))
		if !this.isValidPortBreakoutNumChannels(numChannels) {
			return fmt.Errorf("Number of channels (%d) to breakout is invalid", numChannels)
		}

		channelSpeedChangeItem, err = this.getPortBreakoutChannelSpeedChangeItemFromChangelog(ifname, changelog)
		if err != nil {
			return err
		}
		numChannelsChangeItem = changedItem
	} else if changedItem.Change.Path[cmd.PortBreakoutChanSpeedPathItemIdxC] == cmd.PortBreakoutChanSpeedPathItemC {
		numChannels, err = this.getPortBreakoutNumChannelsFromChangelog(ifname, changelog)
		if err != nil {
			return this.validatePortBreakoutChannSpeedChanging(changedItem, changelog)
		}

		channelSpeed = changedItem.Change.To.(oc.E_OpenconfigIfEthernet_ETHERNET_SPEED)
		if !this.isValidPortBreakoutChannelSpeed(numChannels, channelSpeed) {
			return fmt.Errorf("Speed channel (%d) is invalid", channelSpeed)
		}

		numChannelsChangeItem, err = this.getPortBreakoutNumChannelsChangeItemFromChangelog(ifname, changelog)
		if err != nil {
			return err
		}
		channelSpeedChangeItem = changedItem
	} else {
		return fmt.Errorf("Unable to get port breakout changing")
	}

	log.Infof("Requested changing port %s breakout into mode %d with speed %d", ifname, numChannels, channelSpeed)
	setPortBreakoutCmd := cmd.NewSetPortBreakoutCmdT(numChannelsChangeItem.Change, channelSpeedChangeItem.Change, this.ethSwitchMgmtClient)
	if numChannels == cmd.PortBreakoutModeNoneC {
		for i := 1; i <= 4; i++ {
			slavePort := fmt.Sprintf("%s.%d", ifname, i)
			log.Infof("Composed slave port: %s", slavePort)
			if err := this.transConfigLookupTbl.checkDependenciesForDeletePortBreakout(slavePort); err != nil {
				return fmt.Errorf("Cannot %q because there are dependencies from interface %s:\n%s",
					setPortBreakoutCmd.GetName(), slavePort, err)
			}
		}
	} else {
		if err := this.transConfigLookupTbl.checkDependenciesForDeletePortBreakout(ifname); err != nil {
			return fmt.Errorf("Cannot %q because there are dependencies from interface %s:\n%s",
				setPortBreakoutCmd.GetName(), ifname, err)
		}
	}

	if this.transHasBeenStarted {
		setPortBreakoutCmd := cmd.NewSetPortBreakoutCmdT(numChannelsChangeItem.Change, channelSpeedChangeItem.Change, this.ethSwitchMgmtClient)
		if err = this.appendSetPortBreakoutCmdToTransaction(ifname, setPortBreakoutCmd); err != nil {
			return err
		}

		if numChannels == cmd.PortBreakoutModeNoneC {
			if err := this.transConfigLookupTbl.addNewInterfaceIfItDoesNotExist(ifname); err != nil {
				return err
			}
		} else {
			for i := 1; i <= 4; i++ {
				slavePort := fmt.Sprintf("%s.%d", ifname, i)
				if err := this.transConfigLookupTbl.addNewInterfaceIfItDoesNotExist(slavePort); err != nil {
					return err
				}
			}
		}

		numChannelsChangeItem.MarkAsProcessed()
		channelSpeedChangeItem.MarkAsProcessed()
	}

	return nil
}
