package rdb

import (
	"fmt"
	"strings"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/scaleway/scaleway-cli/v2/internal/core"
	"github.com/scaleway/scaleway-cli/v2/internal/namespaces/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	baseCommand              = "scw rdb instance create node-type=DB-DEV-S is-ha-cluster=false name=%s engine=%s user-name=%s password=%s --wait"
	privateNetworkStaticSpec = " init-endpoints.0.private-network.private-network-id={{ .PN.ID }} init-endpoints.0.private-network.service-ip={{ .IPNet }}"
	privateNetworkIpamSpec   = " init-endpoints.0.private-network.private-network-id={{ .PN.ID }} init-endpoints.0.private-network.enable-ipam=true"
	loadBalancerSpec         = " init-endpoints.1.load-balancer=true"
	publicEndpoint           = "public"
	privateEndpointIpam      = "private IPAM"
	privateEndpointStatic    = "private static"
)

func Test_ListInstance(t *testing.T) {
	t.Run("Simple", core.Test(&core.TestConfig{
		Commands:   GetCommands(),
		BeforeFunc: createInstance("PostgreSQL-12"),
		Cmd:        "scw rdb instance list",
		Check:      core.TestCheckGolden(),
		AfterFunc:  deleteInstance(),
	}))
}

func Test_CloneInstance(t *testing.T) {
	t.Run("Simple", core.Test(&core.TestConfig{
		Commands:   GetCommands(),
		BeforeFunc: createInstance("PostgreSQL-12"),
		Cmd:        "scw rdb instance clone {{ .Instance.ID }} node-type=DB-DEV-M name=foobar --wait",
		Check:      core.TestCheckGolden(),
		AfterFunc:  deleteInstance(),
	}))
}

func Test_CreateInstance(t *testing.T) {
	t.Run("Simple", core.Test(&core.TestConfig{
		Commands: GetCommands(),
		Cmd:      fmt.Sprintf(baseCommand, name, engine, user, password),
		Check: core.TestCheckCombine(
			core.TestCheckGolden(),
			func(t *testing.T, ctx *core.CheckFuncCtx) {
				checkEndpoints(ctx, t, []string{publicEndpoint})
			},
		),
		AfterFunc: core.ExecAfterCmd("scw rdb instance delete {{ .CmdResult.ID }}"),
	}))

	t.Run("With password generator", core.Test(&core.TestConfig{
		Commands: GetCommands(),
		Cmd:      fmt.Sprintf(strings.Replace(baseCommand, "password=%s", "generate-password=true", 1), name, engine, user),
		// do not check the golden as the password generated locally and on CI will necessarily be different
		Check: core.TestCheckCombine(
			core.TestCheckExitCode(0),
			func(t *testing.T, ctx *core.CheckFuncCtx) {
				checkEndpoints(ctx, t, []string{publicEndpoint})
			},
		),
		AfterFunc: core.ExecAfterCmd("scw rdb instance delete {{ .CmdResult.ID }}"),
	}))
}

func Test_CreateInstanceInitEndpoints(t *testing.T) {
	cmds := GetCommands()
	cmds.Merge(vpc.GetCommands())

	t.Run("With static private endpoint", core.Test(&core.TestConfig{
		Commands:   cmds,
		BeforeFunc: createPN(),
		Cmd:        fmt.Sprintf(baseCommand+privateNetworkStaticSpec, name, engine, user, password),
		Check: core.TestCheckCombine(
			core.TestCheckGolden(),
			func(t *testing.T, ctx *core.CheckFuncCtx) {
				checkEndpoints(ctx, t, []string{privateEndpointStatic})
			},
		),
		AfterFunc: core.AfterFuncCombine(
			core.ExecAfterCmd("scw rdb instance delete {{ .CmdResult.ID }} --wait"),
			deletePrivateNetwork(),
		),
	}))

	t.Run("With public and static private endpoint", core.Test(&core.TestConfig{
		Commands:   cmds,
		BeforeFunc: createPN(),
		Cmd:        fmt.Sprintf(baseCommand+privateNetworkStaticSpec+loadBalancerSpec, name, engine, user, password),
		Check: core.TestCheckCombine(
			core.TestCheckGolden(),
			func(t *testing.T, ctx *core.CheckFuncCtx) {
				checkEndpoints(ctx, t, []string{publicEndpoint, privateEndpointStatic})
			},
		),
		AfterFunc: core.AfterFuncCombine(
			core.ExecAfterCmd("scw rdb instance delete {{ .CmdResult.ID }} --wait"),
			deletePrivateNetwork(),
		),
	}))

	t.Run("With IPAM private endpoint", core.Test(&core.TestConfig{
		Commands:   cmds,
		BeforeFunc: createPN(),
		Cmd:        fmt.Sprintf(baseCommand+privateNetworkIpamSpec, name, engine, user, password),
		Check: core.TestCheckCombine(
			core.TestCheckGolden(),
			func(t *testing.T, ctx *core.CheckFuncCtx) {
				checkEndpoints(ctx, t, []string{privateEndpointIpam})
			},
		),
		AfterFunc: core.AfterFuncCombine(
			core.ExecAfterCmd("scw rdb instance delete {{ .CmdResult.ID }} --wait"),
			deletePrivateNetwork(),
		),
	}))

	t.Run("With public and IPAM private endpoint", core.Test(&core.TestConfig{
		Commands:   cmds,
		BeforeFunc: createPN(),
		Cmd:        fmt.Sprintf(baseCommand+privateNetworkIpamSpec+loadBalancerSpec, name, engine, user, password),
		Check: core.TestCheckCombine(
			core.TestCheckGolden(),
			func(t *testing.T, ctx *core.CheckFuncCtx) {
				checkEndpoints(ctx, t, []string{publicEndpoint, privateEndpointIpam})
			},
		),
		AfterFunc: core.AfterFuncCombine(
			core.ExecAfterCmd("scw rdb instance delete {{ .CmdResult.ID }} --wait"),
			deletePrivateNetwork(),
		),
	}))
}

func Test_GetInstance(t *testing.T) {
	t.Run("Simple", core.Test(&core.TestConfig{
		Commands:   GetCommands(),
		BeforeFunc: createInstance("PostgreSQL-12"),
		Cmd:        "scw rdb instance get {{ .Instance.ID }}",
		Check:      core.TestCheckGolden(),
		AfterFunc:  deleteInstance(),
	}))
}

func Test_UpgradeInstance(t *testing.T) {
	t.Run("Simple", core.Test(&core.TestConfig{
		Commands:   GetCommands(),
		BeforeFunc: createInstance("PostgreSQL-12"),
		Cmd:        "scw rdb instance upgrade {{ .Instance.ID }} node-type=DB-DEV-M --wait",
		Check:      core.TestCheckGolden(),
		AfterFunc:  deleteInstance(),
	}))
}

func Test_UpdateInstance(t *testing.T) {
	t.Run("Update instance name", core.Test(&core.TestConfig{
		Commands:   GetCommands(),
		BeforeFunc: createInstance("PostgreSQL-12"),
		Cmd:        "scw rdb instance update {{ .Instance.ID }} name=foo --wait",
		Check: core.TestCheckCombine(
			func(t *testing.T, ctx *core.CheckFuncCtx) {
				assert.Equal(t, "foo", ctx.Result.(*rdb.Instance).Name)
			},
			core.TestCheckGolden(),
			core.TestCheckExitCode(0),
		),
		AfterFunc: deleteInstance(),
	}))

	t.Run("Update instance tags", core.Test(&core.TestConfig{
		Commands:   GetCommands(),
		BeforeFunc: createInstance("PostgreSQL-12"),
		Cmd:        "scw rdb instance update {{ .Instance.ID }} tags.0=a --wait",
		Check: core.TestCheckCombine(
			func(t *testing.T, ctx *core.CheckFuncCtx) {
				assert.Equal(t, "a", ctx.Result.(*rdb.Instance).Tags[0])
			},
			core.TestCheckGolden(),
			core.TestCheckExitCode(0),
		),
		AfterFunc: deleteInstance(),
	}))

	t.Run("Set a timezone", core.Test(&core.TestConfig{
		Commands:   GetCommands(),
		BeforeFunc: createInstance("PostgreSQL-12"),
		Cmd:        "scw rdb instance update {{ .Instance.ID }} settings.0.name=timezone settings.0.value=UTC --wait",
		Check: core.TestCheckCombine(
			func(t *testing.T, ctx *core.CheckFuncCtx) {
				assert.Equal(t, "timezone", ctx.Result.(*rdb.Instance).Settings[5].Name)
				assert.Equal(t, "UTC", ctx.Result.(*rdb.Instance).Settings[5].Value)
			},
			core.TestCheckGolden(),
			core.TestCheckExitCode(0),
		),
		AfterFunc: deleteInstance(),
	}))

	t.Run("Modify default work_mem from 4 to 8 MB", core.Test(&core.TestConfig{
		Commands:   GetCommands(),
		BeforeFunc: createInstance("PostgreSQL-12"),
		Cmd:        "scw rdb instance update {{ .Instance.ID }} settings.0.name=work_mem settings.0.value=8 --wait",
		Check: core.TestCheckCombine(
			func(t *testing.T, ctx *core.CheckFuncCtx) {
				assert.Equal(t, "work_mem", ctx.Result.(*rdb.Instance).Settings[5].Name)
				assert.Equal(t, "8", ctx.Result.(*rdb.Instance).Settings[5].Value)
			},
			core.TestCheckGolden(),
			core.TestCheckExitCode(0),
		),
		AfterFunc: deleteInstance(),
	}))

	t.Run("Modify 3 settings + add a new one", core.Test(&core.TestConfig{
		Commands: GetCommands(),
		BeforeFunc: core.BeforeFuncCombine(
			createInstance("PostgreSQL-12"),
			core.ExecBeforeCmd("scw rdb instance update {{ .Instance.ID }} settings.0.name=work_mem settings.0.value=8"+
				" settings.1.name=max_connections settings.1.value=200"+
				" settings.2.name=effective_cache_size settings.2.value=1000"+
				" name=foo1 --wait"),
		),
		Cmd: "scw rdb instance update {{ .Instance.ID }} settings.0.name=work_mem settings.0.value=16" +
			" settings.1.name=max_connections settings.1.value=150" +
			" settings.2.name=effective_cache_size settings.2.value=1200" +
			" settings.3.name=maintenance_work_mem settings.3.value=200" +
			" name=foo2 --wait",
		Check: core.TestCheckCombine(
			func(t *testing.T, ctx *core.CheckFuncCtx) {
				assert.Equal(t, "effective_cache_size", ctx.Result.(*rdb.Instance).Settings[0].Name)
				assert.Equal(t, "1200", ctx.Result.(*rdb.Instance).Settings[0].Value)
				assert.Equal(t, "maintenance_work_mem", ctx.Result.(*rdb.Instance).Settings[1].Name)
				assert.Equal(t, "200", ctx.Result.(*rdb.Instance).Settings[1].Value)
				assert.Equal(t, "max_connections", ctx.Result.(*rdb.Instance).Settings[2].Name)
				assert.Equal(t, "150", ctx.Result.(*rdb.Instance).Settings[2].Value)
				assert.Equal(t, "work_mem", ctx.Result.(*rdb.Instance).Settings[5].Name)
				assert.Equal(t, "16", ctx.Result.(*rdb.Instance).Settings[5].Value)
				assert.Equal(t, "foo2", ctx.Result.(*rdb.Instance).Name)
			},
			core.TestCheckGolden(),
			core.TestCheckExitCode(0),
		),
		AfterFunc: deleteInstance(),
	}))
}

func Test_Connect(t *testing.T) {
	t.Run("mysql", core.Test(&core.TestConfig{
		Commands: GetCommands(),
		BeforeFunc: core.BeforeFuncCombine(
			core.BeforeFuncStoreInMeta("username", user),
			createInstance("MySQL-8"),
		),
		Cmd: "scw rdb instance connect {{ .Instance.ID }} username={{ .username }}",
		Check: core.TestCheckCombine(
			core.TestCheckGolden(),
			core.TestCheckExitCode(0),
		),
		OverrideExec: core.OverrideExecSimple("mysql --host {{ .Instance.Endpoint.IP }} --port {{ .Instance.Endpoint.Port }} --database rdb --user {{ .username }}", 0),
		AfterFunc:    deleteInstance(),
	}))

	t.Run("psql", core.Test(&core.TestConfig{
		Commands: GetCommands(),
		BeforeFunc: core.BeforeFuncCombine(
			core.BeforeFuncStoreInMeta("username", user),
			createInstance("PostgreSQL-12"),
		),
		Cmd: "scw rdb instance connect {{ .Instance.ID }} username={{ .username }}",
		Check: core.TestCheckCombine(
			core.TestCheckGolden(),
			core.TestCheckExitCode(0),
		),
		OverrideExec: core.OverrideExecSimple("psql --host {{ .Instance.Endpoint.IP }} --port {{ .Instance.Endpoint.Port }} --username {{ .username }} --dbname rdb", 0),
		AfterFunc:    deleteInstance(),
	}))
	t.Run("psql", core.Test(&core.TestConfig{
		Commands: GetCommands(),
		BeforeFunc: core.BeforeFuncCombine(
			core.BeforeFuncStoreInMeta("username", user),
			createPN(),
			createInstanceWithPrivateNetworkAndLoadBalancer("PostgreSQL-14"),
		),
		Cmd: "scw rdb instance connect {{ .Instance.ID }} username={{ .username }}",
		Check: core.TestCheckCombine(
			core.TestCheckGolden(),
			core.TestCheckExitCode(0),
		),
		OverrideExec: core.OverrideExecSimple("psql --host {{ .Instance.Endpoint.IP }} --port {{ .Instance.Endpoint.Port }} --username {{ .username }} --dbname rdb", 0),
		AfterFunc:    deleteInstance(),
	}))
}

func deletePrivateNetwork() core.AfterFunc {
	return core.ExecAfterCmd("scw vpc private-network delete {{ .PN.ID }}")
}

func checkEndpoints(ctx *core.CheckFuncCtx, t *testing.T, expected []string) {
	instance := ctx.Result.(createInstanceResult).Instance
	ipamAPI := ipam.NewAPI(ctx.Client)
	var foundEndpoints = map[string]bool{}

	for _, endpoint := range instance.Endpoints {
		if endpoint.LoadBalancer != nil {
			foundEndpoints[publicEndpoint] = true
		}
		if endpoint.PrivateNetwork != nil {
			ips, err := ipamAPI.ListIPs(&ipam.ListIPsRequest{
				Region:       instance.Region,
				ResourceID:   &instance.ID,
				ResourceType: "rdb_instance",
				IsIPv6:       scw.BoolPtr(false),
			}, scw.WithAllPages())
			if err != nil {
				t.Errorf("could not list IPs: %v", err)
			}
			switch ips.TotalCount {
			case 1:
				foundEndpoints[privateEndpointIpam] = true
			case 0:
				foundEndpoints[privateEndpointStatic] = true
			default:
				t.Errorf("expected no more than 1 IP for instance, got %d", ips.TotalCount)
			}
		}
	}

	// Check that every expected endpoint got found
	for _, e := range expected {
		_, ok := foundEndpoints[e]
		if !ok {
			t.Errorf("expected a %s endpoint but got none", e)
		}
		delete(foundEndpoints, e)
	}
	// Check that no unexpected endpoint was found
	if len(foundEndpoints) > 0 {
		for e := range foundEndpoints {
			t.Errorf("found a %s endpoint when none was expected", e)
		}
	}
}
