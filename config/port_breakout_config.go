package config

import (
	"fmt"
	cmd "opennos-mgmt/config/command"
	"opennos-mgmt/gnmi/modeldata/oc"
	"regexp"

	log "github.com/golang/glog"
	"github.com/r3labs/diff"
)

func (this *ConfigMngrT) findPortBreakoutChanSpeedFromChangelog(ifname string, changelog *DiffChangelogMgmtT) (oc.E_OpenconfigIfEthernet_ETHERNET_SPEED, error) {
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

func (this *ConfigMngrT) findPortBreakoutChanSpeedChangeFromChangelog(ifname string, changelog *DiffChangelogMgmtT) (*DiffChangeMgmtT, error) {
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

func (this *ConfigMngrT) findPortBreakoutNumChannelsFromChangelog(ifname string, changelog *DiffChangelogMgmtT) (cmd.PortBreakoutModeT, error) {
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

func (this *ConfigMngrT) findPortBreakoutNumChannelsChangeFromChangelog(ifname string, changelog *DiffChangelogMgmtT) (*DiffChangeMgmtT, error) {
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

func (this *ConfigMngrT) validatePortBreakoutChannSpeedChange(ch *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
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
		if err := this.appendCmdToTransaction(ifname, setPortBreakoutChanSpeedCmd, setPortBreakoutChanSpeedC, false); err != nil {
			return err
		}
	}

	ch.MarkAsProcessed()

	return nil
}

func (this *ConfigMngrT) isEthIntfGoingToBeAvailableAfterPortBreakout(ifname string) bool {
	if _, exists := this.transConfigLookupTbl.idxByEthIfname[ifname]; exists {
		return true
	}

	return false
}

func (this *ConfigMngrT) isChangedPortBreakout(change *diff.Change) bool {
	if len(change.Path) != cmd.PortBreakoutPathItemsCountC {
		return false
	}

	if (change.Path[cmd.PortBreakoutCompPathItemIdxC] != cmd.PortBreakoutCompPathItemC) || (change.Path[cmd.PortBreakoutPortPathItemIdxC] != cmd.PortBreakoutPortPathItemC) || (change.Path[cmd.PortBreakoutModePathItemIdxC] != cmd.PortBreakoutModePathItemC) || ((change.Path[cmd.PortBreakoutNumChanPathItemIdxC] != cmd.PortBreakoutNumChanPathItemC) && (change.Path[cmd.PortBreakoutChanSpeedPathItemIdxC] != cmd.PortBreakoutChanSpeedPathItemC)) {
		return false
	}

	return true
}

func (this *ConfigMngrT) isChangedPortBreakoutChanSpeed(change *diff.Change) bool {
	if len(change.Path) != cmd.PortBreakoutPathItemsCountC {
		return false
	}

	if (change.Path[cmd.PortBreakoutCompPathItemIdxC] != cmd.PortBreakoutCompPathItemC) || (change.Path[cmd.PortBreakoutPortPathItemIdxC] != cmd.PortBreakoutPortPathItemC) || (change.Path[cmd.PortBreakoutModePathItemIdxC] != cmd.PortBreakoutModePathItemC) || (change.Path[cmd.PortBreakoutChanSpeedPathItemIdxC] != cmd.PortBreakoutChanSpeedPathItemC) {
		return false
	}

	return true
}

func (this *ConfigMngrT) validatePortBreakoutChange(changedItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
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
		channelSpeed, err = this.findPortBreakoutChanSpeedFromChangelog(ifname, changelog)
		if err != nil {
			return err
		}

		numChannels = cmd.PortBreakoutModeT(changedItem.Change.To.(uint8))
		if !this.isValidPortBreakoutNumChannels(numChannels) {
			return fmt.Errorf("Number of channels (%d) to breakout is invalid", numChannels)
		}

		channelSpeedChangeItem, err = this.findPortBreakoutChanSpeedChangeFromChangelog(ifname, changelog)
		if err != nil {
			return err
		}
		numChannelsChangeItem = changedItem
	} else if changedItem.Change.Path[cmd.PortBreakoutChanSpeedPathItemIdxC] == cmd.PortBreakoutChanSpeedPathItemC {
		numChannels, err = this.findPortBreakoutNumChannelsFromChangelog(ifname, changelog)
		if err != nil {
			return this.validatePortBreakoutChannSpeedChange(changedItem, changelog)
		}

		channelSpeed = changedItem.Change.To.(oc.E_OpenconfigIfEthernet_ETHERNET_SPEED)
		if !this.isValidPortBreakoutChannelSpeed(numChannels, channelSpeed) {
			return fmt.Errorf("Speed channel (%d) is invalid", channelSpeed)
		}

		numChannelsChangeItem, err = this.findPortBreakoutNumChannelsChangeFromChangelog(ifname, changelog)
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
			slavePort := fmt.Sprintf("%s/%d", ifname, i)
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
		if err = this.appendCmdToTransaction(ifname, setPortBreakoutCmd, setPortBreakoutC, false); err != nil {
			return err
		}
	}

	// TODO: Because we explicitly create Ethernet interface, that's why we remove this part from here
	// if numChannels == cmd.PortBreakoutModeNoneC {
	// 	if err := this.transConfigLookupTbl.addNewEthIntfIfItDoesNotExist(ifname); err != nil {
	// 		return err
	// 	}
	// } else {
	// 	for i := 1; i <= 4; i++ {
	// 		slavePort := fmt.Sprintf("%s/%d", ifname, i)
	// 		if err := this.transConfigLookupTbl.addNewEthIntfIfItDoesNotExist(slavePort); err != nil {
	// 			return err
	// 		}
	// 	}
	// }

	numChannelsChangeItem.MarkAsProcessed()
	channelSpeedChangeItem.MarkAsProcessed()

	return nil
}

func (this *ConfigMngrT) findSetPortBreakout(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type == diff.UPDATE {
				if this.isChangedPortBreakout(ch.Change) {
					return ch, true
				}
			}
		}
	}

	return nil, false
}

func (this *ConfigMngrT) findSetPortBreakoutChanSpeed(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type == diff.UPDATE {
				if this.isChangedPortBreakoutChanSpeed(ch.Change) {
					return ch, true
				}
			}
		}
	}

	return nil, false
}

func (this *ConfigMngrT) processSetPortBreakoutFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to set port breakout for port
		if change, exists := this.findSetPortBreakout(changelog); exists {
			if err := this.validatePortBreakoutChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (this *ConfigMngrT) processSetPortBreakoutChanSpeedFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to set port breakout channel speed for subports
		if change, exists := this.findSetPortBreakoutChanSpeed(changelog); exists {
			if err := this.validatePortBreakoutChannSpeedChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (this *ConfigMngrT) setPortBreakout(device *oc.Device) error {
	var err error
	for _, ethIfname := range this.configLookupTbl.ethIfnameByIdx {
		// We want to process only not splitted ports
		rgx := regexp.MustCompile(`eth-|/`)
		tokens := rgx.Split(ethIfname, -1)
		if len(tokens) == 4 { // breakout mode enable
			continue
		}

		comp := device.GetComponent(ethIfname)
		if comp == nil {
			continue
		}

		port := comp.GetPort()
		if port == nil {
			continue
		}

		mode := port.GetBreakoutMode()
		if mode == nil {
			continue
		}

		numChannels := mode.GetNumChannels()
		if numChannels == uint8(cmd.PortBreakoutModeInvalidC) {
			return fmt.Errorf("number of channels of port %s is unset", ethIfname)
		}

		if numChannels == uint8(cmd.PortBreakoutModeNoneC) {
			continue
		}

		chanSpeed := mode.GetChannelSpeed()
		if chanSpeed == oc.OpenconfigIfEthernet_ETHERNET_SPEED_UNSET {
			return fmt.Errorf("channel speed of port %s is unset", ethIfname)
		}

		var numChanChange diff.Change
		numChanChange.Type = diff.CREATE
		numChanChange.From = nil
		numChanChange.To = numChannels
		numChanChange.Path = make([]string, cmd.PortBreakoutPathItemsCountC)
		numChanChange.Path[cmd.PortBreakoutCompPathItemIdxC] = cmd.PortBreakoutCompPathItemC
		numChanChange.Path[cmd.PortBreakoutIfnamePathItemIdxC] = ethIfname
		numChanChange.Path[cmd.PortBreakoutPortPathItemIdxC] = cmd.PortBreakoutPortPathItemC
		numChanChange.Path[cmd.PortBreakoutModePathItemIdxC] = cmd.PortBreakoutModePathItemC
		numChanChange.Path[cmd.PortBreakoutNumChanPathItemIdxC] = cmd.PortBreakoutNumChanPathItemC

		var chanSpeedChange diff.Change
		chanSpeedChange.Type = diff.CREATE
		chanSpeedChange.From = nil
		chanSpeedChange.To = chanSpeed
		chanSpeedChange.Path = make([]string, cmd.PortBreakoutPathItemsCountC)
		chanSpeedChange.Path[cmd.PortBreakoutCompPathItemIdxC] = cmd.PortBreakoutCompPathItemC
		chanSpeedChange.Path[cmd.PortBreakoutIfnamePathItemIdxC] = ethIfname
		chanSpeedChange.Path[cmd.PortBreakoutPortPathItemIdxC] = cmd.PortBreakoutPortPathItemC
		chanSpeedChange.Path[cmd.PortBreakoutModePathItemIdxC] = cmd.PortBreakoutModePathItemC
		chanSpeedChange.Path[cmd.PortBreakoutChanSpeedPathItemIdxC] = cmd.PortBreakoutChanSpeedPathItemC

		command := cmd.NewSetPortBreakoutCmdT(&numChanChange, &chanSpeedChange, this.ethSwitchMgmtClient)
		if err = this.appendCmdToTransaction(ethIfname, command, setPortBreakoutC, false); err != nil {
			return err
		}
	}

	return nil
}
