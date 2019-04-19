// Copyright (c) 2018-2019 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package txscript

import (
	"bytes"
	"testing"

	"github.com/btcsuite/btcd/wire"
)

var (
	// manyInputsBenchTx is a transaction that contains a lot of inputs which is
	// useful for benchmarking signature hash calculation.
	manyInputsBenchTx wire.MsgTx

	// A mock previous output script to use in the signing benchmark.
	prevOutScript = hexToBytes("a914f5916158e3e2c4551c1796708db8367207ed13bb87")
)

func init() {
	/*
		// tx 620f57c92cf05a7f7e7f7d28255d5f7089437bc48e34dcfebf7751d08b7fb8f5
		txHex, err := ioutil.ReadFile("data/many_inputs_tx.hex")
		if err != nil {
			panic(fmt.Sprintf("unable to read benchmark tx file: %v", err))
		}

		txBytes := hexToBytes(string(txHex))
		err = manyInputsBenchTx.Deserialize(bytes.NewReader(txBytes))
		if err != nil {
			panic(err)
		}
	*/
}

// BenchmarkCalcSigHash benchmarks how long it takes to calculate the signature
// hashes for all inputs of a transaction with many inputs.
func BenchmarkCalcSigHash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(manyInputsBenchTx.TxIn); j++ {
			_, err := CalcSignatureHash(prevOutScript, SigHashAll,
				&manyInputsBenchTx, j)
			if err != nil {
				b.Fatalf("failed to calc signature hash: %v", err)
			}
		}
	}
}

// genComplexScript returns a script comprised of half as many opcodes as the
// maximum allowed followed by as many max size data pushes fit without
// exceeding the max allowed script size.
func genComplexScript() ([]byte, error) {
	var scriptLen int
	builder := NewScriptBuilder()
	for i := 0; i < MaxOpsPerScript/2; i++ {
		builder.AddOp(OP_TRUE)
		scriptLen++
	}
	maxData := bytes.Repeat([]byte{0x02}, MaxScriptElementSize)
	for i := 0; i < (MaxScriptSize-scriptLen)/MaxScriptElementSize; i++ {
		builder.AddData(maxData)
	}
	return builder.Script()
}

// BenchmarkScriptParsing benchmarks how long it takes to parse a very large
// script.
func BenchmarkScriptParsing(b *testing.B) {
	script, err := genComplexScript()
	if err != nil {
		b.Fatalf("failed to create benchmark script: %v", err)
	}

	const scriptVersion = 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tokenizer := MakeScriptTokenizer(scriptVersion, script)
		for tokenizer.Next() {
			_ = tokenizer.Opcode()
			_ = tokenizer.Data()
			_ = tokenizer.ByteIndex()
		}
		if err := tokenizer.Err(); err != nil {
			b.Fatalf("failed to parse script: %v", err)
		}
	}
}

// BenchmarkDisasmString benchmarks how long it takes to disassemble a very
// large script.
func BenchmarkDisasmString(b *testing.B) {
	script, err := genComplexScript()
	if err != nil {
		b.Fatalf("failed to create benchmark script: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := DisasmString(script)
		if err != nil {
			b.Fatalf("failed to disasm script: %v", err)
		}
	}
}

// BenchmarkIsPayToScriptHash benchmarks how long it takes IsPayToScriptHash to
// analyze a very large script.
func BenchmarkIsPayToScriptHash(b *testing.B) {
	script, err := genComplexScript()
	if err != nil {
		b.Fatalf("failed to create benchmark script: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsPayToScriptHash(script)
	}
}

// BenchmarkIsMultisigScriptLarge benchmarks how long it takes IsMultisigScript
// to analyze a very large script.
func BenchmarkIsMultisigScriptLarge(b *testing.B) {
	script, err := genComplexScript()
	if err != nil {
		b.Fatalf("failed to create benchmark script: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isMultisig, err := IsMultisigScript(script)
		if err != nil {
			b.Fatalf("unexpected err: %v", err)
		}
		if isMultisig {
			b.Fatalf("script should NOT be reported as mutisig script")
		}
	}
}

// BenchmarkIsMultisigScript benchmarks how long it takes IsMultisigScript to
// analyze a 1-of-2 multisig public key script.
func BenchmarkIsMultisigScript(b *testing.B) {
	multisigShortForm := "1 " +
		"DATA_33 " +
		"0x030478aaaa2be30772f1e69e581610f1840b3cf2fe7228ee0281cd599e5746f81e " +
		"DATA_33 " +
		"0x0284f4d078b236a9ff91661f8ffbe012737cd3507566f30fd97d25f2b23539f3cd " +
		"2 CHECKMULTISIG"
	pkScript := mustParseShortForm(multisigShortForm)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isMultisig, err := IsMultisigScript(pkScript)
		if err != nil {
			b.Fatalf("unexpected err: %v", err)
		}
		if !isMultisig {
			b.Fatalf("script should be reported as a mutisig script")
		}
	}
}

// BenchmarkIsMultisigSigScript benchmarks how long it takes IsMultisigSigScript
// to analyze a very large script.
func BenchmarkIsMultisigSigScriptLarge(b *testing.B) {
	script, err := genComplexScript()
	if err != nil {
		b.Fatalf("failed to create benchmark script: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if IsMultisigSigScript(script) {
			b.Fatalf("script should NOT be reported as mutisig sig script")
		}
	}
}

// BenchmarkIsMultisigSigScript benchmarks how long it takes IsMultisigSigScript
// to analyze both a 1-of-2 multisig public key script (which should be false)
// and a signature script comprised of a pay-to-script-hash 1-of-2 multisig
// redeem script (which should be true).
func BenchmarkIsMultisigSigScript(b *testing.B) {
	multisigShortForm := "1 " +
		"DATA_33 " +
		"0x030478aaaa2be30772f1e69e581610f1840b3cf2fe7228ee0281cd599e5746f81e " +
		"DATA_33 " +
		"0x0284f4d078b236a9ff91661f8ffbe012737cd3507566f30fd97d25f2b23539f3cd " +
		"2 CHECKMULTISIG"
	pkScript := mustParseShortForm(multisigShortForm)

	sigHex := "0x304402205795c3ab6ba11331eeac757bf1fc9c34bef0c7e1a9c8bd5eebb8" +
		"82f3b79c5838022001e0ab7b4c7662e4522dc5fa479e4b4133fa88c6a53d895dc1d5" +
		"2eddc7bbcf2801 "
	sigScript := mustParseShortForm("DATA_71 " + sigHex + "DATA_71 " +
		multisigShortForm)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if IsMultisigSigScript(pkScript) {
			b.Fatalf("script should NOT be reported as mutisig sig script")
		}
		if !IsMultisigSigScript(sigScript) {
			b.Fatalf("script should be reported as a mutisig sig script")
		}
	}
}

// BenchmarkGetSigOpCount benchmarks how long it takes to count the signature
// operations of a very large script.
func BenchmarkGetSigOpCount(b *testing.B) {
	script, err := genComplexScript()
	if err != nil {
		b.Fatalf("failed to create benchmark script: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetSigOpCount(script)
	}
}

// BenchmarkGetPreciseSigOpCount benchmarks how long it takes to count the
// signature operations of a very large script using the more precise counting
// method.
func BenchmarkGetPreciseSigOpCount(b *testing.B) {
	redeemScript, err := genComplexScript()
	if err != nil {
		b.Fatalf("failed to create benchmark script: %v", err)
	}

	// Create a fake pay-to-script-hash to pass the necessary checks and create
	// the signature script accordingly by pushing the generated "redeem" script
	// as the final data push so the benchmark will cover the p2sh path.
	scriptHash := "0x0000000000000000000000000000000000000001"
	pkScript := mustParseShortForm("HASH160 DATA_20 " + scriptHash + " EQUAL")
	sigScript, err := NewScriptBuilder().AddFullData(redeemScript).Script()
	if err != nil {
		b.Fatalf("failed to create signature script: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetPreciseSigOpCount(sigScript, pkScript, true)
	}
}

// BenchmarkIsPushOnlyScript benchmarks how long it takes IsPushOnlyScript to
// analyze a very large script.
func BenchmarkIsPushOnlyScript(b *testing.B) {
	script, err := genComplexScript()
	if err != nil {
		b.Fatalf("failed to create benchmark script: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsPushOnlyScript(script)
	}
}

// BenchmarkGetScriptClass benchmarks how long it takes GetScriptClass to
// analyze a very large script.
func BenchmarkGetScriptClass(b *testing.B) {
	script, err := genComplexScript()
	if err != nil {
		b.Fatalf("failed to create benchmark script: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetScriptClass(script)
	}
}

// BenchmarkIsPubKeyScript benchmarks how long it takes to analyze a very large
// script to determine if it is a standard pay-to-pubkey script.
func BenchmarkIsPubKeyScript(b *testing.B) {
	script, err := genComplexScript()
	if err != nil {
		b.Fatalf("failed to create benchmark script: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isPubKeyScript(script)
	}
}

// BenchmarkIsPubKeyHashScript benchmarks how long it takes to analyze a very
// large script to determine if it is a standard pay-to-pubkey-hash script.
func BenchmarkIsPubKeyHashScript(b *testing.B) {
	script, err := genComplexScript()
	if err != nil {
		b.Fatalf("failed to create benchmark script: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isPubKeyHashScript(script)
	}
}

// BenchmarkIsNullDataScript benchmarks how long it takes to analyze a very
// large script to determine if it is a standard nulldata script.
func BenchmarkIsNullDataScript(b *testing.B) {
	script, err := genComplexScript()
	if err != nil {
		b.Fatalf("failed to create benchmark script: %v", err)
	}

	const scriptVersion = 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isNullDataScript(scriptVersion, script)
	}
}

// BenchmarkIsWitnessPubKeyHash benchmarks how long it takes to analyze a very
// large script to determine if it is a standard witness pubkey hash script.
func BenchmarkIsWitnessPubKeyHash(b *testing.B) {
	script, err := genComplexScript()
	if err != nil {
		b.Fatalf("failed to create benchmark script: %v", err)
	}

	const scriptVersion = 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsPayToWitnessPubKeyHash(script)
	}
}
