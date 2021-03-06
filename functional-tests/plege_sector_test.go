package functional

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/filecoin-project/go-filecoin/internal/pkg/constants"
	tf "github.com/filecoin-project/go-filecoin/internal/pkg/testhelpers/testflags"

	"github.com/filecoin-project/go-filecoin/internal/pkg/drand"

	"github.com/stretchr/testify/require"

	"github.com/filecoin-project/go-filecoin/internal/app/go-filecoin/node"
	"github.com/filecoin-project/go-filecoin/internal/pkg/clock"
	gengen "github.com/filecoin-project/go-filecoin/tools/gengen/util"
)

func TestMiningPledgeSector(t *testing.T) {
	t.Skip("This test fails until either the reward actor.LastPerEpochReward verifies its caller, or we relax that condition in the VM")
	tf.FunctionalTest(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wd, _ := os.Getwd()
	genCfgPath := filepath.Join(wd, "..", "fixtures/setup.json")
	presealPath := filepath.Join(wd, "..", "fixtures/genesis-sectors")
	genTime := int64(1000000000)
	blockTime := 1 * time.Second
	fakeClock := clock.NewFake(time.Unix(genTime, 0))

	// Load genesis config fixture.
	genCfg := loadGenesisConfig(t, genCfgPath)
	genCfg.Miners = append(genCfg.Miners, &gengen.CreateStorageMinerConfig{
		Owner:         1,
		SealProofType: constants.DevSealProofType,
	})
	seed := node.MakeChainSeed(t, genCfg)
	chainClock := clock.NewChainClockFromClock(uint64(genTime), blockTime, fakeClock)

	drandImpl := &drand.Fake{
		GenesisTime:   time.Unix(genTime, 0).Add(-1 * blockTime),
		FirstFilecoin: 0,
	}

	bootstrapMiner := makeNode(ctx, t, seed, chainClock, drandImpl)
	_, _, err := initNodeGenesisMiner(ctx, t, bootstrapMiner, seed, genCfg.Miners[0].Owner, presealPath)
	require.NoError(t, err)

	newMiner := makeNode(ctx, t, seed, chainClock, drandImpl)
	seed.GiveKey(t, newMiner, 1)
	_, _ = seed.GiveMiner(t, newMiner, 1)

	err = bootstrapMiner.Start(ctx)
	require.NoError(t, err)
	err = newMiner.Start(ctx)
	require.NoError(t, err)
	defer bootstrapMiner.Stop(ctx)
	defer newMiner.Stop(ctx)

	node.ConnectNodes(t, newMiner, bootstrapMiner)

	// Have bootstrap miner mine continuously so newMiner's pledgeSector can put multiple messages on chain.
	go simulateBlockMining(ctx, t, fakeClock, blockTime, bootstrapMiner)

	err = newMiner.StorageMining.Start(ctx)
	require.NoError(t, err)

	err = newMiner.PieceManager().PledgeSector(ctx)
	require.NoError(t, err)

	// wait while checking to see if the new miner has added any sectors (indicating sealing was successful)
	for i := 0; i < 100; i++ {
		ts, err := newMiner.PorcelainAPI.ChainHead()
		require.NoError(t, err)

		maddr, err := newMiner.BlockMining.BlockMiningAPI.MinerAddress()
		require.NoError(t, err)

		status, err := newMiner.PorcelainAPI.MinerGetStatus(ctx, maddr, ts.Key())
		require.NoError(t, err)

		if status.SectorCount > 0 {
			return
		}

		time.Sleep(2 * time.Second)
	}
	t.Fatal("Did not add sectors in the allotted time")
}
