package iam

import "github.com/scaleway/scaleway-cli/v2/internal/core"

func addSSHKey(metaKey string, key string) core.BeforeFunc {
	return func(ctx *core.BeforeFuncCtx) error {
		ctx.Meta[metaKey] = ctx.ExecuteCmd([]string{
			"scw", "iam", "ssh-key", "create", "public-key=" + key,
		})
		return nil
	}
}
