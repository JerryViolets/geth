package vm

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"os"
	"log"
	"github.com/ethereum/go-ethereum/common/math"
	//"github.com/ethereum/go-ethereum/common/math/big"math
)

// add by ssw
// 记录每次sload和sstore
type StorageInfo struct {
	TokenAddr	string		// 合约地址
	user		string		// 用户地址
	value		*big.Int	// 本次Sload或Sstore的值
}

//karry
type DefiStorageInfo struct {
	address1	string		// 合约地址
	address2		string		// 用户地址
	value		*big.Int	// 本次Sload或Sstore的值
}
var DefiMapMapSloadInfos []DefiStorageInfo
var DefiMapMapSstoreInfos []DefiStorageInfo
//

var TransSloadInfos []StorageInfo	// sload的集合
var TransSstoreInfos []StorageInfo	// sstore的集合
var number = 0

//karry
func appendTransSloadInfo(_TokenAddr string, _user string, _value *big.Int) {
	var storageInfo = StorageInfo{
		TokenAddr: 	_TokenAddr,
		user: 		_user,
		value:		_value,
	}
	TransSloadInfos = append(TransSloadInfos, storageInfo)
}
func appendTransSstoreInfo(_TokenAddr string, _user string, _value *big.Int) {
	var storageInfo = StorageInfo{
		TokenAddr: 	_TokenAddr,
		user: 		_user,
		value:		_value,
	}
	TransSstoreInfos = append(TransSstoreInfos, storageInfo)
}

func appendDefiMapMapSloadInfo(_address1 string, _address2 string, _value *big.Int) {
	var defiStorageInfo = DefiStorageInfo{
		address1: 		_address1,
		address2: 		_address2,
		value:			_value,
	}
	DefiMapMapSloadInfos = append(DefiMapMapSloadInfos, defiStorageInfo)
}

func appendDefiMapMapSstoreInfo(_address1 string, _address2 string, _value *big.Int) {
	var defiStorageInfo = DefiStorageInfo{
		address1: 		_address1,
		address2: 		_address2,
		value:			_value,
	}
	DefiMapMapSstoreInfos = append(DefiMapMapSstoreInfos, defiStorageInfo)
}


// add by ssw
// 输出结构体数组
func printTransSloadInfos() {
	fmt.Println("*****************************Sload*****************************")
	for _, sloadInfo := range TransSloadInfos {
		fmt.Println("TokenAddr:", sloadInfo.TokenAddr, "user:", sloadInfo.user, "value:", sloadInfo.value)
	}
	fmt.Println("***************************************************************")
}
func printTransSstoreInfos() {
	fmt.Println("*****************************Sstore*****************************")
	for _, sstoreInfo := range TransSstoreInfos {
		fmt.Println("TokenAddr:", sstoreInfo.TokenAddr, "user:", sstoreInfo.user, "value:", sstoreInfo.value)
	}
	fmt.Println("****************************************************************")
}
//

type CallInfo struct {

	defiLayer		int		//第几层defi层面的call
	transferLayer	int		//在某一层defi层面的transfer调用

	//karry
	Caller		string		//调用者地址
	CallTo		string		//被调用合约地址
	//

	AccFrom 	string		//转出账户地址
	AccTo 		string		//转入账户地址
	Amount 		map[string]*big.Int	//交易数额

	//karry
	TokenAddr 	[]string		//交易token地址
	//token种类
	TokenNum     int

	FuncName 	string		//函数名
}


var callInfos []CallInfo          //transfer函数信息数组

//var EventInfos []TranInfo


//by Jerry 
//对三个数组进行初始化的功能，在defi的0层进行调用
func initInfos(){
	//karry
	DefiMapMapSloadInfos = make ([]DefiStorageInfo,0)
	DefiMapMapSstoreInfos = make ([]DefiStorageInfo,0)
	//

	TransSloadInfos = make ([]StorageInfo,0)
	TransSstoreInfos = make ([]StorageInfo,0)
	callInfos = make([]CallInfo,0)
}

type ValueInfo struct{
	beforevalue *big.Int
	aftervalue *big.Int
	before_change bool
}


func detectDiff(evm *EVM){
	// File, err := os.OpenFile("/home/jerry/work/defiresult.txt", os.O_WRONLY|os.O_CREATE, 0666)
	// if err != nil {
	// 	log.Println(err)
	// }
	var valueInfos map[string]*ValueInfo
	var user string 
	var temp1 *ValueInfo
	var temp2 *ValueInfo
	//var tempbig = new(big.Int)
	Accfrom := callInfos[0].AccFrom
	Accto := callInfos[0].AccTo
	// valueInfos  := map[string]*ValueInfo{
	// 	Accfrom : &ValueInfo{tempbig,tempbig,false}
	// 	Accto : &ValueInfo{tempbig,tempbig,false}
	// }

	valueInfos = make(map[string]*ValueInfo)
	temp1 = new(ValueInfo)
	temp2 = new(ValueInfo)
	//tempInfo = &ValueInfo{tempbig,tempbig,false}
	temp1.before_change = false
	temp2.before_change = false
	valueInfos[Accfrom] = temp1
	valueInfos[Accto] = temp2

	number += 1
	fmt.Println("Start detect ",number)
	for  _, tokenaddr := range callInfos[0].TokenAddr {
		//beforevalue = make ([string]int)
		//aftervalue = make ([string]int)
		//var tmpvalue *big.Int
		//fmt.Println("tokenaddr:",tokenaddr)

		//karry
		fmt.Println("detecting token :",tokenaddr)
		var defiSload *big.Int =new(big.Int)
		var defiSstore *big.Int =new(big.Int)
		var defi_value *big.Int =new(big.Int)
		for _,v := range DefiMapMapSloadInfos{
			if v.address1 == tokenaddr || v.address2 == tokenaddr{
				defiSload = v.value
				break
			}
		}
		for i:=len(DefiMapMapSstoreInfos)-1;i>=0;i--{
			if DefiMapMapSstoreInfos[i].address1 == tokenaddr || DefiMapMapSstoreInfos[i].address2 == tokenaddr{
				defiSstore = DefiMapMapSstoreInfos[i].value
				break
			}
		}
		if defiSload.Cmp(defiSstore) == 1{
			defi_value.Sub(defiSload,defiSstore)
		}else{
			defi_value.Sub(defiSstore,defiSload)
		}
		//

		for _,sloadInfo := range TransSloadInfos {
			if sloadInfo.TokenAddr == tokenaddr {
				user = sloadInfo.user				
				if valueInfos[user].before_change == false {
					temp1 = valueInfos[user]
					temp1.beforevalue = sloadInfo.value
					temp1.before_change = true
					valueInfos[user] = temp1
				  // first time we meet the user's value ,we store it 
				}

			}
		}
		for _, sstoreInfo := range TransSstoreInfos {
			if sstoreInfo.TokenAddr == tokenaddr {
				user = sstoreInfo.user
				temp2 = valueInfos[user]
				temp2.aftervalue = sstoreInfo.value
				valueInfos[user] = temp2
				// the last time we meet the user's value ,we store it
			}
		}
		// if Accfrom == user {
		// 	//var trans_value *big.Int
		// 	math.U256(aftervalue.Sub(beforevalue,aftervalue))
		// }else if  Accto == user{
		// 	//var trans_value *big.Int
		// 	math.U256(aftervalue.Sub(aftervalue, beforevalue))
		// }else{
		// 	fmt.Println("The wrong user!")
		// }
		//fmt.Println("after1",valueInfos[Accfrom].aftervalue)
		//fmt.Println("before1",valueInfos[Accfrom].beforevalue)
		//fmt.Println("after2",valueInfos[Accto].aftervalue)
		//fmt.Println("before2",valueInfos[Accto].beforevalue)
		// from_before := *valueInfos[Accfrom].beforevalue
		// from_after := *valueInfos[Accfrom].aftervalue
		//from_value := math.U256(from_after.Sub(from_before,from_after))

		from_value := math.U256(valueInfos[Accfrom].aftervalue.Sub(valueInfos[Accfrom].beforevalue,valueInfos[Accfrom].aftervalue))
		to_value := math.U256(valueInfos[Accto].aftervalue.Sub(valueInfos[Accto].aftervalue,valueInfos[Accto].beforevalue))



		// from_before := *valueInfos[Accfrom].beforevalue
		// from_after := *valueInfos[Accfrom].aftervalue
		// from_value := math.U256(from_after.Sub(from_before,from_after))

		// to_before := *valueInfos[Accto].beforevalue
		// to_after := *valueInfos[Accto].aftervalue
		//from_value = valueInfos[Accfrom].aftervalue
		//to_value = valueInfos[Accto].aftervalue
		//fmt.Printf("from %T\n",from_value)
		//fmt.Printf("to %T\n",to_value)
		//fmt.Println()
		if from_value.Cmp(to_value) != 0{
			fmt.Println("Wrong transfer")
			return
		}
		if defi_value.Cmp(from_value) == 0 {
			evm.diff = true
			fmt.Println("this call is consistent",)
			fmt.Println("defi change :",defi_value)
			fmt.Println("transfer change :",from_value)
			//////////
			//fmt.Println("hash:",evm.txhash)
			//fmt.Fprintln(File,tokenaddr)
		}else{
			evm.diff = false
			fmt.Println("this call is NOT consistent")
			fmt.Println("defi change :",defi_value)
			fmt.Println("transfer change :",from_value)
			//fmt.Println("hash:",evm.txhash)
			/////////
			//fmt.Fprintln(File,tokenaddr)
		}

	}
	//fmt.Println("End detect")
}


func getResult(){
	File, err := os.OpenFile("/home/wbs/go/result.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println(err)
	}
	//_, err :=	fmt.Fprintln(testRetFile, "do you love me, my dear")
	fmt.Fprintln(File,"############################Trace###############################")
	fmt.Fprintln(File,"*****************************Sload*****************************")
	for _, sloadInfo := range TransSloadInfos {
	 	fmt.Println("TokenAddr:", sloadInfo.TokenAddr, "user:", sloadInfo.user, "value:", sloadInfo.value)
	 	fmt.Fprintln(File,sloadInfo.TokenAddr)
	 	fmt.Fprintln(File,sloadInfo.user)
	 	fmt.Fprintln(File,sloadInfo.value)
	}
	fmt.Fprintln(File,"***************************************************************")


	fmt.Fprintln(File,"*****************************Sstore*****************************")

	for _, sstoreInfo := range TransSstoreInfos {
		fmt.Println("TokenAddr:", sstoreInfo.TokenAddr, "user:", sstoreInfo.user, "value:", sstoreInfo.value)
	 	fmt.Fprintln(File,sstoreInfo.TokenAddr)
	 	fmt.Fprintln(File,sstoreInfo.user)
	 	fmt.Fprintln(File,sstoreInfo.value)
	}
	fmt.Fprintln(File,"***************************************************************")

	fmt.Fprintln(File,"*****************************Defi*****************************")
		//for _, callInfo := range callInfos {
	fmt.Println("TokenAddr:", callInfos[0].TokenAddr[0], "from:", callInfos[0].AccFrom,"to:" ,callInfos[0].AccTo, "value:", callInfos[0].Amount[callInfos[0].TokenAddr[0]])
	fmt.Fprintln(File,callInfos[0].TokenAddr[0])
	fmt.Fprintln(File,callInfos[0].AccFrom)
	fmt.Fprintln(File,callInfos[0].AccTo)
	fmt.Fprintln(File,callInfos[0].Amount[callInfos[0].TokenAddr[0]])

	fmt.Fprintln(File,"***************************************************************")

	//fmt.Println("End writing ")
	// if err != nil {
	// 	log.Println(err)
	// }
}


func GetFunInfo(input []byte, caller string, conaddr string, evm *EVM) {

	if conaddr == "0x0000000000000000000000000000000000000004" {
		return
	} // solve the precompile contract bug

	inputStr := hex.EncodeToString(input)


	//by jerry
	//检测defi项目函数deposit



	//deposit(address,address,uint256)
	if len(inputStr) ==200  && "8340f549" == inputStr[:8] {
		//fmt.Println("Find depositToken(address,uint256) ")
		amount, ok := new(big.Int).SetString(inputStr[136:200], 16)
		if ok {

			//karry
			addIsDefiAndDefiLayer(evm)
			//
			//fmt.Println("The function  in the defi" ,evm.defiLayer)

			var callinfo = CallInfo{

				defiLayer: 			evm.defiLayer,
				transferLayer: 		evm.transferLayer,

				Caller:			 	caller,//拿到函数调用者地址
				CallTo: 			conaddr,

				//karry
				AccFrom: 			common.HexToAddress(inputStr[8:72]).String(),
				AccTo: 				conaddr,
				//
				TokenNum:            1,
				//TokenAddr:			make([]string, 0)
				//common.HexToAddress(inputStr[72:136]).String(),	//拿到token地址
				Amount: make(map[string]*big.Int),    //input中的第二个参数：amount

				FuncName: "depositToken(address,uint256)",
			}
			callinfo.TokenAddr = append(callinfo.TokenAddr,common.HexToAddress(inputStr[72:136]).String())
			var temp *big.Int
			temp = new(big.Int)
			*temp = *amount;
			callinfo.Amount[callinfo.TokenAddr[0]]=temp
			//临时修改
			fmt.Println("\n","curcalladdr:", conaddr, "Sstore: address1 =", callinfo.TokenAddr[0], "address2 =", callinfo.AccFrom, "value =", amount, "iniVal =", 0)
			callInfos = append(callInfos, callinfo)
		}
	}


	//depositToken(address,uint256) || deposit(address,uint256)
	if len(inputStr) ==136  && ("338b5dea" == inputStr[:8] || "47e7ef24" == inputStr[:8] ){
		//fmt.Println("Find depositToken(address,uint256) ")
		amount, ok := new(big.Int).SetString(inputStr[72:136], 16)
		if ok {

			//karry
			addIsDefiAndDefiLayer(evm)
			//
			//fmt.Println("The function  in the defi" ,evm.defiLayer)

			var callinfo = CallInfo{

				defiLayer: 			evm.defiLayer,
				transferLayer: 		evm.transferLayer,

				Caller:			 	caller,//拿到函数调用者地址
				CallTo: 			conaddr,

				//karry
				AccFrom: 			caller,
				AccTo: 				conaddr,
				//
				TokenNum:            1,
				//TokenAddr:			common.HexToAddress(inputStr[8:72]).String(),	//拿到token地址
				Amount:  make(map[string]*big.Int),    //input中的第二个参数：amount

				FuncName: "depositToken(address,uint256)",
			}
			callinfo.TokenAddr = append(callinfo.TokenAddr,common.HexToAddress(inputStr[8:72]).String())
			var temp *big.Int
			temp = new(big.Int)
			*temp = *amount;
			callinfo.Amount[callinfo.TokenAddr[0]]=temp
			//callinfo.Amount[callinfo.TokenAddr[0]]=amount
			//if(conaddr == "0x4aEa7cf559F67ceDCAD07E12aE6bc00F07E8cf65" ||conaddr =="0x8d12A197cB00D4747a1fe03395095ce2A5CC6819" ){
			fmt.Println("\n","curcalladdr:", conaddr, "Sstore: address1 =", callinfo.TokenAddr[0], "address2 =", callinfo.AccFrom, "value =", amount, "iniVal =", 0)
			//}
			callInfos = append(callInfos, callinfo)
		}
	}

	//depositToken(    address _user, address _assetId, uint256 _amount, uint256 _expectedAmount, uint256 _nonce)
	if len(inputStr) ==328 && "a42d5083" == inputStr[:8] {
		//fmt.Println("Find depositToken(address,address,uint256,uint256,uint256) ")
		amount, ok := new(big.Int).SetString(inputStr[136:200], 16)
		if ok {

			//karry
			addIsDefiAndDefiLayer(evm)
			//
			//fmt.Println("The function  in the defi" ,evm.defiLayer)

			var callinfo = CallInfo{

				defiLayer: 			evm.defiLayer,
				transferLayer: 		evm.transferLayer,

				Caller:			 	caller,//拿到函数调用者地址
				CallTo: 			conaddr,

				//karry
				AccFrom: 			common.HexToAddress(inputStr[8:72]).String(), //_user代表存入者地址
				AccTo: 				conaddr,//存入合约
				//
				TokenNum:            1,
				//TokenAddr:			common.HexToAddress(inputStr[72:136]).String(),	//拿到token地址
				Amount: make(map[string]*big.Int),    //input中的第二个参数：amount

				FuncName: "depositToken(address,address,uint256,uint256,uint256)",
			}
			callinfo.TokenAddr = append(callinfo.TokenAddr,common.HexToAddress(inputStr[72:136]).String())
			var temp *big.Int
			temp = new(big.Int)
			*temp = *amount;
			callinfo.Amount[callinfo.TokenAddr[0]]=temp
			//callinfo.Amount[callinfo.TokenAddr[0]]=amount
			fmt.Println("\n","curcalladdr:", conaddr, "Sstore: address1 =", callinfo.TokenAddr[0], "address2 =", callinfo.AccFrom, "value =", amount, "iniVal =", 0)
			callInfos = append(callInfos, callinfo)
		}
	}



	//withdrawToken(address,uint256)
	if len(inputStr) ==136  && "9e281a98" == inputStr[:8] {
		//fmt.Println("Find withdrawToken(address,uint256) ")
		amount, ok := new(big.Int).SetString(inputStr[72:136], 16)
		if ok {

			//karry
			addIsDefiAndDefiLayer(evm)
			//
			//fmt.Println("The function  in the defi" ,evm.defiLayer)
			var callinfo = CallInfo{

				defiLayer: 			evm.defiLayer,
				transferLayer: 		evm.transferLayer,

				Caller:			caller,//拿到函数调用者地址
				CallTo: 		conaddr,

				//karry
				AccFrom: 		conaddr,
				AccTo: 			caller,
				//
				TokenNum:            1,
				//TokenAddr:		common.HexToAddress(inputStr[8:72]).String(),	//拿到token地址

				Amount: make(map[string]*big.Int),    //input中的第二个参数：amount

				FuncName: "withdrawToken(address,uint256)",
			}
			callinfo.TokenAddr = append(callinfo.TokenAddr,common.HexToAddress(inputStr[8:72]).String())
			var temp *big.Int
			temp = new(big.Int)
			*temp = *amount;
			callinfo.Amount[callinfo.TokenAddr[0]]=temp
			//callinfo.Amount[callinfo.TokenAddr[0]]=amount
			//if(conaddr == "0x4aEa7cf559F67ceDCAD07E12aE6bc00F07E8cf65" ||conaddr =="0x8d12A197cB00D4747a1fe03395095ce2A5CC6819" ){
			//evm.tmp_flag = true
			fmt.Println("\n","curcalladdr:", conaddr, "Sstore: address1 =", callinfo.TokenAddr[0], "address2 =", callinfo.AccFrom, "value =", amount, "iniVal =", 0)
			//}
			callInfos = append(callInfos, callinfo)
		}
	}

	//withdraw(address,uint256)
	if len(inputStr) ==136  && "f3fef3a3" == inputStr[:8] {
		//fmt.Println("Find withdraw(address,uint256) ")
		amount, ok := new(big.Int).SetString(inputStr[72:136], 16)//karry 以“ 0x”或“ 0X”为前缀选择base16
		if ok {

			//karry
			addIsDefiAndDefiLayer(evm)
			//
			//fmt.Println("The function  in the defi" ,evm.defiLayer)
			var callinfo = CallInfo{

				defiLayer: 			evm.defiLayer,
				transferLayer: 		evm.transferLayer,

				Caller:			caller,//拿到函数调用者地址
				CallTo: 		conaddr,

				//karry
				AccFrom: 		conaddr,
				AccTo:			caller,
				//
				TokenNum:            1,
				//TokenAddr:		common.HexToAddress(inputStr[8:72]).String(),	//拿到token地址

				Amount: make(map[string]*big.Int),    //input中的第二个参数：amount

				FuncName: "withdraw(address,uint256)",
			}
			callinfo.TokenAddr = append(callinfo.TokenAddr,common.HexToAddress(inputStr[8:72]).String())
			var temp *big.Int
			temp = new(big.Int)
			*temp = *amount;
			callinfo.Amount[callinfo.TokenAddr[0]]=temp
			//callinfo.Amount[callinfo.TokenAddr[0]]=amount
			//if(conaddr == "0x4aEa7cf559F67ceDCAD07E12aE6bc00F07E8cf65" ||conaddr =="0x8d12A197cB00D4747a1fe03395095ce2A5CC6819" ){
			//evm.tmp_flag = true
			fmt.Println("\n","curcalladdr:", conaddr, "Sstore: address1 =", callinfo.TokenAddr[0], "address2 =", callinfo.AccFrom, "value =", amount, "iniVal =", 0)
			//}
			callInfos = append(callInfos, callinfo)
		}
	}

	//_withdraw()
	//withdraw(address _withdrawer,  address payable _receivingAddress, address _assetId,uint256 _amount,
				//address _feeAssetId,uint256 _feeAmount,uint256 _nonce)
	if len(inputStr) ==456  && "95206540" == inputStr[:8] {
		//fmt.Println("Find _withdraw() ")
		amount, ok := new(big.Int).SetString(inputStr[200:264], 16)//karry 以“ 0x”或“ 0X”为前缀选择base16
		if ok {

			//karry
			addIsDefiAndDefiLayer(evm)
			//
			//fmt.Println("The function  in the defi" ,evm.defiLayer)
			var callinfo = CallInfo{

				defiLayer: 			evm.defiLayer,
				transferLayer: 		evm.transferLayer,

				Caller:			caller,//拿到函数调用者地址
				CallTo: 		conaddr,

				//karry
				AccFrom: 		conaddr,
				AccTo:			common.HexToAddress(inputStr[72:136]).String(),			//来自用户的取款，_receivingAddress才是最终转账地址
				//
				TokenNum:            1,
				//TokenAddr:		common.HexToAddress(inputStr[136:200]).String(),	//拿到token地址

				Amount: make(map[string]*big.Int),    //input中的第二个参数：amount

				FuncName: "_withdraw(）",
			}
			callinfo.TokenAddr = append(callinfo.TokenAddr,common.HexToAddress(inputStr[136:200]).String())
			
			var temp *big.Int
			temp = new(big.Int)
			*temp = *amount;
			callinfo.Amount[callinfo.TokenAddr[0]]=temp
			//callinfo.Amount[callinfo.TokenAddr[0]]=amount
			fmt.Println("\n","curcalladdr:", conaddr, "Sstore: address1 =", callinfo.TokenAddr[0], "address2 =", callinfo.AccFrom, "value =", amount, "iniVal =", 0)
			callInfos = append(callInfos, callinfo)
		}
	}




	//supply(address,uint256)
	if len(inputStr) ==136  && "f2b9fdb8" == inputStr[:8] {
		//fmt.Println("Find supply(address,uint256) ")
		amount, ok := new(big.Int).SetString(inputStr[72:136], 16)//karry 以“ 0x”或“ 0X”为前缀选择base16
		if ok {

			//karry
			addIsDefiAndDefiLayer(evm)
			//
			//fmt.Println("The function  in the defi" ,evm.defiLayer)
			var callinfo = CallInfo{

				defiLayer: 			evm.defiLayer,
				transferLayer: 		evm.transferLayer,

				Caller:			caller,//拿到函数调用者地址
				CallTo: 		conaddr,

				//karry
				AccFrom:		caller,
				AccTo: 			conaddr,
				//
				TokenNum:            1,
				//TokenAddr:		common.HexToAddress(inputStr[8:72]).String(),	//拿到token地址

				Amount: make(map[string]*big.Int),    //input中的第二个参数：amount

				FuncName: "supply(address,uint256)",
			}
			callinfo.TokenAddr = append(callinfo.TokenAddr,common.HexToAddress(inputStr[8:72]).String())
			var temp *big.Int
			temp = new(big.Int)
			*temp = *amount;
			callinfo.Amount[callinfo.TokenAddr[0]]=temp
			//callinfo.Amount[callinfo.TokenAddr[0]]=amount
			fmt.Println("\n","curcalladdr:", conaddr, "Sstore: address1 =", callinfo.TokenAddr[0], "address2 =", callinfo.AccFrom, "value =", amount, "iniVal =", 0)
			callInfos = append(callInfos, callinfo)
		}
	}

	//directwithdrawal(address,uint256)
	if len(inputStr) ==136  && "7330c026" == inputStr[:8] {
		//fmt.Println("Find directwithdrawal(address,uint256) ")
		amount, ok := new(big.Int).SetString(inputStr[72:136], 16)//karry 以“ 0x”或“ 0X”为前缀选择base16
		if ok {

			//karry
			addIsDefiAndDefiLayer(evm)
			//
			//fmt.Println("The function  in the defi" ,evm.defiLayer)
			var withdraw2 = CallInfo{

				defiLayer: 			evm.defiLayer,
				transferLayer: 		evm.transferLayer,

				Caller:			caller,//拿到函数调用者地址
				CallTo: 		conaddr,

				//karry
				AccFrom:		conaddr,
				AccTo: 			caller,
				//
				TokenNum:            1,
				//TokenAddr:		common.HexToAddress(inputStr[8:72]).String(),	//拿到token地址

				Amount: make(map[string]*big.Int),    //input中的第二个参数：amount

				FuncName: "directwithdrawal(address,uint256)",
			}
			withdraw2.TokenAddr = append(withdraw2.TokenAddr,common.HexToAddress(inputStr[8:72]).String())
			
			var temp *big.Int
			temp = new(big.Int)
			*temp = *amount;
			withdraw2.Amount[withdraw2.TokenAddr[0]]=temp
			//withdraw2.Amount[withdraw2.TokenAddr[0]]=amount
			fmt.Println("\n","curcalladdr:", conaddr, "Sstore: address1 =", withdraw2.TokenAddr[0], "address2 =", withdraw2.AccFrom, "value =", amount, "iniVal =", 0)
			callInfos = append(callInfos, withdraw2)
		}
	}




	//*************************************transfer****************************************//
	if (len(inputStr) == 136 || len(inputStr) == 192) && "a9059cbb" == inputStr[:8] {
		//fmt.Println("  Find transfer ")
		amount, ok := new(big.Int).SetString(inputStr[72:136], 16)

		// only if we get the defi function  then continue
		if ok && evm.IsDefi() {

			//karry
			addIsTransferAndTransferLayer(evm)
			//
			fmt.Println("The fucntion in the transfer" ,evm.transferLayer,conaddr)

			var transfer = CallInfo{

				defiLayer: 			evm.defiLayer,
				transferLayer: 		evm.transferLayer,

				Caller: caller,
				CallTo: conaddr,

				AccFrom:   caller,
				AccTo:     common.HexToAddress(inputStr[8:72]).String(),

				TokenNum:        1,
				Amount: make(map[string]*big.Int),

				//TokenAddr: conaddr,

				FuncName: "transfer",
			}
			transfer.TokenAddr =append(transfer.TokenAddr,conaddr)
			var temp *big.Int
			temp = new(big.Int)
			*temp = *amount;
			transfer.Amount[transfer.TokenAddr[0]]=temp
			//transfer.Amount[transfer.TokenAddr[0]]=amount
			callInfos = append(callInfos, transfer)


		}
	}


	if len(inputStr) == 200 && "23b872dd" == inputStr[:8] {
		//fmt.Println("Find transfer")
		amount, ok := new(big.Int).SetString(inputStr[136:200], 16)
		if ok && evm.IsDefi() {

			//karry
			addIsTransferAndTransferLayer(evm)
			//
			fmt.Println("The fucntion in the transfer" ,evm.transferLayer,conaddr)

			//fmt.Println("The transfer in the defi")
			var transfer = CallInfo{

				defiLayer: 			evm.defiLayer,
				transferLayer: 		evm.transferLayer,

				Caller: caller,
				CallTo: conaddr,

				AccFrom:   common.HexToAddress(inputStr[8:72]).String(),
				AccTo:     common.HexToAddress(inputStr[72:136]).String(),
				TokenNum:        1,
				Amount: make(map[string]*big.Int),

				//TokenAddr: conaddr,

				FuncName: "transferFrom",
			}
			transfer.TokenAddr = append(transfer.TokenAddr,conaddr)
				var temp *big.Int
			temp = new(big.Int)
			*temp = *amount;
			transfer.Amount[transfer.TokenAddr[0]]=temp
			//transfer.Amount[transfer.TokenAddr[0]]=amount
			callInfos = append(callInfos, transfer)
		}
	}
}
func addIsDefiAndDefiLayer(evm *EVM){
	if evm.isDefi == false {//karry 这里必定是第一个函数调用
		evm.SetDefi(true)
		evm.SetHasDefi(true)
		evm.SetDefiLayer(0)
		initInfos() //对三个数组进行初始化
	} else {//karry 如果不是第一次调用就让层数+1
		evm.defiLayer += 1
	}
}

func addIsTransferAndTransferLayer(evm *EVM){
	if evm.isTransfer == false {//karry 这里必定是第一个转移函数调用
		evm.SetTransfer(true)
		evm.SetTransferLayer(0)
	} else {//karry 如果不是第一次调用就让层数+1
		evm.transferLayer += 1
	}
}


func redIsDefiAndDefiLayer(input []byte, evm *EVM){
	inputStr := hex.EncodeToString(input)
	if  len([]rune(inputStr)) >= 8{
		nameHash := inputStr[:8]
		if nameHash == "338b5dea" || //depositToken(address,uint256)
			nameHash == "9e281a98"||//withdrawToken(address,uint256)
			nameHash == "f3fef3a3"||//withdraw(address,uint256)
			nameHash == "f2b9fdb8"||//supply(address,uint256)
			nameHash == "7330c026"||//directwithdrawal(address,uint256)
			nameHash == "47e7ef24"||
			nameHash == "95206540"||{
			

			evm.defiLayer -= 1
			//fmt.Println("Normal exit Sub Defi layer",evm.defiLayer)
			if evm.defiLayer == -1 {
				evm.isDefi = false
				//getResult()
				//detectDiff(evm)
			}	

		}
	}
	

}

func redIsTransferAndTransferLayer(input []byte, evm *EVM){
	inputStr := hex.EncodeToString(input)
	if  len([]rune(inputStr)) >= 8{
	nameHash := inputStr[:8]
	if nameHash == "a9059cbb" || nameHash =="23b872dd" {

		evm.transferLayer -= 1
		//fmt.Println("Normal exit Sub trans layer",evm.transferLayer)
		if evm.transferLayer == -1 {
			evm.isTransfer = false
		}

	}
}

}




//
//func GetEventInfo() {
//	// the code is in core/vm/instructions.go : func makeLog(size int) executionFunc {}
//}
//
//func IsCapturedEvent() bool {
//	if len(EventInfos) == 0 {
//		return false
//	}
//	return true
//}
//
//func IsCapturedFun() bool {
//	if len(callInfos) == 0 {
//		return false
//	}
//	return true
//}
//
//func ClearTranInfos() {
//	EventInfos = EventInfos[:0]
//	callInfos = callInfos[:0]
//}

//import (
//"fmt"
//"math/big"
//"encoding/hex"
//"github.com/ethereum/go-ethereum/common"
//)
//
//type TranInfo struct {
//AccFrom 	string		//调用者地址
//AccTo 		string		//被调用者地址
//Amount 		*big.Int	//交易数额
//ConAddr 	string		//合约地址？
//
//FuncName 	string		//函数名
//}
//
//type DepositFun1 struct {
//callerAddr     string  //调用者地址
//tokenAddr		string //存入的代币合约地址
//Amount		*big.Int	//存入的金额（暂用big.Int）
//ConAddr 	string		//合约地址？
//
//FuncName	string	//函数名
//}
//
//type WithdrawFun1 struct {
//callerAddr     string  //调用者地址
//tokenAddr		string //取出的代币合约地址
//Amount		*big.Int	//取出的金额（暂用big.Int）
//ConAddr 	string		//合约地址？
//
//FuncName	string	//函数名
//}
//
//var callInfos []TranInfo		//transfer函数信息数组
//var DepositInfo1 []DepositFun1	//depositToken(address token, uint256 amount)函数类型数组
//var WithdrawInfo1 []WithdrawFun1	//depositToken(address token, uint256 amount)函数类型数组
//
////var EventInfos []TranInfo
//
//func GetFunInfo(input []byte, caller string, conaddr string, evm *EVM ) {
//
//if conaddr == "0x0000000000000000000000000000000000000004" {
//return
//} // solve the precompile contract bug
//
//inputStr := hex.EncodeToString(input)
//
////defi_bool = false
////trans_bool = false
////by jerry
////检测defi项目函数deposit
//
////depositToken()  func
//if len(inputStr) ==136  && "338b5dea" == inputStr[:8] {
//fmt.Println("call Deposit input:", inputStr)
//amount, ok := new(big.Int).SetString(inputStr[72:136], 16)
//if ok {
////set isDefi  true
//evm.SetDefi(true)
////first time layer = 0
//evm.SetDefiLayer(0)
////
//
//var deposit1 =DepositFun1{
//callerAddr:			 caller,//拿到函数调用者地址
//tokenAddr:			common.HexToAddress(inputStr[8:72]).String(),	//拿到token地址
//Amount: amount,    //input中的第二个参数：amount
//ConAddr: conaddr,
//FuncName: "deposit",
//}
//DepositInfo1 = append(DepositInfo1, deposit1)
//}
//}
//
//if len(inputStr) ==136  && "9e281a98" == inputStr[:8] {
//fmt.Println("call Withdraw input:", inputStr)
//amount, ok := new(big.Int).SetString(inputStr[72:136], 16)
//if ok {
////set isDefi  true
//evm.SetDefi(true)
////first time layer = 0
//evm.SetDefiLayer(0)
////
//
//var withdraw1 =WithdrawFun1{
//callerAddr:			 caller,//拿到函数调用者地址
//tokenAddr:			common.HexToAddress(inputStr[8:72]).String(),	//拿到token地址
//Amount: amount,    //input中的第二个参数：amount
//ConAddr: conaddr,
//FuncName: "deposit",
//}
//WithdrawInfo1 = append(WithdrawInfo1, withdraw1)
//}
//}
//
//
//
//
//
//if (len(inputStr) == 136 || len(inputStr) == 192) && "a9059cbb" == inputStr[:8] {
//fmt.Println("  call transfer input:", inputStr)
//amount, ok := new(big.Int).SetString(inputStr[72:136], 16)
//
//// only if we get the defi function  then continue
//if ok && evm.IsDefi() {
//
////set
//evm.SetTransfer(true)
//evm.SetTransLayer(0)
//fmt.Println("The transfer in the defi")
//
//var transfer = TranInfo{
//AccFrom:   caller,
//AccTo:     common.HexToAddress(inputStr[8:72]).String(),
//Amount: amount,
//ConAddr: conaddr,
//FuncName: "transfer",
//}
//callInfos = append(callInfos, transfer)
//}
//}
//
//
//if len(inputStr) == 200 && "23b872dd" == inputStr[:8] {
//fmt.Println("call transfer input:", inputStr)
//amount, ok := new(big.Int).SetString(inputStr[136:200], 16)
//if ok && evm.IsDefi() {
//
////set
//evm.SetTransfer(true)
//evm.SetTransLayer(0)
//fmt.Println("The transfer in the defi")
//var transfer = TranInfo{
//AccFrom:   common.HexToAddress(inputStr[8:72]).String(),
//AccTo:     common.HexToAddress(inputStr[72:136]).String(),
//Amount: amount,
//ConAddr: conaddr,
//FuncName: "transferFrom",
//}
//callInfos = append(callInfos, transfer)
//}
//}
//
//
//
//
//}
////
////func GetEventInfo() {
////	// the code is in core/vm/instructions.go : func makeLog(size int) executionFunc {}
////}
////
////func IsCapturedEvent() bool {
////	if len(EventInfos) == 0 {
////		return false
////	}
////	return true
////}
////
////func IsCapturedFun() bool {
////	if len(callInfos) == 0 {
////		return false
////	}
////	return true
////}
////
////func ClearTranInfos() {
////	EventInfos = EventInfos[:0]
////	callInfos = callInfos[:0]
////}
