package common

const (
	DockerRoot                  = "/var/lib/docker-ece"
	ContainerMountRoot          = DockerRoot + "mnt"
	ContainerWriteLayerRoot     = DockerRoot + "writelayers"
	ContainerReadLayerRoot      = DockerRoot + "readlayers"
	DefaultContainerLogLocation = DockerRoot + "log"
	DefaultContainerInfoDir     = DockerRoot + "info"
)
