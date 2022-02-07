package node

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/simone-trubian/blockchain-tutorial/database"
	"github.com/simone-trubian/blockchain-tutorial/fs"
	"github.com/simone-trubian/blockchain-tutorial/wallet"
)

// The password for testing keystore files:
//
// 	./test_simone--3eb92807f1f91a8d4d85bc908c7f86dcddb1df57
// 	./test_tanya--6fdc0d8d15ae6b4ebf45c52fd2aafbcbb19a65c8
//
// Pre-generated for testing purposes using wallet_test.go.
//
// It's necessary to have pre-existing accounts before a new node
// with fresh new, empty keystore is initialized and booted in order
// to configure the accounts balances in genesis.json
//
// I.e: A quick solution to a chicken-egg problem.
const testKsSimoneAccount = "0x3eb92807f1f91a8d4d85bc908c7f86dcddb1df57"
const testKsTanyaAccount = "0x6fdc0d8d15ae6b4ebf45c52fd2aafbcbb19a65c8"
const testKsSimoneFile = "test_simone--3eb92807f1f91a8d4d85bc908c7f86dcddb1df57"
const testKsTanyaFile = "test_tanya--6fdc0d8d15ae6b4ebf45c52fd2aafbcbb19a65c8"
const testKsAccountsPwd = "security123"

func TestNode_Run(t *testing.T) {
	datadir, err := getTestDataDirPath()
	if err != nil {
		t.Fatal(err)
	}
	err = fs.RemoveDir(datadir)
	if err != nil {
		t.Fatal(err)
	}

	n := New(
		datadir,
		"127.0.0.1",
		8085,
		database.NewAccount(DefaultMiner), PeerNode{})

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	err = n.Run(ctx)
	if err.Error() != "http: Server closed" {
		t.Fatal("node server was suppose to close after 5s")
	}
}

func TestNode_Mining(t *testing.T) {
	tanya := database.NewAccount(testKsTanyaAccount)
	simone := database.NewAccount(testKsSimoneAccount)

	genesisBalances := make(map[common.Address]uint)
	genesisBalances[simone] = 1000000
	genesis := database.Genesis{Balances: genesisBalances}
	genesisJson, err := json.Marshal(genesis)
	if err != nil {
		t.Fatal(err)
	}

	dataDir, err := getTestDataDirPath()
	if err != nil {
		t.Fatal(err)
	}

	err = database.InitDataDirIfNotExists(dataDir, genesisJson)
	defer fs.RemoveDir(dataDir)

	err = copyKeystoreFilesIntoTestDataDirPath(dataDir)
	if err != nil {
		t.Fatal(err)
	}

	// Required for AddPendingTX() to describe
	// from what node the TX came from (local node in this case)
	nInfo := NewPeerNode(
		"127.0.0.1",
		8085,
		false,
		database.NewAccount(""),
		true,
	)

	// Construct a new Node instance and configure
	// Simone as a miner
	n := New(dataDir, nInfo.IP, nInfo.Port, simone, nInfo)

	// Allow the mining to run for 30 mins, in the worst case
	ctx, closeNode := context.WithTimeout(
		context.Background(),
		time.Minute*30,
	)

	// Schedule a new TX in 3 seconds from now, in a separate thread
	// because the n.Run() few lines below is a blocking call
	go func() {
		time.Sleep(time.Second * miningIntervalSeconds / 3)

		tx := database.NewTx(simone, tanya, 1, "")
		signedTx, err := wallet.SignTxWithKeystoreAccount(
			tx, simone, testKsAccountsPwd, wallet.GetKeystoreDirPath(dataDir))
		if err != nil {
			t.Error(err)
			return
		}

		_ = n.AddPendingTX(signedTx, nInfo)
	}()

	// Schedule a new TX in 12 seconds from now simulating
	// that it came in - while the first TX is being mined
	go func() {
		time.Sleep(time.Second*miningIntervalSeconds + 2)

		tx := database.NewTx(simone, tanya, 2, "")
		signedTx, err := wallet.SignTxWithKeystoreAccount(
			tx, simone, testKsAccountsPwd, wallet.GetKeystoreDirPath(dataDir))
		if err != nil {
			t.Error(err)
			return
		}

		_ = n.AddPendingTX(signedTx, nInfo)
	}()

	go func() {
		// Periodically check if we mined the 2 blocks
		ticker := time.NewTicker(10 * time.Second)

		for {
			select {
			case <-ticker.C:
				if n.state.LatestBlock().Header.Number == 1 {
					closeNode()
					return
				}
			}
		}
	}()

	// Run the node, mining and everything in a blocking call
	// (hence the go-routines before)
	_ = n.Run(ctx)

	if n.state.LatestBlock().Header.Number != 1 {
		t.Fatal("2 pending TX not mined into 2 under 30m")
	}
}

// The test logic summary:
//	- Tanya runs the node
//  - Tanya tries to mine 2 TXs
//  	- The mining gets interrupted because a new block from Simone gets synced
//		- Simone will get the block reward for this synced block
//		- The synced block contains 1 of the TXs Tanya tried to mine
//	- Tanya tries to mine 1 TX left
//		- Tanya succeeds and gets her block reward
func TestNode_MiningStopsOnNewSyncedBlock(t *testing.T) {
	tanya := database.NewAccount(testKsTanyaAccount)
	simone := database.NewAccount(testKsSimoneAccount)

	dataDir, err := getTestDataDirPath()
	if err != nil {
		t.Fatal(err)
	}

	genesisBalances := make(map[common.Address]uint)
	genesisBalances[simone] = 1000000
	genesis := database.Genesis{Balances: genesisBalances}
	genesisJson, err := json.Marshal(genesis)
	if err != nil {
		t.Fatal(err)
	}

	err = database.InitDataDirIfNotExists(dataDir, genesisJson)
	defer fs.RemoveDir(dataDir)

	err = copyKeystoreFilesIntoTestDataDirPath(dataDir)
	if err != nil {
		t.Fatal(err)
	}

	// Required for AddPendingTX() to describe
	// from what node the TX came from (local node in this case)
	nInfo := NewPeerNode(
		"127.0.0.1",
		8085,
		false,
		database.NewAccount(""),
		true,
	)

	n := New(dataDir, nInfo.IP, nInfo.Port, tanya, nInfo)

	// Allow the test to run for 30 mins, in the worst case
	ctx, closeNode := context.WithTimeout(context.Background(), time.Minute*30)

	tx1 := database.NewTx(simone, tanya, 1, "")
	tx2 := database.NewTx(simone, tanya, 2, "")

	signedTx1, err := wallet.SignTxWithKeystoreAccount(
		tx1, simone, testKsAccountsPwd, wallet.GetKeystoreDirPath(dataDir))
	if err != nil {
		t.Error(err)
		return
	}

	signedTx2, err := wallet.SignTxWithKeystoreAccount(
		tx2, simone, testKsAccountsPwd, wallet.GetKeystoreDirPath(dataDir))
	if err != nil {
		t.Error(err)
		return
	}
	tx2Hash, err := signedTx2.Hash()
	if err != nil {
		t.Error(err)
		return
	}

	// Pre-mine a valid block without running the `n.Run()`
	// with Simone as a miner who will receive the block reward,
	// to simulate the block came on the fly from another peer
	validPreMinedPb := NewPendingBlock(
		database.Hash{}, 0, simone, []database.SignedTx{signedTx1})
	validSyncedBlock, err := Mine(ctx, validPreMinedPb)
	if err != nil {
		t.Fatal(err)
	}

	// Add 2 new TXs into the Tanya's node, triggers mining
	go func() {
		time.Sleep(time.Second * (miningIntervalSeconds - 2))

		err := n.AddPendingTX(signedTx1, nInfo)
		if err != nil {
			t.Fatal(err)
		}

		err = n.AddPendingTX(signedTx2, nInfo)
		if err != nil {
			t.Fatal(err)
		}
	}()

	// Interrupt the previously started mining with a new synced block
	// BUT this block contains only 1 TX the previous mining activity tried to
	// mine which means the mining will start again for the one pending TX
	// that is left and wasn't in the synced block
	go func() {
		time.Sleep(time.Second * (miningIntervalSeconds + 2))
		if !n.isMining {
			t.Fatal("should be mining")
		}

		_, err := n.state.AddBlock(validSyncedBlock)
		if err != nil {
			t.Fatal(err)
		}
		// Mock the Simone's block came from a network
		n.newSyncedBlocks <- validSyncedBlock

		time.Sleep(time.Second * 2)
		if n.isMining {
			t.Fatal("synced block should have canceled mining")
		}

		// Mined TX1 by Simone should be removed from the Mempool
		_, onlyTX2IsPending := n.pendingTXs[tx2Hash.Hex()]

		if len(n.pendingTXs) != 1 && !onlyTX2IsPending {
			t.Fatal("synced block should have canceled mining of already mined TX")
		}

		time.Sleep(time.Second * (miningIntervalSeconds + 2))
		if !n.isMining {
			t.Fatal("should be mining again the 1 TX not included in synced block")
		}
	}()

	go func() {
		// Regularly check whenever both TXs are now mined
		ticker := time.NewTicker(time.Second * 10)

		for {
			select {
			case <-ticker.C:
				if n.state.LatestBlock().Header.Number == 1 {
					closeNode()
					return
				}
			}
		}
	}()

	go func() {
		time.Sleep(time.Second * 2)

		// Take a snapshot of the DB balances
		// before the mining is finished and the 2 blocks
		// are created.
		startingSimoneBalance := n.state.Balances[simone]
		startingTanyaBalance := n.state.Balances[tanya]

		// Wait until the 30 mins timeout is reached or
		// the 2 blocks got already mined and the closeNode() was triggered
		<-ctx.Done()

		endSimoneBalance := n.state.Balances[simone]
		endTanyaBalance := n.state.Balances[tanya]

		// In TX1 Simone transferred 1 SB token to Tanya
		// In TX2 Simone transferred 2 SB tokens to Tanya
		expectedEndSimoneBalance := startingSimoneBalance - tx1.Value - tx2.Value + database.BlockReward
		expectedEndTanyaBalance := startingTanyaBalance + tx1.Value + tx2.Value + database.BlockReward

		if endSimoneBalance != expectedEndSimoneBalance {
			t.Fatalf("Simone expected end balance is %d not %d", expectedEndSimoneBalance, endSimoneBalance)
		}

		if endTanyaBalance != expectedEndTanyaBalance {
			t.Fatalf("Tanya expected end balance is %d not %d", expectedEndTanyaBalance, endTanyaBalance)
		}

		t.Logf("Starting Simone balance: %d", startingSimoneBalance)
		t.Logf("Starting Tanya balance: %d", startingTanyaBalance)
		t.Logf("Ending Simone balance: %d", endSimoneBalance)
		t.Logf("Ending Tanya balance: %d", endTanyaBalance)
	}()

	_ = n.Run(ctx)

	if n.state.LatestBlock().Header.Number != 1 {
		t.Fatal("was suppose to mine 2 pending TX into 2 valid blocks under 30m")
	}

	if len(n.pendingTXs) != 0 {
		t.Fatal("no pending TXs should be left to mine")
	}
}

// Creates dir like: "/tmp/sb_test945924586"
func getTestDataDirPath() (string, error) {
	return ioutil.TempDir(os.TempDir(), "sb_test")
}

// Copy the pre-generated, commited keystore files from this folder into the new testDataDirPath()
//
// Afterwards the test datadir path will look like:
// 	"/tmp/sb_test945924586/keystore/test_simone--3eb92807f1f91a8d4d85bc908c7f86dcddb1df57"
// 	"/tmp/sb_test945924586/keystore/test_tanya--6fdc0d8d15ae6b4ebf45c52fd2aafbcbb19a65c8"
func copyKeystoreFilesIntoTestDataDirPath(dataDir string) error {
	simoneSrcKs, err := os.Open(testKsSimoneFile)
	if err != nil {
		return err
	}
	defer simoneSrcKs.Close()

	ksDir := filepath.Join(wallet.GetKeystoreDirPath(dataDir))

	err = os.Mkdir(ksDir, 0777)
	if err != nil {
		return err
	}

	simoneDstKs, err := os.Create(filepath.Join(ksDir, testKsSimoneFile))
	if err != nil {
		return err
	}
	defer simoneDstKs.Close()

	_, err = io.Copy(simoneDstKs, simoneSrcKs)
	if err != nil {
		return err
	}

	tanyaSrcKs, err := os.Open(testKsTanyaFile)
	if err != nil {
		return err
	}
	defer tanyaSrcKs.Close()

	tanyaDstKs, err := os.Create(filepath.Join(ksDir, testKsTanyaFile))
	if err != nil {
		return err
	}
	defer tanyaDstKs.Close()

	_, err = io.Copy(tanyaDstKs, tanyaSrcKs)
	if err != nil {
		return err
	}

	return nil
}
