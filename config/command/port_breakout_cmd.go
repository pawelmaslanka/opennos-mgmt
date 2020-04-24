package command

import (
	"context"
	"fmt"
	mgmt "opennos-eth-switch-service/mgmt"
	"opennos-eth-switch-service/mgmt/interfaces"
	"opennos-eth-switch-service/mgmt/platform"
	"opennos-mgmt/gnmi/modeldata/oc"
	"opennos-mgmt/utils"
	"time"

	"github.com/r3labs/diff"
)

type PortBreakoutModeT uint8

const (
	PortBreakoutCompPathItemIdxC      = 0
	PortBreakoutIfnamePathItemIdxC    = 1
	PortBreakoutPortPathItemIdxC      = 2
	PortBreakoutModePathItemIdxC      = 3
	PortBreakoutNumChanPathItemIdxC   = 4
	PortBreakoutChanSpeedPathItemIdxC = 4
	PortBreakoutPathItemsCountC       = 5
	PortBreakoutCompPathItemC         = "Component"
	PortBreakoutPortPathItemC         = "Port"
	PortBreakoutModePathItemC         = "BreakoutMode"
	PortBreakoutNumChanPathItemC      = "NumChannels"
	PortBreakoutChanSpeedPathItemC    = "ChannelSpeed"

	DisabledPortBreakoutC = 1
	EnabledPortBreakoutC  = 4

	PortBreakoutMode4xC      PortBreakoutModeT = 4
	PortBreakoutMode2xC      PortBreakoutModeT = 2
	PortBreakoutModeNoneC    PortBreakoutModeT = 1
	PortBreakoutModeInvalidC PortBreakoutModeT = 0
)

// SetPortBreakoutCmdT implements command for break out front panel port into multiple logical ports
type SetPortBreakoutCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewSetPortBreakoutCmdT create new instance of SetPortBreakoutCmdT type
func NewSetPortBreakoutCmdT(numChansChg *diff.Change, chanSpeedChg *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *SetPortBreakoutCmdT {
	changes := make([]*diff.Change, maxChangePortBreakoutIdxC)
	changes[numChannelsChangeIdxC] = numChansChg
	changes[channelSpeedChangeIdxC] = chanSpeedChg
	return &SetPortBreakoutCmdT{
		commandT: newCommandT("set port breakout", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and breaks out front panel port
// into multiple logical ports
func (this *SetPortBreakoutCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	return this.configurePortBreakout(shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *SetPortBreakoutCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	return this.configurePortBreakout(shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *SetPortBreakoutCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *SetPortBreakoutCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*SetPortBreakoutCmdT)
	return this.equals(otherCmd.commandT)
}

// Append is not supported
func (this *SetPortBreakoutCmdT) Append(cmd CommandI) (bool, error) {
	return false, fmt.Errorf("Unsupported")
}

// SetPortBreakoutChanSpeedCmdT implements command for change speed onto all sub-ports
type SetPortBreakoutChanSpeedCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewSetPortBreakoutCmdT create new instance of SetPortBreakoutChanSpeedCmdT type
func NewSetPortBreakoutChanSpeedCmdT(change *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *SetPortBreakoutChanSpeedCmdT {
	changes := make([]*diff.Change, maxChangePortBreakoutIdxC)
	changes[channelSpeedChangeIdxC] = change
	return &SetPortBreakoutChanSpeedCmdT{
		commandT: newCommandT("set port breakout channel speed", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and set channel speed of all
// sub-ports
func (this *SetPortBreakoutChanSpeedCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	return this.doPortBreakoutChanSpeedCmd(shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *SetPortBreakoutChanSpeedCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	return this.doPortBreakoutChanSpeedCmd(shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *SetPortBreakoutChanSpeedCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *SetPortBreakoutChanSpeedCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*SetPortBreakoutChanSpeedCmdT)
	return this.equals(otherCmd.commandT)
}

// Append is not supported
func (this *SetPortBreakoutChanSpeedCmdT) Append(cmd CommandI) (bool, error) {
	return false, fmt.Errorf("Unsupported")
}

// No exported section
const (
	numChannelsChangeIdxC = iota
	channelSpeedChangeIdxC
	maxChangePortBreakoutIdxC
)

// It cannot work if you would want to be like this: func configurePortBreakout(this *commandT, shouldBeAbleOnlyToUndo bool) error
func (this *commandT) configurePortBreakout(shouldBeAbleOnlyToUndo bool) error {
	if this.isAbleOnlyToUndo() != shouldBeAbleOnlyToUndo {
		return this.createErrorAccordingToExecutionState()
	}

	this.dumpInternalData()
	var err error
	var numChannels platform.PortBreakoutRequest_NumChannels
	var chanSpeedConv uint8

	if this.changes[numChannelsChangeIdxC].Type != diff.DELETE {
		numChannels, err = convertOcNumChanIntoMgmtPortBreakoutReq(PortBreakoutModeT(this.changes[numChannelsChangeIdxC].To.(uint8)))
		if err != nil {
			return err
		}

		chanSpeedConv, err = utils.ConvertGoInterfaceIntoUint8(this.changes[channelSpeedChangeIdxC].To)
		if err != nil {
			return err
		}
	} else {
		numChannels, err = convertOcNumChanIntoMgmtPortBreakoutReq(PortBreakoutModeT(this.changes[numChannelsChangeIdxC].From.(uint8)))
		if err != nil {
			return err
		}

		chanSpeedConv, err = utils.ConvertGoInterfaceIntoUint8(this.changes[channelSpeedChangeIdxC].From)
		if err != nil {
			return err
		}
	}

	channelSpeed, err := convertOcChanSpeedIntoMgmtPortBreakoutReq(oc.E_OpenconfigIfEthernet_ETHERNET_SPEED(chanSpeedConv))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = (*this.ethSwitchMgmt).SetPortBreakout(ctx, &platform.PortBreakoutRequest{
		EthIntf: &interfaces.EthernetIntf{
			Ifname: this.changes[numChannelsChangeIdxC].Path[PortBreakoutIfnamePathItemIdxC],
		},
		NumChannels:  numChannels,
		ChannelSpeed: &channelSpeed,
	})
	if err != nil {
		return err
	}

	this.finalize()
	return nil
}

func (this *commandT) doPortBreakoutChanSpeedCmd(shouldBeAbleOnlyToUndo bool) error {
	if this.isAbleOnlyToUndo() != shouldBeAbleOnlyToUndo {
		return this.createErrorAccordingToExecutionState()
	}

	this.dumpInternalData()
	channelSpeed, err := convertOcChanSpeedIntoMgmtPortBreakoutReq(this.changes[channelSpeedChangeIdxC].To.(oc.E_OpenconfigIfEthernet_ETHERNET_SPEED))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = (*this.ethSwitchMgmt).SetPortBreakoutChanSpeed(ctx, &mgmt.PortBreakoutChanSpeedRequest{
		EthIntf: &interfaces.EthernetIntf{
			Ifname: this.changes[channelSpeedChangeIdxC].Path[PortBreakoutIfnamePathItemIdxC],
		},
		ChannelSpeed: &channelSpeed,
	})
	if err != nil {
		return err
	}

	this.finalize()
	return nil
}

func convertOcNumChanIntoMgmtPortBreakoutReq(numChannels PortBreakoutModeT) (mgmt.PortBreakoutRequest_NumChannels, error) {
	var mode mgmt.PortBreakoutRequest_NumChannels
	switch numChannels {
	case PortBreakoutModeNoneC:
		mode = mgmt.PortBreakoutRequest_MODE_1x
	case PortBreakoutMode4xC:
		mode = mgmt.PortBreakoutRequest_MODE_4x
	default:
		return 0, fmt.Errorf("Failed to convert OC number of channels (%d) into request of management port breakout", numChannels)
	}

	return mode, nil
}

func convertOcChanSpeedIntoMgmtPortBreakoutReq(chanSpeed oc.E_OpenconfigIfEthernet_ETHERNET_SPEED) (mgmt.ChannelSpeed, error) {
	var err error = nil
	var speed mgmt.ChannelSpeed_Mode
	switch chanSpeed {
	case oc.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_10GB:
		speed = mgmt.ChannelSpeed_SPEED_10GB
	case oc.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_100GB:
		speed = mgmt.ChannelSpeed_SPEED_100GB
	default:
		err = fmt.Errorf("Failed to convert OC channel speed (%d) into request of management port breakout", chanSpeed)
	}

	return mgmt.ChannelSpeed{Mode: speed}, err
}
