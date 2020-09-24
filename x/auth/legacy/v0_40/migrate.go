package v040

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	v039auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_39"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
)

func convertBaseAccount(old *v039auth.BaseAccount) *BaseAccount {
	any, err := tx.PubKeyToAny(old.PubKey)
	if err != nil {
		panic(err)
	}

	return &BaseAccount{
		Address:       old.Address,
		PubKey:        any,
		AccountNumber: old.AccountNumber,
		Sequence:      old.Sequence,
	}
}

func convertBaseVestingAccount(old *v039auth.BaseVestingAccount) *BaseVestingAccount {
	baseAccount := convertBaseAccount(old.BaseAccount)

	return &BaseVestingAccount{
		BaseAccount:      baseAccount,
		OriginalVesting:  old.OriginalVesting,
		DelegatedFree:    old.DelegatedFree,
		DelegatedVesting: old.DelegatedVesting,
		EndTime:          old.EndTime,
	}
}

// Migrate accepts exported x/auth genesis state from v0.38/v0.39 and migrates
// it to v0.40 x/auth genesis state. The migration includes:
//
// - Removing coins from account encoding.
func Migrate(authGenState v039auth.GenesisState) *GenesisState {
	// Convert v0.39 accounts to v0.40 ones.
	var v040Accounts = make([]GenesisAccount, len(authGenState.Accounts))
	for i, account := range authGenState.Accounts {
		// set coins to nil and allow the JSON encoding to omit coins.
		if err := account.SetCoins(nil); err != nil {
			panic(fmt.Sprintf("failed to set account coins to nil: %s", err))
		}

		switch account.(type) {
		case *v039auth.BaseAccount:
			{
				v040Accounts[i] = convertBaseAccount(account.(*v039auth.BaseAccount))
			}
		case *v039auth.ModuleAccount:
			{
				v039Account := account.(*v039auth.ModuleAccount)
				v040Accounts[i] = &ModuleAccount{
					BaseAccount: convertBaseAccount(v039Account.BaseAccount),
					Name:        v039Account.Name,
					Permissions: v039Account.Permissions,
				}
			}
		case *v039auth.BaseVestingAccount:
			{
				v040Accounts[i] = convertBaseVestingAccount(account.(*v039auth.BaseVestingAccount))
			}
		case *v039auth.ContinuousVestingAccount:
			{
				v039Account := account.(*v039auth.ContinuousVestingAccount)
				v040Accounts[i] = &ContinuousVestingAccount{
					BaseVestingAccount: convertBaseVestingAccount(v039Account.BaseVestingAccount),
					StartTime:          v039Account.EndTime,
				}
			}
		case *v039auth.DelayedVestingAccount:
			{
				v039Account := account.(*v039auth.DelayedVestingAccount)
				v040Accounts[i] = &DelayedVestingAccount{
					BaseVestingAccount: convertBaseVestingAccount(v039Account.BaseVestingAccount),
				}
			}
		case *v039auth.PeriodicVestingAccount:
			{
				v039Account := account.(*v039auth.PeriodicVestingAccount)
				vestingPeriods := make([]Period, len(v039Account.VestingPeriods))
				for i, period := range v039Account.VestingPeriods {
					vestingPeriods[i] = Period{
						Length: period.Length,
						Amount: period.Amount,
					}
				}
				v040Accounts[i] = &PeriodicVestingAccount{
					BaseVestingAccount: convertBaseVestingAccount(v039Account.BaseVestingAccount),
					StartTime:          v039Account.StartTime,
					VestingPeriods:     vestingPeriods,
				}
			}
		}
	}

	// Convert v0.40 accounts into Anys.
	anys := make([]*codectypes.Any, len(v040Accounts))
	for i, v040Account := range v040Accounts {
		anys[i] = codectypes.UnsafePackAny(v040Account)
	}

	return &GenesisState{
		Params: Params{
			MaxMemoCharacters:      authGenState.Params.MaxMemoCharacters,
			TxSigLimit:             authGenState.Params.TxSigLimit,
			TxSizeCostPerByte:      authGenState.Params.TxSizeCostPerByte,
			SigVerifyCostED25519:   authGenState.Params.SigVerifyCostED25519,
			SigVerifyCostSecp256k1: authGenState.Params.SigVerifyCostSecp256k1,
		},
		Accounts: anys,
	}
}
