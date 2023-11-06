package main

import (


	"io"
	"os"
	"strconv"
	"time"
	"math/rand"
	
	"fmt"
	"log"
	
	"os/exec"
	"bytes"

	"io/ioutil"
	"encoding/json"
	"crypto/sha256"
	"crypto/aes"
	"encoding/hex"
	"bufio"
	"encoding/json"
	"encoding/csv"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/xingzwh/gnark/backend/groth16"
	_"github.com/xingzwh/gnark/backend/witness"

	//"github.com/xingzwh/gnark/frontend/cs/scs"
	"github.com/xingzwh/gnark/frontend/cs/r1cs"
	ccgroth16 "github.com/xingzwh/gnark/curvepp/backend/bn254/groth16"
	//"io/ioutil"
	_"encoding/json"
	"github.com/xingzwh/gnark/frontend"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
    	mt "github.com/txaty/go-merkletree"
)

//circuit define
const (
	samplenum = 10
	degree = 13
	datasetnum = 506
)

type LinearCircuit struct {
	X [samplenum][degree + 1]frontend.Variable `gnark:",public"`	//sample
	L [samplenum]frontend.Variable `gnark:",public"`		//sample label	
	Avgy frontend.Variable `gnark:",public"`			//avg of sample label
	T frontend.Variable `gnark:",public"`				//threshold
	Out frontend.Variable `gnark:",public"`			//output
	W [degree + 1]frontend.Variable `gnark:"w"`			//parameters of LR
}


func (circuit *LinearCircuit) Define(api frontend.API) error {
	ssr:=api.Add(0,0)		//sum (predicty - avgy)^2
	sst:=api.Add(0,0)		//sum (label - avgy)^2
	for i := 0; i < samplenum; i++ {
		//compute predicty 
		predicty:=api.Add(0,0)
		for j := 0; j < degree + 1 ; j++ {
			predicty = api.Add(predicty,api.Mul(circuit.X[i][j],circuit.W[j]))
		}
		//api.Println(predicty)
		//api.Println(circuit.L[i])
		//compute ssr
		mid_ssr:=api.Sub(predicty,api.Mul(circuit.Avgy,1000))
		api.Println(predicty)
		api.Println(mid_ssr)
		ssr = api.Add(ssr,api.Mul(mid_ssr,mid_ssr))
		//api.Println(api.Mul(mid_ssr,mid_ssr))
		//api.Println(ssr)
		//compute sst
		mid_sst:=api.Mul(api.Sub(circuit.L[i],circuit.Avgy),1000)
		//api.Println(mid_sst)	
		sst = api.Add(sst,api.Mul(mid_sst,mid_sst))
	}
	//SST>T.SSR
	//cmp:=api.Select(api.AssertIsLessOrEqual(api.Mul(ssr,circuit.T),api.Mul(sst,100)),1,0)
	
	//api.Println(sst)
	//api.Println(ssr)
	cmp:=api.Cmp(api.Mul(sst,api.Sub(100,circuit.T)),ssr)
	api.AssertIsEqual(circuit.Out,cmp)
	
	
	return nil
}
func GetHash(path string) (hash string) {
	file, _ := os.Open(path)
	h_ob := sha256.New()
	_, err := io.Copy(h_ob, file)
	var hashvalue string
	if err == nil {
		hash := h_ob.Sum(nil)
		hashvalue = hex.EncodeToString(hash)
	} else {
		log.Fatal(err)
	}
	return hashvalue
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
func ReadDataCSV(datapath string) ([][]int,[]int) {
	datafile,err:=os.Open(datapath)
	if err != nil{
		fmt.Println("error %s",err)
	}
	defer datafile.Close()
	
	var LineX [][]int
	var LineY []int
	reader := csv.NewReader(datafile)
	
	//生成不重复随机数组
	rand.Seed(time.Now().UnixNano())
	var randnum [samplenum]int 
	for i:=0;i<datasetnum;i++ {
		num := rand.Intn(506)
		for j:=0;j<i;j++ {
			if randnum[j] == num {
				randnum[j] = rand.Intn(506)
				j--
			}
		}
		randnum[i] = num 
	}
	fmt.Println(randnum)
	//for i:=0;i<samplenum;i++ {
	//var line []string
	for i:=0 ; ; i++{	
		var result bool
		//fmt.Println(result)
		line,err := reader.Read()
		for index:=0 ; index < samplenum ; index++{
			if  i == randnum[index]{
				result = true
			}
		}
		if result {
		
		
			fmt.Println(line)
			//fmt.Println(len(line))
			label := line[degree]


			labelfloat,_:=strconv.ParseFloat(label,64)
			labelvalue := int(labelfloat*1000)

			data:=line[0:degree]
			var dataint []int
			for j:=0;j<len(data);j++{
				newdata,_:=strconv.ParseFloat(data[j],64)
				datat := int(newdata*1000)
				dataint= append(dataint,datat)
			}
			LineY = append(LineY, labelvalue)
			LineX = append(LineX, dataint)		
		}
		if err == io.EOF{
			break
		}else if err!=nil{
			fmt.Println("error %s",err)
		}	
		
		
	}	
	return LineX,LineY

}
func ReadParameterCSV(pmpath string) ([]int) {

	parameterfile,err:=os.Open(pmpath)
	if err != nil{
		fmt.Println("error %s",err)
	}
	defer parameterfile.Close()
	reader := csv.NewReader(parameterfile)
	
	line,err := reader.Read()
	if err == io.EOF{
		fmt.Println("finish")
	}else if err!=nil{
		fmt.Println("error %s",err)
	}
	pm:=line[:]
	var pmint []int
	for j:=0;j<len(pm);j++{
		pmfloat,_:=strconv.ParseFloat(pm[j],64)
		pmtoint := int(pmfloat*1000)
		pmint = append(pmint,pmtoint)
	}
 	//fmt.Println(pmint)
 	return pmint
}

func WriteTo(filepath string,data []byte) {
	
    	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0666)
    	if err != nil {
    	    fmt.Println("文件打开失败", err)
   	 }
    	//及时关闭file句柄
    	defer file.Close()
    	//写入文件时，使用带缓存的 *Writer
   	write := bufio.NewWriter(file)
        write.WriteString(string(data))
    	//Flush将缓存的文件真正写入到文件中
    	write.Flush()
}


type CC_Prover_LR_1 struct {
}

func (t *CC_Prover_LR_1) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success([]byte("successful init"))
}

func (t *CC_Prover_LR_1) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fn, args := stub.GetFunctionAndParameters()	
	if fn == "uploadvalue" {
		return t.UploadValue(stub, args)				
	}else if fn == "getvalue"{
		return t.GetValue(stub, args)		
	}else if fn == "reproduce"{
		return t.Reproduce(stub, args)		
	}else if fn == "setupproof"{
		return t.LRGroth16Setup(stub, args)		
	}else if fn == "challenge"{
		return t.Challenge(stub, args)		
	}else{
		return shim.Error("no such opration")
	}	
}
func (t *CC_Prover_LR_1) GetValue(stub shim.ChaincodeStubInterface, args []string) pb.Response{	

	var keyID = args[0]
	keyvalue,err:=stub.GetState(keyID)
	if err != nil {
		return shim.Error("error getvalue")
	}
	fmt.Println(keyID)
	fmt.Println(keyvalue)
	return shim.Success([]byte(keyvalue))
}
func (t *CC_Prover_LR_1) UploadValue(stub shim.ChaincodeStubInterface, args []string) pb.Response{	
	var keyID = args[0]
	var keyvalue = args[1]
	stub.PutState(keyID,[]byte(keyvalue))
	fmt.Println(keyID)
	fmt.Println(keyvalue)
	return shim.Success([]byte("success put " + keyID + " " + keyvalue))
}

func (t *CC_Prover_LR_1) Reproduce(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	keyID = args[0]
	pyUrl := args[1]      ///xx/xx/xx.py 	具体到文件,py文件路径
	dataUrl = args[2]    ///xx/xx/xx.csv	具体到文件，数据文件路径
	envUrl = args[3]    ///xx/xx/	文件所属文件夹，用于进行hash保存
	parameterUrl = args[4]    ///xx/newxx/  参数存储路径
	
	envpathName := keyID + "envpath"
	parampathName := keyID + "parampath"
	stub.PutState(pypathName,[]byte(envUrl))
	stub.PutState(parampathName,[]byte(parameterUrl))
	//运行完本部分，应该在container对应位置下产生相关的参数文件
	// 分类后的样本与原数据集一起，参数文件另外保存
    	cmd := exec.Command("python3", pyUrl , "10" , dataUrl,envUrl,parameterUrl)
    	
    	var stdout, stderr bytes.Buffer
    	cmd.Stdout = &stdout  // 标准输出
   	cmd.Stderr = &stderr  // 标准错误
   	
    	err := cmd.Run()
    	
    	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
    	fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)
    	if err != nil {
    	    log.Fatalf("cmd.Run() failed with %s\n", err)
    	}	
    	
    	
    	//压缩文件获取整个环境的hashvalue,包含数据集，python文件
	var hashlist []string
	envfiles, _ := ioutil.ReadDir(envUrl)
   	for _ , f1 := range envfiles {
            fmt.Println(f1.Name())
            hashpath := envUrl+f1.Name()
            hashnum := GetHash(hashpath)
            hashlist =append(hashlist,hashnum)     
    	}
	//fmt.Println(hashlist)
	
	hashbyte,_:= json.Marshal(hashlist)
	hashenv := sha256.Sum256(hashbyte)
	hashValue := hex.EncodeToString(hashenv[:])
	fmt.Println(hashValue)
	//upload hashvalue to CC_Arbiter
    	//此部分hash应公开
    	hashName := keyID + "hash"
	hashArgs :=[][]byte{[]byte("uploadvalue"),[]byte(hashName),[]byte(hashValue)}
	hashResponse:=stub.InvokeChaincode("CC_Arbiter",hashArgs,"testch1")
	if hashResponse.Status != shim.OK {
		return shim.Error("failed to query chaincode.got error :"+string(hashResponse.Payload))
	}
	fmt.Printf("upload hashvalue to CC_Seller")
	
	//computehashw
	var parameterlist []string
	parameterfiles, _ := ioutil.ReadDir(parameterUrl) 
   	for i , f2 := range parameterUrl{
            fmt.Println(f2.Name())
            parameterpath := parameterUrl+f2.Name()
            parametertext,err := ioutil.ReadFile(parameterpath)
            parameterlist =append(parameterlist,string(parametertext))    
           
    	}
	//fmt.Println(parameterlist)
	parameterbyte,_:= json.Marshal(parameterlist)
	parameterenv := sha256.Sum256(parameterbyte)
	parameterValue := hex.EncodeToString(parameterenv[:])
	fmt.Println(parameterValue)
	
	
	
    	//upload parameterhashvalue to CC_Arbiter
    	//此部分hash应公开
    	parameterhashName := keyID + "parameterhash"
	parameterhashArgs :=[][]byte{[]byte("uploadvalue"),[]byte(parameterhashName),[]byte(parameterValue)}
	parameterhashResponse:=stub.InvokeChaincode("CC_Arbiter",parameterhashArgs,"testch1")
	if parameterhashResponse.Status != shim.OK {
		return shim.Error("failed to query chaincode.got error :"+string(parameterhashResponse.Payload))
	}
	fmt.Printf("upload parameterhashvalue to CC_Arbiter")
    	
    	//generate SecretKey by Arbiter
    	skName := keyID + "secretkey"
	getSKeyArgs :=[][]byte{[]byte("generatesk"),[]byte(skName)}
	getSKeyResponse:=stub.InvokeChaincode("CC_Arbiter",getSKeyArgs,"testch1")
	if getSKeyResponse.Status != shim.OK {
		return shim.Error("failed to query chaincode.got error :"+string(getSKeyResponse.Payload))
	}
    	SecretKey := getSKeyResponse.Payload
    	
    	
    	//使用k加密w，然后将w上传至私有数据库，将z传至共有链 	
    	parameterfiles, _ := ioutil.ReadDir(parameterUrl)
    	var enum int
   	for i , f2 := range parameterfiles {
            fmt.Println(f2.Name())
            filepath := parameterUrl+f2.Name()
            text,err := ioutil.ReadFile(filepath) //type(text) is []byte
            if err != nil {
            	log.Fatal(err)
            }
            encrypt, err := EncryptAES(string(SecretKey), string(text))
            if err != nil {
            	log.Fatal(err)
            }
            fmt.Println(encrypt)
            
            //z在Arbiter的public链上
            encryptName := keyID + strconv.Itoa(i) + "encryptValue"
            encryptArgs :=[][]byte{[]byte("uploadvalue"),[]byte(encryptName),[]byte(encrypt)}
            encryptResponse:=stub.InvokeChaincode("CC_Arbiter",encryptArgs,"testch1")
            if encryptResponse.Status != shim.OK {
            	return shim.Error("failed to query chaincode.got error :"+string(encryptResponse.Payload))
            }

            //w Arbiter的privacy链上
            /*
            parameterName := keyID + f2.Name() + "originValue"
            parameterArgs :=[][]byte{[]byte("uploadprivatevalue"),[]byte(parameterName),[]byte(text)}
            parameterResponse:=stub.InvokeChaincode("CC_Arbiter",parameterArgs,"testch1")
            if parameterResponse.Status != shim.OK {
            	return shim.Error("failed to query chaincode.got error :"+string(parameterResponse.Payload))
            }    
            */
            enum = i
                    
    	}	
    	ennumName := keyID + "encryptNum"
    	ennumArgs :=[][]byte{[]byte("uploadvalue"),[]byte(ennumName),[]byte(strconv.Itoa(enum))}
    	ennumResponse:=stub.InvokeChaincode("CC_Arbiter",ennumArgs,"testch1")
    	if ennumResponse.Status != shim.OK {
    		return shim.Error("failed to query chaincode.got error :"+string(ennumResponse.Payload))
	}
	return shim.Success([]byte("success reproduce training"))
	//运行完本函数，应该在container对应位置下产生相关的参数文件
}


func (t *CC_Prover_LR_1) LRGroth16Setup(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	key = args[0]
	keyID:=key+"1"//本次证明生成由Seller发起，默认round=1
	datasetUrl := args[1]      ///xx/xx/xx.csv	具体到文件
	//具体到路径，由于保存参数的文件不一样，所以文件在Prover链玛中指定
	parameterUrl := args[2] 	
	thrData := args[3] 
	dataIndex := args[4]
	

 	LineX,LineY := ReadDataCSV(datasetUrl)
 	wfilepath := parameterUrl + "/linear.weight.csv"
 	wdata := ReadParameterCSV(wfilepath)
 	bfilepath := parameterUrl + "/linear.bias.csv"
 	bdata := ReadParameterCSV(bfilepath)
 	//fmt.Println(len(LineX))
 	//fmt.Println(len(LineX[0]))
 	//fmt.Println(len(LineY))
	sum := 0
	for _, val := range LineY {
		//累计求和
		sum += val
	}
	avgy_y := int(sum/len(LineY))
		// compiles our circuit into a R1CS
	start_init := time.Now()
	
	//从dataindex中获取数据index
	var numbers []int
	err := json.Unmarshal([]byte(dataIndex),&numbers)
	if err != nil {
		fmt.Println("dataindex Unmarshal error")
	}	
	fmt.Println("setupproof type is 1.The index is ",numbers)	
 
	//从lineX，lineY中获取数值
	var newLineX [][]int
	var newLineY []int
     	for _,j := range numbers {
     		newLineX = append(newLineX, LineX[j])
     		newLineY = append(newLineY, LineY[j])
     	} 
	
	var circuit LinearCircuit
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	
	if err != nil {
		fmt.Println("circuit compilation error")
	}
	cost_init := time.Since(start_init)
	fmt.Println("init time = ",cost_init)
	

	
	// witness definition
	var assignment LinearCircuit
	//get witness	X [labelnum][degree] W [degree] Y [labelnum]
	
	for i := 0; i < samplenum; i++{
		for j := 0; j < degree ; j++{
			
			assignment.X[i][j] = newLineX[i][j]
		}
		assignment.X[i][degree] = 1
		assignment.L[i] = newLineY[i]
	}
 
	
	for i := 0; i < degree ; i++{
		assignment.W[i] = wdata[i]
	}
	assignment.W[degree] = bdata[0]
	assignment.Avgy = avgy_y
	assignment.T = strconv.Itoa(thrData)
	assignment.Out = 1
	
	//start_set := time.Now()	
	// groth16 zkSNARK: Setup
	pk, vk, err := groth16.Setup(ccs)
	if err != nil {
		log.Fatal(err)
	}
	//cost_set := time.Since(start_set)
	//fmt.Println("Setup time = ",cost_set)
	
	
	//start_proof := time.Now()
	lrwitness, err := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())	
	publicWitness, _ := lrwitness.Public()

	// groth16: Prove & Verify
	proof, err := groth16.Prove(ccs, pk, lrwitness)
	if err != nil {
		log.Fatal(err)
	}
	//cost_proof := time.Since(start_proof)
	//fmt.Println("proof time  = ",cost_proof)
	
	
	//upload proof vk publicWitness to Arbiter
	//proof
	data1,_:= json.Marshal(proof)
	proofName:="lr"+keyID+"proof"
	proofArgs :=[][]byte{[]byte("uploadvalue"),[]byte(proofName),data1}
	proofResponse:=stub.InvokeChaincode("CC_Arbiter",proofArgs,"testch1")
	if proofResponse.Status != shim.OK {
		return shim.Error("failed to query chaincode.got error :"+string(proofResponse.Payload))
	}
	fmt.Printf("upload proof to CC_Arbiter")
	
	//vk
	data2,_:= json.Marshal(vk)	
	vkName:="lr"+keyID+"vk"
	vkArgs :=[][]byte{[]byte("uploadvalue"),[]byte(vkName),data2}
	vkResponse:=stub.InvokeChaincode("CC_Arbiter",vkArgs,"testch1")
	if vkResponse.Status != shim.OK {
		return shim.Error("failed to query chaincode.got error :"+string(vkResponse.Payload))
	}
	fmt.Printf("upload proof to CC_Arbiter")
	
	//witness
	marshal:=publicWitness.MarshalJSON
	data3,err:=marshal()
	if err != nil {
		return shim.Error("Error marshal pw")
	}	
	pwName:="lr"+keyID+"pw"
	pwArgs :=[][]byte{[]byte("uploadvalue"),[]byte(pwName),data3}
	pwResponse:=stub.InvokeChaincode("CC_Arbiter",pwArgs,"testch1")
	if pwResponse.Status != shim.OK {
		return shim.Error("failed to query chaincode.got error :"+string(pwResponse.Payload))
	}
	fmt.Printf("upload proof to CC_Arbiter")
	
	//pw.scc
	var schemacc *schema.Schema
	schemacc = publicWitness.Schema
	data4,_:= json.Marshal(schemacc)
	sccName:="lr"+keyID+"scc"
	sccArgs :=[][]byte{[]byte("uploadvalue"),[]byte(sccName),data4}
	sccResponse:=stub.InvokeChaincode("CC_Arbiter",pwArgs,"testch1")
	if sccResponse.Status != shim.OK {
		return shim.Error("failed to query chaincode.got error :"+string(sccResponse.Payload))
	}
	fmt.Printf("upload proof to CC_Arbiter")
	
	//pw.curve
	var curvecc ecc.ID
	curvecc = publicWitness.CurveID
	data5,_:= json.Marshal(curvecc)
	fmt.Println("curve")
	ecc_name:="lr"+keyID+"ecc"
	eccArgs :=[][]byte{[]byte("uploadvalue"),[]byte(eccName),data5}
	eccResponse:=stub.InvokeChaincode("CC_Arbiter",pwArgs,"testch1")
	if eccResponse.Status != shim.OK {
		return shim.Error("failed to query chaincode.got error :"+string(eccResponse.Payload))
	}
	fmt.Printf("upload proof to CC_Arbiter")
	
	
	return shim.Success([]byte("success generate upload proof vk and pw"))
}
type testData struct {
    data []byte
}

func (q *testData) Serialize() ([]byte, error) {
    return q.data, nil
}

// generate dummy data blocks
func getBlocks(newLineX [][]int) (blocks []mt.DataBlock) {
    for i := 0; i < len(newLineX); i++ {
    	xdata , _ := json.Marshal(newLineX[i])
    	block := &testData{
    		data: xdata,
    	}
        //_, err := rand.Read(block.data)
        //handleError(err)
        blocks = append(blocks, block)
    }
    return
}

func (t *CC_Prover_LR_1) Challenge(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	key = args[0]
	roundID = args[1]//本次证明生成由Buyer发起，默认round>1
	keyID:=key + roundID
	
	//get datapath
	pypathName := keyID + "pypath"
	dataUrl,err := stub.GetState(pypathName)
	if err != nil {
		return shim.Error("Error get data")
	}
	datasetUrl := dataUrl+"/positive.csv"
	
	parampathName := keyID + "parampath"
	paramUrl,err := stub.GetState(pypathName)
	if err != nil {
		return shim.Error("Error get data")
	}
	parameterUrl := string(paramUrl)
	
	//get parampath
	datasetUrl := args[1]      ///xx/xx/xx.py 	具体到文件
	//具体到路径，由于保存参数的文件不一样，所以文件在Prover链玛中指定
	parameterUrl := args[2] 	
	  
	thrName:=keyID + "threshold"	
	thrArgs :=[][]byte{[]byte("getvalue"),[]byte(thrName)}
	thrResponse:=stub.InvokeChaincode("CC_Seller",thrArgs,"testch1")
	if thrResponse.Status != shim.OK {
		return shim.Error("failed to query chaincode.got error :"+string(thrResponse.Payload))
	} 
	thrData := string(thrResponse.Payload)
	
	
	LineX,LineY := ReadDataCSV(datasetUrl)
 	wfilepath := parameterUrl + "/linear.weight.csv"
 	wdata := ReadParameterCSV(wfilepath)
 	bfilepath := parameterUrl + "/linear.bias.csv"
 	bdata := ReadParameterCSV(bfilepath)
 	//fmt.Println(len(LineX))
 	//fmt.Println(len(LineX[0]))
 	//fmt.Println(len(LineY))
	sum := 0
	for _, val := range LineY {
		//累计求和
		sum += val
	}
	avgy_y := int(sum/len(LineY))
		// compiles our circuit into a R1CS
	start_init := time.Now()
	
	//获取样本index,从random中获取数据index
	numbers := []int{}
	for i:=0;i<samplenum;i++ {
		numbers = append(numbers,rand.Intn(datasetnum))
    	}
	fmt.Println("setupproof type is 0.The random index is ",numbers)
	numberName := keyID + "numbers"	
	numbersData,_:= json.Marshal(numbers)
	stub.PutState(numberName,numbersData)
 
	//从lineX，lineY中获取数值
	var newLineX [][]int
	var newLineY []int
     	for _,j := range numbers {
     		newLineX = append(newLineX, LineX[j])
     		newLineY = append(newLineY, LineY[j])
     	} 
	
	var circuit LinearCircuit
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	
	if err != nil {
		fmt.Println("circuit compilation error")
	}
	cost_init := time.Since(start_init)
	fmt.Println("init time = ",cost_init)
	

	
	// witness definition
	var assignment LinearCircuit
	//get witness	X [labelnum][degree] W [degree] Y [labelnum]
	
	for i := 0; i < samplenum; i++{
		for j := 0; j < degree ; j++{
			
			assignment.X[i][j] = newLineX[i][j]
		}
		assignment.X[i][degree] = 1
		assignment.L[i] = newLineY[i]
	}
 
	
	for i := 0; i < degree ; i++{
		assignment.W[i] = wdata[i]
	}
	assignment.W[degree] = bdata[0]
	assignment.Avgy = avgy_y
	assignment.T = strconv.Itoa(thrData)
	assignment.Out = 1
	
	//start_set := time.Now()	
	// groth16 zkSNARK: Setup
	pk, vk, err := groth16.Setup(ccs)
	if err != nil {
		log.Fatal(err)
	}
	//cost_set := time.Since(start_set)
	//fmt.Println("Setup time = ",cost_set)
	
	
	//start_proof := time.Now()
	lrwitness, err := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())	
	publicWitness, _ := lrwitness.Public()

	// groth16: Prove & Verify
	proof, err := groth16.Prove(ccs, pk, lrwitness)
	if err != nil {
		log.Fatal(err)
	}
	//cost_proof := time.Since(start_proof)
	//fmt.Println("proof time  = ",cost_proof)
	
	
	//upload proof vk publicWitness
		//proof
	data1,_:= json.Marshal(proof)
	proofName:="lr"+keyID+"proof"
	proofArgs :=[][]byte{[]byte("uploadvalue"),[]byte(proofName),data1}
	proofResponse:=stub.InvokeChaincode("CC_Arbiter",proofArgs,"testch1")
	if proofResponse.Status != shim.OK {
		return shim.Error("failed to query chaincode.got error :"+string(proofResponse.Payload))
	}
	fmt.Printf("upload proof to CC_Arbiter")
	
	//vk
	data2,_:= json.Marshal(vk)	
	vkName:="lr"+keyID+"vk"
	vkArgs :=[][]byte{[]byte("uploadvalue"),[]byte(vkName),data2}
	vkResponse:=stub.InvokeChaincode("CC_Arbiter",vkArgs,"testch1")
	if vkResponse.Status != shim.OK {
		return shim.Error("failed to query chaincode.got error :"+string(vkResponse.Payload))
	}
	fmt.Printf("upload proof to CC_Arbiter")
	
	//witness
	marshal:=publicWitness.MarshalJSON
	data3,err:=marshal()
	if err != nil {
		return shim.Error("Error marshal pw")
	}	
	pwName:="lr"+keyID+"pw"
	pwArgs :=[][]byte{[]byte("uploadvalue"),[]byte(pwName),data3}
	pwResponse:=stub.InvokeChaincode("CC_Arbiter",pwArgs,"testch1")
	if pwResponse.Status != shim.OK {
		return shim.Error("failed to query chaincode.got error :"+string(pwResponse.Payload))
	}
	fmt.Printf("upload proof to CC_Arbiter")
	
	//pw.scc
	var schemacc *schema.Schema
	schemacc = publicWitness.Schema
	data4,_:= json.Marshal(schemacc)
	sccName:="lr"+keyID+"scc"
	sccArgs :=[][]byte{[]byte("uploadvalue"),[]byte(sccName),data4}
	sccResponse:=stub.InvokeChaincode("CC_Arbiter",pwArgs,"testch1")
	if sccResponse.Status != shim.OK {
		return shim.Error("failed to query chaincode.got error :"+string(sccResponse.Payload))
	}
	fmt.Printf("upload proof to CC_Arbiter")
	
	//pw.curve
	var curvecc ecc.ID
	curvecc = publicWitness.CurveID
	data5,_:= json.Marshal(curvecc)
	fmt.Println("curve")
	ecc_name:="lr"+keyID+"ecc"
	eccArgs :=[][]byte{[]byte("uploadvalue"),[]byte(eccName),data5}
	eccResponse:=stub.InvokeChaincode("CC_Arbiter",pwArgs,"testch1")
	if eccResponse.Status != shim.OK {
		return shim.Error("failed to query chaincode.got error :"+string(eccResponse.Payload))
	}
	fmt.Printf("upload proof to CC_Arbiter")
	
	//对数据生成成员证明	
	blocks := getBlocks(LineX)
   	
    	tree, err := mt.New(nil, blocks)
    	if err != nil {
        	panic(err)
   	}

    	//proof upload
    	merkleProofs := tree.Proofs
    	var newmerkleProof []*mt.Proof
    	//var newblock *testData
    	//var newmerkleblock []mt.DataBlock
    	for _,j := range numbers {
        // if hashFunc is nil, use SHA256 by default
        	//newmerkleblock = append(newmerkleblock, blocks[j])
        	newmerkleProof = append(newmerkleProof, merkleProofs[j])
    	}
    	merkleproofName := keyID+"merkleproof"  
    	dataProof,_:= json.Marshal(newmerkleProof)
    	stub.PutState(merkleproofName,dataProof)
    	
    	
    	
    	
    	merklerootHash := tree.Root 
    	//roothash upload
    	roothashName := keyID+"roothash"  
    	stub.PutState(roothashName,merklerootHash)
 
    	//验证
    	/*
    	var reproof []*mt.Proof
    	errproof := json.Unmarshal(dataproof,&reproof)
    	if errproof != nil {
		fmt.Println("errorproof")
    	}
 	
    	for i := 0 ;i < len(numbers);i++  {
    		ok, err := mt.Verify(newblock[i], reproof[i], rootHash, nil)
    		if err != nil {
        		panic(err)
    		}
    		fmt.Println(ok)
    	}
    	cost_verify := time.Since(start_verify)
    	fmt.Println("verify time = ",cost_verify)	
    	*/
}
func (t *CC_Prover_LR_1) HashCheck(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//将已有的hash与prover中的环境hash重新计算
	keyID = args[0]
	
	hashvalue := args[1]
	envpathName := keyID + "envpath"
	parampathName := keyID + "parampath"
	envUrldata,err1 := stub.GetState(envpathName)
	if err1 != nil {
		return shim.Error("Error download vk")
	}
	envUrl := string(envUrldata)
	
	//parameterUrldata,err2 := stub.GetState(parampathName)
	//if err2 != nil {
	//	return shim.Error("Error download vk")
	//}
	//parameterUrl := string(parameterUrldata)
	
	var hashlist []string
	envfiles, _ := ioutil.ReadDir(envUrl)
   	for _ , f1 := range envfiles {
            fmt.Println(f1.Name())
            hashpath := envUrl+f1.Name()
            hashnum := GetHash(hashpath)
            hashlist =append(hashlist,hashnum)     
    	}
	//fmt.Println(hashlist)
	
	hashbyte,_:= json.Marshal(hashlist)
	hashenv := sha256.Sum256(hashbyte)
	hashValue2 := hex.EncodeToString(hashenv[:])
	//fmt.Println(hashValue)
  	if hashvalue != hashValue2 {
  		return shim.Error("env error")
  	}
  	

}









/*
//直接传x,w,y如何生成证明
	//x[][] w[] y[]
	key:=args[0]
	x_parms_str:=args[1]
	w_parms_str:=args[2]
	y_parms_str:=args[3]
	avgy:=args[4]
	threshold:=args[5]
	
	var x_parms [][]int
	errx := json.Unmarshal([]byte(x_parms_str),&x_parms)
	if errx != nil {
		fmt.Println("circuit compilation error")
	}
	
	var w_parms []int
	errw := json.Unmarshal([]byte(w_parms_str),&w_parms)
	if errw != nil {
		fmt.Println("circuit compilation error")
	}
	
	var y_parms []int
	erry := json.Unmarshal([]byte(y_parms_str),&y_parms)
	if erry != nil {
		fmt.Println("circuit compilation error")
	}
	
	var avgy_parms int
	erravgy := json.Unmarshal([]byte(avgy),&avgy_parms)
	if erravgy != nil {
		fmt.Println("circuit compilation error")
	}
	
	var threshold_parms int
	errts := json.Unmarshal([]byte(threshold),&threshold_parms)
	if errts != nil {
		fmt.Println("circuit compilation error")
	}	
		
	var assignment LinearCircuit
	//get witness	X [labelnum][degree] W [degree] Y [labelnum]
	for i := 0; i < labelnum; i++{
		for j := 0; j <degree ; j++{
			assignment.X[i][j] = x_parms[i*degree+j]
		}
	}
	for i := 0; i < degree; i++{
			assignment.W[i] = w_parms[i]
	}
	for i := 0; i < labelnum; i++{
		assignment.Y[i] = y_parms[i]
	}
	assignment.Avgy = avgy_parms
	assignment.T = threshold_parms
	assignment.Out = 1
*/
	
/*
//如果某些部分过大应该如何分解，vk为例
	//过大证明的分解
	vk_num:="lr"+args[3]+"vknum"
	if len(data2) >= 1000000 {
		num:=len(data2)/1000000
		stub.PutState(vk_num,[]byte(strconv.Itoa(num)))
		
		for i:=0 ; i<num ; i++ {
			slice:=data2[i*1000000:(i+1)*1000000]
			vk_name:="lr"+args[3]+"vkdata"+strconv.Itoa(i)
			stub.PutState(vk_name,slice)
		}
		
		slicefinal := data2[num*1000000:len(data2)]
		final_name:="lr"+args[3]+"vkdata"+strconv.Itoa(num)
		stub.PutState(final_name,slicefinal)
		
	}else{
		vk_name:="lr"+args[3]+"vk"
		stub.PutState(vk_num,[]byte("0"))
		stub.PutState(vk_name,data2)
	}

	//stub.PutState(vk_name,data2)
	fmt.Println("put vk")
	*/
	/*
	vk_num:="lr"+name+"vknum"
	vk_numvalue,err2 := stub.GetState(vk_num)
	if err2 != nil {
		return shim.Error("Error download vk")
	}
	vk_datanum,erratoi:=strconv.Atoi(string(vk_numvalue))
	if erratoi != nil {
		return shim.Error("Error vknum erratoi")
	}
	var vk_data []byte
	var err222 error
	if vk_datanum == 0{
		vk_name:="lr"+name+"vk"
		vk_data,err222 = stub.GetState(vk_name)
		if err222 != nil {
			return shim.Error("Error get vk")
		}
	}else{
		for i:=0 ; i<vk_datanum+1; i++ {
			vk_name:="lr"+args[0]+"vkdata"+strconv.Itoa(i)
			num_data,err:=stub.GetState(vk_name)
			if err != nil {
				return shim.Error("Error get vk"+strconv.Itoa(i))
			}
			fmt.Println(num_data)
			for j:=0 ; j<len(num_data) ; j++{
				vk_data=append(vk_data,num_data[j])
			}
		}
		
	}
*/


func main() {
	if err := shim.Start(new(CC_Prover_LR_1)); err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
