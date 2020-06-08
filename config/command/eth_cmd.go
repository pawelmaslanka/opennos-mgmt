package command

import (
	"context"
	"fmt"
	mgmt "opennos-eth-switch-service/mgmt"
	"opennos-eth-switch-service/mgmt/interfaces"
	"opennos-mgmt/utils"
	"time"

	"github.com/r3labs/diff"
)

const (
	EthIntfInterfacePathItemIdxC = 0
	EthIntfIfnamePathItemIdxC    = 1
	EthIntfNamePathItemIdxC      = 2
	EthIntfEthernetPathItemIdxC  = 2
	EthIntfPathItemsCountC       = 3

	EthIntfInterfacePathItemC = "Interface"
	EthIntfNamePathItemC      = "Name"
	EthIntfEthernetPathItemC  = "Ethernet"
)

const (
	ethChangeIdxC = iota
	maxEthChangeIdxC
)

// SetEthIntfCmdT implements command for creating Ethernet interface
type SetEthIntfCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewSetEthIntfCmdT creates new instance of SetEthIntfCmdT type
func NewSetEthIntfCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *SetEthIntfCmdT {
	changes := make([]*diff.Change, maxEthChangeIdxC)
	changes[ethChangeIdxC] = vlan
	return &SetEthIntfCmdT{
		commandT: newCommandT("set ethernet interface", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and creates Ethernet interface
func (this *SetEthIntfCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	isGoingToBeDeleted := false
	return doEthIntfCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *SetEthIntfCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	isGoingToBeDeleted := true
	return doEthIntfCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *SetEthIntfCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *SetEthIntfCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*SetEthIntfCmdT)
	return this.equals(otherCmd.commandT)
}

// Append is not supported
func (this *SetEthIntfCmdT) Append(cmd CommandI) (bool, error) {
	return false, fmt.Errorf("Unsupported")
}

// DeleteEthIntfCmdT implements command for deletion of Ethernet interface
type DeleteEthIntfCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewDeleteEthIntfCmdT creates new instance of DeleteEthIntfCmdT type
func NewDeleteEthIntfCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *DeleteEthIntfCmdT {
	changes := make([]*diff.Change, maxEthChangeIdxC)
	changes[ethChangeIdxC] = vlan
	return &DeleteEthIntfCmdT{
		commandT: newCommandT("delete ethernet interface", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and deletes Ethernet interface
func (this *DeleteEthIntfCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	isGoingToBeDeleted := true
	return doEthIntfCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *DeleteEthIntfCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	isGoingToBeDeleted := false
	return doEthIntfCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *DeleteEthIntfCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *DeleteEthIntfCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*DeleteEthIntfCmdT)
	return this.equals(otherCmd.commandT)
}

// Append is not supported
func (this *DeleteEthIntfCmdT) Append(cmd CommandI) (bool, error) {
	return false, fmt.Errorf("Unsupported")
}

func doEthIntfCmd(cmd *commandT, isDelete bool, shouldBeAbleOnlyToUndo bool) error {
	if cmd.isAbleOnlyToUndo() != shouldBeAbleOnlyToUndo {
		return cmd.createErrorAccordingToExecutionState()
	}

	cmd.dumpInternalData()

	var err error
	var ifname string
	if isDelete {
		ifname, err = utils.ConvertGoInterfaceIntoString(cmd.changes[0].From)
	} else {
		ifname, err = utils.ConvertGoInterfaceIntoString(cmd.changes[0].To)
	}
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if isDelete {
		_, err = (*cmd.ethSwitchMgmt).DeleteEthernetIntf(ctx, &interfaces.DeleteEthernetIntfRequest{
			EthIntf: &interfaces.EthernetIntf{
				Ifname: ifname,
			},
		})
	} else {
		_, err = (*cmd.ethSwitchMgmt).CreateEthernetIntf(ctx, &interfaces.CreateEthernetIntfRequest{
			EthIntf: &interfaces.EthernetIntf{
				Ifname: ifname,
			},
		})
	}
	if err != nil {
		return err
	}

	cmd.finalize()
	return nil
}
