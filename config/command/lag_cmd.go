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
	AggIntfInterfacePathItemIdxC      = 0
	AggIntfIfnamePathItemIdxC         = 1
	AggIntfNamePathItemIdxC           = 2
	AggIntfPathItemsCountC            = 3
	AggIntfMemberEthernetPathItemIdxC = 2
	AggIntfMemberAggIdPathItemIdxC    = 3
	AggIntfMemberPathItemsCountC      = 4

	AggIntfInterfacePathItemC      = "Interface"
	AggIntfMemberEthernetPathItemC = "Ethernet"
	AggIntfMemberAggIdPathItemC    = "AggregateId"
	AggIntfNamePathItemC           = "Name"
)

const (
	lagChangeIdxC = iota
	maxLagChangeIdxC
)

// SetAggIntfCmdT implements command for creating LAG interface
type SetAggIntfCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewSetAggIntfCmdT creates new instance of SetAggIntfCmdT type
func NewSetAggIntfCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *SetAggIntfCmdT {
	changes := make([]*diff.Change, maxLagChangeIdxC)
	changes[lagChangeIdxC] = vlan
	return &SetAggIntfCmdT{
		commandT: newCommandT("set aggregate interface", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and creates LAG interface
func (this *SetAggIntfCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	isGoingToBeDeleted := false
	return doAggIntfCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *SetAggIntfCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	isGoingToBeDeleted := true
	return doAggIntfCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *SetAggIntfCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *SetAggIntfCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*SetAggIntfCmdT)
	return this.equals(otherCmd.commandT)
}

// DeleteAggIntfCmdT implements command for deletion of LAG interface
type DeleteAggIntfCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewDeleteAggIntfCmdT creates new instance of DeleteAggIntfCmdT type
func NewDeleteAggIntfCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *DeleteAggIntfCmdT {
	changes := make([]*diff.Change, maxLagChangeIdxC)
	changes[lagChangeIdxC] = vlan
	return &DeleteAggIntfCmdT{
		commandT: newCommandT("delete aggregate interface", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and deletes LAG interface
func (this *DeleteAggIntfCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	isGoingToBeDeleted := true
	return doAggIntfCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *DeleteAggIntfCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	isGoingToBeDeleted := false
	return doAggIntfCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *DeleteAggIntfCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *DeleteAggIntfCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*DeleteAggIntfCmdT)
	return this.equals(otherCmd.commandT)
}

// SetAggIntfMemberCmdT implements command for add Ethernet interface to LAG
type SetAggIntfMemberCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewSetAggIntfMemberCmdT creates new instance of SetAggIntfMemberCmdT type
func NewSetAggIntfMemberCmdT(change *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *SetAggIntfMemberCmdT {
	changes := make([]*diff.Change, maxLagChangeIdxC)
	changes[lagChangeIdxC] = change
	return &SetAggIntfMemberCmdT{
		commandT: newCommandT("set aggregate interface member", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and adds Ethernet interface to LAG
func (this *SetAggIntfMemberCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	isGoingToBeDeleted := false
	return doAggIntfMemberCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *SetAggIntfMemberCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	isGoingToBeDeleted := true
	return doAggIntfMemberCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *SetAggIntfMemberCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *SetAggIntfMemberCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*SetAggIntfMemberCmdT)
	return this.equals(otherCmd.commandT)
}

// DeleteAggIntfMemberCmdT implements command for remove Ethernet interface from LAG
type DeleteAggIntfMemberCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewDeleteAggIntfMemberCmdT creates new instance of DeleteAggIntfMemberCmdT type
func NewDeleteAggIntfMemberCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *DeleteAggIntfMemberCmdT {
	changes := make([]*diff.Change, maxLagChangeIdxC)
	changes[lagChangeIdxC] = vlan
	return &DeleteAggIntfMemberCmdT{
		commandT: newCommandT("delete aggregate interface member", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and removes Ethernet interface from LAG
func (this *DeleteAggIntfMemberCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	isGoingToBeDeleted := true
	return doAggIntfMemberCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *DeleteAggIntfMemberCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	isGoingToBeDeleted := false
	return doAggIntfMemberCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *DeleteAggIntfMemberCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *DeleteAggIntfMemberCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*DeleteAggIntfMemberCmdT)
	return this.equals(otherCmd.commandT)
}

func doAggIntfCmd(cmd *commandT, isDelete bool, shouldBeAbleOnlyToUndo bool) error {
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

func doAggIntfMemberCmd(cmd *commandT, isDelete bool, shouldBeAbleOnlyToUndo bool) error {
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

	ethIntfs := make([]*interfaces.EthernetIntf, len(cmd.changes))
	for i, change := range cmd.changes {
		ethIntfs[i] = &interfaces.EthernetIntf{
			Ifname: change.Path[AggIntfIfnamePathItemIdxC],
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
