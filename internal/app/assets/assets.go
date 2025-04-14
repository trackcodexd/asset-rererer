package assets

import (
	"errors"
	"fmt"

	"github.com/kartFr/Asset-Reuploader/internal/app/assets/animation"
	"github.com/kartFr/Asset-Reuploader/internal/app/assets/shared/clientutils"
	"github.com/kartFr/Asset-Reuploader/internal/app/assets/shared/permissions"
	"github.com/kartFr/Asset-Reuploader/internal/app/context"
	"github.com/kartFr/Asset-Reuploader/internal/app/request"
	"github.com/kartFr/Asset-Reuploader/internal/app/response"
	"github.com/kartFr/Asset-Reuploader/internal/color"
	"github.com/kartFr/Asset-Reuploader/internal/console"
	"github.com/kartFr/Asset-Reuploader/internal/roblox"
)

var assetModules = map[string]func(ctx *context.Context, r *request.Request){
	"Animation": animation.Reupload,
}

func NewReuploadHandlerWithType(assetType string, c *roblox.Client, r *request.RawRequest, resp *response.Response) (func(), error) {
	reupload, exists := assetModules[assetType]
	if !exists {
		return func() {}, errors.New(assetType + " module does not exist")
	}

	return func() {
		ctx := context.New(c, resp)

		console.ClearScreen()

		fmt.Println("Getting current place details...")
		req, err := request.FromRawRequest(c, r)
		console.ClearScreen()
		if err != nil {
			color.Error.Println(err)
			return
		}

		fmt.Println("Checking if account can edit universe...")
		err = permissions.CanEditUniverse(ctx, req)
		console.ClearScreen()
		if err != nil {
			clientutils.GetNewCookie(ctx, req, err.Error())
		}

		reupload(ctx, req)
	}, nil
}

func DoesModuleExist(m string) bool {
	_, exists := assetModules[m]
	return exists
}
