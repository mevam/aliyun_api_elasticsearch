package main

import (
	"gopkg.in/olivere/elastic.v3"
	"github.com/denverdino/aliyungo/util"
	"github.com/Unknwon/goconfig"
	"fmt"
	"github.com/denverdino/aliyungo/ecs"
	"time"
	"log"
	"strings"
)



// 阿里云token
const ACCESS_KEY_ID = "6mCH4OKZV18PM7oj"
const ACCESS_KEY_SECRET = "2znZwyKOSVyDmQG8MA8UY6CDW7prTs"

type InstanceMonitorDataType struct {
	InstanceId        string
	CPU               int
	IP 		  string
	IntranetRX        int
	IntranetTX        int
	IntranetBandwidth int
	InternetRX        int
	InternetTX        int
	InternetBandwidth int
	IOPSRead          int
	IOPSWrite         int
	BPSRead           int
	BPSWrite          int
	TimeStamp         util.ISO6801Time
}

//读取conf.ini配置文件

func GetEcsInformation() ([]string, []string){
	cfg, err := goconfig.LoadConfigFile("/home/workspace/aliyun_api/elasticsearch/conf.ini")
	if err != nil {
		log.Println("读取配置文件失败[conf.ini]")

	}
	InstanceId, _ := cfg.GetValue("EcsInformation", "InstanceId")
	IP, _ := cfg.GetValue("EcsInformation", "IP")

	getInstanceId := strings.Split(InstanceId, ",")
	getIP := strings.Split(IP, ",")
	return getInstanceId, getIP
}


//以instanceId查询一小时内监控数据

func GetMonitorInformaiton(instanceId string) (monitorData []ecs.InstanceMonitorDataType, err error){
	client := ecs.NewClient(ACCESS_KEY_ID, ACCESS_KEY_SECRET)
	request := new(ecs.DescribeInstanceMonitorDataArgs)
	timesatrt := time.Now().Add(-10*time.Hour).Format("2006-01-02 15:04:05")
	timeend := time.Now().Add(-9*time.Hour).Format("2006-01-02 15:04:05")
	request.InstanceId = instanceId
	startTime, _ := time.Parse("2006-01-02 15:04:05", timesatrt)
	endTime, _ := time.Parse("2006-01-02 15:04:05", timeend)
	request.StartTime = util.NewISO6801Time(startTime)
	request.EndTime = util.NewISO6801Time(endTime)
	array, err := client.DescribeInstanceMonitorData(request)
	return array, err

}

//初始化elasticsearch连接，并查询索引，没有就新建
func creatindex() (err error){
	client , err := elastic.NewClient(elastic.SetURL("http://10.30.0.32:9200"))
	if err != nil {
		panic(err)
	}

	indexname := "monitor-"+time.Now().Format("2006.01.02")
	logtime := time.Now().Format("2006-01-02 15:04:05")
	searchResult, err := client.Search().
	Index(indexname).   // search in index "twitter"
	Do()                // execute
	if err != nil {
		//panic(err)
		_, err = client.CreateIndex(indexname).Do()
		if err != nil {
			// Handle error
			panic(err)
		}
	fmt.Print(logtime, indexname, "	索引创建成功\n")
	}else {
		fmt.Print(logtime)
		fmt.Printf("	Query took %d milliseconds\n", searchResult.TookInMillis)
	}
	return err
}


//往索引内插入数据
/*func AddIndexDcoument(instanceMonitorDataType interface{}) (err error){
	client , err := elastic.NewClient(elastic.SetURL("http://10.30.0.32:9200"))
	if err != nil {
		panic(err)
	}
	indexname := "monitor-"+time.Now().Format("2006.01.02")
	_, err = client.Index().
	Index(indexname).
	Type("monitor").
	BodyJson(instanceMonitorDataType).
	Do()
	if err != nil {
		// Handle error
		panic(err)
	}
	return err
}*/

//遍历配置文件中的instanceId，得到监控数据
func AddMonitorDcoument() (err error){
	getInstanceId, getInstanceIP := GetEcsInformation()
	logtime := time.Now().Format("2006-01-02 15:04:05")

	client , err := elastic.NewClient(elastic.SetURL("http://10.30.0.32:9200"))
	if err != nil {
		panic(err)
	}
	indexname := "monitor-"+time.Now().Format("2006.01.02")



	for i :=0 ; i < len(getInstanceId); i++ {
		output, _ := GetMonitorInformaiton(getInstanceId[i])
		for m := 0; m < len(output); m++ {
			instanceMonitorDataType := InstanceMonitorDataType{InstanceId:output[m].InstanceId,
				CPU:output[m].CPU,
				IntranetRX:output[m].InternetRX,
				IntranetTX:output[m].IntranetTX,
				IntranetBandwidth:output[m].IntranetBandwidth,
				InternetRX:output[m].InternetRX,
				InternetTX:output[m].InternetTX,
				InternetBandwidth:output[m].InternetBandwidth,
				IOPSRead:output[m].IOPSRead,
				IOPSWrite:output[m].IOPSWrite,
				BPSRead:output[m].BPSRead,
				BPSWrite:output[m].BPSWrite,
				IP:getInstanceIP[i],
				TimeStamp:output[m].TimeStamp}

			//往索引内插入数据
			_, err = client.Index().
			Index(indexname).
			Type("monitor").
			BodyJson(instanceMonitorDataType).
			Do()
			if err != nil {
				// Handle error
				panic(err)
			}


		}
	}
	fmt.Print(logtime, "	数据插入成功\n")
	return err
}

func main() {
	logtime := time.Now().Format("2006-01-02 15:04:05")

	err:=creatindex()
	if err != nil {
		fmt.Print(logtime, "	创建索引失败\n")
	}

	error:=AddMonitorDcoument()
	if error != nil {
		fmt.Print(logtime, "	传输数据失败\n")
	}


}
