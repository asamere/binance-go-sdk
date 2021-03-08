package transaction

import (
	"github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/types/msg"
	"github.com/binance-chain/go-sdk/types/tx"
)

type SendTokenResult struct {
	tx.TxCommitResult
}

func (c *client) SendToken(transfers []msg.Transfer, sync bool, options ...Option) (*SendTokenResult, error) {
	fromAddr := c.keyManager.GetAddr()
	fromCoins := types.Coins{}
	for _, t := range transfers {
		t.Coins = t.Coins.Sort()
		fromCoins = fromCoins.Plus(t.Coins)
	}
	sendMsg := msg.CreateSendMsg(fromAddr, fromCoins, transfers)
	commit, err := c.broadcastMsg(sendMsg, sync, options...)
	if err != nil {
		return nil, err
	}
	return &SendTokenResult{*commit}, err

}

func (c *client) SignTokenTransfer(transfers []msg.Transfer, sync bool, options ...Option) (rawtx []byte, err error) {
	fromAddr := c.keyManager.GetAddr()
	fromCoins := types.Coins{}
	for _, t := range transfers {
		t.Coins = t.Coins.Sort()
		fromCoins = fromCoins.Plus(t.Coins)
	}
	sendMsg := msg.CreateSendMsg(fromAddr, fromCoins, transfers)
	// prepare message to sign
	signMsg := &tx.StdSignMsg{
		ChainID:       c.chainId,
		AccountNumber: -1,
		Sequence:      -1,
		Memo:          "",
		Msgs:          []msg.Msg{sendMsg},
		Source:        tx.Source,
	}

	for _, op := range options {
		signMsg = op(signMsg)
	}

	if signMsg.Sequence == -1 || signMsg.AccountNumber == -1 {
		fromAddr := c.keyManager.GetAddr()
		acc, err := c.queryClient.GetAccount(fromAddr.String())
		if err != nil {
			return nil, err
		}
		signMsg.Sequence = acc.Sequence
		signMsg.AccountNumber = acc.Number
	}

	for _, m := range signMsg.Msgs {
		if err := m.ValidateBasic(); err != nil {
			return nil, err
		}
	}

	rawtx, err = c.keyManager.Sign(*signMsg)
	if err != nil {
		return nil, err
	}

	return rawtx, nil
}
