package common

import (
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/aliyun/aliyun-log-go-sdk/producer"
	"github.com/s3cu1n4/logs/logs"
)

func Sendlog(loglist []map[string]string) {
	var logList []*sls.Log
	producerConfig := producer.GetDefaultProducerConfig()
	producerConfig.Endpoint = Conf.Aliyun.Endpoint
	producerConfig.AccessKeyID = Conf.Aliyun.Ak
	producerConfig.AccessKeySecret = Conf.Aliyun.Sk
	producerInstance := producer.InitProducer(producerConfig)
	producerInstance.Start()
	for k := range loglist {
		aliyunlog := producer.GenerateLog(uint32(time.Now().Unix()), loglist[k])
		logList = append(logList, aliyunlog)
		go sendList(producerInstance, Conf.Aliyun.SlsProject, Conf.Aliyun.Logstore, logList)
	}
}

func sendList(producerInstance *producer.Producer, project, logstore string, logList []*sls.Log) {
	err := producerInstance.SendLogList(project, logstore, "", "", logList)
	if err != nil {
		logs.Error("sendlog err:", err.Error)
	}
}
