package config

import (
	"fmt"
	cmd "opennos-mgmt/config/command"
	"opennos-mgmt/gnmi/modeldata/oc"
	"opennos-mgmt/utils"
	"strings"

	log "github.com/golang/glog"
	"github.com/jinzhu/copier"
	"github.com/r3labs/diff"
)

const (
	idSetAggIntfMemberNameFmt    = "sm-%s"
	idDeleteAggIntfMemberNameFmt = "sm-%s"
)

func isCreateOrDeleteAggIntf(change *diff.Change) bool {
	if len(change.Path) != cmd.AggIntfPathItemsCountC {
		return false
	}

	if change.Path[cmd.AggIntfInterfacePathItemIdxC] != cmd.AggIntfInterfacePathItemC {
		return false
	}

	if !strings.Contains(change.Path[cmd.AggIntfIfnamePathItemIdxC], "ae") {
		return false
	}

	if change.Path[cmd.AggIntfAggregationPathItemIdxC] != cmd.AggIntfAggregationPathItemC {
		return false
	}

	return true
}

func extractAggIdFromEthIntf(ethIfname string, ethIntf *oc.Interface_Ethernet, isDelete bool) ([]diff.Change, error) {
	changes := make([]diff.Change, 0)
	aggIfname := ethIntf.GetAggregateId()
	if len(aggIfname) > 0 {
		fmt.Printf("Found belonging %s to %s\n", ethIfname, aggIfname)
		if isDelete {
			changes = append(changes, *deleteAggIntfMemberDiffChange(ethIfname, aggIfname))
		} else {
			changes = append(changes, *createAggIntfMemberDiffChange(ethIfname, aggIfname))
		}
	}

	return changes, nil
}

func extractAggIntfLagType(ifname string, aggIntf *oc.Interface_Aggregation, isDelete bool) ([]diff.Change, error) {
	changes := make([]diff.Change, 0)
	lagType := aggIntf.GetLagType()
	fmt.Printf("Found lag type %d for %s\n", lagType, ifname)
	if !isDelete {
		changes = append(changes, *createAggIntfLagTypeDiffChange(ifname, lagType))
	}

	return changes, nil
}

func createAggIntfLagTypeDiffChange(ifname string, lagType oc.E_OpenconfigIfAggregate_AggregationType) *diff.Change {
	var ch diff.Change
	ch.Type = diff.CREATE
	ch.From = nil
	ch.To = lagType
	ch.Path = make([]string, cmd.AggIntfMemberPathItemsCountC)
	ch.Path[cmd.AggIntfInterfacePathItemIdxC] = cmd.AggIntfInterfacePathItemC
	ch.Path[cmd.AggIntfIfnamePathItemIdxC] = ifname
	ch.Path[cmd.AggIntfAggregationPathItemIdxC] = cmd.AggIntfAggregationPathItemC
	ch.Path[cmd.AggIntfLagTypePathItemIdxC] = cmd.AggIntfLagTypePathItemC

	fmt.Printf("Aggregate interface diff change:\n%v\n", ch)

	return &ch
}

func createAggIntfMemberDiffChange(ethIfname string, aggIfname string) *diff.Change {
	var ch diff.Change
	ch.Type = diff.CREATE
	ch.From = nil
	ch.To = aggIfname
	ch.Path = make([]string, cmd.AggIntfMemberPathItemsCountC)
	ch.Path[cmd.AggIntfInterfacePathItemIdxC] = cmd.AggIntfInterfacePathItemC
	ch.Path[cmd.AggIntfIfnamePathItemIdxC] = ethIfname
	ch.Path[cmd.AggIntfMemberEthernetPathItemIdxC] = cmd.AggIntfMemberEthernetPathItemC
	ch.Path[cmd.AggIntfMemberAggIdPathItemIdxC] = cmd.AggIntfMemberAggIdPathItemC

	fmt.Printf("Aggregate interface diff change:\n%v\n", ch)

	return &ch
}

func deleteAggIntfMemberDiffChange(ethIfname string, aggIfname string) *diff.Change {
	var ch diff.Change
	ch.Type = diff.DELETE
	ch.From = aggIfname
	ch.To = nil
	ch.Path = make([]string, cmd.AggIntfMemberPathItemsCountC)
	ch.Path[cmd.AggIntfInterfacePathItemIdxC] = cmd.AggIntfInterfacePathItemC
	ch.Path[cmd.AggIntfIfnamePathItemIdxC] = ethIfname
	ch.Path[cmd.AggIntfMemberEthernetPathItemIdxC] = cmd.AggIntfMemberEthernetPathItemC
	ch.Path[cmd.AggIntfMemberAggIdPathItemIdxC] = cmd.AggIntfMemberAggIdPathItemC

	fmt.Printf("Aggregate interface diff change:\n%v\n", ch)

	return &ch
}

func isCreateOrDeleteAggIntfAggregation(change *diff.Change) bool {
	return isCreateOrDeleteAggIntf(change)
}

func isChangedAggIntfMember(change *diff.Change) bool {
	if len(change.Path) != cmd.AggIntfMemberPathItemsCountC {
		return false
	}

	if (change.Path[cmd.AggIntfInterfacePathItemIdxC] == cmd.AggIntfInterfacePathItemC) && (change.Path[cmd.AggIntfMemberEthernetPathItemIdxC] == cmd.AggIntfMemberEthernetPathItemC) && (change.Path[cmd.AggIntfMemberAggIdPathItemIdxC] == cmd.AggIntfMemberAggIdPathItemC) {
		return true
	}

	return false
}

func isChangedAggIntf(change *diff.Change) bool {
	if len(change.Path) != cmd.AggIntfPathItemsCountC {
		return false
	}

	if (change.Path[cmd.AggIntfInterfacePathItemIdxC] == cmd.AggIntfInterfacePathItemC) && strings.Contains(change.Path[cmd.AggIntfIfnamePathItemIdxC], "ae") && (change.Path[cmd.AggIntfNamePathItemIdxC] == cmd.AggIntfNamePathItemC) {
		return true
	}

	return false
}

func isChangedAggIntfLagType(change *diff.Change) bool {
	if len(change.Path) != cmd.AggIntfLagTypePathItemsCountC {
		return false
	}

	if change.Path[cmd.AggIntfInterfacePathItemIdxC] != cmd.AggIntfInterfacePathItemC {
		return false
	}

	if !strings.Contains(change.Path[cmd.AggIntfIfnamePathItemIdxC], "ae") {
		return false
	}

	if change.Path[cmd.AggIntfAggregationPathItemIdxC] != cmd.AggIntfAggregationPathItemC {
		return false
	}

	if change.Path[cmd.AggIntfLagTypePathItemIdxC] != cmd.AggIntfLagTypePathItemC {
		return false
	}

	return true
}

func (this *ConfigMngrT) validateSetAggIntfMemberChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ifname := changeItem.Change.Path[cmd.AggIntfIfnamePathItemIdxC]
	if !this.isEthIntfAvailable(ifname) {
		return fmt.Errorf("Ethernet interface %s is not available", ifname)
	}

	aggIfname, err := utils.ConvertGoInterfaceIntoString(changeItem.Change.To)
	if err != nil {
		return err
	}

	log.Infof("Requested add Ethernet interface %s as LAG member %s", ifname, aggIfname)
	setAggIntfMemberCmd := cmd.NewSetAggIntfMemberCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForSetAggIntfMember(aggIfname, ifname); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from interface %s:\n%s",
			setAggIntfMemberCmd.GetName(), ifname, err)
	}

	if this.transHasBeenStarted {
		id := fmt.Sprintf(idSetAggIntfMemberNameFmt, aggIfname)
		if err := this.appendCmdToTransaction(id, setAggIntfMemberCmd, setAggIntfMemberC, false); err != nil {
			return err
		}
	}

	if err := this.transConfigLookupTbl.setAggIntfMember(aggIfname, ifname); err != nil {
		return err
	}

	changeItem.MarkAsProcessed()

	return nil
}

func (this *ConfigMngrT) validateDeleteAggIntfMemberChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ifname := changeItem.Change.Path[cmd.AggIntfIfnamePathItemIdxC]
	if !this.isEthIntfAvailable(ifname) {
		return fmt.Errorf("Ethernet interface %s is not available", ifname)
	}

	aggIfname, err := utils.ConvertGoInterfaceIntoString(changeItem.Change.From)
	if err != nil {
		return err
	}

	var newChange diff.Change
	needsCreateNewChange := (changeItem.Change.Type == diff.UPDATE) && (changeItem.Change.To != nil)
	if needsCreateNewChange {
		// Update type carries info about old and new LAG. Let's create new change item
		// in order to process new LAG by SetAggIntfMemberCmd
		newAggIfname, err := utils.ConvertGoInterfaceIntoString(changeItem.Change.To)
		if err != nil {
			return err
		}
		if len(newAggIfname) > 0 {
			copier.Copy(&newChange, changeItem.Change)
			newChange.Type = diff.CREATE
			newChange.From = nil
			// Update current change
			changeItem.Change.Type = diff.DELETE
			changeItem.Change.To = nil
		} else {
			needsCreateNewChange = false
		}
	}

	log.Infof("Requested remove Ethernet interface %s from LAG member %s", ifname, aggIfname)
	deleteAggIntfMemberCmd := cmd.NewDeleteAggIntfMemberCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForDeleteAggIntfMember(aggIfname, ifname); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from interface %s:\n%s",
			deleteAggIntfMemberCmd.GetName(), ifname, err)
	}

	if this.transHasBeenStarted {
		id := fmt.Sprintf(idDeleteAggIntfMemberNameFmt, aggIfname)
		if err := this.appendCmdToTransaction(id, deleteAggIntfMemberCmd, deleteAggIntfMemberC, false); err != nil {
			return err
		}
	}

	if err := this.transConfigLookupTbl.deleteAggIntfMember(aggIfname, ifname); err != nil {
		return err
	}

	if needsCreateNewChange {
		changelog.Changes = append(changelog.Changes, NewDiffChangeMgmtT(&newChange))
	}

	changeItem.MarkAsProcessed()

	return nil
}

func (this *ConfigMngrT) validateSetAggIntfLagTypeChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	aggIfname := changeItem.Change.Path[cmd.AggIntfIfnamePathItemIdxC]
	log.Infof("Requested set aggregate interface LAG type %s", aggIfname)
	if changeItem.Change.Type == diff.UPDATE {
		return fmt.Errorf("Dependency error: Transitions LAG type between LACP and STATIC for aggregate interface %s is not supported. Please re-create current aggregate interface", aggIfname)
	}

	changeItem.MarkAsProcessed()

	return nil
}

func (this *ConfigMngrT) validateSetAggIntfChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	log.Infof("validateSetAggIntfChange:\n%v\n", changeItem.Change)
	aggIfname, err := utils.ConvertGoInterfaceIntoString(changeItem.Change.To)
	if err != nil {
		return err
	}

	log.Infof("Requested set LAG interface %s", aggIfname)
	device := (*this.transCandidateConfig).(*oc.Device)
	aggIntf := device.GetInterface(aggIfname)
	if aggIntf == nil {
		return fmt.Errorf("Cannot find aggregate interface %s in config", aggIfname)
	}

	aggregation := aggIntf.GetAggregation()
	if aggregation == nil {
		return fmt.Errorf("Cannot find aggregation settings for aggregate interface %s in config", aggIfname)
	}

	if aggregation.LagType == oc.OpenconfigIfAggregate_AggregationType_UNSET {
		return fmt.Errorf("LAG type for aggregate interface %s cannot be unset", aggIfname)
	} else if aggregation.LagType == oc.OpenconfigIfAggregate_AggregationType_LACP {
		lacp := device.GetLacp()
		if lacp == nil {
			return fmt.Errorf("Cannot find LACP configuration which is required for aggregate interface %s as LACP type", aggIfname)
		}

		lacpIntf := lacp.GetInterface(aggIfname)
		if lacpIntf == nil {
			return fmt.Errorf("Cannot find in config corresponding LACP to aggregate interface %s", aggIfname)
		}

		// Delegate to create aggregate interface to LACP config module
		// changeItem.MarkAsProcessed()
		// return nil
	}

	lagTypeChange, err := this.findAggIntfLagTypeFromChangelog(aggIfname, changelog)
	if err != nil {
		return err
	}

	setAggIntfCmd := cmd.NewSetAggIntfCmdT(changeItem.Change, lagTypeChange.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForSetAggIntf(aggIfname); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from LAG interface %s:\n%s",
			setAggIntfCmd.GetName(), aggIfname, err)
	}

	if this.transHasBeenStarted {
		if err := this.appendCmdToTransaction(aggIfname, setAggIntfCmd, setAggIntfC, false); err != nil {
			return err
		}
	}

	if err := this.transConfigLookupTbl.setAggIntf(aggIfname); err != nil {
		return err
	}

	changeItem.MarkAsProcessed()
	lagTypeChange.MarkAsProcessed()

	return nil
}

func (this *ConfigMngrT) findAggIntfLagTypeFromChangelog(ifname string, changelog *DiffChangelogMgmtT) (*DiffChangeMgmtT, error) {
	var err error = nil
	var changeItem *DiffChangeMgmtT
	lagType := oc.OpenconfigIfAggregate_AggregationType_UNSET
	for _, ch := range changelog.Changes {
		if isChangedAggIntfLagType(ch.Change) {
			if ch.Change.Path[cmd.AggIntfIfnamePathItemIdxC] == ifname {
				lagType = ch.Change.To.(oc.E_OpenconfigIfAggregate_AggregationType)
				changeItem = ch
				break
			}
		}
	}

	if lagType == oc.OpenconfigIfAggregate_AggregationType_UNSET {
		err = fmt.Errorf("Could not found set LAG type request")
	}

	return changeItem, err
}

func (this *ConfigMngrT) validateDeleteAggIntfChange(changeItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	aggIfname, err := utils.ConvertGoInterfaceIntoString(changeItem.Change.From)
	if err != nil {
		return err
	}

	log.Infof("Requested delete LAG interface %s", aggIfname)
	deleteAggIntfCmd := cmd.NewDeleteAggIntfCmdT(changeItem.Change, this.ethSwitchMgmtClient)
	if err := this.transConfigLookupTbl.checkDependenciesForDeleteAggIntf(aggIfname); err != nil {
		return fmt.Errorf("Cannot %q because there are dependencies from LAG interface %s:\n%s",
			deleteAggIntfCmd.GetName(), aggIfname, err)
	}

	if this.transHasBeenStarted {
		if err := this.appendCmdToTransaction(aggIfname, deleteAggIntfCmd, deleteAggIntfC, false); err != nil {
			return err
		}
	}

	if err := this.transConfigLookupTbl.deleteAggIntf(aggIfname); err != nil {
		return err
	}

	changeItem.MarkAsProcessed()

	return nil
}

func findSetAggIntfMemberChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type != diff.DELETE {
				if isChangedAggIntfMember(ch.Change) {
					if ch.Change.To != nil {
						return ch, true
					}
				}
			}
		}
	}

	return nil, false
}

func findDeleteAggIntfMemberChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type != diff.CREATE {
				if isChangedAggIntfMember(ch.Change) {
					if ch.Change.From != nil {
						return ch, true
					}
				}
			}
		}
	}

	return nil, false
}

func findSetAggIntfChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type != diff.DELETE {
				if isChangedAggIntf(ch.Change) {
					if ch.Change.To != nil {
						return ch, true
					}
				}
			}
		}
	}

	return nil, false
}

func findSetAggIntfLagTypeChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type != diff.DELETE {
				if isChangedAggIntfLagType(ch.Change) {
					if ch.Change.To != nil {
						return ch, true
					}
				}
			}
		}
	}

	return nil, false
}

func findDeleteAggIntfChange(changelog *DiffChangelogMgmtT) (change *DiffChangeMgmtT, exists bool) {
	for _, ch := range changelog.Changes {
		if !ch.IsProcessed() {
			if ch.Change.Type != diff.CREATE {
				if isChangedAggIntf(ch.Change) {
					if ch.Change.From != nil {
						return ch, true
					}
				}
			}
		}
	}

	return nil, false
}

func (this *ConfigMngrT) processSetAggIntfMemberFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to set LAG interface member
		if change, exists := findSetAggIntfMemberChange(changelog); exists {
			if err := this.validateSetAggIntfMemberChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (this *ConfigMngrT) processDeleteAggIntfMemberFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to delete LAG interface member
		if change, exists := findDeleteAggIntfMemberChange(changelog); exists {
			if err := this.validateDeleteAggIntfMemberChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (this *ConfigMngrT) processSetAggIntfLagTypeFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to set aggregate interface LAG type
		if change, exists := findSetAggIntfLagTypeChange(changelog); exists {
			if err := this.validateSetAggIntfLagTypeChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (this *ConfigMngrT) processSetAggIntfFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to set LAG interface
		if change, exists := findSetAggIntfChange(changelog); exists {
			if err := this.validateSetAggIntfChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (this *ConfigMngrT) processDeleteAggIntfFromChangelog(changelog *DiffChangelogMgmtT) error {
	if changelog.isProcessed() {
		return nil
	}

	for {
		// Repeat till there is not any change related to delete LAG interface
		if change, exists := findDeleteAggIntfChange(changelog); exists {
			if err := this.validateDeleteAggIntfChange(change, changelog); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (this *ConfigMngrT) setAggIntf(device *oc.Device) error {
	var err error
	for _, aggIfname := range this.configLookupTbl.aggIfnameByIdx {
		var change diff.Change
		change.Type = diff.CREATE
		change.From = nil
		change.To = aggIfname
		change.Path = make([]string, cmd.AggIntfPathItemsCountC)
		change.Path[cmd.AggIntfInterfacePathItemIdxC] = cmd.AggIntfInterfacePathItemC
		change.Path[cmd.AggIntfIfnamePathItemIdxC] = aggIfname
		change.Path[cmd.AggIntfNamePathItemIdxC] = cmd.AggIntfNamePathItemC

		var lagTypeCh diff.Change
		lagTypeCh.Type = diff.CREATE
		lagTypeCh.From = nil
		lagTypeCh.To = device.GetInterface(aggIfname).GetAggregation().LagType
		lagTypeCh.Path = make([]string, cmd.AggIntfLagTypePathItemsCountC)
		lagTypeCh.Path[cmd.AggIntfInterfacePathItemIdxC] = cmd.AggIntfInterfacePathItemC
		lagTypeCh.Path[cmd.AggIntfIfnamePathItemIdxC] = aggIfname
		lagTypeCh.Path[cmd.AggIntfAggregationPathItemIdxC] = cmd.AggIntfAggregationPathItemC
		lagTypeCh.Path[cmd.AggIntfLagTypePathItemIdxC] = cmd.AggIntfLagTypePathItemC

		command := cmd.NewSetAggIntfCmdT(&change, &lagTypeCh, this.ethSwitchMgmtClient)
		if err = this.appendCmdToTransaction(aggIfname, command, setAggIntfC, true); err != nil {
			return err
		}
	}

	return nil
}

func (this *ConfigMngrT) setAggIntfMember(device *oc.Device) error {
	var err error
	for aggIdx, aggIfname := range this.configLookupTbl.aggIfnameByIdx {
		if _, exists := this.configLookupTbl.ethByAgg[aggIdx]; !exists {
			continue
		}

		for _, ethIdx := range this.configLookupTbl.ethByAgg[aggIdx].IdxTs() {
			var change diff.Change
			change.Type = diff.CREATE
			change.From = nil
			change.To = aggIfname
			change.Path = make([]string, cmd.AggIntfMemberPathItemsCountC)
			change.Path[cmd.AggIntfInterfacePathItemIdxC] = cmd.AggIntfInterfacePathItemC
			change.Path[cmd.AggIntfIfnamePathItemIdxC] = this.configLookupTbl.ethIfnameByIdx[ethIdx]
			change.Path[cmd.AggIntfMemberEthernetPathItemIdxC] = cmd.AggIntfMemberEthernetPathItemC
			change.Path[cmd.AggIntfMemberAggIdPathItemIdxC] = cmd.AggIntfMemberAggIdPathItemC

			command := cmd.NewSetAggIntfMemberCmdT(&change, this.ethSwitchMgmtClient)
			id := fmt.Sprintf(idSetAggIntfMemberNameFmt, aggIfname)
			if err = this.appendCmdToTransaction(id, command, setAggIntfMemberC, true); err != nil {
				return err
			}
		}
	}

	return nil
}
