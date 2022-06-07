package initial

import (
	"github.com/astaxie/beego/cache"
)

var GetCache cache.Cache

func InitSesson() {
	GetCache, _ = cache.NewCache("memory", `{"interval":60}`)
}
