package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"io/ioutil"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/simone-trubian/blockchain-tutorial/database"
	"github.com/simone-trubian/blockchain-tutorial/fs"
)

// The password for testing keystore files:
// 	./node/test_simone--3eb92807f1f91a8d4d85bc908c7f86dcddb1df57
// 	./node/test_tanya--6fdc0d8d15ae6b4ebf45c52fd2aafbcbb19a65c8
const testKeystoreAccountsPwd = "security123"

func TestSign(t *testing.T) {
	privKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	pubKey := privKey.PublicKey
	pubKeyBytes := elliptic.Marshal(crypto.S256(), pubKey.X, pubKey.Y)
	pubKeyBytesHash := crypto.Keccak256(pubKeyBytes[1:])

	account := common.BytesToAddress(pubKeyBytesHash[12:])

	msg := []byte("the Web3Coach students are awesome")

	sig, err := Sign(msg, privKey)
	if err != nil {
		t.Fatal(err)
	}

	recoveredPubKey, err := Verify(msg, sig)
	if err != nil {
		t.Fatal(err)
	}

	recoveredPubKeyBytes := elliptic.Marshal(crypto.S256(), recoveredPubKey.X, recoveredPubKey.Y)
	recoveredPubKeyBytesHash := crypto.Keccak256(recoveredPubKeyBytes[1:])
	recoveredAccount := common.BytesToAddress(recoveredPubKeyBytesHash[12:])

	if account.Hex() != recoveredAccount.Hex() {
		t.Fatalf("msg was signed by account %s but signature recovery produced an account %s", account.Hex(), recoveredAccount.Hex())
	}
}

func TestSignTxWithKeystoreAccount(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "wallet_test")
	if err != nil {
		t.Fatal(err)
	}
	defer fs.RemoveDir(tmpDir)

	simone, err := NewKeystoreAccount(tmpDir, testKeystoreAccountsPwd)
	if err != nil {
		t.Error(err)
		return
	}

	tanya, err := NewKeystoreAccount(tmpDir, testKeystoreAccountsPwd)
	if err != nil {
		t.Error(err)
		return
	}

	tx := database.NewTx(simone, tanya, 100, "")

	signedTx, err := SignTxWithKeystoreAccount(tx, simone, testKeystoreAccountsPwd, GetKeystoreDirPath(tmpDir))
	if err != nil {
		t.Error(err)
		return
	}

	spew.Dump(signedTx.Encode())

	ok, err := signedTx.IsAuthentic()
	if err != nil {
		t.Error(err)
		return
	}

	if !ok {
		t.Fatal("the TX was signed by 'from' account and should have been authentic")
	}
}

func TestSignForgedTxWithKeystoreAccount(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "wallet_test")
	if err != nil {
		t.Fatal(err)
	}
	defer fs.RemoveDir(tmpDir)

	hacker, err := NewKeystoreAccount(tmpDir, testKeystoreAccountsPwd)
	if err != nil {
		t.Error(err)
		return
	}

	tanya, err := NewKeystoreAccount(tmpDir, testKeystoreAccountsPwd)
	if err != nil {
		t.Error(err)
		return
	}

	forgedTx := database.NewTx(tanya, hacker, 100, "")

	signedTx, err := SignTxWithKeystoreAccount(forgedTx, hacker, testKeystoreAccountsPwd, GetKeystoreDirPath(tmpDir))
	if err != nil {
		t.Error(err)
		return
	}

	ok, err := signedTx.IsAuthentic()
	if err != nil {
		t.Error(err)
		return
	}

	if ok {
		t.Fatal("the TX 'from' attribute was forged and should have not be authentic")
	}
}
