package coin

type Address = string

type Network struct {
	Name             string
	DepositPossible  bool
	WithdrawPossible bool
}
