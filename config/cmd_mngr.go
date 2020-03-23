package config

import (
	mgmt "opennos-eth-switch-service/mgmt"
)

type CmdMngrT struct {
	ethSwitchMgmt *mgmt.EthSwitchMgmtClient
}

func (mngr *CmdMngrT) AcquireMgmtClient() (*mgmt.EthSwitchMgmtClient, error) {
	return mngr.ethSwitchMgmt, nil
}

func (mngr *CmdMngrT) ReleaseMgmtClient() (*mgmt.EthSwitchMgmtClient, error) {
	return mngr.ethSwitchMgmt, nil
}
