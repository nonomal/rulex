package target

import (
	"errors"
	"fmt"
	"time"

	"github.com/i4de/rulex/core"
	"github.com/i4de/rulex/glogger"
	"github.com/i4de/rulex/typex"
	"github.com/i4de/rulex/utils"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

//
type mqttConfig struct {
	Host      string `json:"host" validate:"required"`
	Port      int    `json:"port" validate:"required"`
	DataTopic string `json:"dataTopic" validate:"required"` // 上报数据的 Topic
	ClientId  string `json:"clientId" validate:"required"`
	Username  string `json:"username" validate:"required"`
	Password  string `json:"password" validate:"required"`
}

//
type mqttOutEndTarget struct {
	typex.XStatus
	client    mqtt.Client
	DataTopic string
}

func NewMqttTarget(e typex.RuleX) typex.XTarget {
	m := new(mqttOutEndTarget)
	m.RuleEngine = e
	return m
}
func (*mqttOutEndTarget) Driver() typex.XExternalDriver {
	return nil
}
func (mm *mqttOutEndTarget) Start(cctx typex.CCTX) error {
	mm.Ctx = cctx.Ctx
	mm.CancelCTX = cctx.CancelCTX
	outEnd := mm.RuleEngine.GetOutEnd(mm.PointId)
	config := outEnd.Config
	var mainConfig mqttConfig
	if err := utils.BindSourceConfig(config, &mainConfig); err != nil {
		return err
	}
	//
	var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
		glogger.GLogger.Infof("Mqtt OutEnd Connected Success")
	}

	var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
		glogger.GLogger.Warnf("Connect lost: %v, try to reconnect\n", err)
	}
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%v", mainConfig.Host, mainConfig.Port))
	opts.SetClientID(mainConfig.ClientId)
	opts.SetUsername(mainConfig.Username)
	opts.SetPassword(mainConfig.Password)
	mm.DataTopic = mainConfig.DataTopic
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	opts.SetPingTimeout(10 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(10 * time.Second)
	mm.client = mqtt.NewClient(opts)
	token := mm.client.Connect()
	token.WaitTimeout(10 * time.Second)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	} else {
		return nil
	}

}

func (mm *mqttOutEndTarget) DataModels() []typex.XDataModel {
	return mm.XDataModels
}

func (mm *mqttOutEndTarget) Stop() {
	if mm.client != nil {
		mm.client.Disconnect(0)
	}
	mm.CancelCTX()

}
func (mm *mqttOutEndTarget) Reload() {

}
func (mm *mqttOutEndTarget) Pause() {

}
func (mm *mqttOutEndTarget) Status() typex.SourceState {
	if mm.client != nil {
		if mm.client.IsConnectionOpen() {
			return typex.SOURCE_UP
		} else {
			return typex.SOURCE_DOWN
		}
	} else {
		return typex.SOURCE_DOWN
	}

}

func (mm *mqttOutEndTarget) Register(outEndId string) error {
	mm.PointId = outEndId
	return nil
}
func (mm *mqttOutEndTarget) Init(outEndId string, cfg map[string]interface{}) error {
	mm.PointId = outEndId
	return nil
}
func (mm *mqttOutEndTarget) Test(outEndId string) bool {
	if mm.client != nil {
		return mm.client.IsConnected()
	}
	return false
}

func (mm *mqttOutEndTarget) Enabled() bool {
	return mm.Enable
}
func (mm *mqttOutEndTarget) Details() *typex.OutEnd {
	return mm.RuleEngine.GetOutEnd(mm.PointId)
}

//
//
//
func (mm *mqttOutEndTarget) To(data interface{}) (interface{}, error) {
	if mm.client != nil {
		return mm.client.Publish(mm.DataTopic, 1, false, data).Error(), nil
	}
	return nil, errors.New("mqtt client is nil")
}

/*
*
* 配置
*
 */
func (*mqttOutEndTarget) Configs() *typex.XConfig {
	return core.GenOutConfig(typex.MQTT_TARGET, "MQTT", httpConfig{})
}
