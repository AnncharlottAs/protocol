/*
	Copyright 2017 - 2018 OneLedger

	Handle setting any options for the node.
*/
package app

import (
	"github.com/Oneledger/protocol/node/comm"
	"github.com/Oneledger/protocol/node/data"
	"github.com/Oneledger/protocol/node/global"
	"github.com/Oneledger/protocol/node/id"
	"github.com/Oneledger/protocol/node/log"
)

// Arguments for registration
type RegisterArguments struct {
	Identity   string
	Chain      string
	PublicKey  string
	PrivateKey string
}

func SetOption(app *Application, key string, value string) bool {
	log.Debug("Setting Application Options", "key", key, "value", value)

	switch key {

	case "Register":
		var arguments RegisterArguments
		result, err := comm.Deserialize([]byte(value), &arguments)
		if err != nil {
			log.Error("Can't set options", "err", err)
			return false
		}
		args := result.(*RegisterArguments)
		publicKey, privateKey := id.GenerateKeys()
		RegisterLocally(app, args.Identity, args.Identity, id.ParseAccountType(args.Chain),
			publicKey, privateKey)

	default:
		log.Warn("Unknown Option", "key", key)
		return false
	}
	return true

}

// Register Identities and Accounts from the user.
func RegisterLocally(app *Application, name string, scope string, chain data.ChainType,
	publicKey id.PublicKey, privateKey id.PrivateKey) bool {

	status := false

	if chain == data.UNKNOWN {
		return status
	}

	// Accounts are relative to a chain
	accountName := name + "-" + scope

	if !app.Accounts.Exists(chain, accountName) {
		log.Debug("Registering New Account", "accountName", accountName)
		account := id.NewAccount(chain, accountName, publicKey, privateKey)
		app.Accounts.Add(account)

		// TODO: This should add to a list
		global.Current.NodeAccountName = accountName

		log.Debug("New Account", "key", account.AccountKey(), "account", account)

		// Fill in the balance
		if !app.Utxo.Exists(account.AccountKey()) {
			balance := data.NewBalance(0, "OLT")
			buffer, _ := comm.Serialize(balance)
			app.Utxo.Delivered.Set(account.AccountKey(), buffer)
			app.Utxo.Commit()
			log.Debug("New Utxo", "key", account.AccountKey(), "balance", balance)
		}
		status = true

	} else {
		log.Debug("Existing Account", "accountName", accountName)
	}

	// Identities are global
	if !app.Identities.Exists(name) {
		log.Debug("Registering a New Identity", "name", name)
		identity := id.NewIdentity(name, "Contact Info", false, global.Current.NodeName)
		app.Identities.Add(identity)
		status = true

	} else {
		log.Debug("Not Registering Existing Identity", "name", name)
		app.Identities.Dump()
	}

	return status
}