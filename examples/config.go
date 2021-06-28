package main

import (
	"github.com/Kotodian/naconfig"
	"log"
	"net/url"
	"reflect"
	"time"
)

func failOnError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func assert(actual interface{}, expected interface{}) {
	if !reflect.DeepEqual(actual, expected) {
		log.Fatalf("expected: %v, actual: %v", expected, actual)
	}
}

type AcOcppTest struct {
	Name string `json:"name"`
}

func main() {
	u := url.URL{Host: "139.198.171.184:8848"}
	config, err := naconfig.NewConfig(naconfig.NewJsonCodec(), "dev", 5000, u)
	failOnError(err)

	// 创建配置文件
	err = config.Create("ac-ocpp", &AcOcppTest{Name: "lqk"})
	failOnError(err)
	// 读取配置文件
	acOcppTest := &AcOcppTest{}
	err = config.Get("ac-ocpp", acOcppTest)
	failOnError(err)

	assert(acOcppTest, &AcOcppTest{Name: "lqk"})
	// 监听配置文件
	acOcppTest2 := &AcOcppTest{}
	err = config.Watch("ac-ocpp", naconfig.DefaultWrapOnChange("dev", "ac-ocpp", config.Codec(), acOcppTest2))
	failOnError(err)

	err = config.Create("ac-ocpp", &AcOcppTest{Name: "lqk2"})
	failOnError(err)

	time.Sleep(5 * time.Second)
	assert(acOcppTest2, &AcOcppTest{Name: "lqk2"})

	// 删除配置文件
	err = config.Delete("ac-ocpp")
	failOnError(err)
}
