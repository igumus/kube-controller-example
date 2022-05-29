package controller

const globalObjectMetadataLabelKey = "mode"
const serviceObjectMetadataLabelVal = "autoExposed"
const deploymentObjectMetadataLabelVal = "autoExpose"

var serviceMetadataLabels = map[string]string{
	globalObjectMetadataLabelKey: serviceObjectMetadataLabelVal,
}
