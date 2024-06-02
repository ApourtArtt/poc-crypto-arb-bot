package broker

type AccountStatus struct {
	CanTrade    bool
	CanDeposit  bool
	CanWithdraw bool
	CanUseSpot  bool
}

func NewAccountStatus(defaultValue bool) AccountStatus {
	return AccountStatus{
		CanTrade:    defaultValue,
		CanDeposit:  defaultValue,
		CanWithdraw: defaultValue,
		CanUseSpot:  defaultValue,
	}
}

func (as AccountStatus) CanTradeSpot() bool {
	return as.CanTrade && as.CanUseSpot
}

func (as AccountStatus) CanBuyAndWithdraw() bool {
	return as.CanTradeSpot() && as.CanWithdraw
}

func (as AccountStatus) CanDepositAndSell() bool {
	return as.CanTradeSpot() && as.CanDeposit
}
