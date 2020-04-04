package command

import (
	mgmt "opennos-eth-switch-service/mgmt"

	"github.com/r3labs/diff"
)

const (
	Ipv4AddrEthIntfPathItemIdxC                       = 0
	Ipv4AddrEthIfnamePathItemIdxC                     = 1
	Ipv4AddrEthSubintfPathItemIdxC                    = 2
	Ipv4AddrEthSubintfIdxPathItemIdxC                 = 3
	Ipv4AddrEthSubintfIpv4PathItemIdxC                = 4
	Ipv4AddrEthSubintfIpv4AddrPathItemIdxC            = 5
	Ipv4AddrEthSubintfIpv4AddrIpPathItemIdxC          = 6
	Ipv4AddrEthSubintfIpv4AddrPartIpPathItemIdxC      = 7
	Ipv4AddrEthSubintfIpv4AddrPartPrfxLenPathItemIdxC = 7
	Ipv4AddrEthPathItemsCountC                        = 8

	Ipv4AddrEthIntfPathItemC                       = "Interface"
	Ipv4AddrEthSubintfPathItemC                    = "Subinterface"
	Ipv4AddrEthSubintfIpv4PathItemC                = "Ipv4Addr"
	Ipv4AddrEthSubintfIpv4AddrPathItemC            = "Address"
	Ipv4AddrEthSubintfIpv4AddrPartIpPathItemC      = "Ip"
	Ipv4AddrEthSubintfIpv4AddrPartPrfxLenPathItemC = "PrefixLength"
)

const (
	ipv4AddrIpChangeIdxC = iota
	ipv4AddrPrfxLenChangeIdxC
	maxChangeIpv4AddrIdxC
)

// SetIpv4AddrEthIntfCmdT implements command for assigning IPv4 address on Ethernet Interface
type SetIpv4AddrEthIntfCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewSetIpv4AddrEthIntfCmdT create new instance of SetIpv4AddrEthIntfCmdT type
func NewSetIpv4AddrEthIntfCmdT(ip *diff.Change, prfxLen *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *SetIpv4AddrEthIntfCmdT {
	changes := make([]*diff.Change, maxChangeIpv4AddrIdxC)
	changes[ipv4AddrIpChangeIdxC] = ip
	changes[ipv4AddrPrfxLenChangeIdxC] = prfxLen
	return &SetIpv4AddrEthIntfCmdT{
		commandT: newCommandT("set ip4 address for ethernet interface", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and assigns IPv4 address for Ethernet interface
func (this *SetIpv4AddrEthIntfCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	return this.configureIpv4AddrCmd(shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *SetIpv4AddrEthIntfCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	return this.configureIpv4AddrCmd(shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *SetIpv4AddrEthIntfCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *SetIpv4AddrEthIntfCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*SetIpv4AddrEthIntfCmdT)
	return this.equals(otherCmd.commandT)
}

// DeleteIpv4AddrEthIntfCmdT implements command for deleting IPv4 address from Ethernet Interface
type DeleteIpv4AddrEthIntfCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewDeleteIpv4AddrEthIntfCmdT create new instance of DeleteIpv4AddrEthIntfCmdT type
func NewDeleteIpv4AddrEthIntfCmdT(ip *diff.Change, prfxLen *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *DeleteIpv4AddrEthIntfCmdT {
	changes := make([]*diff.Change, maxChangeIpv4AddrIdxC)
	changes[ipv4AddrIpChangeIdxC] = ip
	changes[ipv4AddrPrfxLenChangeIdxC] = prfxLen
	return &DeleteIpv4AddrEthIntfCmdT{
		commandT: newCommandT("delete ip4 address from ethernet interface", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and deletes IPv4 address from Ethernet interface
func (this *DeleteIpv4AddrEthIntfCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	return this.configureIpv4AddrCmd(shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *DeleteIpv4AddrEthIntfCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	return this.configureIpv4AddrCmd(shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *DeleteIpv4AddrEthIntfCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *DeleteIpv4AddrEthIntfCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*DeleteIpv4AddrEthIntfCmdT)
	return this.equals(otherCmd.commandT)
}

func (this *commandT) configureIpv4AddrCmd(shouldBeAbleOnlyToUndo bool) error {
	if this.isAbleOnlyToUndo() != shouldBeAbleOnlyToUndo {
		return this.createErrorAccordingToExecutionState()
	}

	this.dumpInternalData()
	// numChannels, err := convertOcNumChanIntoMgmtPortBreakoutReq(PortBreakoutModeT(this.changes[numChannelsChangeIdxC].To.(uint8)))
	// if err != nil {
	// 	return err
	// }

	// channelSpeed, err := convertOcChanSpeedIntoMgmtPortBreakoutReq(this.changes[channelSpeedChangeIdxC].To.(oc.E_OpenconfigIfEthernet_ETHERNET_SPEED))
	// if err != nil {
	// 	return err
	// }

	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// defer cancel()
	// _, err = (*this.ethSwitchMgmt).SetPortBreakout(ctx, &platform.PortBreakoutRequest{
	// 	Ifname:       this.changes[numChannelsChangeIdxC].Path[PortBreakoutIfnamePathItemIdxC],
	// 	NumChannels:  numChannels,
	// 	ChannelSpeed: &channelSpeed,
	// })
	// if err != nil {
	// 	return err
	// }

	this.finalize()
	return nil
}
