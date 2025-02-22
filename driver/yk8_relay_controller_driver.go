package driver

//
// RS232/RS485 控制继电器模块，可利用电脑通过串口（没有串口的可利用 USB 转
// 串口）连接控制器进行对设备的控制，接口采用开关输出，有常开常闭点。
// 资料首页：http://www.yi-kun.com
//
import (
	"encoding/json"
	"errors"

	"github.com/i4de/rulex/common"
	"github.com/i4de/rulex/typex"

	"github.com/goburrow/modbus"
)

type YK8RelayControllerDriver struct {
	state      typex.DriverState
	client     modbus.Client
	device     *typex.Device
	RuleEngine typex.RuleX
}

func NewYK8RelayControllerDriver(d *typex.Device, e typex.RuleX,
	client modbus.Client) typex.XExternalDriver {
	return &YK8RelayControllerDriver{
		state:      typex.DRIVER_STOP,
		device:     d,
		RuleEngine: e,
		client:     client,
	}
}

func (yk8 *YK8RelayControllerDriver) Test() error {
	return nil
}

func (yk8 *YK8RelayControllerDriver) Init(map[string]string) error {
	return nil
}

func (yk8 *YK8RelayControllerDriver) Work() error {
	return nil
}

func (yk8 *YK8RelayControllerDriver) State() typex.DriverState {
	return typex.DRIVER_RUNNING
}

//

type yk08sw struct {
	Sw1 bool `json:"sw1"`
	Sw2 bool `json:"sw2"`
	Sw3 bool `json:"sw3"`
	Sw4 bool `json:"sw4"`
	Sw5 bool `json:"sw5"`
	Sw6 bool `json:"sw6"`
	Sw7 bool `json:"sw7"`
	Sw8 bool `json:"sw8"`
}

/*
*
* 读出来的是个JSON, 记录了8个开关的状态
*
 */
func (yk8 *YK8RelayControllerDriver) Read(data []byte) (int, error) {
	results, err := yk8.client.ReadCoils(0x00, 0x08)
	if err != nil {
		return 0, err
	}
	if len(results) == 1 {
		yks := yk08sw{
			Sw1: common.BitToBool(results[0], 0),
			Sw2: common.BitToBool(results[0], 1),
			Sw3: common.BitToBool(results[0], 2),
			Sw4: common.BitToBool(results[0], 3),
			Sw5: common.BitToBool(results[0], 4),
			Sw6: common.BitToBool(results[0], 5),
			Sw7: common.BitToBool(results[0], 6),
			Sw8: common.BitToBool(results[0], 7),
		}
		bytes, _ := json.Marshal(yks)
		copy(data, bytes)
	}
	return len(data), err
}

//
// 写入数据必须是有8个布尔值的字节数组: [1,1,1,1,1,1,1,1]
//
func (yk8 *YK8RelayControllerDriver) Write(data []byte) (int, error) {
	if len(data) != 8 {
		return 0, errors.New("操作继电器组最少8个布尔值")
	}
	for _, v := range data {
		if v > 1 {
			return 0, errors.New("必须是逻辑值")
		}
	}

	Sw1 := common.ByteToBool(data[0])
	Sw2 := common.ByteToBool(data[1])
	Sw3 := common.ByteToBool(data[2])
	Sw4 := common.ByteToBool(data[3])
	Sw5 := common.ByteToBool(data[4])
	Sw6 := common.ByteToBool(data[5])
	Sw7 := common.ByteToBool(data[6])
	Sw8 := common.ByteToBool(data[7])
	var value byte
	common.SetABitOnByte(&value, 0, Sw1)
	common.SetABitOnByte(&value, 1, Sw2)
	common.SetABitOnByte(&value, 2, Sw3)
	common.SetABitOnByte(&value, 3, Sw4)
	common.SetABitOnByte(&value, 4, Sw5)
	common.SetABitOnByte(&value, 5, Sw6)
	common.SetABitOnByte(&value, 6, Sw7)
	common.SetABitOnByte(&value, 7, Sw8)

	_, err := yk8.client.WriteMultipleCoils(0, 1, []byte{value})
	if err != nil {
		return 0, err
	}
	return 0, err
}

//---------------------------------------------------
func (yk8 *YK8RelayControllerDriver) DriverDetail() typex.DriverDetail {
	return typex.DriverDetail{
		Name:        "YK-08-RELAY CONTROLLER",
		Type:        "UART",
		Description: "一个支持RS232和485的国产8路继电器控制器",
	}
}

func (yk8 *YK8RelayControllerDriver) Stop() error {
	yk8.state = typex.DRIVER_STOP
	return nil
}
