package report

import (
	"controllers"
)

type ReportController struct {
	controllers.BaseController
}

func (c *ReportController) URLMapping() {
	// 首页
	c.Mapping("GetIndex", c.GetIndex)

	c.Mapping("CntrUp", c.CntrUp)
	c.Mapping("CntrDeploy", c.CntrDeploy)
	c.Mapping("VmDeploy", c.VmDeploy)
}