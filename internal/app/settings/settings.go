package settings

import (
	"fmt"
	"time"
)

const (
	CompatiblePluginVersion = "1.0.0"
	Port                    = 51048
	CookieFileName          = "cookie.txt"
)

func GetOutputFileName(reuploadType string) string {
	t := time.Now()
	return fmt.Sprintf("Output_%s_%s.json", reuploadType, t.Format("2006-01-02_15-04-05"))
}
