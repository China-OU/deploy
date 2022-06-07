package ws

import (
	"github.com/astaxie/beego"
)

type WsController struct {
	beego.Controller
	//controllers.BaseController
}

func (c *WsController) URLMapping() {
	c.Mapping("Get", c.Get)
}

// @Title Get
// @Description websocket的get方法，可以传值后台定时刷新，也可以实时交互由前端把要查询的值写到后端
// @Success 200 true or false
// @Failure 403
// @router /ws [get]
func (c *WsController) Get() {
	hub := newHub()
	go hub.run()

	conn, err := upgrader.Upgrade(c.Ctx.ResponseWriter, c.Ctx.Request, nil)
	if err != nil {
		beego.Error(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
