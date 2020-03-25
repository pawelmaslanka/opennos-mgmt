package config

import (
	"fmt"
	cmd "opennos-mgmt/config/command"
	"opennos-mgmt/gnmi/modeldata/oc"

	log "github.com/golang/glog"
	"github.com/r3labs/diff"
)

func (this *ConfigMngrT) getPortBreakoutChannelSpeedFromChangelog(ifname string, changelog *diff.Changelog) (oc.E_OpenconfigIfEthernet_ETHERNET_SPEED, error) {
	var err error = nil
	channelSpeed := oc.OpenconfigIfEthernet_ETHERNET_SPEED_UNSET
	for _, change := range *changelog {
		if this.isChangedPortBreakoutChannelSpeed(&change) {
			log.Infof("Found channel speed request too:\n%+v", change)
			if change.Path[cmd.PortBreakoutIfnamePathItemIdxC] == ifname {
				channelSpeed = change.To.(oc.E_OpenconfigIfEthernet_ETHERNET_SPEED)
				break
			}
		}
	}

	if channelSpeed == oc.OpenconfigIfEthernet_ETHERNET_SPEED_UNSET {
		err = fmt.Errorf("Could not found set channel speed request")
	}

	return channelSpeed, err
}

func (this *ConfigMngrT) getPortBreakoutChannelSpeedChangeItemFromChangelog(ifname string, changelog *diff.Changelog) (*diff.Change, error) {
	var err error = nil
	var changeItem *diff.Change
	channelSpeed := oc.OpenconfigIfEthernet_ETHERNET_SPEED_UNSET
	for _, change := range *changelog {
		if this.isChangedPortBreakoutChannelSpeed(&change) {
			log.Infof("Found channel speed request too:\n%+v", change)
			if change.Path[cmd.PortBreakoutIfnamePathItemIdxC] == ifname {
				channelSpeed = change.To.(oc.E_OpenconfigIfEthernet_ETHERNET_SPEED)
				changeItem = &change
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

func (this *ConfigMngrT) getPortBreakoutNumChannelsFromChangelog(ifname string, changelog *diff.Changelog) (cmd.PortBreakoutModeT, error) {
	var err error = nil
	numChannels := cmd.PortBreakoutModeInvalidC
	for _, change := range *changelog {
		if this.isChangedPortBreakoutNumChannels(&change) {
			log.Infof("Found changing number of channels request too:\n%+v", change)
			if change.Path[cmd.PortBreakoutIfnamePathItemIdxC] == ifname {
				numChannels = cmd.PortBreakoutModeT(change.To.(uint8))
				break
			}
		}
	}

	if !this.isValidPortBreakoutNumChannels(numChannels) {
		err = fmt.Errorf("Number of channels (%d) to breakout is invalid", numChannels)
	}

	return numChannels, err
}

func (this *ConfigMngrT) getPortBreakoutNumChannelsChangeItemFromChangelog(ifname string, changelog *diff.Changelog) (*diff.Change, error) {
	var err error = nil
	var changeItem *diff.Change
	numChannels := cmd.PortBreakoutModeInvalidC
	for _, change := range *changelog {
		if this.isChangedPortBreakoutNumChannels(&change) {
			log.Infof("Found changing number of channels request too:\n%+v", change)
			if change.Path[cmd.PortBreakoutIfnamePathItemIdxC] == ifname {
				numChannels = cmd.PortBreakoutModeT(change.To.(uint8))
				changeItem = &change
				break
			}
		}
	}

	if !this.isValidPortBreakoutNumChannels(numChannels) {
		err = fmt.Errorf("Number of channels (%d) to breakout is invalid", numChannels)
	}

	return changeItem, err
}

func (this *ConfigMngrT) validatePortBreakoutChannSpeedChanging(change *diff.Change, changelog *diff.Changelog) error {
	ifname := change.Path[cmd.PortBreakoutIfnamePathItemIdxC]
	log.Infof("Requested changing of channel speed on subports of port %s", ifname)
	device := this.runningConfig.(*oc.Device)
	numChannels := device.GetComponent(ifname).GetPort().GetBreakoutMode().GetNumChannels()
	mode := cmd.PortBreakoutModeT(numChannels)
	if mode == cmd.PortBreakoutModeNoneC {
		return fmt.Errorf("Unable change channel speed if port %s is not splitted", ifname)
	}

	if !this.isValidPortBreakoutChannelSpeed(mode, change.To.(oc.E_OpenconfigIfEthernet_ETHERNET_SPEED)) {
		return fmt.Errorf("Requested channel speed (%d) on subports of port %s is invalid", change.To, ifname)
	}

	return nil
}

func (this *ConfigMngrT) appendSetPortBreakoutCmdToTransaction(ifname string, cmdToAdd *cmd.SetPortBreakoutCmdT) error {
	setPortBreakoutCmds := this.cmdByIfname[SetPortBreakoutC]
	for _, setPortBreakoutCmd := range setPortBreakoutCmds {
		if setPortBreakoutCmd.Equals(cmdToAdd) {
			return fmt.Errorf("%q already exists in transaction", cmdToAdd.GetName())
		}
	}

	setPortBreakoutCmds[ifname] = cmdToAdd
	return nil
}
