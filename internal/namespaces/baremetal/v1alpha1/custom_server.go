package baremetal

import (
	"context"
	"reflect"
	"time"

	"github.com/scaleway/scaleway-cli/internal/core"
	baremetal "github.com/scaleway/scaleway-sdk-go/api/baremetal/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	serverActionTimeout = 20 * time.Minute
)

func serverWaitCommand() *core.Command {
	type serverWaitRequest struct {
		ServerID string
		Zone     scw.Zone
	}

	return &core.Command{
		Short:     `Wait for a server to reach a stable state`,
		Long:      `Wait for a server to reach a stable state. This is similar to using --wait flag on other action commands, but without requiring a new action on the server.`,
		Namespace: "baremetal",
		Resource:  "server",
		Verb:      "wait",
		ArgsType:  reflect.TypeOf(serverWaitRequest{}),
		Run: func(ctx context.Context, argsI interface{}) (i interface{}, err error) {
			api := baremetal.NewAPI(core.ExtractClient(ctx))
			return api.WaitForServer(&baremetal.WaitForServerRequest{
				ServerID: argsI.(*serverWaitRequest).ServerID,
				Zone:     argsI.(*serverWaitRequest).Zone,
				Timeout:  serverActionTimeout,
			})
		},
		ArgSpecs: core.ArgSpecs{
			{
				Name:     "server-id",
				Short:    `ID of the server affected by the action.`,
				Required: true,
			},
			core.ZoneArgSpec(),
		},
		Examples: []*core.Example{
			{
				Short:   "Wait for a server to reach a stable state",
				Request: `{"server_id": "11111111-1111-1111-1111-111111111111"}`,
			},
		},
	}
}

// serverStartBuilder overrides the baremetalServerStart command
func serverStartBuilder(c *core.Command) *core.Command {
	c.WaitFunc = func(ctx context.Context, argsI, respI interface{}) (interface{}, error) {
		api := baremetal.NewAPI(core.ExtractClient(ctx))
		return api.WaitForServer(&baremetal.WaitForServerRequest{
			Zone:     argsI.(*baremetal.StartServerRequest).Zone,
			ServerID: respI.(*baremetal.StartServerRequest).ServerID,
			Timeout:  serverActionTimeout,
		})
	}

	return c
}

// serverStopBuilder overrides the baremetalServerStop command
func serverStopBuilder(c *core.Command) *core.Command {
	c.WaitFunc = func(ctx context.Context, argsI, respI interface{}) (interface{}, error) {
		api := baremetal.NewAPI(core.ExtractClient(ctx))
		return api.WaitForServer(&baremetal.WaitForServerRequest{
			Zone:     argsI.(*baremetal.StopServerRequest).Zone,
			ServerID: respI.(*baremetal.StopServerRequest).ServerID,
			Timeout:  serverActionTimeout,
		})
	}

	return c
}

// serverRebootBuilder overrides the baremetalServerReboot command
func serverRebootBuilder(c *core.Command) *core.Command {
	c.WaitFunc = func(ctx context.Context, argsI, respI interface{}) (interface{}, error) {
		api := baremetal.NewAPI(core.ExtractClient(ctx))
		return api.WaitForServer(&baremetal.WaitForServerRequest{
			Zone:     argsI.(*baremetal.RebootServerRequest).Zone,
			ServerID: respI.(*baremetal.RebootServerRequest).ServerID,
			Timeout:  serverActionTimeout,
		})
	}

	return c
}