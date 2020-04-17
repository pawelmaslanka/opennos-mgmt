package command

import (
	"context"
	mgmt "opennos-eth-switch-service/mgmt"
	"opennos-eth-switch-service/mgmt/interfaces"
	"opennos-mgmt/utils"
	"time"

	"github.com/r3labs/diff"
)

const (
	LagIntfInterfacePathItemIdxC      = 0
	LagIntfIfnamePathItemIdxC         = 1
	LagIntfNamePathItemIdxC           = 2
	LagIntfPathItemsCountC            = 3
	LagIntfMemberEthernetPathItemIdxC = 2
	LagIntfMemberAggIdPathItemIdxC    = 3
	LagIntfMemberPathItemsCountC      = 4

	LagIntfInterfacePathItemC      = "Interface"
	LagIntfMemberEthernetPathItemC = "Ethernet"
	LagIntfMemberAggIdPathItemC    = "AggregateId"
	LagIntfNamePathItemC           = "Name"
)

const (
	lagChangeIdxC = iota
	maxLagChangeIdxC
)

// SetLagIntfCmdT implements command for creating LAG interface
type SetLagIntfCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewSetLagIntfCmdT creates new instance of SetLagIntfCmdT type
func NewSetLagIntfCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *SetLagIntfCmdT {
	changes := make([]*diff.Change, maxLagChangeIdxC)
	changes[lagChangeIdxC] = vlan
	return &SetLagIntfCmdT{
		commandT: newCommandT("set lag interface", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and creates LAG interface
func (this *SetLagIntfCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	isGoingToBeDeleted := false
	return doLagIntfCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *SetLagIntfCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	isGoingToBeDeleted := true
	return doLagIntfCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *SetLagIntfCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *SetLagIntfCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*SetLagIntfCmdT)
	return this.equals(otherCmd.commandT)
}

// DeleteLagIntfCmdT implements command for deletion of LAG interface
type DeleteLagIntfCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewDeleteLagIntfCmdT creates new instance of DeleteLagIntfCmdT type
func NewDeleteLagIntfCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *DeleteLagIntfCmdT {
	changes := make([]*diff.Change, maxLagChangeIdxC)
	changes[lagChangeIdxC] = vlan
	return &DeleteLagIntfCmdT{
		commandT: newCommandT("delete lag interface", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and deletes LAG interface
func (this *DeleteLagIntfCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	isGoingToBeDeleted := true
	return doLagIntfCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *DeleteLagIntfCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	isGoingToBeDeleted := false
	return doLagIntfCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *DeleteLagIntfCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *DeleteLagIntfCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*DeleteLagIntfCmdT)
	return this.equals(otherCmd.commandT)
}

// SetLagIntfMemberCmdT implements command for add Ethernet interface to LAG
type SetLagIntfMemberCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewSetLagIntfMemberCmdT creates new instance of SetLagIntfMemberCmdT type
func NewSetLagIntfMemberCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *SetLagIntfMemberCmdT {
	changes := make([]*diff.Change, maxLagChangeIdxC)
	changes[lagChangeIdxC] = vlan
	return &SetLagIntfMemberCmdT{
		commandT: newCommandT("set lag interface member", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and adds Ethernet interface to LAG
func (this *SetLagIntfMemberCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	isGoingToBeDeleted := false
	return doLagIntfMemberCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *SetLagIntfMemberCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	isGoingToBeDeleted := true
	return doLagIntfMemberCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *SetLagIntfMemberCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *SetLagIntfMemberCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*SetLagIntfMemberCmdT)
	return this.equals(otherCmd.commandT)
}

// DeleteLagIntfMemberCmdT implements command for remove Ethernet interface from LAG
type DeleteLagIntfMemberCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewDeleteLagIntfMemberCmdT creates new instance of DeleteLagIntfMemberCmdT type
func NewDeleteLagIntfMemberCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *DeleteLagIntfMemberCmdT {
	changes := make([]*diff.Change, maxLagChangeIdxC)
	changes[lagChangeIdxC] = vlan
	return &DeleteLagIntfMemberCmdT{
		commandT: newCommandT("delete lag interface member", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and removes Ethernet interface from LAG
func (this *DeleteLagIntfMemberCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	isGoingToBeDeleted := true
	return doLagIntfMemberCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *DeleteLagIntfMemberCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	isGoingToBeDeleted := false
	return doLagIntfMemberCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *DeleteLagIntfMemberCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *DeleteLagIntfMemberCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*DeleteLagIntfMemberCmdT)
	return this.equals(otherCmd.commandT)
}

func doLagIntfCmd(cmd *commandT, isDelete bool, shouldBeAbleOnlyToUndo bool) error {
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
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if isDelete {
		_, err = (*cmd.ethSwitchMgmt).DeleteAggregateIntf(ctx, &interfaces.DeleteAggregateIntfRequest{
			AggIntf: &interfaces.AggregateIntf{
				Ifname: ifname,
			},
		})
	} else {
		_, err = (*cmd.ethSwitchMgmt).CreateAggregateIntf(ctx, &interfaces.CreateAggregateIntfRequest{
			AggIntf: &interfaces.AggregateIntf{
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

func doLagIntfMemberCmd(cmd *commandT, isDelete bool, shouldBeAbleOnlyToUndo bool) error {
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
		return nil
	}

	ethIntfs := make([]*interfaces.EthernetIntf, len(cmd.changes))
	for i, change := range cmd.changes {
		ethIntfs[i] = &interfaces.EthernetIntf{
			Ifname: change.Path[LagIntfIfnamePathItemIdxC],
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if isDelete {
		_, err = (*cmd.ethSwitchMgmt).RemoveEthernetIntfFromAggregateIntf(ctx, &interfaces.RemoveEthernetIntfFromAggregateIntfRequest{
			AggIntf: &interfaces.AggregateIntf{
				Ifname: ifname,
			},
			EthIntfs: ethIntfs,
		})
	} else {
		_, err = (*cmd.ethSwitchMgmt).AddEthernetIntfToAggregateIntf(ctx, &interfaces.AddEthernetIntfToAggregateIntfRequest{
			AggIntf: &interfaces.AggregateIntf{
				Ifname: ifname,
			},
			EthIntfs: ethIntfs,
		})
	}
	if err != nil {
		return err
	}

	cmd.finalize()
	return nil
}
