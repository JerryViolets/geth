// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package vm

import (
	"bytes"
	"encoding/binary"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	idData "github.com/ethereum/go-ethereum/idData"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"golang.org/x/crypto/sha3"
	"math/big"
	"strings"
)

var (
	bigZero                  = new(big.Int)
	//tt255                    = math.BigPow(2, 255)
)

func opAdd(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()

	// ssw
	if interpreter.evm.IsTxStart() {
		var x_big = x.ToBig()
		var y_big = y.ToBig()

		x_byte32 := common.BigToHash(x_big)
		y_byte32 := common.BigToHash(y_big)

		// 转为int方便后续匹配offset
		x_int := BytesToInt(x_byte32.Bytes())
		y_int := BytesToInt(y_byte32.Bytes())

		var re = new(big.Int).Set(y_big)
		re.Add(x_big, re)
		re_byte32 := common.BigToHash(re)

		if addr_xstr, ok := interpreter.evm.loc2addr[x_byte32]; ok && x_big.Cmp(bigZero) != 0 {    // 如果x对应的字节码有关联的以太坊地址
			if offset, ok := idData.Token2Offset[strings.ToLower(interpreter.evm.curcalladdr)]; ok {
				if y_int == offset {				// 如果其中一个操作数是 token offset
					interpreter.evm.loc2addr[re_byte32] = addr_xstr					// 则将(x+y)也与此地址关联。
					//fmt.Println("Add:", re_byte32, "->", addr_xstr)
				}
			}
			// 下面这段if是检测defi
			if addr_xstr, ok := interpreter.evm.sha32addr[x_byte32]; ok && x_big.Cmp(bigZero) != 0 {	// 如果x在sha32addr中也有关联
				if offset, ok := idData.DeFi2Offset[strings.ToLower(interpreter.evm.curcalladdr)]; ok {
					if y_int == offset {			// 如果其中一个操作数是 defi offset
						interpreter.evm.loc2addr[re_byte32] = addr_xstr
						interpreter.evm.sha32addr[re_byte32] = addr_xstr
						//fmt.Println("\t->", addr_xstr)
					}
				}
			}
		}
		if addr_ystr, ok := interpreter.evm.loc2addr[y_byte32]; ok && y_big.Cmp(bigZero) != 0 {
			if offset, ok := idData.Token2Offset[strings.ToLower(interpreter.evm.curcalladdr)]; ok {
				if x_int == offset {
					interpreter.evm.loc2addr[re_byte32] = addr_ystr
					//fmt.Println("Add:", re_byte32, "->", addr_ystr)
				}
			}
			if addr_ystr, ok := interpreter.evm.sha32addr[y_byte32]; ok && y_big.Cmp(bigZero) != 0 {
				if offset, ok := idData.DeFi2Offset[strings.ToLower(interpreter.evm.curcalladdr)]; ok {
					if y_int == offset {
						interpreter.evm.loc2addr[re_byte32] = addr_ystr
						interpreter.evm.sha32addr[re_byte32] = addr_ystr
						//fmt.Println("\t->", addr_ystr)
					}
				}
			}
		}

	}
	//

	y.Add(&x, y)
	return nil, nil
}

func opSub(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	y.Sub(&x, y)
	return nil, nil
}

func opMul(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	y.Mul(&x, y)
	return nil, nil
}

func opDiv(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	y.Div(&x, y)
	return nil, nil
}

func opSdiv(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	y.SDiv(&x, y)
	return nil, nil
}

func opMod(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	y.Mod(&x, y)
	return nil, nil
}

func opSmod(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	y.SMod(&x, y)
	return nil, nil
}

func opExp(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	base, exponent := callContext.stack.pop(), callContext.stack.peek()
	exponent.Exp(&base, exponent)
	return nil, nil
}

func opSignExtend(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	back, num := callContext.stack.pop(), callContext.stack.peek()
	num.ExtendSign(num, &back)
	return nil, nil
}

func opNot(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	x := callContext.stack.peek()
	x.Not(x)
	return nil, nil
}

func opLt(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	if x.Lt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return nil, nil
}

func opGt(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	if x.Gt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return nil, nil
}

func opSlt(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	if x.Slt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return nil, nil
}

func opSgt(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	if x.Sgt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return nil, nil
}

func opEq(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	if x.Eq(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return nil, nil
}

func opIszero(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	x := callContext.stack.peek()
	if x.IsZero() {
		x.SetOne()
	} else {
		x.Clear()
	}
	return nil, nil
}

func opAnd(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()

	// add by ssw
	x_big := x.ToBig()
	y_big := y.ToBig()

	if interpreter.evm.IsTxStart() {
		address_mask, _ := new(big.Int).SetString("ffffffffffffffffffffffffffffffffffffffff", 16)
		if x_big.Cmp(address_mask) == 0 {	// And操作其中一个操作数为"fffff"

			var temp_y = y_big
			temp_y.And(x_big, temp_y)
			addr_str := common.BytesToAddress(temp_y.Bytes()).String()		// 将操作数y转为以太坊地址
			loc_b32 := common.BigToHash(temp_y)								// 将y转为byte32——EVM字节码

			interpreter.evm.loc2addr[loc_b32] = addr_str					// 将字节码和以太坊地址对应起来
			// 有可能是无效的，因为有无关的And操作，所以这种对应可能不正确。但核心目的是将以太坊地址与字节码关联起来。
			//fmt.Println("And:", loc_b32, "->", addr_str)
		} else if y_big.Cmp(address_mask) == 0 {

			var temp_x = x_big
			temp_x.And(temp_x, y_big)
			addr_str := common.BytesToAddress(temp_x.Bytes()).String()
			loc_b32 := common.BigToHash(temp_x)

			interpreter.evm.loc2addr[loc_b32] = addr_str

			//fmt.Println("And:", loc_b32, "->", addr_str)
		}
	}

	y.And(&x, y)
	return nil, nil
}

func opOr(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	y.Or(&x, y)
	return nil, nil
}

func opXor(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	y.Xor(&x, y)
	return nil, nil
}

func opByte(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	th, val := callContext.stack.pop(), callContext.stack.peek()
	val.Byte(&th)
	return nil, nil
}

func opAddmod(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	x, y, z := callContext.stack.pop(), callContext.stack.pop(), callContext.stack.peek()
	if z.IsZero() {
		z.Clear()
	} else {
		z.AddMod(&x, &y, z)
	}
	return nil, nil
}

func opMulmod(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	x, y, z := callContext.stack.pop(), callContext.stack.pop(), callContext.stack.peek()
	z.MulMod(&x, &y, z)
	return nil, nil
}

// opSHL implements Shift Left
// The SHL instruction (shift left) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the left by arg1 number of bits.
func opSHL(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	// Note, second operand is left in the stack; accumulate result into it, and no need to push it afterwards
	shift, value := callContext.stack.pop(), callContext.stack.peek()
	if shift.LtUint64(256) {
		value.Lsh(value, uint(shift.Uint64()))
	} else {
		value.Clear()
	}
	return nil, nil
}

// opSHR implements Logical Shift Right
// The SHR instruction (logical shift right) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the right by arg1 number of bits with zero fill.
func opSHR(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	// Note, second operand is left in the stack; accumulate result into it, and no need to push it afterwards
	shift, value := callContext.stack.pop(), callContext.stack.peek()
	if shift.LtUint64(256) {
		value.Rsh(value, uint(shift.Uint64()))
	} else {
		value.Clear()
	}
	return nil, nil
}

// opSAR implements Arithmetic Shift Right
// The SAR instruction (arithmetic shift right) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the right by arg1 number of bits with sign extension.
func opSAR(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	shift, value := callContext.stack.pop(), callContext.stack.peek()
	if shift.GtUint64(256) {
		if value.Sign() >= 0 {
			value.Clear()
		} else {
			// Max negative shift: all bits set
			value.SetAllOne()
		}
		return nil, nil
	}
	n := uint(shift.Uint64())
	value.SRsh(value, n)
	return nil, nil
}

func opSha3(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	offset, size := callContext.stack.pop(), callContext.stack.peek()
	data := callContext.memory.GetPtr(int64(offset.Uint64()), int64(size.Uint64()))

	if interpreter.hasher == nil {
		interpreter.hasher = sha3.NewLegacyKeccak256().(keccakState)
	} else {
		interpreter.hasher.Reset()
	}
	interpreter.hasher.Write(data)
	interpreter.hasher.Read(interpreter.hasherBuf[:])

	evm := interpreter.evm
	if evm.vmConfig.EnablePreimageRecording {
		evm.StateDB.AddPreimage(interpreter.hasherBuf, data)
	}

	var hash_res = size.SetBytes(interpreter.hasherBuf[:])
	var hash_big = hash_res.ToBig()

	// add by ssw
	if interpreter.evm.IsTxStart() {
		loc_byte32 := common.BigToHash(hash_big)	// 把SHA3的结果转成byte32
		var data1 []byte
		var data2 []byte
		data1 = make([]byte,0)
		data2 = make([]byte,0)
		if len(data) == 64 {
			data1 = data[0 : 32]	// 取sha3的第一个操作数
			data2 = data[32 : 64]	// 取sha3的第二个操作数
			//fmt.Println("SHA3:", common.BytesToHash(data1), common.BytesToHash(data2))
			tempId_2 := BytesToInt(data2)	// 第二个操作数可能是id，先将其转化为int
			tempId_1 := BytesToInt(data1)	// 第一个操作数也可能是id，因为token中有这种pattern：sha3(ID, x)
			//fmt.Println("tempId:", tempId_2)
			//fmt.Println("curcalladdr", interpreter.evm.curcalladdr)

			// 以下是检测defi的代码
			if addr1, ok := interpreter.evm.loc2addr[common.BytesToHash(data1)]; ok { // 第一个操作数如果关联了addr
				// 把culcalladdr转为小写，再去DeFi2Id里面找
				if id, ok := idData.DeFi2Id[strings.ToLower(interpreter.evm.curcalladdr)]; ok {
					if id == tempId_2 { // 第二个操作数是想找的id
						interpreter.evm.loc2addr[loc_byte32] = addr1		// 则sha3的结果也关联此addr
						//fmt.Println("Sha3_1:", loc_byte32, "->", addr1)
					} else if addr2, ok := interpreter.evm.loc2addr[common.BytesToHash(data2)]; ok { // 如果第二操作数不是id，但关联了某addr
						interpreter.evm.loc2addr[loc_byte32] = addr1		// 则sha3的结果关联两个addr
						//fmt.Println("Sha3_1:", loc_byte32, "->", addr1)
						interpreter.evm.sha32addr[loc_byte32] = addr2		// 注意是在sha32addr里关联
						//fmt.Println("Sha3_2:", loc_byte32, "->", addr2)
					}
				}
			}
			// 以下是检测token的代码
			if addr, ok := interpreter.evm.loc2addr[common.BytesToHash(data1)]; ok {	// 第一个操作数如果关联了addr
				// 把culcalladdr转为小写，再去Token2Id里面找
				if id, ok := idData.Token2Id[strings.ToLower(interpreter.evm.curcalladdr)]; ok {
					if id == tempId_2 { // 第二个操作数是想找的id
						interpreter.evm.loc2addr[loc_byte32] = addr		// 则sha3的结果也关联此addr
						//fmt.Println("Sha3_1:", loc_byte32, "->", addr)
					}
				}
			} else if addr, ok := interpreter.evm.loc2addr[common.BytesToHash(data2)]; ok {	// 第二个操作数如果关联了addr。因为token中有这种pattern：sha3(ID, x)
				// 把culcalladdr转为小写，再去Token2Id里面找
				if id, ok := idData.Token2Id[strings.ToLower(interpreter.evm.curcalladdr)]; ok {
					if id == tempId_1 { // 第一个操作数是想找的id
						interpreter.evm.loc2addr[loc_byte32] = addr		// 则sha3的结果也关联此addr
						//fmt.Println("Sha3_1:", loc_byte32, "->", addr)
					}
				}
			}
		}
	}

	return nil, nil
}
func opAddress(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	callContext.stack.push(new(uint256.Int).SetBytes(callContext.contract.Address().Bytes()))
	return nil, nil
}

func opBalance(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	slot := callContext.stack.peek()
	address := common.Address(slot.Bytes20())
	slot.SetFromBig(interpreter.evm.StateDB.GetBalance(address))
	return nil, nil
}

func opOrigin(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	callContext.stack.push(new(uint256.Int).SetBytes(interpreter.evm.Origin.Bytes()))
	return nil, nil
}
func opCaller(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	callContext.stack.push(new(uint256.Int).SetBytes(callContext.contract.Caller().Bytes()))

	// add by ssw
	if interpreter.evm.IsTxStart() {
		temp_bytes := callContext.contract.Caller().Bytes()
		addr_str := callContext.contract.Caller().String()
		temp_bytes32 := common.BytesToHash(temp_bytes)

		interpreter.evm.loc2addr[temp_bytes32] = addr_str	// caller地址(32 byte) -> caller的以太坊地址(String)
		//fmt.Println("Caller:", temp_bytes32, "->", addr_str)
	}

	return nil, nil
}

func opCallValue(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	v, _ := uint256.FromBig(callContext.contract.value)
	callContext.stack.push(v)
	return nil, nil
}

func opCallDataLoad(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	x := callContext.stack.peek()
	if offset, overflow := x.Uint64WithOverflow(); !overflow {
		data := getData(callContext.contract.Input, offset, 32)
		x.SetBytes(data)
	} else {
		x.Clear()
	}
	return nil, nil
}

func opCallDataSize(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	callContext.stack.push(new(uint256.Int).SetUint64(uint64(len(callContext.contract.Input))))
	return nil, nil
}

func opCallDataCopy(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	var (
		memOffset  = callContext.stack.pop()
		dataOffset = callContext.stack.pop()
		length     = callContext.stack.pop()
	)
	dataOffset64, overflow := dataOffset.Uint64WithOverflow()
	if overflow {
		dataOffset64 = 0xffffffffffffffff
	}
	// These values are checked for overflow during gas cost calculation
	memOffset64 := memOffset.Uint64()
	length64 := length.Uint64()
	callContext.memory.Set(memOffset64, length64, getData(callContext.contract.Input, dataOffset64, length64))

	return nil, nil
}

func opReturnDataSize(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	callContext.stack.push(new(uint256.Int).SetUint64(uint64(len(interpreter.returnData))))
	return nil, nil
}

func opReturnDataCopy(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	var (
		memOffset  = callContext.stack.pop()
		dataOffset = callContext.stack.pop()
		length     = callContext.stack.pop()
	)

	offset64, overflow := dataOffset.Uint64WithOverflow()
	if overflow {
		return nil, ErrReturnDataOutOfBounds
	}
	// we can reuse dataOffset now (aliasing it for clarity)
	var end = dataOffset
	end.Add(&dataOffset, &length)
	end64, overflow := end.Uint64WithOverflow()
	if overflow || uint64(len(interpreter.returnData)) < end64 {
		return nil, ErrReturnDataOutOfBounds
	}
	callContext.memory.Set(memOffset.Uint64(), length.Uint64(), interpreter.returnData[offset64:end64])
	return nil, nil
}

func opExtCodeSize(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	slot := callContext.stack.peek()
	slot.SetUint64(uint64(interpreter.evm.StateDB.GetCodeSize(slot.Bytes20())))
	return nil, nil
}

func opCodeSize(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	l := new(uint256.Int)
	l.SetUint64(uint64(len(callContext.contract.Code)))
	callContext.stack.push(l)
	return nil, nil
}

func opCodeCopy(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	var (
		memOffset  = callContext.stack.pop()
		codeOffset = callContext.stack.pop()
		length     = callContext.stack.pop()
	)
	uint64CodeOffset, overflow := codeOffset.Uint64WithOverflow()
	if overflow {
		uint64CodeOffset = 0xffffffffffffffff
	}
	codeCopy := getData(callContext.contract.Code, uint64CodeOffset, length.Uint64())
	callContext.memory.Set(memOffset.Uint64(), length.Uint64(), codeCopy)

	return nil, nil
}

func opExtCodeCopy(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	var (
		stack      = callContext.stack
		a          = stack.pop()
		memOffset  = stack.pop()
		codeOffset = stack.pop()
		length     = stack.pop()
	)
	uint64CodeOffset, overflow := codeOffset.Uint64WithOverflow()
	if overflow {
		uint64CodeOffset = 0xffffffffffffffff
	}
	addr := common.Address(a.Bytes20())
	codeCopy := getData(interpreter.evm.StateDB.GetCode(addr), uint64CodeOffset, length.Uint64())
	callContext.memory.Set(memOffset.Uint64(), length.Uint64(), codeCopy)

	return nil, nil
}

// opExtCodeHash returns the code hash of a specified account.
// There are several cases when the function is called, while we can relay everything
// to `state.GetCodeHash` function to ensure the correctness.
//   (1) Caller tries to get the code hash of a normal contract account, state
// should return the relative code hash and set it as the result.
//
//   (2) Caller tries to get the code hash of a non-existent account, state should
// return common.Hash{} and zero will be set as the result.
//
//   (3) Caller tries to get the code hash for an account without contract code,
// state should return emptyCodeHash(0xc5d246...) as the result.
//
//   (4) Caller tries to get the code hash of a precompiled account, the result
// should be zero or emptyCodeHash.
//
// It is worth noting that in order to avoid unnecessary create and clean,
// all precompile accounts on mainnet have been transferred 1 wei, so the return
// here should be emptyCodeHash.
// If the precompile account is not transferred any amount on a private or
// customized chain, the return value will be zero.
//
//   (5) Caller tries to get the code hash for an account which is marked as suicided
// in the current transaction, the code hash of this account should be returned.
//
//   (6) Caller tries to get the code hash for an account which is marked as deleted,
// this account should be regarded as a non-existent account and zero should be returned.
func opExtCodeHash(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	slot := callContext.stack.peek()
	address := common.Address(slot.Bytes20())
	if interpreter.evm.StateDB.Empty(address) {
		slot.Clear()
	} else {
		slot.SetBytes(interpreter.evm.StateDB.GetCodeHash(address).Bytes())
	}
	return nil, nil
}

func opGasprice(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	v, _ := uint256.FromBig(interpreter.evm.GasPrice)
	callContext.stack.push(v)
	return nil, nil
}

func opBlockhash(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	num := callContext.stack.peek()
	num64, overflow := num.Uint64WithOverflow()
	if overflow {
		num.Clear()
		return nil, nil
	}
	var upper, lower uint64
	upper = interpreter.evm.Context.BlockNumber.Uint64()
	if upper < 257 {
		lower = 0
	} else {
		lower = upper - 256
	}
	if num64 >= lower && num64 < upper {
		num.SetBytes(interpreter.evm.Context.GetHash(num64).Bytes())
	} else {
		num.Clear()
	}
	return nil, nil
}

func opCoinbase(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	callContext.stack.push(new(uint256.Int).SetBytes(interpreter.evm.Context.Coinbase.Bytes()))
	return nil, nil
}

func opTimestamp(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	v, _ := uint256.FromBig(interpreter.evm.Context.Time)
	callContext.stack.push(v)
	return nil, nil
}

func opNumber(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	v, _ := uint256.FromBig(interpreter.evm.Context.BlockNumber)
	callContext.stack.push(v)
	return nil, nil
}

func opDifficulty(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	v, _ := uint256.FromBig(interpreter.evm.Context.Difficulty)
	callContext.stack.push(v)
	return nil, nil
}

func opGasLimit(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	callContext.stack.push(new(uint256.Int).SetUint64(interpreter.evm.Context.GasLimit))
	return nil, nil
}

func opPop(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	callContext.stack.pop()
	return nil, nil
}

func opMload(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	v := callContext.stack.peek()
	offset := int64(v.Uint64())
	v.SetBytes(callContext.memory.GetPtr(offset, 32))
	return nil, nil
}

func opMstore(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	// pop value of the stack
	mStart, val := callContext.stack.pop(), callContext.stack.pop()
	callContext.memory.Set32(mStart.Uint64(), &val)
	return nil, nil
}

func opMstore8(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	off, val := callContext.stack.pop(), callContext.stack.pop()
	callContext.memory.store[off.Uint64()] = byte(val.Uint64())
	return nil, nil
}

func opSload(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	loc := callContext.stack.peek()
	hash := common.Hash(loc.Bytes32())
	val := interpreter.evm.StateDB.GetState(callContext.contract.Address(), hash)
	loc.SetBytes(val.Bytes())
	return nil, nil
}

func opSstore(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	loc := callContext.stack.pop()
	val := callContext.stack.pop()

	//Add by jerry
	loc_big := loc.ToBig()
	val_big := val.ToBig()

	// add by ssw
	if interpreter.evm.IsTxStart() {
		// 通过loc取得sstore之前的初始value
		hash := common.Hash(loc.Bytes32())
		iniVal := interpreter.evm.StateDB.GetState(callContext.contract.Address(), hash)
		var tempIniVal = new(big.Int).Set(iniVal.Big()) // 初始value

		// ！！！！下面的两个if，如果存在一个合约，既是token又是defi，并且带有offset，会出现问题，但现在没有这种情况，暂时这样写。！！！！
		// 如果是当前合约是token的第二种pattern（带有offset），就把loc_big变为(loc_big - offset)，目的是只输出指定offset的变量
		if offset, ok := idData.Token2Offset[strings.ToLower(interpreter.evm.curcalladdr)]; ok {
			offset_big := big.NewInt(int64(offset))
			loc_big.Sub(loc_big, offset_big)
		}
		// 如果当前合约是defi，并且有offset，就把就把loc_big变为(loc_big - offset)，目的是只输出指定offset的变量
		if offset, ok := idData.DeFi2Offset[strings.ToLower(interpreter.evm.curcalladdr)]; ok {
			offset_big := big.NewInt(int64(offset))
			loc_big.Sub(loc_big, offset_big)
		}
		//

		temploc := common.BigToHash(loc_big)
		var tempval = val_big

		if addr1, ok := interpreter.evm.loc2addr[temploc]; ok {
			if addr2, ok := interpreter.evm.sha32addr[temploc]; ok {   // Sstore的参数location，在loc2addr和sha32addr中都有对应的以太坊地址
				if _, ok := idData.DeFi2Id[strings.ToLower(interpreter.evm.curcalladdr)]; ok {     // 只有curcalladdr在DeFi列表里时才输出
					// 先将输出暂存到 outMsg
					interpreter.evm.SetHasDefi(true)
					interpreter.evm.outMsg += "curcalladdr: " + interpreter.evm.curcalladdr + " Sstore: address1 = " + addr1 + " address2 = " + addr2 + " value = " + tempval.String() + " iniVal = " + tempIniVal.String() + "\n"
					//fmt.Println("curcalladdr:", interpreter.evm.curcalladdr, "Sstore: address1 =", addr1, "address2 =", addr2, "value =", tempval, "iniVal =", tempIniVal)
				}
			} else {													// Sstore的参数location，只在loc2addr有对应的以太坊地址
				if _, ok := idData.Token2Id[strings.ToLower(interpreter.evm.curcalladdr)]; ok {    // 只有curcalladdr在token列表里时才输出
					// 先将输出暂存到 outMsg
					interpreter.evm.SetHasToken(true)
					interpreter.evm.outMsg += "curcalladdr: " + interpreter.evm.curcalladdr + " Sstore: address = " + addr1 + " value = " + tempval.String() + " iniVal = " + tempIniVal.String() + "\n"
					//fmt.Println("curcalladdr:", interpreter.evm.curcalladdr, "Sstore: address =", addr1, "value =", tempval, "iniVal =", tempIniVal)
				}
			}
		}
	}
	//

	interpreter.evm.StateDB.SetState(callContext.contract.Address(),
		loc.Bytes32(), val.Bytes32())
	return nil, nil
}

func opJump(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	pos := callContext.stack.pop()
	if !callContext.contract.validJumpdest(&pos) {
		return nil, ErrInvalidJump
	}
	*pc = pos.Uint64()
	return nil, nil
}

func opJumpi(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	pos, cond := callContext.stack.pop(), callContext.stack.pop()
	if !cond.IsZero() {
		if !callContext.contract.validJumpdest(&pos) {
			return nil, ErrInvalidJump
		}
		*pc = pos.Uint64()
	} else {
		*pc++
	}
	return nil, nil
}

func opJumpdest(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	return nil, nil
}

func opBeginSub(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	return nil, ErrInvalidSubroutineEntry
}

func opJumpSub(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	if len(callContext.rstack.data) >= 1023 {
		return nil, ErrReturnStackExceeded
	}
	pos := callContext.stack.pop()
	if !pos.IsUint64() {
		return nil, ErrInvalidJump
	}
	posU64 := pos.Uint64()
	if !callContext.contract.validJumpSubdest(posU64) {
		return nil, ErrInvalidJump
	}
	callContext.rstack.push(uint32(*pc))
	*pc = posU64 + 1
	return nil, nil
}

func opReturnSub(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	if len(callContext.rstack.data) == 0 {
		return nil, ErrInvalidRetsub
	}
	// Other than the check that the return stack is not empty, there is no
	// need to validate the pc from 'returns', since we only ever push valid
	//values onto it via jumpsub.
	*pc = uint64(callContext.rstack.pop()) + 1
	return nil, nil
}

func opPc(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	callContext.stack.push(new(uint256.Int).SetUint64(*pc))
	return nil, nil
}

func opMsize(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	callContext.stack.push(new(uint256.Int).SetUint64(uint64(callContext.memory.Len())))
	return nil, nil
}

func opGas(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	callContext.stack.push(new(uint256.Int).SetUint64(callContext.contract.Gas))
	return nil, nil
}

func opCreate(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	var (
		value        = callContext.stack.pop()
		offset, size = callContext.stack.pop(), callContext.stack.pop()
		input        = callContext.memory.GetCopy(int64(offset.Uint64()), int64(size.Uint64()))
		gas          = callContext.contract.Gas
	)
	if interpreter.evm.chainRules.IsEIP150 {
		gas -= gas / 64
	}
	// reuse size int for stackvalue
	stackvalue := size

	callContext.contract.UseGas(gas)
	//TODO: use uint256.Int instead of converting with toBig()
	var bigVal = big0
	if !value.IsZero() {
		bigVal = value.ToBig()
	}

	res, addr, returnGas, suberr := interpreter.evm.Create(callContext.contract, input, gas, bigVal)
	// Push item on the stack based on the returned error. If the ruleset is
	// homestead we must check for CodeStoreOutOfGasError (homestead only
	// rule) and treat as an error, if the ruleset is frontier we must
	// ignore this error and pretend the operation was successful.
	if interpreter.evm.chainRules.IsHomestead && suberr == ErrCodeStoreOutOfGas {
		stackvalue.Clear()
	} else if suberr != nil && suberr != ErrCodeStoreOutOfGas {
		stackvalue.Clear()
	} else {
		stackvalue.SetBytes(addr.Bytes())
	}
	callContext.stack.push(&stackvalue)
	callContext.contract.Gas += returnGas

	if suberr == ErrExecutionReverted {
		return res, nil
	}
	return nil, nil
}

func opCreate2(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	var (
		endowment    = callContext.stack.pop()
		offset, size = callContext.stack.pop(), callContext.stack.pop()
		salt         = callContext.stack.pop()
		input        = callContext.memory.GetCopy(int64(offset.Uint64()), int64(size.Uint64()))
		gas          = callContext.contract.Gas
	)

	// Apply EIP150
	gas -= gas / 64
	callContext.contract.UseGas(gas)
	// reuse size int for stackvalue
	stackvalue := size
	//TODO: use uint256.Int instead of converting with toBig()
	bigEndowment := big0
	if !endowment.IsZero() {
		bigEndowment = endowment.ToBig()
	}
	res, addr, returnGas, suberr := interpreter.evm.Create2(callContext.contract, input, gas,
		bigEndowment, &salt)
	// Push item on the stack based on the returned error.
	if suberr != nil {
		stackvalue.Clear()
	} else {
		stackvalue.SetBytes(addr.Bytes())
	}
	callContext.stack.push(&stackvalue)
	callContext.contract.Gas += returnGas

	if suberr == ErrExecutionReverted {
		return res, nil
	}
	return nil, nil
}

func opCall(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	stack := callContext.stack
	// Pop gas. The actual gas in interpreter.evm.callGasTemp.
	// We can use this as a temporary value
	temp := stack.pop()
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.Address(addr.Bytes20())
	// Get the arguments from the memory.
	args := callContext.memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	var bigVal = big0
	//TODO: use uint256.Int instead of converting with toBig()
	// By using big0 here, we save an alloc for the most common case (non-ether-transferring contract calls),
	// but it would make more sense to extend the usage of uint256.Int
	if !value.IsZero() {
		gas += params.CallStipend
		bigVal = value.ToBig()
	}

	ret, returnGas, err := interpreter.evm.Call(callContext.contract, toAddr, args, gas, bigVal)

	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.push(&temp)
	if err == nil || err == ErrExecutionReverted {
		callContext.memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	callContext.contract.Gas += returnGas

	return ret, nil
}

func opCallCode(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	// Pop gas. The actual gas is in interpreter.evm.callGasTemp.
	stack := callContext.stack
	// We use it as a temporary value
	temp := stack.pop()
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.Address(addr.Bytes20())
	// Get arguments from the memory.
	args := callContext.memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	//TODO: use uint256.Int instead of converting with toBig()
	var bigVal = big0
	if !value.IsZero() {
		gas += params.CallStipend
		bigVal = value.ToBig()
	}

	ret, returnGas, err := interpreter.evm.CallCode(callContext.contract, toAddr, args, gas, bigVal)
	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.push(&temp)
	if err == nil || err == ErrExecutionReverted {
		callContext.memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	callContext.contract.Gas += returnGas

	return ret, nil
}

func opDelegateCall(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	stack := callContext.stack
	// Pop gas. The actual gas is in interpreter.evm.callGasTemp.
	// We use it as a temporary value
	temp := stack.pop()
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.Address(addr.Bytes20())
	// Get arguments from the memory.
	args := callContext.memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	ret, returnGas, err := interpreter.evm.DelegateCall(callContext.contract, toAddr, args, gas)
	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.push(&temp)
	if err == nil || err == ErrExecutionReverted {
		callContext.memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	callContext.contract.Gas += returnGas

	return ret, nil
}

func opStaticCall(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	// Pop gas. The actual gas is in interpreter.evm.callGasTemp.
	stack := callContext.stack
	// We use it as a temporary value
	temp := stack.pop()
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.Address(addr.Bytes20())
	// Get arguments from the memory.
	args := callContext.memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	ret, returnGas, err := interpreter.evm.StaticCall(callContext.contract, toAddr, args, gas)
	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.push(&temp)
	if err == nil || err == ErrExecutionReverted {
		callContext.memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	callContext.contract.Gas += returnGas

	return ret, nil
}

func opReturn(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	offset, size := callContext.stack.pop(), callContext.stack.pop()
	ret := callContext.memory.GetPtr(int64(offset.Uint64()), int64(size.Uint64()))

	return ret, nil
}

func opRevert(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	offset, size := callContext.stack.pop(), callContext.stack.pop()
	ret := callContext.memory.GetPtr(int64(offset.Uint64()), int64(size.Uint64()))

	return ret, nil
}

func opStop(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	return nil, nil
}

func opSuicide(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	beneficiary := callContext.stack.pop()
	balance := interpreter.evm.StateDB.GetBalance(callContext.contract.Address())
	interpreter.evm.StateDB.AddBalance(beneficiary.Bytes20(), balance)
	interpreter.evm.StateDB.Suicide(callContext.contract.Address())
	return nil, nil
}

// following functions are used by the instruction jump  table

// make log instruction function
func makeLog(size int) executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
		topics := make([]common.Hash, size)
		stack := callContext.stack
		mStart, mSize := stack.pop(), stack.pop()
		for i := 0; i < size; i++ {
			addr := stack.pop()
			topics[i] = addr.Bytes32()
		}

		d := callContext.memory.GetCopy(int64(mStart.Uint64()), int64(mSize.Uint64()))
		interpreter.evm.StateDB.AddLog(&types.Log{
			Address: callContext.contract.Address(),
			Topics:  topics,
			Data:    d,
			// This is a non-consensus field, but assigned here because
			// core/state doesn't know the current block number.
			BlockNumber: interpreter.evm.Context.BlockNumber.Uint64(),
		})

		return nil, nil
	}
}

// opPush1 is a specialized version of pushN
func opPush1(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
	var (
		codeLen = uint64(len(callContext.contract.Code))
		integer = new(uint256.Int)
	)
	*pc += 1
	if *pc < codeLen {
		callContext.stack.push(integer.SetUint64(uint64(callContext.contract.Code[*pc])))
	} else {
		callContext.stack.push(integer.Clear())
	}
	return nil, nil
}

// make push instruction function
func makePush(size uint64, pushByteSize int) executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
		codeLen := len(callContext.contract.Code)

		startMin := codeLen
		if int(*pc+1) < startMin {
			startMin = int(*pc + 1)
		}

		endMin := codeLen
		if startMin+pushByteSize < endMin {
			endMin = startMin + pushByteSize
		}

		integer := new(uint256.Int)
		callContext.stack.push(integer.SetBytes(common.RightPadBytes(
			callContext.contract.Code[startMin:endMin], pushByteSize)))

		*pc += size
		return nil, nil
	}
}

// make dup instruction function
func makeDup(size int64) executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
		callContext.stack.dup(int(size))
		return nil, nil
	}
}

// make swap instruction function
func makeSwap(size int64) executionFunc {
	// switch n + 1 otherwise n would be swapped with n
	size++
	return func(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error) {
		callContext.stack.swap(int(size))
		return nil, nil
	}
}

// add by ssw
func BytesToInt(bys []byte) int {
	bys_temp := bys[24:]
	bytebuff := bytes.NewBuffer(bys_temp)
	var data int64
	binary.Read(bytebuff, binary.BigEndian, &data)
	return int(data)
}