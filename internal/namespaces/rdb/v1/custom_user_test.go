package rdb

import (
	"fmt"
	"testing"

	"github.com/scaleway/scaleway-cli/v2/internal/core"
)

func Test_ListUser(t *testing.T) {
	t.Run("Simple", core.Test(&core.TestConfig{
		Commands:   GetCommands(),
		BeforeFunc: createInstance("PostgreSQL-12"),
		Cmd:        "scw rdb user list instance-id={{ .Instance.ID }}",
		Check:      core.TestCheckGolden(),
		AfterFunc:  deleteInstance(),
	}))
}

func Test_CreateUser(t *testing.T) {
	t.Run("Simple", core.Test(&core.TestConfig{
		Commands:   GetCommands(),
		BeforeFunc: createInstance("PostgreSQL-12"),
		Cmd:        fmt.Sprintf("scw rdb user create instance-id={{ $.Instance.Instance.ID }} name=%s password=%s", name, password),
		Check:      core.TestCheckGolden(),
		AfterFunc:  deleteInstance(),
	}))

	t.Run("With password generator", core.Test(&core.TestConfig{
		Commands:   GetCommands(),
		BeforeFunc: createInstance("PostgreSQL-12"),
		Cmd:        fmt.Sprintf("scw rdb user create instance-id={{ $.Instance.Instance.ID }} name=%s generate-password=true", name),
		// do not check the golden as the password generated locally and on CI will necessarily be different
		Check:     core.TestCheckExitCode(0),
		AfterFunc: deleteInstance(),
	}))
}

func Test_UpdateUser(t *testing.T) {
	t.Run("Simple", core.Test(&core.TestConfig{
		Commands: GetCommands(),
		BeforeFunc: core.BeforeFuncCombine(
			createInstance("PostgreSQL-12"),
			core.ExecBeforeCmd(fmt.Sprintf("scw rdb user create instance-id={{ $.Instance.Instance.ID }} name=%s password=%s", name, password)),
		),
		Cmd:       fmt.Sprintf("scw rdb user update instance-id={{ $.Instance.Instance.ID }} name=%s password=Newp1ssw0rd! is-admin=true", name),
		Check:     core.TestCheckGolden(),
		AfterFunc: deleteInstance(),
	}))

	t.Run("With password generator", core.Test(&core.TestConfig{
		Commands: GetCommands(),
		BeforeFunc: core.BeforeFuncCombine(
			createInstance("PostgreSQL-12"),
			core.ExecBeforeCmd(fmt.Sprintf("scw rdb user create instance-id={{ $.Instance.Instance.ID }} name=%s password=%s", name, password)),
		),
		Cmd: fmt.Sprintf("scw rdb user update instance-id={{ $.Instance.Instance.ID }} name=%s generate-password=true is-admin=true", name),
		// do not check the golden as the password generated locally and on CI will necessarily be different
		Check:     core.TestCheckExitCode(0),
		AfterFunc: deleteInstance(),
	}))
}
