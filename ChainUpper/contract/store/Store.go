// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package Store

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/FISCO-BCOS/go-sdk/v3/abi"
	"github.com/FISCO-BCOS/go-sdk/v3/abi/bind"
	"github.com/FISCO-BCOS/go-sdk/v3/types"
	"github.com/ethereum/go-ethereum/common"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = abi.U256
	_ = bind.Bind
	_ = common.Big1
)

// StoreABI is the input ABI used to generate the binding from.
const StoreABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"key\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"value\",\"type\":\"bytes\"}],\"name\":\"ItemSet\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"key\",\"type\":\"bytes32\"}],\"name\":\"getItem\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"items\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"key\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"value\",\"type\":\"bytes\"}],\"name\":\"setItem\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"keys\",\"type\":\"bytes32[]\"},{\"internalType\":\"bytes[]\",\"name\":\"values\",\"type\":\"bytes[]\"}],\"name\":\"setItems\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// StoreBin is the compiled bytecode used for deploying new contracts.
var StoreBin = "0x608060405234801561001057600080fd5b50610a4c806100206000396000f3fe608060405234801561001057600080fd5b506004361061004c5760003560e01c806348f343f3146100515780637c50bba014610081578063aa0372e71461009d578063ff78e5f5146100cd575b600080fd5b61006b600480360381019061006691906104b5565b6100e9565b604051610078919061057b565b60405180910390f35b61009b60048036038101906100969190610602565b610189565b005b6100b760048036038101906100b291906104b5565b6101ea565b6040516100c4919061057b565b60405180910390f35b6100e760048036038101906100e2919061070e565b61028e565b005b60006020528060005260406000206000915090508054610108906107be565b80601f0160208091040260200160405190810160405280929190818152602001828054610134906107be565b80156101815780601f1061015657610100808354040283529160200191610181565b820191906000526020600020905b81548152906001019060200180831161016457829003601f168201915b505050505081565b818160008086815260200190815260200160002091906101aa9291906103d2565b50827f5b60432ae9d9b7811073954a6ab7fc5ba55a88a4eb30e23c8ebf9ff1440e077d83836040516101dd92919061082c565b60405180910390a2505050565b60606000808381526020019081526020016000208054610209906107be565b80601f0160208091040260200160405190810160405280929190818152602001828054610235906107be565b80156102825780601f1061025757610100808354040283529160200191610282565b820191906000526020600020905b81548152906001019060200180831161026557829003601f168201915b50505050509050919050565b8181905084849050146102d6576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016102cd906108d3565b60405180910390fd5b60005b848490508110156103cb578282828181106102f7576102f66108f3565b5b90506020028101906103099190610931565b60008088888681811061031f5761031e6108f3565b5b90506020020135815260200190815260200160002091906103419291906103d2565b50848482818110610355576103546108f3565b5b905060200201357f5b60432ae9d9b7811073954a6ab7fc5ba55a88a4eb30e23c8ebf9ff1440e077d8484848181106103905761038f6108f3565b5b90506020028101906103a29190610931565b6040516103b092919061082c565b60405180910390a280806103c3906109cd565b9150506102d9565b5050505050565b8280546103de906107be565b90600052602060002090601f0160209004810192826104005760008555610447565b82601f1061041957803560ff1916838001178555610447565b82800160010185558215610447579182015b8281111561044657823582559160200191906001019061042b565b5b5090506104549190610458565b5090565b5b80821115610471576000816000905550600101610459565b5090565b600080fd5b600080fd5b6000819050919050565b6104928161047f565b811461049d57600080fd5b50565b6000813590506104af81610489565b92915050565b6000602082840312156104cb576104ca610475565b5b60006104d9848285016104a0565b91505092915050565b600081519050919050565b600082825260208201905092915050565b60005b8381101561051c578082015181840152602081019050610501565b8381111561052b576000848401525b50505050565b6000601f19601f8301169050919050565b600061054d826104e2565b61055781856104ed565b93506105678185602086016104fe565b61057081610531565b840191505092915050565b600060208201905081810360008301526105958184610542565b905092915050565b600080fd5b600080fd5b600080fd5b60008083601f8401126105c2576105c161059d565b5b8235905067ffffffffffffffff8111156105df576105de6105a2565b5b6020830191508360018202830111156105fb576105fa6105a7565b5b9250929050565b60008060006040848603121561061b5761061a610475565b5b6000610629868287016104a0565b935050602084013567ffffffffffffffff81111561064a5761064961047a565b5b610656868287016105ac565b92509250509250925092565b60008083601f8401126106785761067761059d565b5b8235905067ffffffffffffffff811115610695576106946105a2565b5b6020830191508360208202830111156106b1576106b06105a7565b5b9250929050565b60008083601f8401126106ce576106cd61059d565b5b8235905067ffffffffffffffff8111156106eb576106ea6105a2565b5b602083019150836020820283011115610707576107066105a7565b5b9250929050565b6000806000806040858703121561072857610727610475565b5b600085013567ffffffffffffffff8111156107465761074561047a565b5b61075287828801610662565b9450945050602085013567ffffffffffffffff8111156107755761077461047a565b5b610781878288016106b8565b925092505092959194509250565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b600060028204905060018216806107d657607f821691505b602082108114156107ea576107e961078f565b5b50919050565b82818337600083830152505050565b600061080b83856104ed565b93506108188385846107f0565b61082183610531565b840190509392505050565b600060208201905081810360008301526108478184866107ff565b90509392505050565b600082825260208201905092915050565b7f4b65797320616e642076616c756573206172726179206d757374206265206f6660008201527f2073616d65206c656e6774680000000000000000000000000000000000000000602082015250565b60006108bd602c83610850565b91506108c882610861565b604082019050919050565b600060208201905081810360008301526108ec816108b0565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b600080fd5b600080fd5b600080fd5b6000808335600160200384360303811261094e5761094d610922565b5b80840192508235915067ffffffffffffffff8211156109705761096f610927565b5b60208301925060018202360383131561098c5761098b61092c565b5b509250929050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000819050919050565b60006109d8826109c3565b91507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff821415610a0b57610a0a610994565b5b60018201905091905056fea26469706673582212207fce0ff25f5ddba6b4034a43c3c9169167a92b55ed90024be83b94a14fe5be7c64736f6c634300080b0033"
var StoreSMBin = "0x"

// DeployStore deploys a new contract, binding an instance of Store to it.
func DeployStore(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Receipt, *Store, error) {
	parsed, err := abi.JSON(strings.NewReader(StoreABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	var bytecode []byte
	if backend.SMCrypto() {
		bytecode = common.FromHex(StoreSMBin)
	} else {
		bytecode = common.FromHex(StoreBin)
	}
	if len(bytecode) == 0 {
		return common.Address{}, nil, nil, fmt.Errorf("cannot deploy empty bytecode")
	}
	address, receipt, contract, err := bind.DeployContract(auth, parsed, bytecode, StoreABI, backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, receipt, &Store{StoreCaller: StoreCaller{contract: contract}, StoreTransactor: StoreTransactor{contract: contract}, StoreFilterer: StoreFilterer{contract: contract}}, nil
}

func AsyncDeployStore(auth *bind.TransactOpts, handler func(*types.Receipt, error), backend bind.ContractBackend) (*types.Transaction, error) {
	parsed, err := abi.JSON(strings.NewReader(StoreABI))
	if err != nil {
		return nil, err
	}

	var bytecode []byte
	if backend.SMCrypto() {
		bytecode = common.FromHex(StoreSMBin)
	} else {
		bytecode = common.FromHex(StoreBin)
	}
	if len(bytecode) == 0 {
		return nil, fmt.Errorf("cannot deploy empty bytecode")
	}
	tx, err := bind.AsyncDeployContract(auth, handler, parsed, bytecode, StoreABI, backend)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// Store is an auto generated Go binding around a Solidity contract.
type Store struct {
	StoreCaller     // Read-only binding to the contract
	StoreTransactor // Write-only binding to the contract
	StoreFilterer   // Log filterer for contract events
}

// StoreCaller is an auto generated read-only Go binding around a Solidity contract.
type StoreCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StoreTransactor is an auto generated write-only Go binding around a Solidity contract.
type StoreTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StoreFilterer is an auto generated log filtering Go binding around a Solidity contract events.
type StoreFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StoreSession is an auto generated Go binding around a Solidity contract,
// with pre-set call and transact options.
type StoreSession struct {
	Contract     *Store            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// StoreCallerSession is an auto generated read-only Go binding around a Solidity contract,
// with pre-set call options.
type StoreCallerSession struct {
	Contract *StoreCaller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// StoreTransactorSession is an auto generated write-only Go binding around a Solidity contract,
// with pre-set transact options.
type StoreTransactorSession struct {
	Contract     *StoreTransactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// StoreRaw is an auto generated low-level Go binding around a Solidity contract.
type StoreRaw struct {
	Contract *Store // Generic contract binding to access the raw methods on
}

// StoreCallerRaw is an auto generated low-level read-only Go binding around a Solidity contract.
type StoreCallerRaw struct {
	Contract *StoreCaller // Generic read-only contract binding to access the raw methods on
}

// StoreTransactorRaw is an auto generated low-level write-only Go binding around a Solidity contract.
type StoreTransactorRaw struct {
	Contract *StoreTransactor // Generic write-only contract binding to access the raw methods on
}

// NewStore creates a new instance of Store, bound to a specific deployed contract.
func NewStore(address common.Address, backend bind.ContractBackend) (*Store, error) {
	contract, err := bindStore(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Store{StoreCaller: StoreCaller{contract: contract}, StoreTransactor: StoreTransactor{contract: contract}, StoreFilterer: StoreFilterer{contract: contract}}, nil
}

// NewStoreCaller creates a new read-only instance of Store, bound to a specific deployed contract.
func NewStoreCaller(address common.Address, caller bind.ContractCaller) (*StoreCaller, error) {
	contract, err := bindStore(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &StoreCaller{contract: contract}, nil
}

// NewStoreTransactor creates a new write-only instance of Store, bound to a specific deployed contract.
func NewStoreTransactor(address common.Address, transactor bind.ContractTransactor) (*StoreTransactor, error) {
	contract, err := bindStore(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &StoreTransactor{contract: contract}, nil
}

// NewStoreFilterer creates a new log filterer instance of Store, bound to a specific deployed contract.
func NewStoreFilterer(address common.Address, filterer bind.ContractFilterer) (*StoreFilterer, error) {
	contract, err := bindStore(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &StoreFilterer{contract: contract}, nil
}

// bindStore binds a generic wrapper to an already deployed contract.
func bindStore(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(StoreABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Store *StoreRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Store.Contract.StoreCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Store *StoreRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, *types.Receipt, error) {
	return _Store.Contract.StoreTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Store *StoreRaw) TransactWithResult(opts *bind.TransactOpts, result interface{}, method string, params ...interface{}) (*types.Transaction, *types.Receipt, error) {
	return _Store.Contract.StoreTransactor.contract.TransactWithResult(opts, result, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Store *StoreCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Store.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Store *StoreTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, *types.Receipt, error) {
	return _Store.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Store *StoreTransactorRaw) TransactWithResult(opts *bind.TransactOpts, result interface{}, method string, params ...interface{}) (*types.Transaction, *types.Receipt, error) {
	return _Store.Contract.contract.TransactWithResult(opts, result, method, params...)
}

// GetItem is a free data retrieval call binding the contract method 0xaa0372e7.
//
// Solidity: function getItem(bytes32 key) constant returns(bytes)
func (_Store *StoreCaller) GetItem(opts *bind.CallOpts, key [32]byte) ([]byte, error) {
	var (
		ret0 = new([]byte)
	)
	out := ret0
	err := _Store.contract.Call(opts, out, "getItem", key)
	return *ret0, err
}

// GetItem is a free data retrieval call binding the contract method 0xaa0372e7.
//
// Solidity: function getItem(bytes32 key) constant returns(bytes)
func (_Store *StoreSession) GetItem(key [32]byte) ([]byte, error) {
	return _Store.Contract.GetItem(&_Store.CallOpts, key)
}

// GetItem is a free data retrieval call binding the contract method 0xaa0372e7.
//
// Solidity: function getItem(bytes32 key) constant returns(bytes)
func (_Store *StoreCallerSession) GetItem(key [32]byte) ([]byte, error) {
	return _Store.Contract.GetItem(&_Store.CallOpts, key)
}

// Items is a free data retrieval call binding the contract method 0x48f343f3.
//
// Solidity: function items(bytes32 ) constant returns(bytes)
func (_Store *StoreCaller) Items(opts *bind.CallOpts, arg0 [32]byte) ([]byte, error) {
	var (
		ret0 = new([]byte)
	)
	out := ret0
	err := _Store.contract.Call(opts, out, "items", arg0)
	return *ret0, err
}

// Items is a free data retrieval call binding the contract method 0x48f343f3.
//
// Solidity: function items(bytes32 ) constant returns(bytes)
func (_Store *StoreSession) Items(arg0 [32]byte) ([]byte, error) {
	return _Store.Contract.Items(&_Store.CallOpts, arg0)
}

// Items is a free data retrieval call binding the contract method 0x48f343f3.
//
// Solidity: function items(bytes32 ) constant returns(bytes)
func (_Store *StoreCallerSession) Items(arg0 [32]byte) ([]byte, error) {
	return _Store.Contract.Items(&_Store.CallOpts, arg0)
}

// SetItem is a paid mutator transaction binding the contract method 0x7c50bba0.
//
// Solidity: function setItem(bytes32 key, bytes value) returns()
func (_Store *StoreTransactor) SetItem(opts *bind.TransactOpts, key [32]byte, value []byte) (*types.Transaction, *types.Receipt, error) {
	var ()
	out := &[]interface{}{}
	transaction, receipt, err := _Store.contract.TransactWithResult(opts, out, "setItem", key, value)
	return transaction, receipt, err
}

func (_Store *StoreTransactor) AsyncSetItem(handler func(*types.Receipt, error), opts *bind.TransactOpts, key [32]byte, value []byte) (*types.Transaction, error) {
	return _Store.contract.AsyncTransact(opts, handler, "setItem", key, value)
}

// SetItem is a paid mutator transaction binding the contract method 0x7c50bba0.
//
// Solidity: function setItem(bytes32 key, bytes value) returns()
func (_Store *StoreSession) SetItem(key [32]byte, value []byte) (*types.Transaction, *types.Receipt, error) {
	return _Store.Contract.SetItem(&_Store.TransactOpts, key, value)
}

func (_Store *StoreSession) AsyncSetItem(handler func(*types.Receipt, error), key [32]byte, value []byte) (*types.Transaction, error) {
	return _Store.Contract.AsyncSetItem(handler, &_Store.TransactOpts, key, value)
}

// SetItem is a paid mutator transaction binding the contract method 0x7c50bba0.
//
// Solidity: function setItem(bytes32 key, bytes value) returns()
func (_Store *StoreTransactorSession) SetItem(key [32]byte, value []byte) (*types.Transaction, *types.Receipt, error) {
	return _Store.Contract.SetItem(&_Store.TransactOpts, key, value)
}

func (_Store *StoreTransactorSession) AsyncSetItem(handler func(*types.Receipt, error), key [32]byte, value []byte) (*types.Transaction, error) {
	return _Store.Contract.AsyncSetItem(handler, &_Store.TransactOpts, key, value)
}

// SetItems is a paid mutator transaction binding the contract method 0xff78e5f5.
//
// Solidity: function setItems(bytes32[] keys, bytes[] values) returns()
func (_Store *StoreTransactor) SetItems(opts *bind.TransactOpts, keys [][32]byte, values [][]byte) (*types.Transaction, *types.Receipt, error) {
	var ()
	out := &[]interface{}{}
	transaction, receipt, err := _Store.contract.TransactWithResult(opts, out, "setItems", keys, values)
	return transaction, receipt, err
}

func (_Store *StoreTransactor) AsyncSetItems(handler func(*types.Receipt, error), opts *bind.TransactOpts, keys [][32]byte, values [][]byte) (*types.Transaction, error) {
	return _Store.contract.AsyncTransact(opts, handler, "setItems", keys, values)
}

// SetItems is a paid mutator transaction binding the contract method 0xff78e5f5.
//
// Solidity: function setItems(bytes32[] keys, bytes[] values) returns()
func (_Store *StoreSession) SetItems(keys [][32]byte, values [][]byte) (*types.Transaction, *types.Receipt, error) {
	return _Store.Contract.SetItems(&_Store.TransactOpts, keys, values)
}

func (_Store *StoreSession) AsyncSetItems(handler func(*types.Receipt, error), keys [][32]byte, values [][]byte) (*types.Transaction, error) {
	return _Store.Contract.AsyncSetItems(handler, &_Store.TransactOpts, keys, values)
}

// SetItems is a paid mutator transaction binding the contract method 0xff78e5f5.
//
// Solidity: function setItems(bytes32[] keys, bytes[] values) returns()
func (_Store *StoreTransactorSession) SetItems(keys [][32]byte, values [][]byte) (*types.Transaction, *types.Receipt, error) {
	return _Store.Contract.SetItems(&_Store.TransactOpts, keys, values)
}

func (_Store *StoreTransactorSession) AsyncSetItems(handler func(*types.Receipt, error), keys [][32]byte, values [][]byte) (*types.Transaction, error) {
	return _Store.Contract.AsyncSetItems(handler, &_Store.TransactOpts, keys, values)
}

// StoreItemSet represents a ItemSet event raised by the Store contract.
type StoreItemSet struct {
	Key   [32]byte
	Value []byte
	Raw   types.Log // Blockchain specific contextual infos
}

// WatchItemSet is a free log subscription operation binding the contract event 0x5b60432ae9d9b7811073954a6ab7fc5ba55a88a4eb30e23c8ebf9ff1440e077d.
//
// Solidity: event ItemSet(bytes32 indexed key, bytes value)
func (_Store *StoreFilterer) WatchItemSet(fromBlock *int64, handler func(int, []types.Log), key [32]byte) (string, error) {
	return _Store.contract.WatchLogs(fromBlock, handler, "ItemSet", key)
}

func (_Store *StoreFilterer) WatchAllItemSet(fromBlock *int64, handler func(int, []types.Log)) (string, error) {
	return _Store.contract.WatchLogs(fromBlock, handler, "ItemSet")
}

// ParseItemSet is a log parse operation binding the contract event 0x5b60432ae9d9b7811073954a6ab7fc5ba55a88a4eb30e23c8ebf9ff1440e077d.
//
// Solidity: event ItemSet(bytes32 indexed key, bytes value)
func (_Store *StoreFilterer) ParseItemSet(log types.Log) (*StoreItemSet, error) {
	event := new(StoreItemSet)
	if err := _Store.contract.UnpackLog(event, "ItemSet", log); err != nil {
		return nil, err
	}
	return event, nil
}

// WatchItemSet is a free log subscription operation binding the contract event 0x5b60432ae9d9b7811073954a6ab7fc5ba55a88a4eb30e23c8ebf9ff1440e077d.
//
// Solidity: event ItemSet(bytes32 indexed key, bytes value)
func (_Store *StoreSession) WatchItemSet(fromBlock *int64, handler func(int, []types.Log), key [32]byte) (string, error) {
	return _Store.Contract.WatchItemSet(fromBlock, handler, key)
}

func (_Store *StoreSession) WatchAllItemSet(fromBlock *int64, handler func(int, []types.Log)) (string, error) {
	return _Store.Contract.WatchAllItemSet(fromBlock, handler)
}

// ParseItemSet is a log parse operation binding the contract event 0x5b60432ae9d9b7811073954a6ab7fc5ba55a88a4eb30e23c8ebf9ff1440e077d.
//
// Solidity: event ItemSet(bytes32 indexed key, bytes value)
func (_Store *StoreSession) ParseItemSet(log types.Log) (*StoreItemSet, error) {
	return _Store.Contract.ParseItemSet(log)
}
