package driver

import (
	"encoding/binary"
	"encoding/json"

	"github.com/i4de/rulex/common"
	"github.com/i4de/rulex/typex"

	"github.com/goburrow/modbus"
)

/*
*
* Modbus RTU
*
 */
type modBusRtuDriver struct {
	state      typex.DriverState
	handler    *modbus.RTUClientHandler
	client     modbus.Client
	RuleEngine typex.RuleX
	Registers  []common.RegisterRW
	device     *typex.Device
}

func NewModBusRtuDriver(
	d *typex.Device,
	e typex.RuleX,
	Registers []common.RegisterRW,
	handler *modbus.RTUClientHandler,
	client modbus.Client) typex.XExternalDriver {
	return &modBusRtuDriver{
		state:      typex.DRIVER_RUNNING,
		device:     d,
		RuleEngine: e,
		client:     client,
		handler:    handler,
		Registers:  Registers,
	}

}
func (d *modBusRtuDriver) Test() error {
	return nil
}

func (d *modBusRtuDriver) Init(map[string]string) error {
	return nil
}

func (d *modBusRtuDriver) Work() error {
	return nil
}

func (d *modBusRtuDriver) State() typex.DriverState {
	return d.state
}

func (d *modBusRtuDriver) Read(data []byte) (int, error) {
	datas := map[string]common.RegisterRW{}
	for _, r := range d.Registers {
		d.handler.SlaveId = r.SlaverId
		if r.Function == common.READ_COIL {
			results, err := d.client.ReadCoils(r.Address, r.Quantity)
			if err != nil {
				return 0, err
			}
			value := common.RegisterRW{
				Tag:      r.Tag,
				Function: r.Function,
				SlaverId: r.SlaverId,
				Address:  r.Address,
				Quantity: r.Quantity,
				Value:    string(results),
			}
			datas[r.Tag] = value
		}
		if r.Function == common.READ_DISCRETE_INPUT {
			results, err := d.client.ReadDiscreteInputs(r.Address, r.Quantity)
			if err != nil {
				return 0, err
			}
			value := common.RegisterRW{
				Tag:      r.Tag,
				Function: r.Function,
				SlaverId: r.SlaverId,
				Address:  r.Address,
				Quantity: r.Quantity,
				Value:    string(results),
			}
			datas[r.Tag] = value

		}
		if r.Function == common.READ_HOLDING_REGISTERS {
			results, err := d.client.ReadHoldingRegisters(r.Address, r.Quantity)
			if err != nil {
				return 0, err
			}
			value := common.RegisterRW{
				Tag:      r.Tag,
				Function: r.Function,
				SlaverId: r.SlaverId,
				Address:  r.Address,
				Quantity: r.Quantity,
				Value:    string(results),
			}
			datas[r.Tag] = value
		}
		if r.Function == common.READ_INPUT_REGISTERS {
			results, err := d.client.ReadInputRegisters(r.Address, r.Quantity)
			if err != nil {
				return 0, err
			}
			value := common.RegisterRW{
				Tag:      r.Tag,
				Function: r.Function,
				SlaverId: r.SlaverId,
				Address:  r.Address,
				Quantity: r.Quantity,
				Value:    string(results),
			}
			datas[r.Tag] = value
		}

	}
	bytes, _ := json.Marshal(datas)
	copy(data, bytes)
	return len(bytes), nil

}

func (d *modBusRtuDriver) Write(data []byte) (int, error) {
	datas := []common.RegisterRW{}
	if err := json.Unmarshal(data, &datas); err != nil {
		return 0, err
	}
	for _, r := range d.Registers {
		if r.Function == common.WRITE_SINGLE_COIL {
			_, err := d.client.WriteSingleCoil(r.Address, binary.BigEndian.Uint16([]byte(r.Value)[0:2]))
			if err != nil {
				return 0, err
			}
		}
		if r.Function == common.WRITE_MULTIPLE_COILS {
			_, err := d.client.WriteMultipleCoils(r.Address, r.Quantity, []byte(r.Value))
			if err != nil {
				return 0, err
			}
		}
		if r.Function == common.WRITE_SINGLE_HOLDING_REGISTER {
			_, err := d.client.WriteSingleRegister(r.Address, binary.BigEndian.Uint16([]byte(r.Value)[0:2]))
			if err != nil {
				return 0, err
			}
		}
		if r.Function == common.WRITE_MULTIPLE_HOLDING_REGISTERS {
			_, err := d.client.WriteMultipleRegisters(r.Address, r.Quantity, []byte(r.Value))
			if err != nil {
				return 0, err
			}
		}
	}
	return 0, nil
}

func (d *modBusRtuDriver) DriverDetail() typex.DriverDetail {
	return typex.DriverDetail{
		Name:        "Generic ModBus RTU Driver",
		Type:        "UART",
		Description: "Generic ModBus RTU Driver",
	}
}

func (d *modBusRtuDriver) Stop() error {
	d.handler.Close()
	d = nil
	return nil
}
