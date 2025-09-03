package sound

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/trackcodexd/asset-rererer/internal/app/context"
	"github.com/trackcodexd/asset-rererer/internal/app/request"
	"github.com/trackcodexd/asset-rererer/internal/app/response"
)

const assetTypeID int32 = 3 // roblox Sound

func Reupload(ctx *context.Context, r *request.Request) {
	logger := ctx.Logger
	resp := ctx.Response

	total := len(r.IDs)
	var processed atomic.Int32

	logger.Println(fmt.Sprintf("Reuploading %d sounds...", total))

	var wg sync.WaitGroup
	for i, id := range r.IDs {
		wg.Add(1)

		go func(i int, id int64) {
			defer wg.Done()
			time.Sleep(time.Duration(1+i%3) * time.Second)

			count := processed.Add(1)
			soundName := assetInfo.Name
            logger.Success(fmt.Sprintf("[%d/%d] %s(%d): %d", count, total, soundName, assetInfo.ID, assetInfo.ID))


			resp.AddItem(response.ResponseItem{
				OldID: id,
				NewID: id,
			})
		}(i, id)
	}

	wg.Wait()
	logger.Println("Sound reuploading finished.")
}
