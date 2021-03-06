package gengen_test

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/filecoin-project/specs-actors/actors/builtin"
	ds "github.com/ipfs/go-datastore"
	blockstore "github.com/ipfs/go-ipfs-blockstore"

	"github.com/filecoin-project/go-filecoin/internal/pkg/constants"
	th "github.com/filecoin-project/go-filecoin/internal/pkg/testhelpers"
	tf "github.com/filecoin-project/go-filecoin/internal/pkg/testhelpers/testflags"
	. "github.com/filecoin-project/go-filecoin/tools/gengen/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testConfig(t *testing.T) *GenesisCfg {
	fiftyCommCfgs, err := MakeCommitCfgs(50)
	require.NoError(t, err)
	tenCommCfgs, err := MakeCommitCfgs(10)
	require.NoError(t, err)

	return &GenesisCfg{
		KeysToGen:         4,
		PreallocatedFunds: []string{"1000000", "500000"},
		Miners: []*CreateStorageMinerConfig{
			{
				Owner:            0,
				CommittedSectors: fiftyCommCfgs,
				SealProofType:    constants.DevSealProofType,
			},
			{
				Owner:            1,
				CommittedSectors: tenCommCfgs,
				SealProofType:    constants.DevSealProofType,
			},
		},
		Network: "gfctest",
		Seed:    defaultSeed,
		Time:    defaultTime,
	}
}

const defaultSeed = 4
const defaultTime = 123456789

func TestGenGenLoading(t *testing.T) {
	tf.IntegrationTest(t)

	fi, err := ioutil.TempFile("", "gengentest")
	assert.NoError(t, err)

	_, err = GenGenesisCar(testConfig(t), fi)
	assert.NoError(t, err)
	assert.NoError(t, fi.Close())

	td := th.NewDaemon(t, th.GenesisFile(fi.Name())).Start()
	defer td.ShutdownSuccess()

	o := td.Run("actor", "ls").AssertSuccess()

	stdout := o.ReadStdout()
	assert.Contains(t, stdout, builtin.StoragePowerActorCodeID.String())
	assert.Contains(t, stdout, builtin.StorageMarketActorCodeID.String())
	assert.Contains(t, stdout, builtin.InitActorCodeID.String())
}

func TestGenGenDeterministic(t *testing.T) {
	tf.IntegrationTest(t)

	ctx := context.Background()
	var info *RenderedGenInfo
	for i := 0; i < 5; i++ {
		bstore := blockstore.NewBlockstore(ds.NewMapDatastore())
		inf, err := GenGen(ctx, testConfig(t), bstore)
		assert.NoError(t, err)
		if info == nil {
			info = inf
		} else {
			assert.Equal(t, info, inf)
		}
	}
}
