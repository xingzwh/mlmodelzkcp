package main

import (
	"fmt"
	"encoding/json"
	_"strconv"
	"crypto/rand"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/xingzwh/gnark/backend/groth16"
	ccgroth16 "github.com/xingzwh/gnark/curvepp/backend/bn254/groth16"
	"github.com/xingzwh/gnark/backend/witness"
	"github.com/xingzwh/gnark/frontend/schema"
	"github.com/xingzwh/gnark/frontend"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)


type CC_Arbiter struct {
}

func (t *CC_Arbiter) Init(stub shim.ChaincodeStubInterface) pb.Response {
	stub.PutState("seller",[]byte("0"))
	stub.PutState("buyer",[]byte("100"))
	return shim.Success([]byte("successful init"))
}

func (t *CC_Arbiter) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fn, args := stub.GetFunctionAndParameters()	
	if fn == "uploadvalue" {
		return t.UploadValue(stub, args)
						
	}else if fn == "getvalue"{
		return t.GetValue(stub, args)			
	}else if fn == "getprivatevalue"{
		return t.GetPrivateValue(stub, args)			
	}else if fn == "uploadprivatevalue"{
		return t.UploadPrivateValue(stub, args)			
	}else if fn == "generatesk"{
		return t.GenerateSecretKey(stub, args)				
	}else if fn == "verifyproof"{
		return t.LRGroth16Verify(stub, args)			
	}else if fn == "hashcheck"{
		return t.HashCheck(stub, args)			
	}else{
		return shim.Error("no such opration")
	}	
}

func (t *CC_Arbiter) GetValue(stub shim.ChaincodeStubInterface, args []string) pb.Response{	

	keyID := args[0]
	keyvalue,err:=stub.GetState(keyID)
	if err != nil {
		return shim.Error("error getvalue")
	}
	fmt.Println(keyID)
	fmt.Println(keyvalue)
	return shim.Success([]byte(keyvalue))
}
func (t *CC_Arbiter) UploadValue(stub shim.ChaincodeStubInterface, args []string) pb.Response{	
	keyID := args[0]
	keyvalue := args[1]
	stub.PutState(keyID,[]byte(keyvalue))
	fmt.Println(keyID)
	fmt.Println(keyvalue)
	return shim.Success([]byte("success put " + keyID + " " + keyvalue))
}
func (t *CC_Arbiter) GetPrivateValue(stub shim.ChaincodeStubInterface, args []string) pb.Response{	
	keyID := args[0]
	keyvalue, err := stub.GetPrivateData("PrivateKeyCollection",keyID)
	if err != nil {
		return shim.Error("read private parms error")
	}
	return shim.Success([]byte(keyvalue))
}
func (t *CC_Arbiter) UploadPrivateValue(stub shim.ChaincodeStubInterface, args []string) pb.Response{	
	keyID := args[0]
	keyvalue := args[1]
	err := stub.PutPrivateData("PrivateKeyCollection",keyID, []byte(keyvalue))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("success put privatedata" ))
}


func (t *CC_Arbiter) GenerateSecretKey(stub shim.ChaincodeStubInterface, args []string) pb.Response{
	keyID := args[0]
	skey := make([]byte, 32)
	_, err := rand.Read(skey)
	if err != nil {
		return nil, err
	}
	
	//secretkey 上传Arbiter 私有数据集
	err := stub.PutPrivateData("PrivateKeyCollection",keyID, skey)
	if err != nil {
		return shim.Error(err.Error())
	}
	
	
	return shim.Success([]byte(skey))
}


func (t *CC_Arbiter) LRGroth16Verify(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	keyID := args[0]
	
	//proof
	proofName:="lr"+keyID+"proof"	
	proofData,_:=stub.GetState(proofName) 
	
	var proof *ccgroth16.Proof
	errppf := json.Unmarshal(proofResponse.Payload,&pproof)
	if errppf != nil {
		return shim.Error("Error download pproof")
	}
	
	//vk
	vkName:="lr"+keyID+"vk"
	vkData,_:=stub.GetState(vkName)  
	
	var vk *ccgroth16.VerifyingKey
	errpvk := json.Unmarshal(vkResponse.Payload,&pvk)
	if errpvk != nil {
		return shim.Error("Error download pvk")
	}
	

	
	//scc
	sccName:="lr"+keyID+"scc"
	sccData,_:=stub.GetState(sccName)
	var schemacc *schema.Schema
	errsc := json.Unmarshal(sccData,&schemacc)
	if errsc != nil {
		return shim.Error("Error download schame")
	}
	//ecc
	eccName:="lr"+keyID+"ecc"
	eccData,_:=stub.GetState(eccName)	
	var curvecc ecc.ID
	errec := json.Unmarshal(eccData,&curvecc)
	if errec != nil {
		return shim.Error("Error unmarshal curve")
	}
	
	//pw
	pwName:="lr"+keyID+"pw"
	pwData,_:=stub.GetState(pwName)
	witnesscc:=witness.Witness{CurveID:curvecc,Schema:schemacc}
	unmarshal:=witnesscc.UnmarshalJSON
	errpw := unmarshal(pwData)
	if errpw != nil {
		return shim.Error("Error download pw")
	}
	
	
	err := groth16.Verify(pproof, pvk, &witnesscc)
	if err != nil {
		return shim.Error("Error verify")
	}
	return shim.Success([]byte("success verify proof"))
}
func (t *CC_Arbiter) HashCheck(stub shim.ChaincodeStubInterface, args []string) pb.Response {
//将已有的hash与prover中的环境hash重新计算
	keyID = args[0]
	
	chaincodeId:= args[1]
	hashvalue := args[2]
	hashcheckArgs :=[][]byte{[]byte("verifyhash"),[]byte(keyID),[]byte(chaincodeId),[]byte(hashvalue)}
	hashcheckResponse:=stub.InvokeChaincode(chaincodeId,hashcheckArgs ,"testch1")
	if hashcheckResponse.Status != shim.OK {
		return shim.Error("invoke chaincode failed:"+string(hashcheckResponse.Status))
	}
	fmt.Printf(string(hashcheckResponse.Payload))
	fmt.Printf("verify hash in " + chaincodeId)

	return shim.Success([]byte("challenge success"))

}

func EncryptAES(key string, plainText string) (string, error) {

    cipher, err := aes.NewCipher([]byte(key))

    if err != nil {
        return "", err
    }

    out := make([]byte, len(plainText))

    cipher.Encrypt(out, []byte(plainText))

    return hex.EncodeToString(out), nil
}

func DecryptAES(key string, encryptText string) (string, error) {
    decodeText, _ := hex.DecodeString(encryptText)

    cipher, err := aes.NewCipher([]byte(key))
    if err != nil {
        return "", err
    }

    out := make([]byte, len(decodeText))
    cipher.Decrypt(out, decodeText)

    return string(out[:]), nil
}
func (t *CC_Buyer) TradeProcess(stub shim.ChaincodeStubInterface, args []string) pb.Response{
 	keyID := args[0]
	//get z
	ennumName := keyID + "encryptNum"
	ennumdata,err:=stub.GetState(ennumName)
	if err != nil {
		return shim.Error("error getvalue")
	}
	ennum,err:=strconv.Atoi(string(ennumdata))
	if err != nil {
		return shim.Error("Error")
	}
	var z []string
	for i:=0;i<ennum;i++ {
		encryptName := keyID + strconv.Itoa(i) + "encryptValue"
		encryptdata,err:=stub.GetState(encryptName)
		if err != nil {
			return shim.Error("error getvalue")
		}
		z = append(z,string(encryptdata))
	}
        
	//get k
	sk, err := stub.GetPrivateData("PrivateKeyCollection",keyID)
	if err != nil {
		return shim.Error("read private parms error")
	}
	
	//dec w
	var parameter []string
	for i:=0;i<ennum;i++ {
		decrypt, err := DecryptAES(string(sk), z[i])
            	if err != nil {
            		log.Fatal(err)
            	}  
		parameter = append(parameter,decrypt)
	}
	
	//compute w hash 
	whashbyte,_:= json.Marshal(w)
	whashenv := sha256.Sum256(whashbyte)
	whashValue := hex.EncodeToString(whashenv[:])
	fmt.Println(whashValue)
	
	//verify whash 
	verifyhashName := keyID+"parameterhash"
	verifyhash, err := stub.GetPrivateData("PrivateKeyCollection",verifyhashName)
	if err != nil {
		return shim.Error("read private parms error")
	}
	//检测
	if string(verifyhash) != whashValue {
		return shim.Error("parameter hash check error")
	}else{
		skArgs :=[][]byte{[]byte("uploadvalue"),[]byte(keyID),[]byte(sk)}
		skResponse:=stub.InvokeChaincode("CC_Buyer",skArgs,"testch1")
		if skResponse.Status != shim.OK {
            		return shim.Error("failed to query chaincode.got error :"+string(skResponse.Payload))
            	}
            	fmt.Println("upload sk to buyer")
            	
            	tokensName := keyID + "tokens"
            	tokensdata,err:=stub.GetState(tokensName)
		if err != nil {
			return shim.Error("error getvalue")
		}
		tokens,err:=strconv.Atoi(string(tokensdata))
		if err != nil {
			return shim.Error("Error")
		}
		
		coinArgs :=[][]byte{[]byte("getvalue"),[]byte("seller")}
		coinResponse:=stub.InvokeChaincode("CC_Seller",coinArgs,"testch1")
		if coinResponse.Status != shim.OK {
            		return shim.Error("failed to query chaincode.got error :"+string(coinResponse.Payload))
            	}
		sellercoin,err:= strconv.Atoi(string(coinResponse.Payload))
		if err != nil {
			return shim.Error("Error")
		}
		
		tradecoin :=sellercoin + tokens
		tradeArgs :=[][]byte{[]byte("uploadvalue"),[]byte("seller"),[]byte(string(tradecoin))}
		tradeResponse:=stub.InvokeChaincode("CC_Seller",tradeArgs,"testch1")
		if tradeResponse.Status != shim.OK {
            		return shim.Error("failed to query chaincode.got error :"+string(tradeResponse.Payload))
            	}
            	stub.PutState(tokensName,[]byte("0"))
            	
		return shim.Success([]byte("finish trade"))
            	
            	
	}
	
	return shim.Success([]byte("success hashcheck training"))
}



func main() {
	if err := shim.Start(new(CC_Arbiter)); err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
