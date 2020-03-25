package command

import (
	"context"
	"fmt"
	mgmt "opennos-eth-switch-service/mgmt"
	"opennos-mgmt/gnmi/modeldata/oc"
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
	changes := make([]*diff.Change, maxChangeIdxC)
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
	return this.doPortBreakoutCmd(shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *SetPortBreakoutCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	return this.doPortBreakoutCmd(shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *SetPortBreakoutCmdT) GetName() string {
	return this.name
}

func (this *SetPortBreakoutCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*SetPortBreakoutCmdT)
	return this.equals(otherCmd.commandT)
}

// WithdrawPortBreakoutCmdT implements command for combine Ccmbine multiple logical ports into
// single port
// TODO: Consider if it is needed because we don't removing any parameters, we just edit port breakout mode
type WithdrawPortBreakoutCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewSetPortBreakoutCmdT create new instance of WithdrawPortBreakoutCmdT type
func NewWithdrawPortBreakoutCmdT(numChansChg *diff.Change, chanSpeedChg *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *WithdrawPortBreakoutCmdT {
	changes := make([]*diff.Change, maxChangeIdxC)
	changes[numChannelsChangeIdxC] = numChansChg
	changes[channelSpeedChangeIdxC] = chanSpeedChg
	return &WithdrawPortBreakoutCmdT{
		commandT: newCommandT("withdraw port breakout", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and combines multiple logical
// ports into single port
func (this *WithdrawPortBreakoutCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	return this.doPortBreakoutCmd(shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *WithdrawPortBreakoutCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	return this.doPortBreakoutCmd(shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *WithdrawPortBreakoutCmdT) GetName() string {
	return this.name
}

// SetPortBreakoutSpeedCmdT implements command for change speed onto multiple logical ports
type SetPortBreakoutSpeedCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// No exported section
const (
	numChannelsChangeIdxC = iota
	channelSpeedChangeIdxC
	maxChangeIdxC
)

// It cannot work if you would want to be like this: func doPortBreakoutCmd(this *commandT, shouldBeAbleOnlyToUndo bool) error
func (this *commandT) doPortBreakoutCmd(shouldBeAbleOnlyToUndo bool) error {
	if this.isAbleOnlyToUndo() != shouldBeAbleOnlyToUndo {
		return this.createErrorAccordingToExecutionState()
	}

	this.dumpInternalData()
	numChannels, err := convertOcNumChanIntoMgmtPortBreakoutReq(PortBreakoutModeT(this.changes[numChannelsChangeIdxC].To.(uint8)))
	if err != nil {
		return err
	}

	channelSpeed, err := convertOcChanSpeedIntoMgmtPortBreakoutReq(this.changes[channelSpeedChangeIdxC].To.(oc.E_OpenconfigIfEthernet_ETHERNET_SPEED))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = (*this.ethSwitchMgmt).SetPortBreakout(ctx, &mgmt.PortBreakoutRequest{
		Ifname:       this.changes[numChannelsChangeIdxC].Path[PortBreakoutIfnamePathItemIdxC],
		NumChannels:  numChannels,
		ChannelSpeed: channelSpeed,
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

func convertOcChanSpeedIntoMgmtPortBreakoutReq(chanSpeed oc.E_OpenconfigIfEthernet_ETHERNET_SPEED) (mgmt.PortBreakoutRequest_ChannelSpeed, error) {
	var speed mgmt.PortBreakoutRequest_ChannelSpeed
	switch chanSpeed {
	case oc.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_10GB:
		speed = mgmt.PortBreakoutRequest_SPEED_10GB
	case oc.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_100GB:
		speed = mgmt.PortBreakoutRequest_SPEED_100GB
	default:
		return 0, fmt.Errorf("Failed to convert OC channel speed (%d) into request of management port breakout", chanSpeed)
	}

	return speed, nil
}
