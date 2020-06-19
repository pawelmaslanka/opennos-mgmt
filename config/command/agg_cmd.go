package command

import (
	"context"
	"fmt"
	mgmt "opennos-eth-switch-service/mgmt"
	"opennos-eth-switch-service/mgmt/interfaces"
	"opennos-mgmt/gnmi/modeldata/oc"
	"opennos-mgmt/utils"
	"time"

	"github.com/r3labs/diff"
)

const (
	// Common for all subtrees changes of aggregate interface
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

	// Aggregation subtree change
	AggIntfAggregationPathItemIdxC = 2
	AggIntfAggregationPathItemC    = "Aggregation"

	AggIntfAggregationPathItemsCountC = 3

	// LAG type change
	AggIntfLagTypePathItemIdxC = 3
	AggIntfLagTypePathItemC    = "LagType"

	AggIntfLagTypePathItemsCountC = 4
)

const (
	aggIntfChangeIdxC = iota
	aggIntfLagTypeChangeIdxC
	maxAggIntfChangeIdxC
)

const (
	aggIntfMemberChangeIdxC = iota
	maxAggIntfMemberChangeIdxC
)

const (
	maxAggIntfLagTypeChangeIdxC = 1
)

// SetAggIntfCmdT implements command for creating LAG interface
type SetAggIntfCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewSetAggIntfCmdT creates new instance of SetAggIntfCmdT type
func NewSetAggIntfCmdT(aggIntfChange *diff.Change, lagTypeChange *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *SetAggIntfCmdT {
	changes := make([]*diff.Change, maxAggIntfChangeIdxC)
	changes[aggIntfChangeIdxC] = aggIntfChange
	changes[aggIntfLagTypeChangeIdxC] = lagTypeChange
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

// Append is not supported
func (this *SetAggIntfCmdT) Append(cmd CommandI) (bool, error) {
	return false, fmt.Errorf("Unsupported")
}

// DeleteAggIntfCmdT implements command for deletion of LAG interface
type DeleteAggIntfCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewDeleteAggIntfCmdT creates new instance of DeleteAggIntfCmdT type
func NewDeleteAggIntfCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *DeleteAggIntfCmdT {
	changes := make([]*diff.Change, maxAggIntfChangeIdxC)
	changes[aggIntfChangeIdxC] = vlan
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

// Append is not supported
func (this *DeleteAggIntfCmdT) Append(cmd CommandI) (bool, error) {
	return false, fmt.Errorf("Unsupported")
}

// SetAggIntfMemberCmdT implements command for add Ethernet interface to LAG
type SetAggIntfMemberCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewSetAggIntfMemberCmdT creates new instance of SetAggIntfMemberCmdT type
func NewSetAggIntfMemberCmdT(change *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *SetAggIntfMemberCmdT {
	changes := make([]*diff.Change, maxAggIntfMemberChangeIdxC)
	changes[aggIntfMemberChangeIdxC] = change
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

// Append extracts internal data of 'other' and attach them to 'this'
func (this *SetAggIntfMemberCmdT) Append(other CommandI) (bool, error) {
	return this.append(other)
}

// DeleteAggIntfMemberCmdT implements command for remove Ethernet interface from LAG
type DeleteAggIntfMemberCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewDeleteAggIntfMemberCmdT creates new instance of DeleteAggIntfMemberCmdT type
func NewDeleteAggIntfMemberCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *DeleteAggIntfMemberCmdT {
	changes := make([]*diff.Change, maxAggIntfMemberChangeIdxC)
	changes[aggIntfMemberChangeIdxC] = vlan
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

// Append extracts internal data of 'other' and attach them to 'this'
func (this *DeleteAggIntfMemberCmdT) Append(other CommandI) (bool, error) {
	return this.append(other)
}

func doAggIntfCmd(cmd *commandT, isDelete bool, shouldBeAbleOnlyToUndo bool) error {
	if cmd.isAbleOnlyToUndo() != shouldBeAbleOnlyToUndo {
		return cmd.createErrorAccordingToExecutionState()
	}

	cmd.dumpInternalData()

	var err error
	var ifname string
	var lagType int64
	if isDelete {
		ifname, err = utils.ConvertGoInterfaceIntoString(cmd.changes[aggIntfChangeIdxC].From)
	} else {
		ifname, err = utils.ConvertGoInterfaceIntoString(cmd.changes[aggIntfChangeIdxC].To)
		if err != nil {
			return err
		}

		lagType, err = utils.ConvertGoInterfaceIntoInt64(cmd.changes[aggIntfLagTypeChangeIdxC].To)
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
		var lagTypeReq interfaces.CreateAggregateIntfRequest_AggregationType
		if lagType == int64(oc.OpenconfigIfAggregate_AggregationType_LACP) {
			lagTypeReq = interfaces.CreateAggregateIntfRequest_LACP
		} else {
			lagTypeReq = interfaces.CreateAggregateIntfRequest_STATIC
		}

		_, err = (*cmd.ethSwitchMgmt).CreateAggregateIntf(ctx, &interfaces.CreateAggregateIntfRequest{
			AggIntf: &interfaces.AggregateIntf{
				Ifname: ifname,
			},
			AggType: lagTypeReq,
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
		ifname, err = utils.ConvertGoInterfaceIntoString(cmd.changes[aggIntfMemberChangeIdxC].From)
	} else {
		ifname, err = utils.ConvertGoInterfaceIntoString(cmd.changes[aggIntfMemberChangeIdxC].To)
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
