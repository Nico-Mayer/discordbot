package lol

import (
	"os"

	"github.com/KnutZuidema/golio"
	"github.com/KnutZuidema/golio/api"
)

var GolioClient *golio.Client

func init() {
	GolioClient = golio.NewClient(
		os.Getenv("RIOT_API_KEY"),
		golio.WithRegion(api.RegionEuropeWest),
	)
}
