package vm

//go:generate go run golang.org/x/tools/cmd/stringer -type=Opcode,GPR,Privilege,Priority,offset,literal,vector -output=strings_gen.go -linecomment
