package test

import (
	"bytes"
	"math/rand"
	"sort"
	"testing"
	"time"

	ic "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	pt "github.com/libp2p/go-libp2p-core/test"
	"github.com/textileio/go-textile-core/crypto"
	"github.com/textileio/go-textile-core/thread"
	tstore "github.com/textileio/go-textile-core/threadstore"
)

var logKeyBookSuite = map[string]func(kb tstore.LogKeyBook) func(*testing.T){
	"AddGetPrivKey":         testKeyBookPrivKey,
	"AddGetPubKey":          testKeyBookPubKey,
	"AddGetReadKey":         testKeyBookReadKey,
	"AddGetFollowKey":       testKeyBookFollowKey,
	"LogsWithKeys":          testKeyBookLogs,
	"ThreadsFromKeys":       testKeyBookThreads,
	"PubKeyAddedOnRetrieve": testInlinedPubKeyAddedOnRetrieve,
}

type LogKeyBookFactory func() (tstore.LogKeyBook, func())

func LogKeyBookTest(t *testing.T, factory LogKeyBookFactory) {
	for name, test := range logKeyBookSuite {
		// Create a new book.
		kb, closeFunc := factory()

		// Run the test.
		t.Run(name, test(kb))

		// Cleanup.
		if closeFunc != nil {
			closeFunc()
		}
	}
}

func testKeyBookPrivKey(kb tstore.LogKeyBook) func(t *testing.T) {
	return func(t *testing.T) {
		tid := thread.NewIDV1(thread.Raw, 24)

		if logs := kb.LogsWithKeys(tid); len(logs) > 0 {
			t.Error("expected logs to be empty on init")
		}

		priv, _, err := pt.RandTestKeyPair(ic.RSA, 512)
		if err != nil {
			t.Error(err)
		}

		id, err := peer.IDFromPrivateKey(priv)
		if err != nil {
			t.Error(err)
		}

		if res := kb.LogPrivKey(tid, id); res != nil {
			t.Error("retrieving private key should have failed")
		}

		err = kb.AddLogPrivKey(tid, id, priv)
		if err != nil {
			t.Error(err)
		}

		if res := kb.LogPrivKey(tid, id); !priv.Equals(res) {
			t.Error("retrieved private key did not match stored private key")
		}

		if logs := kb.LogsWithKeys(tid); len(logs) != 1 || logs[0] != id {
			t.Error("list of logs did not include test log")
		}
	}
}

func testKeyBookPubKey(kb tstore.LogKeyBook) func(t *testing.T) {
	return func(t *testing.T) {
		tid := thread.NewIDV1(thread.Raw, 24)

		if logs := kb.LogsWithKeys(tid); len(logs) > 0 {
			t.Error("expected logs to be empty on init")
		}

		_, pub, err := pt.RandTestKeyPair(ic.RSA, 512)
		if err != nil {
			t.Error(err)
		}

		id, err := peer.IDFromPublicKey(pub)
		if err != nil {
			t.Error(err)
		}

		if res := kb.LogPubKey(tid, id); res != nil {
			t.Error("retrieving public key should have failed")
		}

		err = kb.AddLogPubKey(tid, id, pub)
		if err != nil {
			t.Error(err)
		}

		if res := kb.LogPubKey(tid, id); !pub.Equals(res) {
			t.Error("retrieved public key did not match stored public key")
		}

		if logs := kb.LogsWithKeys(tid); len(logs) != 1 || logs[0] != id {
			t.Error("list of logs did not include test log")
		}
	}
}

func testKeyBookReadKey(kb tstore.LogKeyBook) func(t *testing.T) {
	return func(t *testing.T) {
		tid := thread.NewIDV1(thread.Raw, 24)

		if logs := kb.LogsWithKeys(tid); len(logs) > 0 {
			t.Error("expected logs to be empty on init")
		}

		_, pub, err := pt.RandTestKeyPair(ic.RSA, 512)
		if err != nil {
			t.Error(err)
		}

		id, err := peer.IDFromPublicKey(pub)
		if err != nil {
			t.Error(err)
		}

		key, err := crypto.GenerateAESKey()
		if err != nil {
			t.Error(err)
		}

		err = kb.AddLogReadKey(tid, id, key)
		if err != nil {
			t.Error(err)
		}

		if res := kb.LogReadKey(tid, id); !bytes.Equal(res, key) {
			t.Error("retrieved read key did not match stored read key")
		}
	}
}

func testKeyBookFollowKey(kb tstore.LogKeyBook) func(t *testing.T) {
	return func(t *testing.T) {
		tid := thread.NewIDV1(thread.Raw, 24)

		if logs := kb.LogsWithKeys(tid); len(logs) > 0 {
			t.Error("expected logs to be empty on init")
		}

		_, pub, err := pt.RandTestKeyPair(ic.RSA, 512)
		if err != nil {
			t.Error(err)
		}

		id, err := peer.IDFromPublicKey(pub)
		if err != nil {
			t.Error(err)
		}

		key, err := crypto.GenerateAESKey()
		if err != nil {
			t.Error(err)
		}

		err = kb.AddLogFollowKey(tid, id, key)
		if err != nil {
			t.Error(err)
		}

		if res := kb.LogFollowKey(tid, id); !bytes.Equal(res, key) {
			t.Error("retrieved read key did not match stored read key")
		}
	}
}

func testKeyBookLogs(kb tstore.LogKeyBook) func(t *testing.T) {
	return func(t *testing.T) {
		tid := thread.NewIDV1(thread.Raw, 24)

		if logs := kb.LogsWithKeys(tid); len(logs) > 0 {
			t.Error("expected logs to be empty on init")
		}

		logs := make(peer.IDSlice, 0)
		for i := 0; i < 10; i++ {
			// Add a public key.
			_, pub, _ := pt.RandTestKeyPair(ic.RSA, 512)
			p1, _ := peer.IDFromPublicKey(pub)
			_ = kb.AddLogPubKey(tid, p1, pub)

			// Add a private key.
			priv, _, _ := pt.RandTestKeyPair(ic.RSA, 512)
			p2, _ := peer.IDFromPrivateKey(priv)
			_ = kb.AddLogPrivKey(tid, p2, priv)

			logs = append(logs, []peer.ID{p1, p2}...)
		}

		kbLogs := kb.LogsWithKeys(tid)
		sort.Sort(kbLogs)
		sort.Sort(logs)

		for i, p := range kbLogs {
			if p != logs[i] {
				t.Errorf("mismatch of log at index %d", i)
			}
		}
	}
}

func testKeyBookThreads(kb tstore.LogKeyBook) func(t *testing.T) {
	return func(t *testing.T) {
		if threads := kb.ThreadsFromKeys(); len(threads) > 0 {
			t.Error("expected threads to be empty on init")
		}

		threads := thread.IDSlice{
			thread.NewIDV1(thread.Raw, 16),
			thread.NewIDV1(thread.Raw, 24),
			thread.NewIDV1(thread.Textile, 32),
		}
		rand.Seed(time.Now().Unix())
		for i := 0; i < 10; i++ {
			// Choose a random thread.
			tid := threads[rand.Intn(len(threads))]
			// Add a public key.
			_, pub, _ := pt.RandTestKeyPair(ic.RSA, 512)
			p1, _ := peer.IDFromPublicKey(pub)
			_ = kb.AddLogPubKey(tid, p1, pub)

			// Add a private key.
			priv, _, _ := pt.RandTestKeyPair(ic.RSA, 512)
			p2, _ := peer.IDFromPrivateKey(priv)
			_ = kb.AddLogPrivKey(tid, p2, priv)
		}

		kbThreads := kb.ThreadsFromKeys()
		sort.Sort(kbThreads)
		sort.Sort(threads)

		for i, p := range kbThreads {
			if p != threads[i] {
				t.Errorf("mismatch of thread at index %d", i)
			}
		}
	}
}

func testInlinedPubKeyAddedOnRetrieve(kb tstore.LogKeyBook) func(t *testing.T) {
	return func(t *testing.T) {
		t.Skip("key inlining disabled for now: see libp2p/specs#111")

		tid := thread.NewIDV1(thread.Raw, 24)

		if logs := kb.LogsWithKeys(tid); len(logs) > 0 {
			t.Error("expected logs to be empty on init")
		}

		// Key small enough for inlining.
		_, pub, err := ic.GenerateKeyPair(ic.Ed25519, 256)
		if err != nil {
			t.Error(err)
		}

		id, err := peer.IDFromPublicKey(pub)
		if err != nil {
			t.Error(err)
		}

		pubKey := kb.LogPubKey(tid, id)
		if !pubKey.Equals(pub) {
			t.Error("mismatch between original public key and keybook-calculated one")
		}
	}
}

var logKeybookBenchmarkSuite = map[string]func(kb tstore.LogKeyBook) func(*testing.B){
	"LogPubKey":     benchmarkPubKey,
	"AddLogPubKey":  benchmarkAddPubKey,
	"LogPrivKey":    benchmarkPrivKey,
	"AddLogPrivKey": benchmarkAddPrivKey,
	"LogsWithKeys":  benchmarkLogsWithKeys,
}

func BenchmarkLogKeyBook(b *testing.B, factory LogKeyBookFactory) {
	ordernames := make([]string, 0, len(logKeybookBenchmarkSuite))
	for name := range logKeybookBenchmarkSuite {
		ordernames = append(ordernames, name)
	}
	sort.Strings(ordernames)
	for _, name := range ordernames {
		bench := logKeybookBenchmarkSuite[name]
		kb, closeFunc := factory()

		b.Run(name, bench(kb))

		if closeFunc != nil {
			closeFunc()
		}
	}
}

func benchmarkPubKey(kb tstore.LogKeyBook) func(*testing.B) {
	return func(b *testing.B) {
		tid := thread.NewIDV1(thread.Raw, 24)

		_, pub, err := pt.RandTestKeyPair(ic.RSA, 512)
		if err != nil {
			b.Error(err)
		}

		id, err := peer.IDFromPublicKey(pub)
		if err != nil {
			b.Error(err)
		}

		err = kb.AddLogPubKey(tid, id, pub)
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			kb.LogPubKey(tid, id)
		}
	}
}

func benchmarkAddPubKey(kb tstore.LogKeyBook) func(*testing.B) {
	return func(b *testing.B) {
		tid := thread.NewIDV1(thread.Raw, 24)

		_, pub, err := pt.RandTestKeyPair(ic.RSA, 512)
		if err != nil {
			b.Error(err)
		}

		id, err := peer.IDFromPublicKey(pub)
		if err != nil {
			b.Error(err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = kb.AddLogPubKey(tid, id, pub)
		}
	}
}

func benchmarkPrivKey(kb tstore.LogKeyBook) func(*testing.B) {
	return func(b *testing.B) {
		tid := thread.NewIDV1(thread.Raw, 24)

		priv, _, err := pt.RandTestKeyPair(ic.RSA, 512)
		if err != nil {
			b.Error(err)
		}

		id, err := peer.IDFromPrivateKey(priv)
		if err != nil {
			b.Error(err)
		}

		err = kb.AddLogPrivKey(tid, id, priv)
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			kb.LogPrivKey(tid, id)
		}
	}
}

func benchmarkAddPrivKey(kb tstore.LogKeyBook) func(*testing.B) {
	return func(b *testing.B) {
		tid := thread.NewIDV1(thread.Raw, 24)

		priv, _, err := pt.RandTestKeyPair(ic.RSA, 512)
		if err != nil {
			b.Error(err)
		}

		id, err := peer.IDFromPrivateKey(priv)
		if err != nil {
			b.Error(err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = kb.AddLogPrivKey(tid, id, priv)
		}
	}
}

func benchmarkLogsWithKeys(kb tstore.LogKeyBook) func(*testing.B) {
	return func(b *testing.B) {
		tid := thread.NewIDV1(thread.Raw, 24)
		for i := 0; i < 10; i++ {
			priv, pub, err := pt.RandTestKeyPair(ic.RSA, 512)
			if err != nil {
				b.Error(err)
			}

			id, err := peer.IDFromPublicKey(pub)
			if err != nil {
				b.Error(err)
			}

			err = kb.AddLogPubKey(tid, id, pub)
			if err != nil {
				b.Fatal(err)
			}
			err = kb.AddLogPrivKey(tid, id, priv)
			if err != nil {
				b.Fatal(err)
			}
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			kb.LogsWithKeys(tid)
		}
	}
}