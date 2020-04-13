package config

import (
	"fmt"
	"math"
	"opennos-mgmt/gnmi/modeldata/oc"
	"reflect"

	"github.com/r3labs/diff"
)

const (
	mgmtTransManagementPathItemIdxC           = 0
	mgmtTransTransactionPathItemIdxC          = 1
	mgmtTransDefaultConfigActionPathItemIdxC  = 2
	mgmtTransConfigActionPathItemIdxC         = 2
	mgmtTransCommitConfirmTimeoutPathItemIdxC = 2
	mgmtTransPathItemsCountC                  = 3

	mgmtTransManagementPathItemC           = "Management"
	mgmtTransTransactionPathItemC          = "Transaction"
	mgmtTransDefaultConfigActionPathItemC  = "DefaultConfigAction"
	mgmtTransConfigActionPathItemC         = "ConfigAction"
	mgmtTransCommitConfirmTimeoutPathItemC = "CommitConfirmTimeout"
)

func (cfgMngr *ConfigMngrT) getCurrentTransDefaultConfigAction() oc.E_OpenconfigManagement_TRANS_TYPE {
	device := cfgMngr.runningConfig.(*oc.Device)
	return device.GetOrCreateManagement().GetOrCreateTransaction().GetDefaultConfigAction()
}

func checkMgmtTransParamIfItIsGoingToBeUnset(value interface{}) (bool, error) {
	if (value == nil) || (reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil()) {
		return true, nil
	}

	switch v := value.(type) {
	case oc.E_OpenconfigManagement_TRANS_TYPE:
		return v == oc.OpenconfigManagement_TRANS_TYPE_UNSET, nil
	case uint16:
		return v == 0, nil
	default:
		return false, fmt.Errorf("Cannot convert %v to any of [uint16, E_OpenconfigManagement_TRANS_TYPE], unsupported type, got: %T", v, v)
	}
}

func findDisallowedManagementTreeNodeDeleteOperation(changelog *DiffChangelogMgmtT) (*diff.Change, bool) {
	for _, ch := range changelog.Changes {
		if (len(ch.Change.Path) > 0) && (ch.Change.Path[mgmtTransManagementPathItemIdxC] == mgmtTransManagementPathItemC) {
			if ch.Change.Type != diff.CREATE {
				if unset, err := checkMgmtTransParamIfItIsGoingToBeUnset(ch.Change.To); err != nil {
					return nil, false
				} else if unset {
					return ch.Change, true
				}
			}
		}
	}

	return nil, false
}

func (cfgMngr *ConfigMngrT) findTransDefaultConfigActionChange(changelog *DiffChangelogMgmtT) (oc.E_OpenconfigManagement_TRANS_TYPE, error) {
	// Find the latest one request of change this parameter
	defaultConfigAction := oc.OpenconfigManagement_TRANS_TYPE_UNSET
	for _, ch := range changelog.Changes {
		if ch.Change.Type == diff.UPDATE {
			if len(ch.Change.Path) == mgmtTransPathItemsCountC {
				if (ch.Change.Path[mgmtTransManagementPathItemIdxC] == mgmtTransManagementPathItemC) && (ch.Change.Path[mgmtTransTransactionPathItemIdxC] == mgmtTransTransactionPathItemC) && (ch.Change.Path[mgmtTransDefaultConfigActionPathItemIdxC] == mgmtTransDefaultConfigActionPathItemC) {
					ch.MarkAsProcessed()
					defaultConfigAction = ch.Change.To.(oc.E_OpenconfigManagement_TRANS_TYPE)
				}
			}
		}
	}

	if defaultConfigAction != oc.OpenconfigManagement_TRANS_TYPE_UNSET {
		return defaultConfigAction, nil
	}

	return cfgMngr.getCurrentTransDefaultConfigAction(), nil
}

func (cfgMngr *ConfigMngrT) findTransConfigActionChange(changelog *DiffChangelogMgmtT) (oc.E_OpenconfigManagement_TRANS_TYPE, error) {
	// Find the latest one request of change this parameter
	configAction := oc.OpenconfigManagement_TRANS_TYPE_UNSET
	for _, ch := range changelog.Changes {
		if ch.Change.Type == diff.UPDATE {
			if len(ch.Change.Path) == mgmtTransPathItemsCountC {
				if (ch.Change.Path[mgmtTransManagementPathItemIdxC] == mgmtTransManagementPathItemC) && (ch.Change.Path[mgmtTransTransactionPathItemIdxC] == mgmtTransTransactionPathItemC) && (ch.Change.Path[mgmtTransConfigActionPathItemIdxC] == mgmtTransConfigActionPathItemC) {
					ch.MarkAsProcessed()
					configAction = ch.Change.To.(oc.E_OpenconfigManagement_TRANS_TYPE)
				}
			}
		}
	}

	if configAction != oc.OpenconfigManagement_TRANS_TYPE_UNSET {
		return configAction, nil
	}

	return oc.OpenconfigManagement_TRANS_TYPE_UNSET, nil
}

func (cfgMngr *ConfigMngrT) findTransCommitConfirmTimeoutChange(changelog *DiffChangelogMgmtT) (uint16, error) {
	// Find the latest one request of change this parameter
	var timeout uint16 = math.MaxUint16
	for _, ch := range changelog.Changes {
		if ch.Change.Type == diff.UPDATE {
			if len(ch.Change.Path) == mgmtTransPathItemsCountC {
				if (ch.Change.Path[mgmtTransManagementPathItemIdxC] == mgmtTransManagementPathItemC) && (ch.Change.Path[mgmtTransTransactionPathItemIdxC] == mgmtTransTransactionPathItemC) && (ch.Change.Path[mgmtTransCommitConfirmTimeoutPathItemIdxC] == mgmtTransCommitConfirmTimeoutPathItemC) {
					ch.MarkAsProcessed()
					timeout = ch.Change.To.(uint16)
				}
			}
		}
	}

	if timeout != math.MaxUint16 {
		return timeout, nil
	}

	device := cfgMngr.runningConfig.(*oc.Device)
	return device.GetOrCreateManagement().GetOrCreateTransaction().GetCommitConfirmTimeout(), nil
}
