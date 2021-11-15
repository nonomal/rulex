package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"rulex/core"
	"rulex/statistics"
	"rulex/typex"
	"rulex/utils"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ngaut/log"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
)


//
// Get all plugins
//
func Plugins(c *gin.Context, hh *HttpApiServer, e typex.RuleX) {
	data := []interface{}{}
	for _, v := range e.AllPlugins() {
		data = append(data, v.XPluginMetaInfo())
	}
	c.JSON(http.StatusOK, Result{
		Code: http.StatusOK,
		Msg:  "Success",
		Data: data,
	})
}
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

//
// Get system infomation
//
func System(c *gin.Context, hh *HttpApiServer, e typex.RuleX) {
	cpuPercent, _ := cpu.Percent(time.Millisecond, true)
	parts, _ := disk.Partitions(true)
	diskInfo, _ := disk.Usage(parts[0].Mountpoint)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	c.JSON(http.StatusOK, Result{
		Code: http.StatusOK,
		Msg:  "Success",
		Data: gin.H{
			"version":    e.Version().Version,
			"diskInfo":   int(diskInfo.UsedPercent),
			"system":     bToMb(m.Sys),
			"alloc":      bToMb(m.Alloc),
			"total":      bToMb(m.TotalAlloc),
			"cpuPercent": cpuPercent,
			"osArch":     runtime.GOOS + "-" + runtime.GOARCH,
		},
	})
}

//
// Get all inends
//
func InEnds(c *gin.Context, hh *HttpApiServer, e typex.RuleX) {
	data := []interface{}{}
	for _, v := range e.AllInEnd() {
		data = append(data, v)
	}
	c.JSON(http.StatusOK, Result{
		Code: http.StatusOK,
		Msg:  "Success",
		Data: data,
	})
}

//
// Get all Drivers
//
func Drivers(c *gin.Context, hh *HttpApiServer, e typex.RuleX) {
	data := []interface{}{}
	for _, v := range e.AllInEnd() {
		if v.Resource.Driver() != nil {
			data = append(data, v.Resource.Driver().DriverDetail())
		}
	}
	c.JSON(200, Result{
		Code: 200,
		Msg:  "Success",
		Data: data,
	})
}

//
// Get all outends
//
func OutEnds(c *gin.Context, hh *HttpApiServer, e typex.RuleX) {
	data := []interface{}{}
	for _, v := range e.AllOutEnd() {
		data = append(data, v)
	}
	c.JSON(http.StatusOK, Result{
		Code: http.StatusOK,
		Msg:  "Success",
		Data: data,
	})
}

//
// Get all rules
//
func Rules(c *gin.Context, hh *HttpApiServer, e typex.RuleX) {
	data := []interface{}{}
	for _, v := range e.AllRule() {
		data = append(data, v)
	}
	c.JSON(http.StatusOK, Result{
		Code: http.StatusOK,
		Msg:  "Success",
		Data: data,
	})
}

//
// Get statistics data
//
func Statistics(c *gin.Context, hh *HttpApiServer, e typex.RuleX) {
	c.JSON(http.StatusOK, Result{
		Code: http.StatusOK,
		Msg:  "Success",
		Data: statistics.AllStatistics(),
	})
}

//
// Get statistics data
//
func ResourceCount(c *gin.Context, hh *HttpApiServer, e typex.RuleX) {
	c.JSON(http.StatusOK, Result{
		Code: http.StatusOK,
		Msg:  "Success",
		Data: map[string]int{
			"inends":  len(e.AllInEnd()),
			"outends": len(e.AllOutEnd()),
			"rules":   len(e.AllRule()),
			"plugins": len(e.AllPlugins()),
		},
	})
}

//
// All Users
//
func Users(c *gin.Context, hh *HttpApiServer, e typex.RuleX) {
	users := hh.AllMUser()
	c.JSON(http.StatusOK, Result{
		Code: http.StatusOK,
		Msg:  "Success",
		Data: users,
	})
}

//
// Create InEnd
//
func CreateInend(c *gin.Context, hh *HttpApiServer, e typex.RuleX) {
	type Form struct {
		Type        string                 `json:"type" binding:"required"`
		Name        string                 `json:"name" binding:"required"`
		Description string                 `json:"description"`
		Config      map[string]interface{} `json:"config" binding:"required"`
	}
	form := Form{}

	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(200, Error400(err))
		return
	}
	configJson, err1 := json.Marshal(form.Config)
	if err1 != nil {
		c.JSON(200, Error400(err1))
		return
	}
	uuid := utils.MakeUUID("INEND")
	hh.InsertMInEnd(&MInEnd{
		UUID:        uuid,
		Type:        form.Type,
		Name:        form.Name,
		Description: form.Description,
		Config:      string(configJson),
	})
	if err := hh.LoadNewestInEnd(uuid); err != nil {
		log.Error(err)
		c.JSON(200, Error400(err))
		return
	} else {
		c.JSON(http.StatusOK, Ok())
		return
	}

}

//
// Create OutEnd
//
func CreateOutEnd(c *gin.Context, hh *HttpApiServer, e typex.RuleX) {
	type Form struct {
		Type        string                 `json:"type" binding:"required"`
		Name        string                 `json:"name" binding:"required"`
		Description string                 `json:"description"`
		Config      map[string]interface{} `json:"config" binding:"required"`
	}
	form := Form{}
	err0 := c.ShouldBindJSON(&form)
	if err0 != nil {
		c.JSON(200, gin.H{"msg": err0.Error()})
		return
	} else {
		configJson, err1 := json.Marshal(form.Config)
		if err1 != nil {
			c.JSON(200, gin.H{"msg": err1.Error()})
			return
		} else {
			uuid := utils.MakeUUID("OUTEND")
			hh.InsertMOutEnd(&MOutEnd{
				UUID:        uuid,
				Type:        form.Type,
				Name:        form.Name,
				Description: form.Description,
				Config:      string(configJson),
			})
			err := hh.LoadNewestOutEnd(uuid)
			if err != nil {
				c.JSON(200, Error400(err))
				return
			} else {
				c.JSON(200, Ok())
				return
			}
		}
	}
}

//
// Create rule
//
func CreateRule(c *gin.Context, hh *HttpApiServer, e typex.RuleX) {
	type Form struct {
		From        []string `json:"from" binding:"required"`
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Actions     string   `json:"actions"`
		Success     string   `json:"success"`
		Failed      string   `json:"failed"`
	}
	form := Form{}

	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(200, Error400(err))
		return
	}

	if len(form.From) > 0 {
		for _, id := range form.From {
			if e.GetInEnd(id) == nil {
				c.JSON(200, errors.New(`"inend not exists:" `+id))
				return
			}
		}

		tmpRule := typex.NewRule(nil,
			"",
			form.Name,
			form.Description,
			nil,
			form.Success,
			form.Actions,
			form.Failed)

		if err := core.VerifyCallback(tmpRule); err != nil {
			c.JSON(200, Error400(err))
			return
		} else {
			mRule := &MRule{
				UUID:        utils.MakeUUID("RULE"),
				Name:        form.Name,
				Description: form.Description,
				From:        form.From,
				Success:     form.Success,
				Failed:      form.Failed,
				Actions:     form.Actions,
			}
			if err := hh.InsertMRule(mRule); err != nil {
				c.JSON(200, gin.H{"msg": err.Error()})
				return
			}
			rule := typex.NewRule(hh.ruleEngine,
				mRule.UUID,
				mRule.Name,
				mRule.Description,
				mRule.From,
				mRule.Success,
				mRule.Actions,
				mRule.Failed)
			if err := e.LoadRule(rule); err != nil {
				c.JSON(200, Error400(err))
			} else {
				c.JSON(200, Ok())
			}
			return
		}
	} else {
		c.JSON(200, Error400(errors.New("from can't empty")))
		return
	}

}

//
// Delete inend by UUID
//
func DeleteInend(c *gin.Context, hh *HttpApiServer, e typex.RuleX) {
	uuid, _ := c.GetQuery("uuid")
	_, err := hh.GetMInEnd(uuid)
	if err != nil {
		c.JSON(200, Error400(err))
		return
	}
	if err := hh.DeleteMInEnd(uuid); err != nil {
		c.JSON(200, Error400(err))
	} else {
		e.RemoveInEnd(uuid)
		c.JSON(200, Ok())
	}

}

//
// Delete outend by UUID
//
func DeleteOutend(c *gin.Context, hh *HttpApiServer, e typex.RuleX) {
	uuid, _ := c.GetQuery("uuid")
	_, err := hh.GetMOutEnd(uuid)
	if err != nil {
		c.JSON(200, gin.H{"msg": err.Error()})
	} else {
		if err := hh.DeleteMOutEnd(uuid); err != nil {
			e.RemoveOutEnd(uuid)
			c.JSON(200, gin.H{"msg": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"msg": "remove success"})
		}
	}
}

//
// Delete rule by UUID
//
func DeleteRule(c *gin.Context, hh *HttpApiServer, e typex.RuleX) {
	uuid, _ := c.GetQuery("uuid")
	_, err0 := hh.GetMRule(uuid)
	if err0 != nil {
		c.JSON(200, Error400(err0))
		return
	}
	if err1 := hh.DeleteMRule(uuid); err1 != nil {
		c.JSON(200, Error400(err1))
	} else {
		e.RemoveRule(uuid)
		c.JSON(200, Ok())
	}

}

//
// CreateUser
//
func CreateUser(c *gin.Context, hh *HttpApiServer, e typex.RuleX) {
	type Form struct {
		Role        string `json:"role" binding:"required"`
		Username    string `json:"username" binding:"required"`
		Password    string `json:"password" binding:"required"`
		Description string `json:"description"`
	}
	form := Form{}
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, Result{
			Code: http.StatusBadGateway,
			Msg:  err.Error(),
			Data: nil,
		})
		return
	}

	if user, err := hh.GetMUser(form.Username, form.Password); err != nil {
		c.JSON(http.StatusOK, Result{
			Code: http.StatusBadGateway,
			Msg:  err.Error(),
			Data: nil,
		})
		return
	} else {
		if user.ID > 0 {
			c.JSON(http.StatusOK, Result{
				Code: http.StatusBadGateway,
				Msg:  "用户已存在:" + user.Username,
				Data: nil,
			})
			return
		} else {
			hh.InsertMUser(&MUser{
				Role:        form.Role,
				Username:    form.Username,
				Password:    form.Password,
				Description: form.Description,
			})
			c.JSON(http.StatusOK, Result{
				Code: http.StatusOK,
				Msg:  "用户创建成功",
				Data: form.Username,
			})
			return
		}
	}
}

//
// Auth
//
func Auth(c *gin.Context, hh *HttpApiServer, e typex.RuleX) {
	// type Form struct {
	// 	Username string `json:"username" binding:"required"`
	// 	Password string `json:"password" binding:"required"`
	// }
	// form := Form{}
	// err0 := c.ShouldBindJSON(&form)
	// if err0 != nil {
	// 	c.JSON(http.StatusOK, Result{
	// 		Code: http.StatusBadGateway,
	// 		Msg:  err0.Error(),
	// 		Data: nil,
	// 	})
	// } else {
	// 	user, err1 := hh.GetMUser(form.Username, form.Password)
	// 	if err1 != nil {
	// 		c.JSON(http.StatusOK, Result{
	// 			Code: http.StatusBadGateway,
	// 			Msg:  err1.Error(),
	// 			Data: nil,
	// 		})
	// 	} else {
	// 		c.JSON(http.StatusOK, Result{
	// 			Code: http.StatusOK,
	// 			Msg:  "Auth Success",
	// 			Data: user.Username,
	// 		})
	// 	}
	// }
	c.JSON(http.StatusOK, Result{
		Code: http.StatusOK,
		Msg:  "Auth Success",
		Data: map[string]interface{}{
			"token":  "defe7c05fea849c78cec647273427ee7",
			"avatar": "rulex",
			"name":   "rulex",
		},
	})
}
func Info(c *gin.Context, hh *HttpApiServer, e typex.RuleX) {
	c.JSON(http.StatusOK, Result{
		Code: http.StatusOK,
		Msg:  "Auth Success",
		Data: map[string]interface{}{
			"token":  "defe7c05fea849c78cec647273427ee7",
			"avatar": "rulex",
			"name":   "rulex",
		},
	})
}
func Logs(c *gin.Context, hh *HttpApiServer, e typex.RuleX) {
	type Data struct {
		Id      int    `json:"id" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	logs := []Data{}
	for i, s := range core.LogSlot {
		if s != "" {
			logs = append(logs, Data{i, s})
		}
	}
	c.JSON(http.StatusOK, Result{
		Code: http.StatusOK,
		Msg:  "Success",
		Data: logs,
	})
}

func LogOut(c *gin.Context, hh *HttpApiServer, e typex.RuleX) {
	c.JSON(http.StatusOK, Ok())
}
