package reader

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	eu "github.com/tranvictor/ethutils"
)

var (
	DEFAULT_ADDRESS string = "0x0000000000000000000000000000000000000000"
)

type EthReader struct {
	chain             string
	nodes             map[string]EthereumNode
	latestGasPrice    float64
	gasPriceTimestamp int64
	gpmu              sync.Mutex
}

func newEthReaderGeneric(nodes map[string]string, chain string) *EthReader {
	ns := map[string]EthereumNode{}
	for name, c := range nodes {
		ns[name] = NewOneNodeReader(name, c)
	}
	return &EthReader{
		chain:             chain,
		nodes:             ns,
		latestGasPrice:    0.0,
		gasPriceTimestamp: 0,
		gpmu:              sync.Mutex{},
	}
}

func NewKovanReaderWithCustomNodes(nodes map[string]string) *EthReader {
	return newEthReaderGeneric(nodes, "kovan")
}

func NewRinkebyReaderWithCustomNodes(nodes map[string]string) *EthReader {
	return newEthReaderGeneric(nodes, "rinkeby")
}

func NewRopstenReaderWithCustomNodes(nodes map[string]string) *EthReader {
	return newEthReaderGeneric(nodes, "ropsten")
}

func NewKovanReader() *EthReader {
	nodes := map[string]string{
		"kovan-infura": "https://kovan.infura.io/v3/247128ae36b6444d944d4c3793c8e3f5",
	}
	return NewKovanReaderWithCustomNodes(nodes)
}

func NewRinkebyReader() *EthReader {
	nodes := map[string]string{
		"rinkeby-infura": "https://rinkeby.infura.io/v3/247128ae36b6444d944d4c3793c8e3f5",
	}
	return NewRinkebyReaderWithCustomNodes(nodes)
}

func NewRopstenReader() *EthReader {
	nodes := map[string]string{
		"ropsten-infura": "https://ropsten.infura.io/v3/247128ae36b6444d944d4c3793c8e3f5",
	}
	return NewRopstenReaderWithCustomNodes(nodes)
}

func NewTomoReaderWithCustomNodes(nodes map[string]string) *EthReader {
	return newEthReaderGeneric(nodes, "tomo")
}

func NewTomoReader() *EthReader {
	nodes := map[string]string{
		"mainnet-tomo": "https://rpc.tomochain.com",
	}
	return NewTomoReaderWithCustomNodes(nodes)
}

func NewEthReaderWithCustomNodes(nodes map[string]string) *EthReader {
	return newEthReaderGeneric(nodes, "ethereum")
}

func NewEthReader() *EthReader {
	nodes := map[string]string{
		"mainnet-alchemy": "https://eth-mainnet.alchemyapi.io/jsonrpc/YP5f6eM2wC9c2nwJfB0DC1LObdSY7Qfv",
		"mainnet-infura":  "https://mainnet.infura.io/v3/247128ae36b6444d944d4c3793c8e3f5",
	}
	return NewEthReaderWithCustomNodes(nodes)
}

// gas station response
type abiresponse struct {
	Status  string `json:"string"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

func (self *EthReader) GetEthereumABIString(address string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=contract&action=getabi&address=%s&apikey=UBB257TI824FC7HUSPT66KZUMGBPRN3IWV", address))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	abiresp := abiresponse{}
	err = json.Unmarshal(body, &abiresp)
	if err != nil {
		return "", err
	}
	return abiresp.Result, err
}

func (self *EthReader) GetEthereumABI(address string) (*abi.ABI, error) {
	body, err := self.GetEthereumABIString(address)
	if err != nil {
		return nil, err
	}
	result, err := abi.JSON(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// gas station response
type tomoabiresponse struct {
	Contract struct {
		ABICode string `json:"abiCode"`
	} `json:"contract"`
}

func (self *EthReader) GetTomoABIString(address string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://scan.tomochain.com/api/accounts/%s", address))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	abiresp := tomoabiresponse{}
	err = json.Unmarshal(body, &abiresp)
	if err != nil {
		return "", err
	}
	return abiresp.Contract.ABICode, nil
}

func (self *EthReader) GetTomoABI(address string) (*abi.ABI, error) {
	body, err := self.GetTomoABIString(address)
	if err != nil {
		return nil, err
	}
	result, err := abi.JSON(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (self *EthReader) GetRinkebyABIString(address string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api-rinkeby.etherscan.io/api?module=contract&action=getabi&address=%s&apikey=UBB257TI824FC7HUSPT66KZUMGBPRN3IWV", address))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	abiresp := abiresponse{}
	err = json.Unmarshal(body, &abiresp)
	if err != nil {
		return "", err
	}
	return abiresp.Result, err
}

func (self *EthReader) GetRinkebyABI(address string) (*abi.ABI, error) {
	body, err := self.GetRinkebyABIString(address)
	if err != nil {
		return nil, err
	}
	result, err := abi.JSON(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (self *EthReader) GetKovanABIString(address string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api-kovan.etherscan.io/api?module=contract&action=getabi&address=%s&apikey=UBB257TI824FC7HUSPT66KZUMGBPRN3IWV", address))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	abiresp := abiresponse{}
	err = json.Unmarshal(body, &abiresp)
	if err != nil {
		return "", err
	}
	return abiresp.Result, err
}

func (self *EthReader) GetKovanABI(address string) (*abi.ABI, error) {
	body, err := self.GetKovanABIString(address)
	if err != nil {
		return nil, err
	}
	result, err := abi.JSON(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (self *EthReader) GetRopstenABIString(address string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api-ropsten.etherscan.io/api?module=contract&action=getabi&address=%s&apikey=UBB257TI824FC7HUSPT66KZUMGBPRN3IWV", address))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	abiresp := abiresponse{}
	err = json.Unmarshal(body, &abiresp)
	if err != nil {
		return "", err
	}
	return abiresp.Result, err
}

func (self *EthReader) GetRopstenABI(address string) (*abi.ABI, error) {
	body, err := self.GetRopstenABIString(address)
	if err != nil {
		return nil, err
	}
	result, err := abi.JSON(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (self *EthReader) GetABIString(address string) (string, error) {
	switch self.chain {
	case "ethereum":
		return self.GetEthereumABIString(address)
	case "ropsten":
		return self.GetRopstenABIString(address)
	case "kovan":
		return self.GetKovanABIString(address)
	case "rinkeby":
		return self.GetRinkebyABIString(address)
	case "tomo":
		return self.GetTomoABIString(address)
	}
	return "", fmt.Errorf("'%s' chain is not supported", self.chain)
}

func (self *EthReader) GetABI(address string) (*abi.ABI, error) {
	switch self.chain {
	case "ethereum":
		return self.GetEthereumABI(address)
	case "ropsten":
		return self.GetRopstenABI(address)
	case "kovan":
		return self.GetKovanABI(address)
	case "rinkeby":
		return self.GetRinkebyABI(address)
	case "tomo":
		return self.GetTomoABI(address)
	}
	return nil, fmt.Errorf("'%s' chain is not supported", self.chain)
}

func errorInfo(errs []error) string {
	estrs := []string{}
	for i, e := range errs {
		estrs = append(estrs, fmt.Sprintf("%d. %s", i+1, e))
	}
	return strings.Join(estrs, "\n")
}

func wrapError(e error, name string) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("%s: %s", name, e)
}

type estimateGasResult struct {
	Gas   uint64
	Error error
}

func (self *EthReader) EstimateGas(from, to string, priceGwei, value float64, data []byte) (uint64, error) {
	resCh := make(chan estimateGasResult, len(self.nodes))
	for i, _ := range self.nodes {
		n := self.nodes[i]
		go func() {
			gas, err := n.EstimateGas(from, to, priceGwei, value, data)
			resCh <- estimateGasResult{
				Gas:   gas,
				Error: wrapError(err, n.NodeName()),
			}
		}()
	}
	errs := []error{}
	for i := 0; i < len(self.nodes); i++ {
		result := <-resCh
		if result.Error == nil {
			return result.Gas, result.Error
		}
		errs = append(errs, result.Error)
	}
	return 0, fmt.Errorf("Couldn't read from any nodes: %s", errorInfo(errs))
}

type getCodeResponse struct {
	Code  []byte
	Error error
}

func (self *EthReader) GetCode(address string) (code []byte, err error) {
	resCh := make(chan getCodeResponse, len(self.nodes))
	for i, _ := range self.nodes {
		n := self.nodes[i]
		go func() {
			code, err := n.GetCode(address)
			resCh <- getCodeResponse{
				Code:  code,
				Error: wrapError(err, n.NodeName()),
			}
		}()
	}
	errs := []error{}
	for i := 0; i < len(self.nodes); i++ {
		result := <-resCh
		if result.Error == nil {
			return result.Code, result.Error
		}
		errs = append(errs, result.Error)
	}
	return nil, fmt.Errorf("Couldn't read from any nodes: %s", errorInfo(errs))
}

func (self *EthReader) TxInfoFromHash(tx string) (eu.TxInfo, error) {
	txObj, isPending, err := self.TransactionByHash(tx)
	if err != nil {
		return eu.TxInfo{"error", nil, nil, nil}, err
	}
	if txObj == nil {
		return eu.TxInfo{"notfound", nil, nil, nil}, nil
	} else {
		if isPending {
			return eu.TxInfo{"pending", txObj, nil, nil}, nil
		} else {
			receipt, _ := self.TransactionReceipt(tx)
			if receipt == nil {
				return eu.TxInfo{"pending", txObj, nil, nil}, nil
			} else {
				// only byzantium has status field at the moment
				// mainnet, ropsten are byzantium, other chains such as
				// devchain, kovan are not.
				// if PostState is a hash, it is pre-byzantium and all
				// txs with PostState are considered done
				if len(receipt.PostState) == len(common.Hash{}) {
					return eu.TxInfo{"done", txObj, []eu.InternalTx{}, receipt}, nil
				} else {
					if receipt.Status == 1 {
						// successful tx
						return eu.TxInfo{"done", txObj, []eu.InternalTx{}, receipt}, nil
					}
					// failed tx
					return eu.TxInfo{"reverted", txObj, []eu.InternalTx{}, receipt}, nil
				}
			}
		}
	}
}

// gas station response
type gsresponse struct {
	Average float64 `json:"average"`
	Fast    float64 `json:"fast"`
	Fastest float64 `json:"fastest"`
	SafeLow float64 `json:"safeLow"`
}

func (self *EthReader) RecommendedGasPriceKovan() (float64, error) {
	return 50, nil
}

func (self *EthReader) RecommendedGasPriceRinkeby() (float64, error) {
	return 50, nil
}

func (self *EthReader) RecommendedGasPriceRopsten() (float64, error) {
	return 50, nil
}

func (self *EthReader) RecommendedGasPriceTomo() (float64, error) {
	return 1, nil
}

func (self *EthReader) RecommendedGasPriceEthereum() (float64, error) {
	self.gpmu.Lock()
	defer self.gpmu.Unlock()
	if self.latestGasPrice == 0 || time.Now().Unix()-self.gasPriceTimestamp > 30 {
		resp, err := http.Get("https://ethgasstation.info/json/ethgasAPI.json")
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return 0, err
		}
		prices := gsresponse{}
		err = json.Unmarshal(body, &prices)
		if err != nil {
			return 0, err
		}
		self.latestGasPrice = float64(prices.Fast) / 10.0
		self.gasPriceTimestamp = time.Now().Unix()
	}
	return self.latestGasPrice, nil
}

// return gwei
func (self *EthReader) RecommendedGasPrice() (float64, error) {
	switch self.chain {
	case "ethereum":
		return self.RecommendedGasPriceEthereum()
	case "ropsten":
		return self.RecommendedGasPriceRopsten()
	case "kovan":
		return self.RecommendedGasPriceKovan()
	case "rinkeby":
		return self.RecommendedGasPriceRinkeby()
	case "tomo":
		return self.RecommendedGasPriceTomo()
	}
	return 0, fmt.Errorf("'%s' chain is not supported", self.chain)
}

type getBalanceResponse struct {
	Balance *big.Int
	Error   error
}

func (self *EthReader) GetBalance(address string) (balance *big.Int, err error) {
	resCh := make(chan getBalanceResponse, len(self.nodes))
	for i, _ := range self.nodes {
		n := self.nodes[i]
		go func() {
			balance, err := n.GetBalance(address)
			resCh <- getBalanceResponse{
				Balance: balance,
				Error:   wrapError(err, n.NodeName()),
			}
		}()
	}
	errs := []error{}
	for i := 0; i < len(self.nodes); i++ {
		result := <-resCh
		if result.Error == nil {
			return result.Balance, result.Error
		}
		errs = append(errs, result.Error)
	}
	return nil, fmt.Errorf("Couldn't read from any nodes: %s", errorInfo(errs))
}

type getNonceResponse struct {
	Nonce uint64
	Error error
}

func (self *EthReader) GetMinedNonce(address string) (nonce uint64, err error) {
	resCh := make(chan getNonceResponse, len(self.nodes))
	for i, _ := range self.nodes {
		n := self.nodes[i]
		go func() {
			nonce, err := n.GetMinedNonce(address)
			resCh <- getNonceResponse{
				Nonce: nonce,
				Error: wrapError(err, n.NodeName()),
			}
		}()
	}
	errs := []error{}
	for i := 0; i < len(self.nodes); i++ {
		result := <-resCh
		if result.Error == nil {
			return result.Nonce, result.Error
		}
		errs = append(errs, result.Error)
	}
	return 0, fmt.Errorf("Couldn't read from any nodes: %s", errorInfo(errs))
}

func (self *EthReader) GetPendingNonce(address string) (nonce uint64, err error) {
	resCh := make(chan getNonceResponse, len(self.nodes))
	for i, _ := range self.nodes {
		n := self.nodes[i]
		go func() {
			nonce, err := n.GetPendingNonce(address)
			resCh <- getNonceResponse{
				Nonce: nonce,
				Error: wrapError(err, n.NodeName()),
			}
		}()
	}
	errs := []error{}
	for i := 0; i < len(self.nodes); i++ {
		result := <-resCh
		if result.Error == nil {
			return result.Nonce, result.Error
		}
		errs = append(errs, result.Error)
	}
	return 0, fmt.Errorf("Couldn't read from any nodes: %s", errorInfo(errs))
}

type transactionReceiptResponse struct {
	Receipt *types.Receipt
	Error   error
}

func (self *EthReader) TransactionReceipt(txHash string) (receipt *types.Receipt, err error) {
	resCh := make(chan transactionReceiptResponse, len(self.nodes))
	for i, _ := range self.nodes {
		n := self.nodes[i]
		go func() {
			receipt, err := n.TransactionReceipt(txHash)
			resCh <- transactionReceiptResponse{
				Receipt: receipt,
				Error:   wrapError(err, n.NodeName()),
			}
		}()
	}
	errs := []error{}
	for i := 0; i < len(self.nodes); i++ {
		result := <-resCh
		if result.Error == nil {
			return result.Receipt, result.Error
		}
		errs = append(errs, result.Error)
	}
	return nil, fmt.Errorf("Couldn't read from any nodes: %s", errorInfo(errs))
}

type transactionByHashResponse struct {
	Tx        *eu.Transaction
	IsPending bool
	Error     error
}

func (self *EthReader) TransactionByHash(txHash string) (tx *eu.Transaction, isPending bool, err error) {
	resCh := make(chan transactionByHashResponse, len(self.nodes))
	for i, _ := range self.nodes {
		n := self.nodes[i]
		go func() {
			tx, ispending, err := n.TransactionByHash(txHash)
			resCh <- transactionByHashResponse{
				Tx:        tx,
				IsPending: ispending,
				Error:     wrapError(err, n.NodeName()),
			}
		}()
	}
	errs := []error{}
	for i := 0; i < len(self.nodes); i++ {
		result := <-resCh
		if result.Error == nil {
			return result.Tx, result.IsPending, result.Error
		}
		errs = append(errs, result.Error)
	}
	return nil, false, fmt.Errorf("Couldn't read from any nodes: %s", errorInfo(errs))
}

// TODO: this method can't utilize all of the nodes because the result reference
// will be written in parallel and it is not thread safe
// func (self *EthReader) Call(result interface{}, method string, args ...interface{}) error {
// 	for _, node := range self.nodes {
// 		return node.Call(result, method, args...)
// 	}
// 	return fmt.Errorf("no nodes to call")
// }

type readContractToBytesResponse struct {
	Data  []byte
	Error error
}

func (self *EthReader) ReadContractToBytes(atBlock int64, from string, caddr string, abi *abi.ABI, method string, args ...interface{}) ([]byte, error) {
	resCh := make(chan readContractToBytesResponse, len(self.nodes))
	for i, _ := range self.nodes {
		n := self.nodes[i]
		go func() {
			data, err := n.ReadContractToBytes(atBlock, from, caddr, abi, method, args...)
			resCh <- readContractToBytesResponse{
				Data:  data,
				Error: wrapError(err, n.NodeName()),
			}
		}()
	}
	errs := []error{}
	for i := 0; i < len(self.nodes); i++ {
		result := <-resCh
		if result.Error == nil {
			return result.Data, result.Error
		}
		errs = append(errs, result.Error)
	}
	return nil, fmt.Errorf("Couldn't read from any nodes: %s", errorInfo(errs))
}

func (self *EthReader) ReadHistoryContractWithABI(atBlock int64, result interface{}, caddr string, abi *abi.ABI, method string, args ...interface{}) error {
	responseBytes, err := self.ReadContractToBytes(
		int64(atBlock), DEFAULT_ADDRESS, caddr, abi, method, args...)
	if err != nil {
		return err
	}
	return abi.UnpackToInterface(result, method, responseBytes)
}

func (self *EthReader) ReadContractWithABIAndFrom(result interface{}, from string, caddr string, abi *abi.ABI, method string, args ...interface{}) error {
	responseBytes, err := self.ReadContractToBytes(-1, from, caddr, abi, method, args...)
	if err != nil {
		return err
	}
	return abi.UnpackToInterface(result, method, responseBytes)
}

func (self *EthReader) ReadContractWithABI(result interface{}, caddr string, abi *abi.ABI, method string, args ...interface{}) error {
	responseBytes, err := self.ReadContractToBytes(-1, DEFAULT_ADDRESS, caddr, abi, method, args...)
	if err != nil {
		return err
	}
	return abi.UnpackToInterface(result, method, responseBytes)
}

func (self *EthReader) ReadHistoryContract(atBlock int64, result interface{}, caddr string, method string, args ...interface{}) error {
	abi, err := self.GetABI(caddr)
	if err != nil {
		return err
	}
	return self.ReadHistoryContractWithABI(atBlock, result, caddr, abi, method, args...)
}

func (self *EthReader) ReadContract(result interface{}, caddr string, method string, args ...interface{}) error {
	abi, err := self.GetABI(caddr)
	if err != nil {
		return err
	}
	return self.ReadContractWithABI(result, caddr, abi, method, args...)
}

func (self *EthReader) HistoryERC20Balance(atBlock int64, caddr string, user string) (*big.Int, error) {
	abi, err := eu.GetERC20ABI()
	if err != nil {
		return nil, err
	}
	result := big.NewInt(0)
	err = self.ReadHistoryContractWithABI(atBlock, &result, caddr, abi, "balanceOf", eu.HexToAddress(user))
	return result, err
}

func (self *EthReader) ERC20Balance(caddr string, user string) (*big.Int, error) {
	abi, err := eu.GetERC20ABI()
	if err != nil {
		return nil, err
	}
	result := big.NewInt(0)
	err = self.ReadContractWithABI(&result, caddr, abi, "balanceOf", eu.HexToAddress(user))
	return result, err
}

func (self *EthReader) HistoryERC20Decimal(atBlock int64, caddr string) (int64, error) {
	abi, err := eu.GetERC20ABI()
	if err != nil {
		return 0, err
	}
	var result uint8
	err = self.ReadHistoryContractWithABI(atBlock, &result, caddr, abi, "decimals")
	return int64(result), err
}

func (self *EthReader) ERC20Decimal(caddr string) (int64, error) {
	abi, err := eu.GetERC20ABI()
	if err != nil {
		return 0, err
	}
	var result uint8
	err = self.ReadContractWithABI(&result, caddr, abi, "decimals")
	return int64(result), err
}

type headerByNumberResponse struct {
	Header *types.Header
	Error  error
}

func (self *EthReader) HeaderByNumber(number int64) (*types.Header, error) {
	resCh := make(chan headerByNumberResponse, len(self.nodes))
	for i, _ := range self.nodes {
		n := self.nodes[i]
		go func() {
			header, err := n.HeaderByNumber(number)
			resCh <- headerByNumberResponse{
				Header: header,
				Error:  wrapError(err, n.NodeName()),
			}
		}()
	}
	errs := []error{}
	for i := 0; i < len(self.nodes); i++ {
		result := <-resCh
		if result.Error == nil {
			return result.Header, result.Error
		}
		errs = append(errs, result.Error)
	}
	return nil, fmt.Errorf("Couldn't read from any nodes: %s", errorInfo(errs))
}

func (self *EthReader) HistoryERC20Allowance(atBlock int64, caddr string, owner string, spender string) (*big.Int, error) {
	abi, err := eu.GetERC20ABI()
	if err != nil {
		return nil, err
	}
	result := big.NewInt(0)
	err = self.ReadHistoryContractWithABI(
		atBlock,
		&result, caddr, abi,
		"allowance",
		eu.HexToAddress(owner),
		eu.HexToAddress(spender),
	)
	return result, err
}

func (self *EthReader) ERC20Allowance(caddr string, owner string, spender string) (*big.Int, error) {
	abi, err := eu.GetERC20ABI()
	if err != nil {
		return nil, err
	}
	result := big.NewInt(0)
	err = self.ReadContractWithABI(
		&result, caddr, abi,
		"allowance",
		eu.HexToAddress(owner),
		eu.HexToAddress(spender),
	)
	return result, err
}

func (self *EthReader) AddressFromContract(contract string, method string) (*common.Address, error) {
	result := common.Address{}
	err := self.ReadContract(&result, contract, method)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

type getLogsResponse struct {
	Logs  []types.Log
	Error error
}

// if toBlock < 0, it will query to the latest block
func (self *EthReader) GetLogs(fromBlock, toBlock int, addresses []string, topic string) ([]types.Log, error) {
	resCh := make(chan getLogsResponse, len(self.nodes))
	for i, _ := range self.nodes {
		n := self.nodes[i]
		go func() {
			logs, err := n.GetLogs(fromBlock, toBlock, addresses, topic)
			resCh <- getLogsResponse{
				Logs:  logs,
				Error: wrapError(err, n.NodeName()),
			}
		}()
	}
	errs := []error{}
	for i := 0; i < len(self.nodes); i++ {
		result := <-resCh
		if result.Error == nil {
			return result.Logs, result.Error
		}
		errs = append(errs, result.Error)
	}
	return nil, fmt.Errorf("Couldn't read from any nodes: %s", errorInfo(errs))
}

type getBlockResponse struct {
	Block uint64
	Error error
}

func (self *EthReader) CurrentBlock() (uint64, error) {
	resCh := make(chan getBlockResponse, len(self.nodes))
	for i, _ := range self.nodes {
		n := self.nodes[i]
		go func() {
			block, err := n.CurrentBlock()
			resCh <- getBlockResponse{
				Block: block,
				Error: wrapError(err, n.NodeName()),
			}
		}()
	}
	errs := []error{}
	for i := 0; i < len(self.nodes); i++ {
		result := <-resCh
		if result.Error == nil {
			return result.Block, result.Error
		}
		errs = append(errs, result.Error)
	}
	return 0, fmt.Errorf("Couldn't read from any nodes: %s", errorInfo(errs))
}
