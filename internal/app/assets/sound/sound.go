package sound

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kartFr/Asset-Reuploader/internal/app/assets/shared/assetutils"
	"github.com/kartFr/Asset-Reuploader/internal/app/assets/shared/clientutils"
	"github.com/kartFr/Asset-Reuploader/internal/app/assets/shared/uploaderror"
	"github.com/kartFr/Asset-Reuploader/internal/app/context"
	"github.com/kartFr/Asset-Reuploader/internal/app/request"
	"github.com/kartFr/Asset-Reuploader/internal/app/response"
	"github.com/kartFr/Asset-Reuploader/internal/atomicarray"
	"github.com/kartFr/Asset-Reuploader/internal/retry"
	"github.com/kartFr/Asset-Reuploader/internal/roblox/assetdelivery"
	"github.com/kartFr/Asset-Reuploader/internal/roblox/develop"
	"github.com/kartFr/Asset-Reuploader/internal/roblox/games"
	"github.com/kartFr/Asset-Reuploader/internal/roblox/ide"
	"github.com/kartFr/Asset-Reuploader/internal/shardedmap"
	"github.com/kartFr/Asset-Reuploader/internal/taskqueue"
)

const assetTypeID int32 = 3 // Roblox internal type ID for Sound/Audio

var ErrUnauthorized = errors.New("authentication required to access asset")

func Reupload(ctx *context.Context, r *request.Request) {
	client := ctx.Client
	logger := ctx.Logger
	pauseController := ctx.PauseController
	resp := ctx.Response

	idsToUpload := len(r.IDs)
	var idsProcessed atomic.Int32

	var groupID int64
	if r.IsGroup {
		groupID = r.CreatorID
	}

	filter := assetutils.NewFilter(ctx, r, assetTypeID)

	uploadQueue := taskqueue.New[int64](time.Minute, 3000)

	logger.Println("Reuploading sounds...")

	newUploadError := func(m string, assetInfo *develop.AssetInfo, err any) {
		newValue := idsProcessed.Add(1)
		logger.Error(uploaderror.New(int(newValue), idsToUpload, m, assetInfo, err))
	}

	uploadAsset := func(wg *sync.WaitGroup, assetInfo *develop.AssetInfo, location string) {
		defer wg.Done()

		oldName := assetInfo.Name
		assetData, err := clientutils.GetRequest(client, location)
		if err != nil {
			newUploadError("Failed to get asset data", assetInfo, err)
			return
		}

		uploadHandler, err := ide.NewUploadAudioHandler(client, assetInfo.Name, "", assetData, groupID)
		if err != nil {
			newUploadError("Failed to get upload handler", assetInfo, err)
			return
		}

		res := <-uploadQueue.QueueTask(func() (int64, error) {
			return retry.Do(
				retry.NewOptions(retry.Tries(3)),
				func(try int) (int64, error) {
					pauseController.WaitIfPaused()
					if try > 1 {
						uploadQueue.Limiter.Wait()
					}
					id, err := uploadHandler()
					if err == nil {
						return id, nil
					}
					if err == ide.UploadAudioErrors.ErrNotLoggedIn {
						clientutils.GetNewCookie(ctx, r, "cookie expired")
					} else if err == ide.UploadAudioErrors.ErrInappropriateName {
						assetInfo.Name = fmt.Sprintf("(%s) [Censored]", assetInfo.Name)
					} else {
						if _, ok := err.(*net.OpError); ok {
							uploadQueue.Limiter.Decrement()
						}
					}
					return 0, &retry.ContinueRetry{Err: err}
				},
			)
		})

		if err := res.Error; err != nil {
			assetInfo.Name = oldName
			newUploadError("Failed to upload", assetInfo, err)
			return
		}

		newID := res.Result
		newValue := idsProcessed.Add(1)
		logger.Success(uploaderror.New(int(newValue), idsToUpload, "", assetInfo, newID))
		resp.AddItem(response.ResponseItem{OldID: assetInfo.ID, NewID: newID})
	}

	// similar batching logic to animation.go...
	tasks := assetutils.GetAssetsInfoInChunks(ctx, r)
	var wg sync.WaitGroup
	wg.Add(len(tasks))
	for _, task := range tasks {
		go func(resChan <-chan assetutils.AssetsInfoResult) {
			defer wg.Done()
			res := <-resChan
			if err := res.Error; err != nil {
				logger.Error(uploaderror.NewBatch(0, 0, idsToUpload, "Failed to get assets info", err))
				return
			}
			assetsInfo := filter(res.Result)
			var uploadWG sync.WaitGroup
			for _, assetInfo := range assetsInfo {
				locations, _ := assetdelivery.NewBatchHandler(client, []*assetdelivery.AssetRequestItem{
					{AssetID: assetInfo.ID, AssetTypeID: assetTypeID},
				}, r.PlaceID)()
				if len(locations) == 0 || len(locations[0].Locations) == 0 {
					newUploadError("Failed to get asset location", assetInfo, "no locations")
					continue
				}
				uploadWG.Add(1)
				go uploadAsset(&uploadWG, assetInfo, locations[0].Locations[0].Location)
			}
			uploadWG.Wait()
		}(task)
	}
	wg.Wait()
}
