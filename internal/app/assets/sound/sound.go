package sound

import (
	"fmt"

	"github.com/trackcodexd/asset-rererer/internal/app/context"
	"github.com/trackcodexd/asset-rererer/internal/app/request"
	"github.com/trackcodexd/asset-rererer/internal/app/response"
)

func Reupload(ctx *context.Context, r *request.Request) {
	// just log for now
	fmt.Println("Sound reupload stub running...")
	for _, id := range r.IDs {
		ctx.Response.AddItem(response.ResponseItem{
			OldID: id,
			NewID: id,
		})
	}
}
