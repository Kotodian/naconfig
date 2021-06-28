package naconfig

import (
	"encoding/json"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"gopkg.in/yaml.v2"
	"net/url"
	"strconv"
)

type Config struct {
	client config_client.IConfigClient
	codec  Codec
}

type Codec interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

type OnChange func(namespace, group, dataId, data string)

func DefaultWrapOnChange(ns, dId string, codec Codec, v interface{}) OnChange {
	return func(namespace, group, dataId, data string) {
		if ns == namespace && dId == dataId {
			_ = codec.Unmarshal([]byte(data), v)
		}
	}
}

type jsonCodec struct {
}

func NewJsonCodec() Codec {
	return &jsonCodec{}
}

func (j *jsonCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (j *jsonCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

type yamlCodec struct {
}

func NewYamlCodec() Codec {
	return &yamlCodec{}
}
func (y *yamlCodec) Marshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

func (y *yamlCodec) Unmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}

func NewConfig(codec Codec, namespace string, timeout uint64, urls ...url.URL) (*Config, error) {
	if len(urls) == 0 {
		return nil, nil
	}
	clientConfig := *constant.NewClientConfig(
		constant.WithNamespaceId(namespace),
		constant.WithTimeoutMs(timeout),
	)
	serverConfigs := make([]constant.ServerConfig, 0)
	for _, u := range urls {
		port, _ := strconv.ParseUint(u.Port(), 10, 64)
		serverConfig := *constant.NewServerConfig(u.Hostname(), port)
		serverConfigs = append(serverConfigs, serverConfig)
	}
	client, err := clients.CreateConfigClient(map[string]interface{}{
		constant.KEY_SERVER_CONFIGS: serverConfigs,
		constant.KEY_CLIENT_CONFIG:  clientConfig,
	})
	if err != nil {
		return nil, err
	}
	config := &Config{
		client: client,
		codec:  codec,
	}
	return config, nil
}

func (c *Config) Codec() Codec {
	return c.codec
}

func (c *Config) Create(dataId string, data interface{}) error {
	content, err := c.codec.Marshal(data)
	if err != nil {
		return err
	}
	param := vo.ConfigParam{
		DataId:  dataId,
		Content: string(content),
		Group:   constant.DEFAULT_GROUP,
	}
	_, err = c.client.PublishConfig(param)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) Get(dataId string, v interface{}) error {
	param := vo.ConfigParam{DataId: dataId, Group: constant.DEFAULT_GROUP}
	config, err := c.client.GetConfig(param)
	if err != nil {
		return err
	}
	return c.codec.Unmarshal([]byte(config), v)
}

func (c *Config) Watch(dataId string, onChange OnChange) error {
	param := vo.ConfigParam{
		DataId:   dataId,
		OnChange: onChange,
		Group:    constant.DEFAULT_GROUP,
	}
	err := c.client.ListenConfig(param)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) CancelWatch(dataId string) error {
	param := vo.ConfigParam{DataId: dataId, Group: constant.DEFAULT_GROUP}
	err := c.client.CancelListenConfig(param)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) Delete(dataId string) error {
	param := vo.ConfigParam{DataId: dataId, Group: constant.DEFAULT_GROUP}
	_, err := c.client.DeleteConfig(param)
	if err != nil {
		return err
	}
	return nil
}
