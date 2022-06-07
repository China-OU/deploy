package models

type OprVMUpgrade struct {
    ID              int     `gorm:"column:id" json:"id"`
    UnitID          int     `gorm:"column:unit_id" json:"unit_id"`
    UnitName        string  `gorm:"column:unit_name" json:"unit_name"`                  // 发布单元名，中文名(英文名)
    Operation       string  `gorm:"column:operation" json:"operation"`                  // 操作类型，upgrade, restart
    ArtifactURL     string  `gorm:"column:artifact_url" json:"artifact_url"`            // 升级包下载地址
    OldArtifactHash string  `gorm:"column:old_artifact_hash" json:"old_artifact_hash"`  // 升级前的版本hash，已备份的最新文件为准
    NewArtifactHash string  `gorm:"column:new_artifact_hash" json:"new_artifact_hash"`  // 升级后的版本hash
    Logs            string  `gorm:"column:logs" json:"logs"`
    Operator        string  `gorm:"column:operator" json:"operator"`
    Status          int     `gorm:"column:status" json:"status"`                        // 0:失败，1:成功，2:进行中，3:未开始
    Duration        int     `gorm:"column:duration" json:"duration"`                    // 操作耗时
    CreateTime      string  `gorm:"column:create_time" json:"create_time"`
    UpdateTime      string  `gorm:"column:update_time" json:"update_time"`
    IsDelete        int     `gorm:"column:is_delete" json:"is_delete"`
}

func (OprVMUpgrade) TableName() string {
    return "opr_vm_upgrade"
}

type OprVMVersion struct {
    ID              int     `gorm:"column:id" json:"id"`
    UnitID          int     `gorm:"column:unit_id" json:"unit_id"`
    Version         string  `gorm:"column:version" json:"version"`                  // 版本号，如20200101
    ArtifactFile    string  `gorm:"column:artifact_file" json:"artifact_file"`      // 当前版本部署包位置，以备份的文件为准
    ArtifactHash    string  `gorm:"column:artifact_hash" json:"artifact_hash"`      // 当前版本文件的SHA256校验值
    HashMethod      string  `gorm:"column:hash_method" json:"hash_method"`          // hash校验类型，SHA256，MD5
    CreateTime      string  `gorm:"column:create_time" json:"create_time"`
    UpdateTime      string  `gorm:"column:update_time" json:"update_time"`
    IsDelete        int     `gorm:"column:is_delete" json:"is_delete"`
}

func (OprVMVersion) TableName() string {
    return "opr_vm_version"
}