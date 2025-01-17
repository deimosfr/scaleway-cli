package redis

import (
	"github.com/scaleway/scaleway-cli/v2/internal/core"
	"github.com/scaleway/scaleway-cli/v2/internal/human"
	"github.com/scaleway/scaleway-sdk-go/api/redis/v1"
)

func GetCommands() *core.Commands {
	cmds := GetGeneratedCommands()

	human.RegisterMarshalerFunc(redis.Cluster{}, redisClusterGetMarshalerFunc)
	human.RegisterMarshalerFunc(redis.Cluster{}.Endpoints, redisEndpointsClusterGetMarshalerFunc)

	cmds.Merge(core.NewCommands(clusterWaitCommand()))
	cmds.MustFind("redis", "cluster", "create").Override(clusterCreateBuilder)
	cmds.MustFind("redis", "cluster", "delete").Override(clusterDeleteBuilder)
	cmds.MustFind("redis", "acl", "add").Override(ACLAddListBuilder)

	return cmds
}
