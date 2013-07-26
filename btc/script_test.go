package btc

import (
	"bytes"
	"errors"
	"strconv"
	"testing"
	"strings"
	"math/big"
	"io/ioutil"
	"encoding/hex"
	"encoding/json"
	"encoding/binary"
)

var (
	cnt int
	valid bool
)

func RawToStack(sig []byte) ([]byte) {
	if len(sig)==1 {
		if sig[0]==0x81 {
			return []byte{OP_1NEGATE}
		}
		if sig[0]==0x80 || sig[0]==0x00 {
			return []byte{OP_0}
		}
		if sig[0]<=16 {
			return []byte{OP_1-1+sig[0]}
		}
	}
	bb := new(bytes.Buffer)
	if len(sig) < OP_PUSHDATA1 {
		bb.Write([]byte{byte(len(sig))})
	} else if len(sig) <= 0xff {
		bb.Write([]byte{OP_PUSHDATA1})
		bb.Write([]byte{byte(len(sig))})
	} else if len(sig) <= 0xffff {
		bb.Write([]byte{OP_PUSHDATA2})
		binary.Write(bb, binary.LittleEndian, uint16(len(sig)))
	} else {
		bb.Write([]byte{OP_PUSHDATA4})
		binary.Write(bb, binary.LittleEndian, uint32(len(sig)))
	}
	bb.Write(sig)
	return bb.Bytes()
}


func int2scr(v int64) ([]byte) {
	if v==-1 || v>=1 && v<=16 {
		return []byte{byte(v + OP_1 - 1)}
	}

	neg := v<0
	if neg {
		v = -v
	}
	bn := big.NewInt(v)
	bts := bn.Bytes()
	if (bts[0]&0x80)!=0 {
		if neg {
			bts = append([]byte{0x80}, bts...)
		} else {
			bts = append([]byte{0x00}, bts...)
		}
	} else if neg {
		bts[0] |= 0x80
	}

	sig := make([]byte, len(bts))
	for i := range bts {
		sig[len(bts)-i-1] = bts[i]
	}

	return RawToStack(sig)
}


func pk2hex(pk string) (out []byte, e error) {
	xx := strings.Split(pk, " ")
	for i := range xx {
		v, er := strconv.ParseInt(xx[i], 10, 64)
		if er==nil {
			switch {
				case v==-1: out = append(out, 0x4f)
				case v==0: out = append(out, 0x0)
				case v>0 && v<=16: out = append(out, 0x50+byte(v))
				default:
					out = append(out, int2scr(v)...)
			}
		} else if len(xx[i])>2 && xx[i][:2]=="0x" {
			d, _ := hex.DecodeString(xx[i][2:])
			out = append(out, d...)
		} else {
			if len(xx[i])>=2 && xx[i][0]=='\'' && xx[i][len(xx[i])-1]=='\'' {
				out = append(out, RawToStack([]byte(xx[i][1:len(xx[i])-1]))...)
			} else {
				if len(xx[i])>3 && xx[i][:3]=="OP_" {
					xx[i] = xx[i][3:]
				}
				switch(xx[i]) {
					case "NOP": out = append(out, 0x61)
					case "VER": out = append(out, 0x62)
					case "IF": out = append(out, 0x63)
					case "NOTIF": out = append(out, 0x64)
					case "VERIF": out = append(out, 0x65)
					case "VERNOTIF": out = append(out, 0x66)
					case "ELSE": out = append(out, 0x67)
					case "ENDIF": out = append(out, 0x68)
					case "VERIFY": out = append(out, 0x69)
					case "RETURN": out = append(out, 0x6a)
					case "TOALTSTACK": out = append(out, 0x6b)
					case "FROMALTSTACK": out = append(out, 0x6c)
					case "2DROP": out = append(out, 0x6d)
					case "2DUP": out = append(out, 0x6e)
					case "3DUP": out = append(out, 0x6f)
					case "2OVER": out = append(out, 0x70)
					case "2ROT": out = append(out, 0x71)
					case "2SWAP": out = append(out, 0x72)
					case "IFDUP": out = append(out, 0x73)
					case "DEPTH": out = append(out, 0x74)
					case "DROP": out = append(out, 0x75)
					case "DUP": out = append(out, 0x76)
					case "NIP": out = append(out, 0x77)
					case "OVER": out = append(out, 0x78)
					case "PICK": out = append(out, 0x79)
					case "ROLL": out = append(out, 0x7a)
					case "ROT": out = append(out, 0x7b)
					case "SWAP": out = append(out, 0x7c)
					case "TUCK": out = append(out, 0x7d)
					case "CAT": out = append(out, 0x7e)
					case "SUBSTR": out = append(out, 0x7f)
					case "LEFT": out = append(out, 0x80)
					case "RIGHT": out = append(out, 0x81)
					case "SIZE": out = append(out, 0x82)
					case "INVERT": out = append(out, 0x83)
					case "AND": out = append(out, 0x84)
					case "OR": out = append(out, 0x85)
					case "XOR": out = append(out, 0x86)
					case "EQUAL": out = append(out, 0x87)
					case "EQUALVERIFY": out = append(out, 0x88)
					case "RESERVED1": out = append(out, 0x89)
					case "RESERVED2": out = append(out, 0x8a)
					case "1ADD": out = append(out, 0x8b)
					case "1SUB": out = append(out, 0x8c)
					case "2MUL": out = append(out, 0x8d)
					case "2DIV": out = append(out, 0x8e)
					case "NEGATE": out = append(out, 0x8f)
					case "ABS": out = append(out, 0x90)
					case "NOT": out = append(out, 0x91)
					case "0NOTEQUAL": out = append(out, 0x92)
					case "ADD": out = append(out, 0x93)
					case "SUB": out = append(out, 0x94)
					case "MUL": out = append(out, 0x95)
					case "DIV": out = append(out, 0x96)
					case "MOD": out = append(out, 0x97)
					case "LSHIFT": out = append(out, 0x98)
					case "RSHIFT": out = append(out, 0x99)
					case "BOOLAND": out = append(out, 0x9a)
					case "BOOLOR": out = append(out, 0x9b)
					case "NUMEQUAL": out = append(out, 0x9c)
					case "NUMEQUALVERIFY": out = append(out, 0x9d)
					case "NUMNOTEQUAL": out = append(out, 0x9e)
					case "LESSTHAN": out = append(out, 0x9f)
					case "GREATERTHAN": out = append(out, 0xa0)
					case "LESSTHANOREQUAL": out = append(out, 0xa1)
					case "GREATERTHANOREQUAL": out = append(out, 0xa2)
					case "MIN": out = append(out, 0xa3)
					case "MAX": out = append(out, 0xa4)
					case "WITHIN": out = append(out, 0xa5)
					case "RIPEMD160": out = append(out, 0xa6)
					case "SHA1": out = append(out, 0xa7)
					case "SHA256": out = append(out, 0xa8)
					case "HASH160": out = append(out, 0xa9)
					case "HASH256": out = append(out, 0xaa)
					case "CODESEPARATOR": out = append(out, 0xab)
					case "CHECKSIG": out = append(out, 0xac)
					case "CHECKSIGVERIFY": out = append(out, 0xad)
					case "CHECKMULTISIG": out = append(out, 0xae)
					case "CHECKMULTISIGVERIFY": out = append(out, 0xaf)
					case "NOP1": out = append(out, 0xb0)
					case "NOP2": out = append(out, 0xb1)
					case "NOP3": out = append(out, 0xb2)
					case "NOP4": out = append(out, 0xb3)
					case "NOP5": out = append(out, 0xb4)
					case "NOP6": out = append(out, 0xb5)
					case "NOP7": out = append(out, 0xb6)
					case "NOP8": out = append(out, 0xb7)
					case "NOP9": out = append(out, 0xb8)
					case "NOP10": out = append(out, 0xb9)
					case "": out = append(out, []byte{}...)
					default:
						return nil, errors.New("Syntax error: "+xx[i])
				}
			}
		}
	}
	return
}


func testit(txhex string, i int, pk_script string) bool {
	cnt++

	rd, er := hex.DecodeString(txhex)
	if er != nil {
		println("Cannot decode tx raw data from vector", cnt)
		return false
	}
	pk, er := pk2hex(pk_script)
	if er!=nil {
		return false
	}

	tx, _ := NewTx(rd)
	if tx==nil {
		println("Cannot decode tx at vector", cnt)
		return false
	}
	tx.Size = uint32(len(rd))
	ha := Sha2Sum(rd)
	tx.Hash = NewUint256(ha[:])
	//println(tx.Hash.String(), len(tx.TxIn), len(tx.TxOut), i)

	var ss []byte
	if i>=0 {
		ss = tx.TxIn[i].ScriptSig
	}
	ok := VerifyTxScript(ss, pk, i, tx)
	return ok==valid
}


func TestTransactions(t *testing.T) {
	valid = true
	//DbgSwitch(DBG_SCRIPT, true)
	// It is of particular interest because it contains an invalidly-encoded signature which OpenSSL accepts
	if !testit("0100000001b14bdcbc3e01bdaad36cc08e81e69c82e1060bc14e518db2b49aa43ad90ba26000000000490047304402203f16c6f40162ab686621ef3000b04e75418a0c0cb2d8aebeac894ae360ac1e780220ddc15ecdfc3507ac48e1681a33eb60996631bf6bf5bc0a0682c4db743ce7ca2b01ffffffff0140420f00000000001976a914660d4ef3a743e3e696ad990364e555c271ad504b88ac00000000",
		0, "1 0x41 0x04cc71eb30d653c0c3163990c47b976f3fb3f37cccdcbedb169a1dfef58bbfbfaff7d8a473e7e2e6d317b87bafe8bde97e3cf8f065dec022b51d11fcdd0d348ac4 0x41 0x0461cbdcc5409fb4b4d42b51d33381354d80e550078cb532a34bfa2fcfdeb7d76519aecc62770f5b0e4ef8551946d8a540911abe3e7854a26f39f58b25c15342af 2 OP_CHECKMULTISIG") {
		t.Error("Error")
	}

	// It has an arbitrary extra byte stuffed into the signature at pos length - 2
	if !testit("0100000001b14bdcbc3e01bdaad36cc08e81e69c82e1060bc14e518db2b49aa43ad90ba260000000004A0048304402203f16c6f40162ab686621ef3000b04e75418a0c0cb2d8aebeac894ae360ac1e780220ddc15ecdfc3507ac48e1681a33eb60996631bf6bf5bc0a0682c4db743ce7ca2bab01ffffffff0140420f00000000001976a914660d4ef3a743e3e696ad990364e555c271ad504b88ac00000000",
		0, "1 0x41 0x04cc71eb30d653c0c3163990c47b976f3fb3f37cccdcbedb169a1dfef58bbfbfaff7d8a473e7e2e6d317b87bafe8bde97e3cf8f065dec022b51d11fcdd0d348ac4 0x41 0x0461cbdcc5409fb4b4d42b51d33381354d80e550078cb532a34bfa2fcfdeb7d76519aecc62770f5b0e4ef8551946d8a540911abe3e7854a26f39f58b25c15342af 2 OP_CHECKMULTISIG") {
		t.Error("Error")
	}

	// It is of interest because it contains a 0-sequence as well as a signature of SIGHASH type 0 (which is not a real type)
	if !testit("01000000010276b76b07f4935c70acf54fbf1f438a4c397a9fb7e633873c4dd3bc062b6b40000000008c493046022100d23459d03ed7e9511a47d13292d3430a04627de6235b6e51a40f9cd386f2abe3022100e7d25b080f0bb8d8d5f878bba7d54ad2fda650ea8d158a33ee3cbd11768191fd004104b0e2c879e4daf7b9ab68350228c159766676a14f5815084ba166432aab46198d4cca98fa3e9981d0a90b2effc514b76279476550ba3663fdcaff94c38420e9d5000000000100093d00000000001976a9149a7b0f3b80c6baaeedce0a0842553800f832ba1f88ac00000000",
		0, "DUP HASH160 0x14 0xdc44b1164188067c3a32d4780f5996fa14a4f2d9 EQUALVERIFY CHECKSIG") {
		t.Error("Error")
	}

	// A nearly-standard transaction with CHECKSIGVERIFY 1 instead of CHECKSIG
	if !testit("01000000010001000000000000000000000000000000000000000000000000000000000000000000006a473044022067288ea50aa799543a536ff9306f8e1cba05b9c6b10951175b924f96732555ed022026d7b5265f38d21541519e4a1e55044d5b9e17e15cdbaf29ae3792e99e883e7a012103ba8c8b86dea131c22ab967e6dd99bdae8eff7a1f75a2c35f1f944109e3fe5e22ffffffff010000000000000000015100000000",
		0, "DUP HASH160 0x14 0x5b6462475454710f3c22f5fdf0b40704c92f25c3 EQUALVERIFY CHECKSIGVERIFY 1") {
		t.Error("Error")
	}

	// Same as above, but with the signature duplicated in the scriptPubKey with the proper pushdata prefix
	if !testit("01000000010001000000000000000000000000000000000000000000000000000000000000000000006a473044022067288ea50aa799543a536ff9306f8e1cba05b9c6b10951175b924f96732555ed022026d7b5265f38d21541519e4a1e55044d5b9e17e15cdbaf29ae3792e99e883e7a012103ba8c8b86dea131c22ab967e6dd99bdae8eff7a1f75a2c35f1f944109e3fe5e22ffffffff010000000000000000015100000000",
		0, "DUP HASH160 0x14 0x5b6462475454710f3c22f5fdf0b40704c92f25c3 EQUALVERIFY CHECKSIGVERIFY 1 0x47 0x3044022067288ea50aa799543a536ff9306f8e1cba05b9c6b10951175b924f96732555ed022026d7b5265f38d21541519e4a1e55044d5b9e17e15cdbaf29ae3792e99e883e7a01") {
		t.Error("Error")
	}

	// It caught a bug in the workaround for 23b397edccd3740a74adb603c9756370fafcde9bcc4483eb271ecad09a94dd63 in an overly simple implementation
	if !testit("01000000023d6cf972d4dff9c519eff407ea800361dd0a121de1da8b6f4138a2f25de864b4000000008a4730440220ffda47bfc776bcd269da4832626ac332adfca6dd835e8ecd83cd1ebe7d709b0e022049cffa1cdc102a0b56e0e04913606c70af702a1149dc3b305ab9439288fee090014104266abb36d66eb4218a6dd31f09bb92cf3cfa803c7ea72c1fc80a50f919273e613f895b855fb7465ccbc8919ad1bd4a306c783f22cd3227327694c4fa4c1c439affffffff21ebc9ba20594737864352e95b727f1a565756f9d365083eb1a8596ec98c97b7010000008a4730440220503ff10e9f1e0de731407a4a245531c9ff17676eda461f8ceeb8c06049fa2c810220c008ac34694510298fa60b3f000df01caa244f165b727d4896eb84f81e46bcc4014104266abb36d66eb4218a6dd31f09bb92cf3cfa803c7ea72c1fc80a50f919273e613f895b855fb7465ccbc8919ad1bd4a306c783f22cd3227327694c4fa4c1c439affffffff01f0da5200000000001976a914857ccd42dded6df32949d4646dfa10a92458cfaa88ac00000000",
		0, "DUP HASH160 0x14 0xbef80ecf3a44500fda1bc92176e442891662aed2 EQUALVERIFY CHECKSIG") {
		t.Error("Error")
	}
	if !testit("01000000023d6cf972d4dff9c519eff407ea800361dd0a121de1da8b6f4138a2f25de864b4000000008a4730440220ffda47bfc776bcd269da4832626ac332adfca6dd835e8ecd83cd1ebe7d709b0e022049cffa1cdc102a0b56e0e04913606c70af702a1149dc3b305ab9439288fee090014104266abb36d66eb4218a6dd31f09bb92cf3cfa803c7ea72c1fc80a50f919273e613f895b855fb7465ccbc8919ad1bd4a306c783f22cd3227327694c4fa4c1c439affffffff21ebc9ba20594737864352e95b727f1a565756f9d365083eb1a8596ec98c97b7010000008a4730440220503ff10e9f1e0de731407a4a245531c9ff17676eda461f8ceeb8c06049fa2c810220c008ac34694510298fa60b3f000df01caa244f165b727d4896eb84f81e46bcc4014104266abb36d66eb4218a6dd31f09bb92cf3cfa803c7ea72c1fc80a50f919273e613f895b855fb7465ccbc8919ad1bd4a306c783f22cd3227327694c4fa4c1c439affffffff01f0da5200000000001976a914857ccd42dded6df32949d4646dfa10a92458cfaa88ac00000000",
		1, "DUP HASH160 0x14 0xbef80ecf3a44500fda1bc92176e442891662aed2 EQUALVERIFY CHECKSIG") {
		t.Error("Error")
	}

	// The following tests for the presence of a bug in the handling of SIGHASH_SINGLE
	if !testit("01000000020002000000000000000000000000000000000000000000000000000000000000000000000151ffffffff0001000000000000000000000000000000000000000000000000000000000000000000006b483045022100c9cdd08798a28af9d1baf44a6c77bcc7e279f47dc487c8c899911bc48feaffcc0220503c5c50ae3998a733263c5c0f7061b483e2b56c4c41b456e7d2f5a78a74c077032102d5c25adb51b61339d2b05315791e21bbe80ea470a49db0135720983c905aace0ffffffff010000000000000000015100000000",
		1, "DUP HASH160 0x14 0xe52b482f2faa8ecbf0db344f93c84ac908557f33 EQUALVERIFY CHECKSIG") {
		t.Error("Error")
	}

	valid = false
	// An invalid P2SH Transaction
	if !testit("010000000100010000000000000000000000000000000000000000000000000000000000000000000009085768617420697320ffffffff010000000000000000015100000000",
		0, "HASH160 0x14 0x7a052c840ba73af26755de42cf01cc9e0a49fef0 EQUAL") {
		t.Error("Error")
	}

	valid = true
	// A valid P2SH Transaction using the standard transaction type put forth in BIP 16
	if !testit("01000000010001000000000000000000000000000000000000000000000000000000000000000000006e493046022100c66c9cdf4c43609586d15424c54707156e316d88b0a1534c9e6b0d4f311406310221009c0fe51dbc9c4ab7cc25d3fdbeccf6679fe6827f08edf2b4a9f16ee3eb0e438a0123210338e8034509af564c62644c07691942e0c056752008a173c89f60ab2a88ac2ebfacffffffff010000000000000000015100000000",
		0, "HASH160 0x14 0x8febbed40483661de6958d957412f82deed8e2f7 EQUAL") {
		t.Error("Error")
	}

	// MAX_MONEY output
	if !testit("01000000010001000000000000000000000000000000000000000000000000000000000000000000006e493046022100e1eadba00d9296c743cb6ecc703fd9ddc9b3cd12906176a226ae4c18d6b00796022100a71aef7d2874deff681ba6080f1b278bac7bb99c61b08a85f4311970ffe7f63f012321030c0588dc44d92bdcbf8e72093466766fdc265ead8db64517b0c542275b70fffbacffffffff010040075af0750700015100000000",
		0, "HASH160 0x14 0x32afac281462b822adbec5094b8d4d337dd5bd6a EQUAL") {
		t.Error("Error")
	}

	// MAX_MONEY output + 0 output
	if !testit("01000000010001000000000000000000000000000000000000000000000000000000000000000000006d483045022027deccc14aa6668e78a8c9da3484fbcd4f9dcc9bb7d1b85146314b21b9ae4d86022100d0b43dece8cfb07348de0ca8bc5b86276fa88f7f2138381128b7c36ab2e42264012321029bb13463ddd5d2cc05da6e84e37536cb9525703cfd8f43afdb414988987a92f6acffffffff020040075af075070001510000000000000000015100000000",
		0, "HASH160 0x14 0xb558cbf4930954aa6a344363a15668d7477ae716 EQUAL") {
		t.Error("Error")
	}

	// Coinbase of size 2
	if !testit("01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff025151ffffffff010000000000000000015100000000",
		-1, "1") {
		t.Error("Error")
	}

	// Coinbase of size 100
	if !testit("01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff6451515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151ffffffff010000000000000000015100000000",
		-1, "1") {
		t.Error("Error")
	}

	// Simple transaction with first input is signed with SIGHASH_ALL, second with SIGHASH_ANYONECANPAY
	if !testit("010000000200010000000000000000000000000000000000000000000000000000000000000000000049483045022100d180fd2eb9140aeb4210c9204d3f358766eb53842b2a9473db687fa24b12a3cc022079781799cd4f038b85135bbe49ec2b57f306b2bb17101b17f71f000fcab2b6fb01ffffffff0002000000000000000000000000000000000000000000000000000000000000000000004847304402205f7530653eea9b38699e476320ab135b74771e1c48b81a5d041e2ca84b9be7a802200ac8d1f40fb026674fe5a5edd3dea715c27baa9baca51ed45ea750ac9dc0a55e81ffffffff010100000000000000015100000000",
		0, "0x21 0x035e7f0d4d0841bcd56c39337ed086b1a633ee770c1ffdd94ac552a95ac2ce0efc CHECKSIG") {
		t.Error("Error")
	}
	if !testit("010000000200010000000000000000000000000000000000000000000000000000000000000000000049483045022100d180fd2eb9140aeb4210c9204d3f358766eb53842b2a9473db687fa24b12a3cc022079781799cd4f038b85135bbe49ec2b57f306b2bb17101b17f71f000fcab2b6fb01ffffffff0002000000000000000000000000000000000000000000000000000000000000000000004847304402205f7530653eea9b38699e476320ab135b74771e1c48b81a5d041e2ca84b9be7a802200ac8d1f40fb026674fe5a5edd3dea715c27baa9baca51ed45ea750ac9dc0a55e81ffffffff010100000000000000015100000000",
		1, "0x21 0x035e7f0d4d0841bcd56c39337ed086b1a633ee770c1ffdd94ac552a95ac2ce0efc CHECKSIG") {
		t.Error("Error")
	}

	//Same as above, but we change the sequence number of the first input to check that SIGHASH_ANYONECANPAY is being followed
	if !testit("01000000020001000000000000000000000000000000000000000000000000000000000000000000004948304502203a0f5f0e1f2bdbcd04db3061d18f3af70e07f4f467cbc1b8116f267025f5360b022100c792b6e215afc5afc721a351ec413e714305cb749aae3d7fee76621313418df101010000000002000000000000000000000000000000000000000000000000000000000000000000004847304402205f7530653eea9b38699e476320ab135b74771e1c48b81a5d041e2ca84b9be7a802200ac8d1f40fb026674fe5a5edd3dea715c27baa9baca51ed45ea750ac9dc0a55e81ffffffff010100000000000000015100000000",
		0, "0x21 0x035e7f0d4d0841bcd56c39337ed086b1a633ee770c1ffdd94ac552a95ac2ce0efc CHECKSIG") {
		t.Error("Error")
	}
	if !testit("01000000020001000000000000000000000000000000000000000000000000000000000000000000004948304502203a0f5f0e1f2bdbcd04db3061d18f3af70e07f4f467cbc1b8116f267025f5360b022100c792b6e215afc5afc721a351ec413e714305cb749aae3d7fee76621313418df101010000000002000000000000000000000000000000000000000000000000000000000000000000004847304402205f7530653eea9b38699e476320ab135b74771e1c48b81a5d041e2ca84b9be7a802200ac8d1f40fb026674fe5a5edd3dea715c27baa9baca51ed45ea750ac9dc0a55e81ffffffff010100000000000000015100000000",
		1, "0x21 0x035e7f0d4d0841bcd56c39337ed086b1a633ee770c1ffdd94ac552a95ac2ce0efc CHECKSIG") {
		t.Error("Error")
	}

	//afd9c17f8913577ec3509520bd6e5d63e9c0fd2a5f70c787993b097ba6ca9fae which has several SIGHASH_SINGLE signatures
	if !testit("010000000370ac0a1ae588aaf284c308d67ca92c69a39e2db81337e563bf40c59da0a5cf63000000006a4730440220360d20baff382059040ba9be98947fd678fb08aab2bb0c172efa996fd8ece9b702201b4fb0de67f015c90e7ac8a193aeab486a1f587e0f54d0fb9552ef7f5ce6caec032103579ca2e6d107522f012cd00b52b9a65fb46f0c57b9b8b6e377c48f526a44741affffffff7d815b6447e35fbea097e00e028fb7dfbad4f3f0987b4734676c84f3fcd0e804010000006b483045022100c714310be1e3a9ff1c5f7cacc65c2d8e781fc3a88ceb063c6153bf950650802102200b2d0979c76e12bb480da635f192cc8dc6f905380dd4ac1ff35a4f68f462fffd032103579ca2e6d107522f012cd00b52b9a65fb46f0c57b9b8b6e377c48f526a44741affffffff3f1f097333e4d46d51f5e77b53264db8f7f5d2e18217e1099957d0f5af7713ee010000006c493046022100b663499ef73273a3788dea342717c2640ac43c5a1cf862c9e09b206fcb3f6bb8022100b09972e75972d9148f2bdd462e5cb69b57c1214b88fc55ca638676c07cfc10d8032103579ca2e6d107522f012cd00b52b9a65fb46f0c57b9b8b6e377c48f526a44741affffffff0380841e00000000001976a914bfb282c70c4191f45b5a6665cad1682f2c9cfdfb88ac80841e00000000001976a9149857cc07bed33a5cf12b9c5e0500b675d500c81188ace0fd1c00000000001976a91443c52850606c872403c0601e69fa34b26f62db4a88ac00000000",
		0, "DUP HASH160 0x14 0xdcf72c4fd02f5a987cf9b02f2fabfcac3341a87d EQUALVERIFY CHECKSIG") {
		t.Error("Error")
	}
	if !testit("010000000370ac0a1ae588aaf284c308d67ca92c69a39e2db81337e563bf40c59da0a5cf63000000006a4730440220360d20baff382059040ba9be98947fd678fb08aab2bb0c172efa996fd8ece9b702201b4fb0de67f015c90e7ac8a193aeab486a1f587e0f54d0fb9552ef7f5ce6caec032103579ca2e6d107522f012cd00b52b9a65fb46f0c57b9b8b6e377c48f526a44741affffffff7d815b6447e35fbea097e00e028fb7dfbad4f3f0987b4734676c84f3fcd0e804010000006b483045022100c714310be1e3a9ff1c5f7cacc65c2d8e781fc3a88ceb063c6153bf950650802102200b2d0979c76e12bb480da635f192cc8dc6f905380dd4ac1ff35a4f68f462fffd032103579ca2e6d107522f012cd00b52b9a65fb46f0c57b9b8b6e377c48f526a44741affffffff3f1f097333e4d46d51f5e77b53264db8f7f5d2e18217e1099957d0f5af7713ee010000006c493046022100b663499ef73273a3788dea342717c2640ac43c5a1cf862c9e09b206fcb3f6bb8022100b09972e75972d9148f2bdd462e5cb69b57c1214b88fc55ca638676c07cfc10d8032103579ca2e6d107522f012cd00b52b9a65fb46f0c57b9b8b6e377c48f526a44741affffffff0380841e00000000001976a914bfb282c70c4191f45b5a6665cad1682f2c9cfdfb88ac80841e00000000001976a9149857cc07bed33a5cf12b9c5e0500b675d500c81188ace0fd1c00000000001976a91443c52850606c872403c0601e69fa34b26f62db4a88ac00000000",
		1, "DUP HASH160 0x14 0xdcf72c4fd02f5a987cf9b02f2fabfcac3341a87d EQUALVERIFY CHECKSIG") {
		t.Error("Error")
	}
	if !testit("010000000370ac0a1ae588aaf284c308d67ca92c69a39e2db81337e563bf40c59da0a5cf63000000006a4730440220360d20baff382059040ba9be98947fd678fb08aab2bb0c172efa996fd8ece9b702201b4fb0de67f015c90e7ac8a193aeab486a1f587e0f54d0fb9552ef7f5ce6caec032103579ca2e6d107522f012cd00b52b9a65fb46f0c57b9b8b6e377c48f526a44741affffffff7d815b6447e35fbea097e00e028fb7dfbad4f3f0987b4734676c84f3fcd0e804010000006b483045022100c714310be1e3a9ff1c5f7cacc65c2d8e781fc3a88ceb063c6153bf950650802102200b2d0979c76e12bb480da635f192cc8dc6f905380dd4ac1ff35a4f68f462fffd032103579ca2e6d107522f012cd00b52b9a65fb46f0c57b9b8b6e377c48f526a44741affffffff3f1f097333e4d46d51f5e77b53264db8f7f5d2e18217e1099957d0f5af7713ee010000006c493046022100b663499ef73273a3788dea342717c2640ac43c5a1cf862c9e09b206fcb3f6bb8022100b09972e75972d9148f2bdd462e5cb69b57c1214b88fc55ca638676c07cfc10d8032103579ca2e6d107522f012cd00b52b9a65fb46f0c57b9b8b6e377c48f526a44741affffffff0380841e00000000001976a914bfb282c70c4191f45b5a6665cad1682f2c9cfdfb88ac80841e00000000001976a9149857cc07bed33a5cf12b9c5e0500b675d500c81188ace0fd1c00000000001976a91443c52850606c872403c0601e69fa34b26f62db4a88ac00000000",
		2, "DUP HASH160 0x14 0xdcf72c4fd02f5a987cf9b02f2fabfcac3341a87d EQUALVERIFY CHECKSIG") {
		t.Error("Error")
	}


	// The negative tests...

	valid = false
	//DbgSwitch(DBG_SCRIPT, true)
	// 0e1b5688cf179cd9f7cbda1fac0090f6e684bbf8cd946660120197c3f3681809 but with extra junk appended to the end of the scriptPubKey
	if !testit("010000000127587a10248001f424ad94bb55cd6cd6086a0e05767173bdbdf647187beca76c000000004948304502201b822ad10d6adc1a341ae8835be3f70a25201bbff31f59cbb9c5353a5f0eca18022100ea7b2f7074e9aa9cf70aa8d0ffee13e6b45dddabf1ab961bda378bcdb778fa4701ffffffff0100f2052a010000001976a914fc50c5907d86fed474ba5ce8b12a66e0a4c139d888ac00000000",
		0, "0x41 0x043b640e983c9690a14c039a2037ecc3467b27a0dcd58f19d76c7bc118d09fec45adc5370a1c5bf8067ca9f5557a4cf885fdb0fe0dcc9c3a7137226106fbc779a5 CHECKSIG VERIFY 1") {
		t.Error("Error")
	}

	// This is the nearly-standard transaction with CHECKSIGVERIFY 1 instead of CHECKSIG from tx_valid.json
	// but with the signature duplicated in the scriptPubKey with a non-standard pushdata prefix
	// "See FindAndDelete, which will only remove if it uses the same pushdata prefix as is standard
	if !testit("01000000010001000000000000000000000000000000000000000000000000000000000000000000006a473044022067288ea50aa799543a536ff9306f8e1cba05b9c6b10951175b924f96732555ed022026d7b5265f38d21541519e4a1e55044d5b9e17e15cdbaf29ae3792e99e883e7a012103ba8c8b86dea131c22ab967e6dd99bdae8eff7a1f75a2c35f1f944109e3fe5e22ffffffff010000000000000000015100000000",
		0, "DUP HASH160 0x14 0x5b6462475454710f3c22f5fdf0b40704c92f25c3 EQUALVERIFY CHECKSIGVERIFY 1 0x4c 0x47 0x3044022067288ea50aa799543a536ff9306f8e1cba05b9c6b10951175b924f96732555ed022026d7b5265f38d21541519e4a1e55044d5b9e17e15cdbaf29ae3792e99e883e7a01") {
		t.Error("Error")
	}

	// Same as above, but with the sig in the scriptSig also pushed with the same non-standard OP_PUSHDATA
	if !testit("01000000010001000000000000000000000000000000000000000000000000000000000000000000006b4c473044022067288ea50aa799543a536ff9306f8e1cba05b9c6b10951175b924f96732555ed022026d7b5265f38d21541519e4a1e55044d5b9e17e15cdbaf29ae3792e99e883e7a012103ba8c8b86dea131c22ab967e6dd99bdae8eff7a1f75a2c35f1f944109e3fe5e22ffffffff010000000000000000015100000000",
		0, "DUP HASH160 0x14 0x5b6462475454710f3c22f5fdf0b40704c92f25c3 EQUALVERIFY CHECKSIGVERIFY 1 0x4c 0x47 0x3044022067288ea50aa799543a536ff9306f8e1cba05b9c6b10951175b924f96732555ed022026d7b5265f38d21541519e4a1e55044d5b9e17e15cdbaf29ae3792e99e883e7a01") {
		t.Error("Error")
	}


	// The remainig tests are mostly not aplicable to gocoin architecture...

	// An invalid P2SH Transaction
	if !testit("010000000100010000000000000000000000000000000000000000000000000000000000000000000009085768617420697320ffffffff010000000000000000015100000000",
		0, "HASH160 0x14 0x7a052c840ba73af26755de42cf01cc9e0a49fef0 EQUAL") {
		t.Error("Error")
	}

	// No inputs
	if !testit("0100000000010000000000000000015100000000",
		-1, "HASH160 0x14 0x7a052c840ba73af26755de42cf01cc9e0a49fef0 EQUAL") {
		t.Error("Error")
	}

	// No outputs
	if !testit("01000000010001000000000000000000000000000000000000000000000000000000000000000000006d483045022100f16703104aab4e4088317c862daec83440242411b039d14280e03dd33b487ab802201318a7be236672c5c56083eb7a5a195bc57a40af7923ff8545016cd3b571e2a601232103c40e5d339df3f30bf753e7e04450ae4ef76c9e45587d1d993bdc4cd06f0651c7acffffffff0000000000",
		0, "HASH160 0x14 0x05ab9e14d983742513f0f451e105ffb4198d1dd4 EQUAL") {
		t.Error("Error")
	}

	// Negative output
	/*
	It's a stupid test (cannot happen in a real life)
	if !testit("01000000010001000000000000000000000000000000000000000000000000000000000000000000006d4830450220063222cbb128731fc09de0d7323746539166544d6c1df84d867ccea84bcc8903022100bf568e8552844de664cd41648a031554327aa8844af34b4f27397c65b92c04de0123210243ec37dee0e2e053a9c976f43147e79bc7d9dc606ea51010af1ac80db6b069e1acffffffff01ffffffffffffffff015100000000",
		0, "HASH160 0x14 0xae609aca8061d77c5e111f6bb62501a6bbe2bfdb EQUAL") {
		t.Error("Error")
	}
	*/

	// MAX_MONEY + 1 output
	if !testit("01000000010001000000000000000000000000000000000000000000000000000000000000000000006e493046022100e1eadba00d9296c743cb6ecc703fd9ddc9b3cd12906176a226ae4c18d6b00796022100a71aef7d2874deff681ba6080f1b278bac7bb99c61b08a85f4311970ffe7f63f012321030c0588dc44d92bdcbf8e72093466766fdc265ead8db64517b0c542275b70fffbacffffffff010140075af0750700015100000000",
		0, "HASH160 0x14 0x32afac281462b822adbec5094b8d4d337dd5bd6a EQUAL") {
		t.Error("Error")
	}

	// MAX_MONEY output + 1 output
	if !testit("01000000010001000000000000000000000000000000000000000000000000000000000000000000006d483045022027deccc14aa6668e78a8c9da3484fbcd4f9dcc9bb7d1b85146314b21b9ae4d86022100d0b43dece8cfb07348de0ca8bc5b86276fa88f7f2138381128b7c36ab2e42264012321029bb13463ddd5d2cc05da6e84e37536cb9525703cfd8f43afdb414988987a92f6acffffffff020040075af075070001510001000000000000015100000000",
		0, "HASH160 0x14 0xb558cbf4930954aa6a344363a15668d7477ae716 EQUAL") {
		t.Error("Error")
	}

	// Duplicate inputs
	/*
	This is normally handled on a different level
	if !testit("01000000020001000000000000000000000000000000000000000000000000000000000000000000006c47304402204bb1197053d0d7799bf1b30cd503c44b58d6240cccbdc85b6fe76d087980208f02204beeed78200178ffc6c74237bb74b3f276bbb4098b5605d814304fe128bf1431012321039e8815e15952a7c3fada1905f8cf55419837133bd7756c0ef14fc8dfe50c0deaacffffffff0001000000000000000000000000000000000000000000000000000000000000000000006c47304402202306489afef52a6f62e90bf750bbcdf40c06f5c6b138286e6b6b86176bb9341802200dba98486ea68380f47ebb19a7df173b99e6bc9c681d6ccf3bde31465d1f16b3012321039e8815e15952a7c3fada1905f8cf55419837133bd7756c0ef14fc8dfe50c0deaacffffffff010000000000000000015100000000",
		0, "HASH160 0x14 0x236d0639db62b0773fd8ac34dc85ae19e9aba80a EQUAL") {
		t.Error("Error")
	}
	*/

	// Coinbase of size 1
	/*
	This function does not check coinbase
	if !testit("01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0151ffffffff010000000000000000015100000000",
		0, "1") {
		t.Error("Error")
	}
	*/

	// Coinbase of size 101
	/*
	This function does not check coinbase
	if !testit("01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff655151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151515151ffffffff010000000000000000015100000000",
		0, "1") {
		t.Error("Error")
	}
	*/

	// Null txin
	/*
	We wont find a null txin in out UTXO db
	if !testit("01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff6e49304602210086f39e028e46dafa8e1e3be63906465f4cf038fbe5ed6403dc3e74ae876e6431022100c4625c675cfc5c7e3a0e0d7eaec92ac24da20c73a88eb40d09253e51ac6def5201232103a183ddc41e84753aca47723c965d1b5c8b0e2b537963518355e6dd6cf8415e50acffffffff010000000000000000015100000000",
		0, "HASH160 0x14 0x02dae7dbbda56097959cba59b1989dd3e47937bf EQUAL") {
		t.Error("Error")
	}
	*/

	// Same as the transactions in valid with one input SIGHASH_ALL and one SIGHASH_ANYONECANPAY, but we set the _ANYONECANPAY sequence number, invalidating the SIGHASH_ALL signature
	if !testit("01000000020001000000000000000000000000000000000000000000000000000000000000000000004948304502203a0f5f0e1f2bdbcd04db3061d18f3af70e07f4f467cbc1b8116f267025f5360b022100c792b6e215afc5afc721a351ec413e714305cb749aae3d7fee76621313418df10101000000000200000000000000000000000000000000000000000000000000000000000000000000484730440220201dc2d030e380e8f9cfb41b442d930fa5a685bb2c8db5906671f865507d0670022018d9e7a8d4c8d86a73c2a724ee38ef983ec249827e0e464841735955c707ece98101000000010100000000000000015100000000",
		0, "0x21 0x035e7f0d4d0841bcd56c39337ed086b1a633ee770c1ffdd94ac552a95ac2ce0efc CHECKSIG") {
		t.Error("Error")
	}
}


func TestScritps(t *testing.T) {
	// use some dummy tx
	rd, _ := hex.DecodeString("0100000001b14bdcbc3e01bdaad36cc08e81e69c82e1060bc14e518db2b49aa43ad90ba26000000000490047304402203f16c6f40162ab686621ef3000b04e75418a0c0cb2d8aebeac894ae360ac1e780220ddc15ecdfc3507ac48e1681a33eb60996631bf6bf5bc0a0682c4db743ce7ca2b01ffffffff0140420f00000000001976a914660d4ef3a743e3e696ad990364e555c271ad504b88ac00000000")
	tx, _ := NewTx(rd)
	tx.Size = uint32(len(rd))
	ha := Sha2Sum(rd)
	tx.Hash = NewUint256(ha[:])

	dat, er := ioutil.ReadFile("test/script_valid.json")
	if er != nil {
		t.Error(er.Error())
		return
	}
	var vecs [][]string
	er = json.Unmarshal(dat, &vecs)
	if er != nil {
		t.Error(er.Error())
		return
	}

	tot := 0
	for i := range vecs {
		if len(vecs[i])>=2 {
			tot++

			s1, e := pk2hex(vecs[i][0])
			if e!=nil {
				t.Error(tot, "error A in", vecs[i][0], "->", vecs[i][1])
				return
			}
			s2, e := pk2hex(vecs[i][1])
			if e!=nil {
				t.Error(tot, "error B in", vecs[i][0], "->", vecs[i][1])
				return
			}

			res := VerifyTxScript(s1, s2, 0, tx)
			if !res {
				t.Error(tot, "VerifyTxScript failed in", vecs[i][0], "->", vecs[i][1])
				return
			}
		}
	}
	//t.Info(tot, "valid vectors processed")

	dat, er = ioutil.ReadFile("test/script_invalid.json")
	if er != nil {
		t.Error(er.Error())
		return
	}
	er = json.Unmarshal(dat, &vecs)
	if er != nil {
		t.Error(er.Error())
		return
	}

	//DbgSwitch(DBG_SCRIPT, true)
	tot = 0
	for i := range vecs {
		if len(vecs[i])>=2 {
			tot++

			s1, e := pk2hex(vecs[i][0])
			if e!=nil {
				t.Error(tot, "error A in", vecs[i][0], "->", vecs[i][1])
				return
			}
			s2, e := pk2hex(vecs[i][1])
			if e!=nil {
				t.Error(tot, "error B in", vecs[i][0], "->", vecs[i][1])
				return
			}

			res := VerifyTxScript(s1, s2, 0, tx)
			if res {
				t.Error(tot, "VerifyTxScript NOT failed in", vecs[i][0], "->", vecs[i][1])
				return
			}
		}
	}
	//println(tot, "invalid vectors processed")
}
