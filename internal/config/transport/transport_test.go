package transport

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/toknowwhy/theunit-oracle/pkg/ethereum"
	"github.com/toknowwhy/theunit-oracle/pkg/ethereum/mocks"
	"github.com/toknowwhy/theunit-oracle/pkg/log/null"
	"github.com/toknowwhy/theunit-oracle/pkg/transport"
	"github.com/toknowwhy/theunit-oracle/pkg/transport/local"
	"github.com/toknowwhy/theunit-oracle/pkg/transport/messages"
	"github.com/toknowwhy/theunit-oracle/pkg/transport/p2p"
)

func TestTransport_P2P_EmptyConfig(t *testing.T) {
	prevP2PTransportFactory := p2pTransportFactory
	defer func() { p2pTransportFactory = prevP2PTransportFactory }()

	feeds := []ethereum.Address{ethereum.HexToAddress("0x07a35a1d4b751a818d93aa38e615c0df23064881")}
	signer := &mocks.Signer{}
	logger := null.New()

	config := Transport{
		P2P: P2P{
			PrivKeySeed:      "",
			ListenAddrs:      nil,
			BootstrapAddrs:   nil,
			DirectPeersAddrs: nil,
			BlockedAddrs:     nil,
			DisableDiscovery: false,
		},
	}

	p2pTransportFactory = func(ctx context.Context, cfg p2p.Config) (transport.Transport, error) {
		assert.NotNil(t, ctx)
		assert.NotNil(t, cfg.PeerPrivKey)
		assert.Len(t, cfg.ListenAddrs, 0)
		assert.Len(t, cfg.BootstrapAddrs, 0)
		assert.Len(t, cfg.DirectPeersAddrs, 0)
		assert.Len(t, cfg.BlockedAddrs, 0)
		assert.Equal(t, map[string]transport.Message{messages.PriceMessageName: (*messages.Price)(nil)}, cfg.Topics)
		assert.Equal(t, true, cfg.Discovery)
		assert.Equal(t, "spire", cfg.AppName)
		assert.Equal(t, feeds, cfg.FeedersAddrs)
		assert.Same(t, signer, cfg.Signer)
		assert.Same(t, logger, cfg.Logger)

		return local.New(context.Background(), 0, nil), nil
	}

	tra, err := config.Configure(Dependencies{
		Context: context.Background(),
		Signer:  signer,
		Feeds:   feeds,
		Logger:  logger,
	})
	require.NoError(t, err)
	assert.NotNil(t, tra)
}

func TestTransport_P2P_CustomValues(t *testing.T) {
	prevP2PTransportFactory := p2pTransportFactory
	defer func() { p2pTransportFactory = prevP2PTransportFactory }()

	feeds := []ethereum.Address{ethereum.HexToAddress("0x07a35a1d4b751a818d93aa38e615c0df23064881")}
	signer := &mocks.Signer{}
	logger := null.New()
	privKeySeed := "d382e2b16d8a2e770dd8e0b65554a2ce7a072ac67d4ca6f34052771dfdcdac07"
	listenAddrs := []string{"/ip4/0.0.0.0/tcp/8000"}
	bootstrapAddrs := []string{"/ip4/1.1.1.1/tcp/8000/p2p/abc"}
	directPeersAddrs := []string{"/ip4/1.1.1.2/tcp/8000/p2p/abc"}
	blockedAddrs := []string{"/ip4/1.1.1.3/tcp/8000/p2p/abc"}

	config := Transport{
		P2P: P2P{
			PrivKeySeed:      privKeySeed,
			ListenAddrs:      listenAddrs,
			BootstrapAddrs:   bootstrapAddrs,
			DirectPeersAddrs: directPeersAddrs,
			BlockedAddrs:     blockedAddrs,
			DisableDiscovery: true,
		},
	}

	p2pTransportFactory = func(ctx context.Context, cfg p2p.Config) (transport.Transport, error) {
		assert.NotNil(t, ctx)
		assert.NotNil(t, cfg.PeerPrivKey)
		assert.Equal(t, listenAddrs, cfg.ListenAddrs)
		assert.Equal(t, bootstrapAddrs, cfg.BootstrapAddrs)
		assert.Equal(t, directPeersAddrs, cfg.DirectPeersAddrs)
		assert.Equal(t, blockedAddrs, cfg.BlockedAddrs)
		assert.Equal(t, map[string]transport.Message{messages.PriceMessageName: (*messages.Price)(nil)}, cfg.Topics)
		assert.Equal(t, false, cfg.Discovery)
		assert.Equal(t, "spire", cfg.AppName)
		assert.Equal(t, feeds, cfg.FeedersAddrs)
		assert.Same(t, signer, cfg.Signer)
		assert.Same(t, logger, cfg.Logger)

		return local.New(context.Background(), 0, nil), nil
	}

	tra, err := config.Configure(Dependencies{
		Context: context.Background(),
		Signer:  signer,
		Feeds:   feeds,
		Logger:  logger,
	})
	require.NoError(t, err)
	assert.NotNil(t, tra)
}

func TestTransport_P2P_InvalidSeed(t *testing.T) {
	prevP2PTransportFactory := p2pTransportFactory
	defer func() { p2pTransportFactory = prevP2PTransportFactory }()

	config := Transport{
		P2P: P2P{
			PrivKeySeed:      "invalid",
			ListenAddrs:      nil,
			BootstrapAddrs:   nil,
			DirectPeersAddrs: nil,
			BlockedAddrs:     nil,
			DisableDiscovery: false,
		},
	}

	feeds := []ethereum.Address{ethereum.HexToAddress("0x07a35a1d4b751a818d93aa38e615c0df23064881")}
	signer := &mocks.Signer{}
	logger := null.New()

	p2pTransportFactory = func(ctx context.Context, cfg p2p.Config) (transport.Transport, error) {
		assert.NotNil(t, ctx)
		return local.New(context.Background(), 0, nil), nil
	}

	_, err := config.Configure(Dependencies{
		Context: context.Background(),
		Signer:  signer,
		Feeds:   feeds,
		Logger:  logger,
	})
	require.Error(t, err)
}
