package main

import (
	"fmt"
	_"os"
	_"io"
	_"encoding/csv"
	_"encoding/json"
	_"reflect"
	_"bytes"
	"strconv"


	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type CC_Buyer struct {
}

func (t *CC_Buyer) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success([]byte("successful init"))
}

func (t *CC_Buyer) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fn, args := stub.GetFunctionAndParameters()	
	if fn == "starttrade" {
		return t.StartTrade(stub, args)
						
	//}else if fn == "challenge"{
	//	return t.Challenge(stub, args)	
				
	}else if fn == "uploadvalue" {
		return t.UploadValue(stub, args)				
	}else if fn == "getvalue"{
		return t.GetValue(stub, args)		
	}else if fn == "challenge"{
		return t.Challenge(stub, args)		
	}else if fn == "challengecheck"{
		return t.ChallengeCheck(stub, args)		
	}else if fn == "trade"{
		return t.TradeProcess(stub, args)		
	}else{
		return shim.Error("no such opration")
	}
	
}
func (t *CC_Buyer) StartTrade(stub shim.ChaincodeStubInterface, args []string) pb.Response{	
	//trade ID : keyID
	keyID := args[0]	
	
	//threshold requirements:thresholdReq
	//a float num must be a int (0.8 -> 80)	
	//a string must convert to a int
	thrReq := args[1]	
	
	//Trade payment:tradePay
	//for example:"1000",a string must convert to a int
	tradePay := args[2]
	
	//threshold key in CC_Seller
	thrName:= keyID + "threshold"
	//upload threshold requirements to CC_Seller
	thrArgs :=[][]byte{[]byte("uploadvalue"),[]byte(thrName),[]byte(thrReq)}
	thrResponse:=stub.InvokeChaincode("CC_Seller",thrArgs,"testch1")
	if thrResponse.Status != shim.OK {
		return shim.Error("failed to query chaincode.got error :"+string(thrResponse.Payload))
	}
	fmt.Printf("upload threshold to CC_Seller")
	
	
	
	//payment in CC_Seller
	paymentName:= keyID + "payment"
	//upload Trade payment to CC_Seller
	paymentArgs :=[][]byte{[]byte("uploadvalue"),[]byte(paymentName),[]byte(tradePay)}
	paymentResponse:=stub.InvokeChaincode("CC_Seller",paymentArgs,"testch1")
	if paymentResponse.Status != shim.OK {
		return shim.Error("failed to query chaincode.got error :"+string(paymentResponse.Payload))
	}
	fmt.Printf("upload payment to CC_Seller")
	
		
	return shim.Success([]byte("success put threshold to seller"))
}




func (t *CC_Buyer) TradeProcess(stub shim.ChaincodeStubInterface, args []string) pb.Response{
	keyID := args[0]
	tokens := args[1]
	
	//自减少tokens
	coin,err := stub.GetState("buyer")
	if err != nil {
		return shim.Error("error getcoin")
	}
	keycoin , _ := strconv.Atoi(string(coin))
	tokenscoin , _ := strconv.Atoi(tokens)
	newCoin := keycoin - tokenscoin
	if newCoin < 0 {
		return shim.Error("coin is not enough")
	}else{
		stub.PutState("buyer",[]byte(strconv.Itoa(newCoin)))
	}
	
	//验证payment
	//get pay from CC_Arbiter
	tokensname := keyID + "tokens"
	queryArgs :=[][]byte{[]byte("UploadValue"),[]byte(payname),[]byte(tokens)}
	response:=stub.InvokeChaincode("CC_Arbiter",queryArgs,"testch1")
	if response.Status != shim.OK {
		return shim.Error("failed to query chaincode.got error :"+string(response.Payload))
	}
	
	
	
	tradeArgs :=[][]byte{[]byte("tradeprocess"),[]byte(keyID)}
	response:=stub.InvokeChaincode("CC_Arbiter",queryArgs,"testch1")
	if response.Status != shim.OK {
		return shim.Error("failed to query chaincode.got error :"+string(response.Payload))
	}
	fmt.Printf(string(response.Payload))
	

/*
	//compare coin and pay 
	if coin != payment {
		return shim.Error("coin is false")
	}
	
	//update coin and knock in CC_Trade
	keycoin,err := stub.GetState("buyer")
	if err != nil {
		return shim.Error("error getcoin")
	}
	keycoinint,_ := strconv.Atoi(string(keycoin))
	coinint,_:=strconv.Atoi(coin)
	newCoin := keycoinint - coinint
	if newCoin < 0 {
		return shim.Error("coin is not enough")
	}
	stub.PutState("buyer",[]byte(strconv.Itoa(newCoin)))
	fmt.Println(newCoin)
	fmt.Println(keycoinint)
	fmt.Println(coinint)
	
	//updata payment in CC_Verifier
	knockcoin := keyID+"knock"
	queryArgs2 := [][]byte{[]byte("UploadValue"),[]byte(knockcoin),[]byte(coin)}
	response2 := stub.InvokeChaincode("CC_Verifier",queryArgs2,"testch1")
	if response2.Status != shim.OK {
		return shim.Error("failed to query chaincode.got error:"+string(response2.Payload))
	}
	fmt.Printf("knock coin in CC_Trade")
*/	
	return shim.Success([]byte("success trade"))
}


func (t *CC_Buyer) Challenge(stub shim.ChaincodeStubInterface, args []string) pb.Response{


	//验证首次生成的证明，roundID默认为1
	keyID := args[0]
	chaincodeId := args[1]//对哪个模型提出挑战
	round := args[2] //轮次验证

	//parameterUrl := args[2]
	//datasetUrl := args[3]	//xx/xx/xx.py 	具体到文件
	//dataIndex := args[4]	//默认为“0”，如果不为0则使用index的数值
	//指定阈值
	firstKey := keyID + "1"

	firstverifyArgs :=[][]byte{[]byte("verifyproof"),[]byte(firstKey)}
	firstverifyResponse:=stub.InvokeChaincode("CC_Arbiter",firstverifyArgs,"testch1")
	if firstverifyResponse.Status != shim.OK {
		return shim.Error("failed to query chaincode.got error :"+string(firstverifyResponse.Payload))
	}
	fmt.Printf("verify first proof in CC_Arbiter")
	
	
	
	//从Arbiter的public链上获取hash数值，存档
	hashName := keyID + "hash"
	hashArgs := [][]byte{[]byte("GetValue"),[]byte(hashName)}
	hashResponse:=stub.InvokeChaincode("CC_Arbiter",hashArgs,"testch1")
	if hashResponse.Status != shim.OK {
		return shim.Error("failed to query chaincode.got error :"+string(hashResponse.Payload))
	}
	stub.PutState(hashName,hashResponse.Payload)
	
	strconv.Itoa(i)
	//首次验证服务正确性后提出多次验证挑战
	roundnum , _ := strconv.Atoi(round)
	
	if roundnum <= 1{
		return shim.Error("round < 1 ,error.")
	}else{
		for i:=0;i<roundnum;i++{
			roundID:=i+1
			challengeArgs := [][]byte{[]byte("challenge"),[]byte(keyID),[]byte(strconv.Itoa(roundID))}
		challengeResponse:=stub.InvokeChaincode(chaincodeId,challengeArgs,"testch1")
			if challengeResponse.Status != shim.OK {
				return shim.Error("failed to query chaincode.got error :"+string(challengeResponse.Payload))
			}
	
	
		} 
	
	}
	

	return shim.Success([]byte("success challenge"))
}
func (t *CC_Buyer) ChallengeCheck(stub shim.ChaincodeStubInterface, args []string) pb.Response{
//验证之前的多轮验证
//验证成员证明
	keyID := args[0]
	chaincodeId := args[1]//对哪个模型提出挑战
	round := args[2] //轮次验证

	//多轮验证
	roundnum , _ := strconv.Atoi(round)
	if roundnum <= 1{
		return shim.Error("round < 1 ,error")
	}else{
		for i:=0;i<roundnum;i++{
			roundID := i+1
			roundkey := keyID + strconv.Itoa(roundID)
			verifyArgs := [][]byte{[]byte("verifyproof"),[]byte(roundKey)}
			verifyResponse:=stub.InvokeChaincode(chaincodeId,verifyArgs,"testch1")
			if verifyResponse.Status != shim.OK {
				return shim.Error("failed to query chaincode.got error :"+string(verifyResponse.Payload))
			}
		} 
	}
	fmt.Printf("verify all round proof in CC_Arbiter")

	//验证环境hash
	//get hash
	hashName := key + "hash"
	hashArgs := [][]byte{[]byte("GetValue"),[]byte(hashName)}
	hashResponse:=stub.InvokeChaincode("CC_Arbiter",hashArgs,"testch1")
	if hashResponse.Status != shim.OK {
		return shim.Error("failed to query chaincode.got  error :"+string(hashResponse.Payload))
	}
	fmt.Printf(string(hashResponse.Payload))
	fmt.Printf("get hash in CC_Arbiter")
	
	//hashcheck
	//verify hash and parms
	hashcheckArgs :=[][]byte{[]byte("verifyhash"),[]byte(keyID),[]byte(chaincodeId),[]byte(hashResponse.Payload)}
	hashcheckResponse:=stub.InvokeChaincode("CC_Arbiter",hashcheckArgs ,"testch1")
	if hashcheckResponse.Status != shim.OK {
		return shim.Error("invoke chaincode failed:"+string(hashcheckResponse.Status))
	}
	fmt.Printf(string(hashcheckResponse.Payload))
	fmt.Printf("verify hash in CC_Arbiter")
	return shim.Success([]byte("challenge success"))

}

func (t *CC_Buyer) UploadValue(stub shim.ChaincodeStubInterface, args []string) pb.Response{	
	key := args[0]
	value := args[1]
	stub.PutState(key,[]byte(value))
	fmt.Println(key)
	fmt.Println(value)
	return shim.Success([]byte("success put " + key + " "+value))
}
func (t *CC_Buyer) GetValue(stub shim.ChaincodeStubInterface, args []string) pb.Response{	
	key := args[0]
	keyvalue,err := stub.GetState(key)
	if err != nil {
		return shim.Error("error getcoin")
	}
	fmt.Println(key)
	fmt.Println(string(keyvalue))
	return shim.Success([]byte(keyvalue))
}


func main() {
	if err := shim.Start(new(CC_Buyer)); err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
