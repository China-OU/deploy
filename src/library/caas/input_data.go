package caas

type ServiceUpgradeInput struct {
	StartFirst     string     `json:"startFirst"`
	LaunchConfig   LConfig    `json:"launchconfig"`
	InitContainers []string   `json:"initContainers"`
}

type LConfig struct {
	ImageUuid    string     `json:"imageUuid"`
}
