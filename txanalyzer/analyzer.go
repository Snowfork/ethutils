package txanalyzer

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/tranvictor/ethutils"
	"github.com/tranvictor/ethutils/reader"
)

type TxAnalyzer struct {
	reader *reader.EthReader
	addrdb AddressDatabase
}

func (self *TxAnalyzer) setBasicTxInfo(txinfo ethutils.TxInfo, result *TxResult) {
	// fmt.Printf("From: %s\n", )
	result.From = AddressResult{
		Address: txinfo.Tx.Extra.From.Hex(),
		Name:    self.addrdb.GetName(txinfo.Tx.Extra.From.Hex()),
	}
	// fmt.Printf("Value: %f eth\n", )
	result.Value = fmt.Sprintf("%f", ethutils.BigToFloat(txinfo.Tx.Value(), 18))
	// fmt.Printf("To: %s\n", txinfo.Tx.To().Hex())
	result.To = AddressResult{
		Address: txinfo.Tx.To().Hex(),
		Name:    self.addrdb.GetName(txinfo.Tx.To().Hex()),
	}
	// fmt.Printf("Nonce: %d\n", txinfo.Tx.Nonce())
	result.Nonce = fmt.Sprintf("%d", txinfo.Tx.Nonce())
	// fmt.Printf("Gas price: %f gwei\n", ethutils.BigToFloat(txinfo.Tx.GasPrice(), 9))
	result.GasPrice = fmt.Sprintf("%f", ethutils.BigToFloat(txinfo.Tx.GasPrice(), 9))
	// fmt.Printf("Gas limit: %d\n", txinfo.Tx.Gas())
	result.GasLimit = fmt.Sprintf("%d", txinfo.Tx.Gas())
}

func (self *TxAnalyzer) nonArrayParamAsString(t abi.Type, value interface{}) string {
	switch t.T {
	case abi.StringTy: // variable arrays are written at the end of the return bytes
		return fmt.Sprintf("%s", value.(string))
	case abi.IntTy, abi.UintTy:
		return fmt.Sprintf("%d (0x%x)", value, value)
	case abi.BoolTy:
		return fmt.Sprintf("%t", value.(bool))
	case abi.AddressTy:
		return fmt.Sprintf("%s - (%s)", value.(common.Address).Hex(), self.addrdb.GetName(value.(common.Address).Hex()))
	case abi.HashTy:
		return fmt.Sprintf("%s", value.(common.Hash).Hex())
	case abi.BytesTy:
		return fmt.Sprintf("0x%s", common.Bytes2Hex(value.([]byte)))
	case abi.FixedBytesTy:
		word := []byte{}
		for i := 0; i < int(reflect.TypeOf(value).Size()); i++ {
			word = append(word, byte(0))
		}
		reflect.Copy(reflect.ValueOf(word), reflect.ValueOf(value))
		return fmt.Sprintf("0x%s", common.Bytes2Hex(word))
	case abi.FunctionTy:
		return fmt.Sprintf("0x%s", common.Bytes2Hex(value.([]byte)))
	default:
		return fmt.Sprintf("%v", value)
	}
}

func (self *TxAnalyzer) paramAsString(t abi.Type, value interface{}) string {
	switch t.T {
	case abi.SliceTy:
		realVal := reflect.ValueOf(value)
		result := ""
		for i := 0; i < realVal.Len(); i++ {
			result += fmt.Sprintf("\n%d. %v", i, self.paramAsString(*t.Elem, realVal.Index(i).Interface()))
		}
		return result
	case abi.ArrayTy:
		realVal := reflect.ValueOf(value)
		result := ""
		for i := 0; i < realVal.Len(); i++ {
			result += fmt.Sprintf("\n%d. %v", i, self.paramAsString(*t.Elem, realVal.Index(i).Interface()))
		}
		return result
	default:
		return self.nonArrayParamAsString(t, value)
	}
}

func findEventById(a *abi.ABI, topic []byte) (*abi.Event, error) {
	for _, event := range a.Events {
		if bytes.Equal(event.ID().Bytes(), topic) {
			return &event, nil
		}
	}
	return nil, fmt.Errorf("no event with id: %#x", topic)
}

func (self *TxAnalyzer) AnalyzeMethodCall(abi *abi.ABI, data []byte) (method string, params []ParamResult, err error) {
	m, err := abi.MethodById(data)
	if err != nil {
		return "", []ParamResult{}, err
	}
	method = m.Name
	ps, err := m.Inputs.UnpackValues(data[4:])
	if err != nil {
		return method, []ParamResult{}, err
	}
	params = []ParamResult{}
	for i, input := range m.Inputs {
		params = append(params, ParamResult{
			Name:  input.Name,
			Type:  input.Type.String(),
			Value: self.paramAsString(input.Type, ps[i]),
		})
	}
	return method, params, nil
}

func (self *TxAnalyzer) AnalyzeLog(abi *abi.ABI, l *types.Log) (LogResult, error) {
	logResult := LogResult{
		Name:   "",
		Topics: []TopicResult{},
		Data:   []ParamResult{},
	}

	event, err := findEventById(abi, l.Topics[0].Bytes())
	if err != nil {
		return logResult, err
	}
	logResult.Name = event.Name

	iArgs, niArgs := SplitEventArguments(event.Inputs)
	for j, topic := range l.Topics[1:] {
		logResult.Topics = append(logResult.Topics, TopicResult{
			Name:  iArgs[j].Name,
			Value: topic.Hex(),
		})
	}

	params, err := niArgs.UnpackValues(l.Data)
	if err != nil {
		return logResult, err
	}
	for i, input := range niArgs {
		logResult.Data = append(logResult.Data, ParamResult{
			Name:  input.Name,
			Type:  input.Type.String(),
			Value: self.paramAsString(input.Type, params[i]),
		})
	}
	return logResult, nil
}

func (self *TxAnalyzer) analyzeContractTx(txinfo ethutils.TxInfo, abi *abi.ABI, result *TxResult) {
	result.Contract.Address = txinfo.Tx.To().Hex()
	result.Contract.Name = self.addrdb.GetName(result.Contract.Address)
	// fmt.Printf("------------------------------------------Contract call info-------------------------------------------------------------\n")
	data := txinfo.Tx.Data()
	methodName, params, err := self.AnalyzeMethodCall(abi, data)
	if err != nil {
		result.Error = fmt.Sprintf("Cannot analyze the method call: %s", err)
		return
	}

	result.Method = methodName
	result.Params = append(result.Params, params...)

	m, err := abi.MethodById(data)
	if err != nil {
		result.Error = fmt.Sprintf("Cannot get corresponding method from the ABI: %s", err)
		return
	}

	ps, err := m.Inputs.UnpackValues(data[4:])
	if err != nil {
		result.Error = fmt.Sprintf("Cannot parse params: %s", err)
		return
	}

	logs := txinfo.Receipt.Logs
	// fmt.Printf("Event logs:\n")
	for _, l := range logs {
		logResult, err := self.AnalyzeLog(abi, l)
		if err != nil {
			result.Error += fmt.Sprintf("%s", err)
		}
		result.Logs = append(result.Logs, logResult)
		// event, err := findEventById(abi, l.Topics[0].Bytes())
		// if err == nil {
		// 	// fmt.Printf("Log %d:\n", i)
		// 	// fmt.Printf("    Event name: %s\n", event.Name)
		// 	logResult := LogResult{
		// 		Name:   event.Name,
		// 		Topics: []TopicResult{},
		// 		Data:   []ParamResult{},
		// 	}
		// 	iArgs, niArgs := SplitEventArguments(event.Inputs)
		// 	for j, topic := range l.Topics[1:] {
		// 		// fmt.Printf("    Topic %d - %s: %s\n", j+1, iArgs[j].Name, topic.Hex())
		// 		logResult.Topics = append(logResult.Topics, TopicResult{
		// 			Name:  iArgs[j].Name,
		// 			Value: topic.Hex(),
		// 		})
		// 	}
		// 	// fmt.Printf("    Data: %s\n", common.Bytes2Hex(l.Data))
		// 	// fmt.Printf("    Inputs: %+v\n", event.Inputs)
		// 	params, err := niArgs.UnpackValues(l.Data)
		// 	if err != nil {
		// 		result.Error += fmt.Sprintf("Cannot parse params: %s", err)
		// 	} else {
		// 		for i, input := range niArgs {
		// 			// fmt.Printf("    %s (%s): ", input.Name, input.Type)
		// 			logResult.Data = append(logResult.Data, ParamResult{
		// 				Name:  input.Name,
		// 				Type:  input.Type.String(),
		// 				Value: self.paramAsString(input.Type, params[i]),
		// 			})
		// 		}
		// 	}
		// 	result.Logs = append(result.Logs, logResult)
		// } else {
		// 	result.Error += fmt.Sprintf("Cannot find event of topic %s in the abi.", l.Topics[0].Hex())
		// }
	}
	if isGnosisMultisig(m, ps) {
		// fmt.Printf("    ==> Gnosis Multisig init data:\n")
		result.GnosisInit = &GnosisResult{
			Contract: AddressResult{},
			Method:   "",
			Params:   []ParamResult{},
		}
		self.setGnosisMultisigInitData(m.Inputs, ps, result)
	}
}

// this function only compares thee function and param names against
// standard gnosis multisig. No bytecode is compared.
func isGnosisMultisig(method *abi.Method, params []interface{}) bool {
	if method.Name != "submitTransaction" {
		return false
	}
	expectedParams := []string{"destination", "value", "data"}
	for i, input := range method.Inputs {
		if input.Name != expectedParams[i] {
			return false
		}
	}
	return true
}

func (self *TxAnalyzer) setGnosisMultisigInitData(inputs []abi.Argument, params []interface{}, result *TxResult) {
	contract := params[0].(common.Address)
	// fmt.Printf("    Contract: %s (%s)\n", contract.Hex(), "TODO")
	result.GnosisInit.Contract = AddressResult{
		Address: contract.Hex(),
		Name:    self.addrdb.GetName(contract.Hex()),
	}
	data := params[2].([]byte)
	abi, err := self.reader.GetABI(contract.Hex())
	if err != nil {
		result.Error = fmt.Sprintf("Cannot get abi of the contract: %s", err)
		return
	}
	method, err := abi.MethodById(data)
	if err != nil {
		result.Error = fmt.Sprintf("Cannot get corresponding method from the ABI: %s", err)
		return
	}
	// fmt.Printf("    Method: %s\n", method.Name)
	result.GnosisInit.Method = method.Name
	ps, err := method.Inputs.UnpackValues(data[4:])
	if err != nil {
		result.Error = fmt.Sprintf("Cannot parse params: %s", err)
		return
	}
	// fmt.Printf("    Params:\n")
	for i, input := range method.Inputs {
		result.GnosisInit.Params = append(result.GnosisInit.Params, ParamResult{
			Name:  input.Name,
			Type:  input.Type.String(),
			Value: self.paramAsString(input.Type, ps[i]),
		})
		// fmt.Printf("        %s (%s): ", input.Name, input.Type)
	}
}

func (self *TxAnalyzer) AnalyzeOffline(txinfo *ethutils.TxInfo, abi *abi.ABI, isContract bool) *TxResult {
	result := NewTxResult()
	// fmt.Printf("==========================================Transaction info===============================================================\n")
	// fmt.Printf("tx hash: %s\n", tx)
	result.Hash = txinfo.Tx.Hash().Hex()
	// fmt.Printf("mining status: %s\n", txinfo.Status)
	result.Status = txinfo.Status
	if txinfo.Status == "done" || txinfo.Status == "reverted" {
		self.setBasicTxInfo(*txinfo, result)
		if !isContract {
			// fmt.Printf("tx type: normal\n")
			result.TxType = "normal"
		} else {
			// fmt.Printf("tx type: contract call\n")
			result.TxType = "contract call"
			self.analyzeContractTx(*txinfo, abi, result)
		}
	}
	// fmt.Printf("=========================================================================================================================\n")
	return result
}

// print all info on the tx
func (self *TxAnalyzer) Analyze(tx string) *TxResult {
	txinfo, err := self.reader.TxInfoFromHash(tx)
	if err != nil {
		return &TxResult{
			Error: fmt.Sprintf("getting tx info failed: %s", err),
		}
	}

	code, err := self.reader.GetCode(txinfo.Tx.To().Hex())
	if err != nil {
		return &TxResult{
			Error: fmt.Sprintf("checking tx type failed: %s", err),
		}
	}
	isContract := len(code) > 0

	if isContract {
		abi, err := self.reader.GetABI(txinfo.Tx.To().Hex())
		if err != nil {
			return &TxResult{
				Error: fmt.Sprintf("Cannot get abi of the contract: %s", err),
			}
		}
		return self.AnalyzeOffline(&txinfo, abi, true)
	} else {
		return self.AnalyzeOffline(&txinfo, nil, false)
	}
}

func (self *TxAnalyzer) SetAddressDatabase(db AddressDatabase) {
	self.addrdb = db
}

func NewAnalyzer() *TxAnalyzer {
	return &TxAnalyzer{
		reader.NewEthReader(),
		NewDefaultAddressDatabase(),
	}
}

func NewRopstenAnalyzer() *TxAnalyzer {
	return &TxAnalyzer{
		reader.NewRopstenReader(),
		NewDefaultAddressDatabase(),
	}
}

func NewTomoAnalyzer() *TxAnalyzer {
	return &TxAnalyzer{
		reader.NewTomoReader(),
		NewDefaultAddressDatabase(),
	}
}
