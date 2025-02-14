package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/fiatjaf/eventstore/lmdb"
	"github.com/fiatjaf/khatru"
	"github.com/joho/godotenv"
	"github.com/nbd-wtf/go-nostr"
)

var config Config
var relay *khatru.Relay
var walletKinds = []int{
	nostr.KindNWCWalletInfo,
	nostr.KindNWCWalletRequest,
	nostr.KindNWCWalletResponse,
	nostr.KindNutZap,
	nostr.KindNutZapInfo,
	nostr.KindZap,
	nostr.KindZapRequest,
	17375,
	7375,
	7376,
	7376,
}

type Config struct {
	RelayName        string
	RelayPubkey      string
	RelayDescription string
	RelayIcon        string
	RelaySoftware    string
	RelayVersion     string
	RelayPort        string
	LmdbMapSize      int64
	LmdbPath         string
}

func main() {
	relay = khatru.NewRelay()
	config = LoadConfig()

	db := lmdb.LMDBBackend{
		Path:    config.LmdbPath,
		MapSize: config.LmdbMapSize,
	}

	if err := db.Init(); err != nil {
		panic(err)
	}

	relay.StoreEvent = append(relay.StoreEvent, db.SaveEvent)
	relay.QueryEvents = append(relay.QueryEvents, db.QueryEvents)

	relay.RejectFilter = append(relay.RejectFilter, func(ctx context.Context, filter nostr.Filter) (bool, string) {
		if !containsOnlyWalletKids(filter.Kinds) {
			return true, "invalid-filter: only wallet kinds are allowed"
		}

		return false, ""
	})

	relay.RejectEvent = append(relay.RejectEvent, func(ctx context.Context, event *nostr.Event) (bool, string) {
		if !containsOnlyWalletKids([]int{event.Kind}) {
			return true, "invalid-event: only wallet kinds are allowed"
		}

		return false, ""
	})

	addr := fmt.Sprintf("%s:%s", "0.0.0.0", config.RelayPort)

	log.Printf("🔗 listening at %s", addr)
	err := http.ListenAndServe(addr, relay)
	if err != nil {
		log.Fatal(err)
	}
}

func LoadConfig() Config {
	_ = godotenv.Load(".env")

	config = Config{
		RelayName:        os.Getenv("RELAY_NAME"),
		RelayPubkey:      os.Getenv("RELAY_PUBKEY"),
		RelayDescription: os.Getenv("RELAY_DESCRIPTION"),
		RelayIcon:        os.Getenv("RELAY_ICON"),
		RelaySoftware:    "https://github.com/bitvora/wallet-relay",
		RelayVersion:     "0.1.0",
		RelayPort:        os.Getenv("RELAY_PORT"),
		LmdbPath:         os.Getenv("LMDB_PATH"),
	}

	relay.Info.Name = config.RelayName
	relay.Info.PubKey = config.RelayPubkey
	relay.Info.Description = config.RelayDescription
	relay.Info.Icon = config.RelayIcon
	relay.Info.Software = config.RelaySoftware
	relay.Info.Version = config.RelayVersion

	return config
}

func containsOnlyWalletKids(kinds []int) bool {
	walletKindSet := make(map[int]struct{})

	for _, walletKind := range walletKinds {
		walletKindSet[walletKind] = struct{}{}
	}
	for _, kind := range kinds {
		if _, exists := walletKindSet[kind]; !exists {
			return false
		}
	}

	return true
}
