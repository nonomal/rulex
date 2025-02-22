package httpserver

import (
	"context"
	"embed"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/i4de/rulex/glogger"
	"github.com/i4de/rulex/typex"
	"github.com/i4de/rulex/utils"

	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"gopkg.in/ini.v1"

	_ "github.com/mattn/go-sqlite3"

	"gorm.io/gorm"
)

const _API_V1_ROOT string = "/api/v1/"
const _DEFAULT_DB_PATH string = "./rulex.db"

// 启动时间
var StartedTime = time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05")

type _serverConfig struct {
	Enable bool   `ini:"enable"`
	Host   string `ini:"host"`
	DbPath string `ini:"dbpath"`
	Port   int    `ini:"port"`
}
type HttpApiServer struct {
	Port       int
	Host       string
	sqliteDb   *gorm.DB
	dbPath     string
	ginEngine  *gin.Engine
	ruleEngine typex.RuleX
}

func NewHttpApiServer() *HttpApiServer {
	return &HttpApiServer{}
}

//
func (hh *HttpApiServer) Init(config *ini.Section) error {
	gin.SetMode(gin.ReleaseMode)
	hh.ginEngine = gin.New()
	var mainConfig _serverConfig
	if err := utils.InIMapToStruct(config, &mainConfig); err != nil {
		return err
	}
	hh.Host = mainConfig.Host
	hh.dbPath = mainConfig.DbPath
	hh.Port = mainConfig.Port
	configHttpServer(hh)
	//
	// Http server
	//
	go func(ctx context.Context, port int) {
		if err := hh.ginEngine.Run(":" + strconv.Itoa(port)); err != nil {
			glogger.GLogger.Fatalf("httpserver listen error: %s\n", err)
		}
	}(typex.GCTX, hh.Port)
	//
	// WebSocket server
	//
	socketioServer := socketio.NewServer(nil)
	configSocketIO(socketioServer)
	hh.ginEngine.GET("/socket.io", gin.WrapH(socketioServer))
	go func(ctx context.Context) {
		if err := socketioServer.Serve(); err != nil {
			glogger.GLogger.Fatalf("socketio listen error: %s\n", err)
		}
	}(context.Background())
	return nil
}

//
// HttpApiServer Start
//
func (hh *HttpApiServer) Start(r typex.RuleX) error {
	hh.ruleEngine = r
	hh.ginEngine.GET("/", hh.addRoute(func(c *gin.Context, has *HttpApiServer, rx typex.RuleX) {
		c.Request.URL.Path = "/static/"
		hh.ginEngine.HandleContext(c)
	}))

	//
	// Get all plugins
	//
	hh.ginEngine.GET(url("plugins"), hh.addRoute(Plugins))
	//
	// Get system information
	//
	hh.ginEngine.GET(url("system"), hh.addRoute(System))
	//
	// Ping -> Pong
	//
	hh.ginEngine.GET(url("ping"), hh.addRoute(Ping))
	//
	//
	//
	hh.ginEngine.GET(url("sourceCount"), hh.addRoute(SourceCount))
	//
	//
	//
	hh.ginEngine.GET(url("logs"), hh.addRoute(Logs))
	//
	//
	//
	hh.ginEngine.POST(url("logout"), hh.addRoute(LogOut))
	//
	// Get all inends
	//
	hh.ginEngine.GET(url("inends"), hh.addRoute(InEnds))
	//
	//
	//
	hh.ginEngine.GET(url("drivers"), hh.addRoute(Drivers))
	//
	// Get all outends
	//
	hh.ginEngine.GET(url("outends"), hh.addRoute(OutEnds))
	//
	// Get all rules
	//
	hh.ginEngine.GET(url("rules"), hh.addRoute(Rules))
	//
	// Get statistics data
	//
	hh.ginEngine.GET(url("statistics"), hh.addRoute(Statistics))
	//
	// Auth
	//
	hh.ginEngine.POST(url("users"), hh.addRoute(CreateUser))
	//
	//
	//
	hh.ginEngine.POST(url("login"), hh.addRoute(Login))
	//
	//
	//
	hh.ginEngine.GET(url("info"), hh.addRoute(Info))
	//
	// Create InEnd
	//
	hh.ginEngine.POST(url("inends"), hh.addRoute(CreateInend))
	//
	// 配置表
	//
	hh.ginEngine.GET(url("inends/config"), hh.addRoute(GetInEndConfig))
	//
	// 数据模型表
	//
	hh.ginEngine.GET(url("inends/models"), hh.addRoute(GetInEndModels))
	//
	// Create OutEnd
	//
	hh.ginEngine.POST(url("outends"), hh.addRoute(CreateOutEnd))
	//
	// Create rule
	//
	hh.ginEngine.POST(url("rules"), hh.addRoute(CreateRule))
	//
	// Delete inend by UUID
	//
	hh.ginEngine.DELETE(url("inends"), hh.addRoute(DeleteInEnd))
	//
	// Delete outEnd by UUID
	//
	hh.ginEngine.DELETE(url("outends"), hh.addRoute(DeleteOutEnd))
	//
	// Delete rule by UUID
	//
	hh.ginEngine.DELETE(url("rules"), hh.addRoute(DeleteRule))
	//
	// 验证 lua 语法
	//
	hh.ginEngine.POST(url("validateRule"), hh.addRoute(ValidateLuaSyntax))
	//
	// 获取配置表
	//
	hh.ginEngine.GET(url("rType"), hh.addRoute(RType))
	hh.ginEngine.GET(url("tType"), hh.addRoute(TType))
	//
	// 串口列表
	//
	hh.ginEngine.GET(url("uarts"), hh.addRoute(GetUarts))
	//
	// 获取服务启动时间
	//
	hh.ginEngine.GET(url("startedAt"), hh.addRoute(StartedAt))
	//
	// 设备管理
	//
	hh.ginEngine.GET(url("devices"), hh.addRoute(Devices))
	hh.ginEngine.POST(url("devices"), hh.addRoute(CreateDevice))
	hh.ginEngine.PUT(url("devices"), hh.addRoute(UpdateDevice))
	hh.ginEngine.DELETE(url("devices"), hh.addRoute(DeleteDevice))
	// 外挂管理
	hh.ginEngine.GET(url("goods"), hh.addRoute(Goods))
	hh.ginEngine.POST(url("goods"), hh.addRoute(CreateGoods))
	hh.ginEngine.PUT(url("goods"), hh.addRoute(UpdateGoods))
	hh.ginEngine.DELETE(url("goods"), hh.addRoute(DeleteGoods))
	glogger.GLogger.Infof("Http server started on http://0.0.0.0:%v", hh.Port)
	return nil
}

func (hh *HttpApiServer) Stop() error {
	return nil
}

func (hh *HttpApiServer) Db() *gorm.DB {
	return hh.sqliteDb
}
func (hh *HttpApiServer) PluginMetaInfo() typex.XPluginMetaInfo {
	return typex.XPluginMetaInfo{
		Name:     "Rulex Base Api Server",
		Version:  typex.DefaultVersion.Version,
		Homepage: "https://rulex.pages.dev",
		HelpLink: "https://rulex.pages.dev",
		Author:   "wwhai",
		Email:    "cnwwhai@gmail.com",
		License:  "MIT",
	}
}

//go:embed  www/*
var files embed.FS

type customFS struct {
	efs fs.FS
}

// Open 实现fs接口
func (c *customFS) Open(name string) (fs.File, error) {
	if strings.Contains(name, "/") {
		name = "static/" + name
	}
	return c.efs.Open(name)

}

func configHttpServer(hh *HttpApiServer) {
	hh.ginEngine.Use(hh.Authorize())
	hh.ginEngine.Use(Cros())
	www, err := fs.Sub(files, "www")

	if err == nil {
		hh.ginEngine.StaticFS("static", http.FS(&customFS{www}))
	}

	if hh.dbPath == "" {
		hh.InitDb(_DEFAULT_DB_PATH)
	} else {
		hh.InitDb(hh.dbPath)
	}
}

/*
*
* 拼接URL
*
 */
func url(path string) string {
	return _API_V1_ROOT + path
}
