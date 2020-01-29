package binance

import (
	"encoding/json"
	"io/ioutil"
	"strconv"

	"github.com/pkg/errors"
)

type rawTransaction struct {
	TransactionId       int     `json:"tranId"`
}

// Margin account borrow (MARGIN)
// Post /sapi/v1/margin/loan

// Name	Type	Mandatory	Description
// asset	STRING	YES
// amount	DECIMAL	YES
// recvWindow	LONG	NO
// timestamp	LONG	YES

func (as *apiService) Borrow(bor BorrowRequest){
	params := make(map[string]string)
	params["asset"] = bor.Asset
	params["amount"] = strconv.FormatFloat(bor.Amount, 'f', -1, 64)
	if bor.RecvWindow != 0 {
		params["recvWindow"] = strconv.FormatInt(recvWindow(bor.RecvWindow), 10)
	}
	params["timestamp"] = strconv.FormatInt(unixMillis(bor.Timestamp), 10)

	res, err := as.request("POST", "sapi/v1/margin/loan", params, true, true)
	if err != nil {
		return
	}
	textRes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return
	}

	rawResponse := rawTransaction{}
	if err := json.Unmarshal(textRes, &rawResponse); err != nil {
		return
	}

	// t, err := timeFromUnixTimestampFloat(rawResponse.TransactionId)
	// if err != nil {
	// 	return
	// }

	return
}

// Margin account repay (MARGIN)
// Post /sapi/v1/margin/repay
// Repay loan for margin account.
//
// Weight: 1
//
// Parameters:
//
// Name	Type	Mandatory	Description
// asset	STRING	YES
// amount	DECIMAL	YES
// recvWindow	LONG	NO
// timestamp	LONG	YES
func (as *apiService) Repay(bor BorrowRequest){
	params := make(map[string]string)
	params["asset"] = bor.Asset
	params["amount"] = strconv.FormatFloat(bor.Amount, 'f', -1, 64)
	if bor.RecvWindow != 0 {
		params["recvWindow"] = strconv.FormatInt(recvWindow(bor.RecvWindow), 10)
	}
	params["timestamp"] = strconv.FormatInt(unixMillis(bor.Timestamp), 10)

	res, err := as.request("POST", "sapi/v1/margin/repay", params, true, true)
	if err != nil {
		return
	}
	textRes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return
	}

	rawResponse := rawTransaction{}
	if err := json.Unmarshal(textRes, &rawResponse); err != nil {
		return
	}
	return
}

// Query margin account details (USER_DATA)
// Get /sapi/v1/margin/account
// Weight: 5
//
// Parameters:
//
// None
func (as *apiService) MarginAccount(ar AccountRequest) (*MarginAccount, error) {
	params := make(map[string]string)
	params["timestamp"] = strconv.FormatInt(unixMillis(ar.Timestamp), 10)
	if ar.RecvWindow != 0 {
		params["recvWindow"] = strconv.FormatInt(recvWindow(ar.RecvWindow), 10)
	}

	res, err := as.request("GET", "/sapi/v1/margin/account", params, true, true)
	if err != nil {
		return nil, err
	}
	textRes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read response from account.get")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, as.handleError(textRes)
	}

	rawAccount := struct {
		BorrowEnabled    		bool 		`json:"borrowEnabled"`
		MarginLevel  				string `json:"marginLevel"`
		TotalAssetOfBtc  		string `json:"totalAssetOfBtc"`
		TotalLiabilityOfBtc string `json:"totalLiabilityOfBtc"`
		TotalNetAssetOfBtc  string `json:"totalNetAssetOfBtc"`
		TradeEnabled      	bool  	`json:"tradeEnabled"`
		TransferEnabled     bool  	`json:"transferEnabled"`
		UserAssets         []struct {
			Asset  		string `json:"asset"`
			Borrowed 	string `json:"borrowed"`
			Free   		string `json:"free"`
			Interest 	string `json:"interest"`
			Locked 		string `json:"locked"`
			NetAsset 	string `json:"netAsset"`
		}
	}{}
	if err := json.Unmarshal(textRes, &rawAccount); err != nil {
		return nil, errors.Wrap(err, "rawAccount unmarshal failed")
	}
	ml, err := floatFromString(rawAccount.MarginLevel)
	if err != nil {
		return nil, err
	}
	tab, err := floatFromString(rawAccount.TotalAssetOfBtc)
	if err != nil {
		return nil, err
	}
	tlb, err := floatFromString(rawAccount.TotalLiabilityOfBtc)
	if err != nil {
		return nil, err
	}
	tnab, err := floatFromString(rawAccount.TotalNetAssetOfBtc)
	if err != nil {
		return nil, err
	}
	acc := &MarginAccount{
		BorrowEnabled:  			rawAccount.BorrowEnabled,
		MarginLevel:  				ml,
		TotalAssetOfBtc:  		tab,
		TotalLiabilityOfBtc: 	tlb,
		TotalNetAssetOfBtc:   tnab,
		TradeEnabled:     		rawAccount.TradeEnabled,
		TransferEnabled:      rawAccount.TransferEnabled,
	}
	for _, b := range rawAccount.UserAssets {
		f, err := floatFromString(b.Free)
		if err != nil {
			return nil, err
		}
		l, err := floatFromString(b.Locked)
		if err != nil {
			return nil, err
		}
		bo, err := floatFromString(b.Borrowed)
		if err != nil {
			return nil, err
		}
		i, err := floatFromString(b.Interest)
		if err != nil {
			return nil, err
		}
		n, err := floatFromString(b.NetAsset)
		if err != nil {
			return nil, err
		}
		acc.UserAssets = append(acc.UserAssets, &UserAsset{
			Asset:  	b.Asset,
			Borrowed: bo,
			Free:   	f,
			Interest: i,
			Locked: 	l,
			NetAsset: n,
		})
	}

	return acc, nil
}
// Query margin asset (MARKET_DATA)
// Get /sapi/v1/margin/asset
// Weight: 5
//
// Parameters:
//
// Name	Type	Mandatory	Description
// asset	STRING	YES
type rawMarginExecutedOrder struct {
	Symbol        string  `json:"symbol"`
	OrderID       int     `json:"orderId"`
	ClientOrderID string  `json:"clientOrderId"`
	Price         string  `json:"price"`
	OrigQty       string  `json:"origQty"`
	ExecutedQty   string  `json:"executedQty"`
	Status        string  `json:"status"`
	TimeInForce   string  `json:"timeInForce"`
	Type          string  `json:"type"`
	Side          string  `json:"side"`
	StopPrice     string  `json:"stopPrice"`
	IcebergQty    string  `json:"icebergQty"`
	Time          float64 `json:"time"`
}

func (as *apiService) MarginNewOrder(or NewOrderRequest) (*ProcessedOrder, error) {
	params := make(map[string]string)
	params["symbol"] = or.Symbol
	params["side"] = string(or.Side)
	params["type"] = string(or.Type)
	params["quantity"] = strconv.FormatFloat(or.Quantity, 'f', 3, 32)
	if string(or.Type) != "MARKET"{
		params["price"] = strconv.FormatFloat(or.Price, 'f', 3, 32)
	}
	params["timestamp"] = strconv.FormatInt(unixMillis(or.Timestamp), 10)
	if or.NewClientOrderID != "" {
		params["newClientOrderId"] = or.NewClientOrderID
	}
	if or.TimeInForce != ""{
		params["timeInForce"] = string(or.TimeInForce)
	}
	if or.StopPrice != 0 {
		params["stopPrice"] = strconv.FormatFloat(or.StopPrice, 'f', -1, 64)
	}
	if or.IcebergQty != 0 {
		params["icebergQty"] = strconv.FormatFloat(or.IcebergQty, 'f', -1, 64)
	}
	if or.RecvWindow != 0 {
		params["recvWindow"] = strconv.FormatInt(recvWindow(or.RecvWindow), 10)
	}

	res, err := as.request("POST", "sapi/v1/margin/order", params, true, true)
	if err != nil {
		return nil, err
	}
	textRes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read response from Ticker/24hr")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, as.handleError(textRes)
	}

	rawOrder := struct {
		Symbol        string  `json:"symbol"`
		OrderID       int64   `json:"orderId"`
		ClientOrderID string  `json:"clientOrderId"`
		TransactTime  float64 `json:"transactTime"`
	}{}
	if err := json.Unmarshal(textRes, &rawOrder); err != nil {
		return nil, errors.Wrap(err, "rawOrder unmarshal failed")
	}

	t, err := timeFromUnixTimestampFloat(rawOrder.TransactTime)
	if err != nil {
		return nil, err
	}

	return &ProcessedOrder{
		Symbol:        rawOrder.Symbol,
		OrderID:       rawOrder.OrderID,
		ClientOrderID: rawOrder.ClientOrderID,
		TransactTime:  t,
	}, nil
}
