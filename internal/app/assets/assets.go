package assets

import (
	"errors"
	"fmt"

	"github.com/trackcodexd/asset-rererer/internal/app/assets/animation"
    "github.com/trackcodexd/asset-rererer/internal/app/assets/sound"
	"github.com/trackcodexd/asset-rererer/internal/app/assets/shared/clientutils"
	"github.com/trackcodexd/asset-rererer/internal/app/assets/shared/permissions"
	"github.com/trackcodexd/asset-rererer/internal/app/context"
	"github.com/trackcodexd/asset-rererer/internal/app/request"
	"github.com/trackcodexd/asset-rererer/internal/app/response"
	"github.com/trackcodexd/asset-rererer/internal/console"
	"github.com/trackcodexd/asset-rererer/internal/roblox"
)

var assetModules = map[string]func(ctx *context.Context, r *request.Request){
	"Animation": animation.Reupload,
	"Sound": sound.Reupload,
}

func NewReuploadHandlerWithType(assetType string, c *roblox.Client, r *request.RawRequest, resp *response.Response) (func() error, error) {
	reupload, exists := assetModules[assetType]
	if !exists {
		return func() error { return nil }, errors.New(assetType + " module does not exist")
	}

	return func() error {
		ctx := context.New(c, resp)

		console.ClearScreen()

		fmt.Println("Getting current place details...")
		req, err := request.FromRawRequest(c, r)
		console.ClearScreen()
		if err != nil {
			return err
		}

		fmt.Println("Checking if account can edit universe...")
		err = permissions.CanEditUniverse(ctx, req)
		console.ClearScreen()
		if err != nil {
			clientutils.GetNewCookie(ctx, req, err.Error())
		}

		reupload(ctx, req)
		return nil
	}, nil
}

func DoesModuleExist(m string) bool {
	_, exists := assetModules[m]
	return exists
}
