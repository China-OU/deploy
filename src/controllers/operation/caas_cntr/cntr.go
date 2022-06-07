package caas_cntr

import "controllers"

type CntrController struct {
	controllers.BaseController
	IsEdit bool
}

func (c *CntrController) URLMapping() {
	c.Mapping("InitService", c.InitService)
	c.Mapping("ReInitService", c.ReInitService)
	c.Mapping("ListCntrInit", c.ListCntrInit)
	c.Mapping("GetCntrInit", c.GetCntrInit)
	c.Mapping("ListAppUnit", c.ListAppUnit)
	c.Mapping("DeleteInitRecord", c.DeleteInitRecord)
	c.Mapping("ListOprLog", c.ListOprLog)
	c.Mapping("CntrConfigSync", c.CntrConfigSync)
	c.Mapping("SyncCntrInitStatus", c.SyncCntrInitStatus)
	c.Mapping("CntrConfigScan", c.CntrConfigScan)

}