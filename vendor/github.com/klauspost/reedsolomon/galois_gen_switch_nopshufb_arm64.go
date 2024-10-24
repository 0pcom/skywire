// Code generated by command: go generate gen.go. DO NOT EDIT.

//go:build !appengine && !noasm && gc && !nogen && nopshufb
// +build !appengine,!noasm,gc,!nogen,nopshufb

package reedsolomon

const (
	codeGen              = false
	codeGenMaxGoroutines = 16
	codeGenMaxInputs     = 10
	codeGenMaxOutputs    = 10
	minCodeGenSize       = 64
)

func (r *reedSolomon) hasCodeGen(byteCount int, inputs, outputs int) (_, _ *func(matrix []byte, in, out [][]byte, start, stop int) int, ok bool) {
	return nil, nil, false
}

func (r *reedSolomon) canGFNI(byteCount int, inputs, outputs int) (_, _ *func(matrix []uint64, in, out [][]byte, start, stop int) int, ok bool) {
	return nil, nil, false
}