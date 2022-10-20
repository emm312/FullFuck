package main

/*
BSD 2-Clause License

Copyright (c) 2022, funnsam
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
   list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

import (
	"encoding/hex"
	"fmt"
	"os"
)

type Token struct {
	Lable  uint64 // Lable is for the register for the loop count
	ID     uint32 // ID is used for identifying tokens
	LoopID uint64 // LoopID is uniqe for each loop
}

var Uerr error
var InputFile []byte
var LoopID uint64
var LoopLayer uint64
var TokenList []Token
var ParsingHex uint8
var ParsingHexChars []uint8
var LoopLoopsTimes = []int16{0}
var t []uint8
var OutputFile []byte

var FullFuckToURCLTable = []string{
	"INC R1 R1\n",
	"DEC R1 R1\n",
	"OUT %d R1\n",
	"IN R1 %d\n",
	"BRZ .loop%d_e R1\nMOV R%d %d\n.loop%d\n",
	"DEC R%d R%d\nBNZ .loop%d R%d\n.loop%d_e\n",
}

func main() {
	if len(os.Args) <= 2 {
		fmt.Print("\x1b[1;31m>:(\n\x1b[1;0m")
		os.Exit(-1)
	}

	InputFile, Uerr = os.ReadFile(os.Args[1])
	checkUErr()

	for _, element := range InputFile {
		if ParsingHex == 1 || ParsingHex == 2 {
			ParsingHexChars = append(ParsingHexChars, element)
			ParsingHex++
			continue
		}
		switch element {
		case '+':
			TokenList = append(TokenList, Token{LoopLayer, 0, LoopID})
		case '-':
			TokenList = append(TokenList, Token{LoopLayer, 1, LoopID})
		case '>':
			if ParsingHex == 3 {
				t, Uerr = hex.DecodeString(string(ParsingHexChars))
				checkUErr()

				TokenList = append(TokenList, Token{LoopLayer, 2 | uint32(t[0])<<8, LoopID})

				ParsingHex = 0
				ParsingHexChars = make([]uint8, 0, 2)
			} else {
				TokenList = append(TokenList, Token{LoopLayer, 258, LoopID})
			}
		case '<':
			if ParsingHex == 3 {
				t, Uerr = hex.DecodeString(string(ParsingHexChars))
				checkUErr()

				TokenList = append(TokenList, Token{LoopLayer, 3 | uint32(t[0])<<8, LoopID})

				ParsingHex = 0
				ParsingHexChars = make([]uint8, 0, 2)
			} else {
				TokenList = append(TokenList, Token{LoopLayer, 259, LoopID})
			}
		case '0':
			ParsingHex = 1
		case '[':
			LoopID++
			LoopLayer++

			TokenList = append(TokenList, Token{LoopLayer, 4, LoopID})
			t, Uerr = hex.DecodeString(string(ParsingHexChars))
			checkUErr()

			if ParsingHex != 3 {
				LoopLoopsTimes = append(LoopLoopsTimes, -1)
			} else {
				LoopLoopsTimes = append(LoopLoopsTimes, int16(t[0]))
			}
			ParsingHex = 0
			ParsingHexChars = make([]uint8, 0, 2)
		case ']':
			TokenList = append(TokenList, Token{LoopLayer, 5, LoopID})
			LoopLayer--
		}
	}

	// Compile to URCL now
	for _, element := range TokenList {
		var resultAppend string
		switch element.ID & 0xFF {
		case 2:
			resultAppend = fmt.Sprintf(FullFuckToURCLTable[2], element.ID>>8)
		case 3:
			resultAppend = fmt.Sprintf(FullFuckToURCLTable[3], element.ID>>8)
		case 4:
			if LoopLoopsTimes[element.LoopID] == -1 {
				resultAppend = fmt.Sprintf(".loop%d\n", element.LoopID)
			} else {
				resultAppend = fmt.Sprintf(FullFuckToURCLTable[4], element.LoopID, element.Lable+1, LoopLoopsTimes[element.LoopID], element.LoopID)
			}
		case 5:
			if LoopLoopsTimes[element.LoopID] == -1 {
				resultAppend = fmt.Sprintf("JMP .loop%d\n", element.LoopID)
			} else {
				resultAppend = fmt.Sprintf(FullFuckToURCLTable[5], element.Lable+1, element.Lable+1, element.LoopID, element.Lable+1, element.LoopID)
			}
		default:
			resultAppend = FullFuckToURCLTable[element.ID]
		}
		OutputFile = append(OutputFile, []byte(resultAppend)...)
	}
	OutputFile = append(OutputFile, []byte("HLT\n")...)
	OutputFile = append(OutputFile, []byte("// This is generated by FullFuck\n")...)
	os.WriteFile(os.Args[2], []byte(OutputFile), 0664)
}

func checkUErr() {
	if Uerr != nil {
		panic(Uerr)
	}
}
