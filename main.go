package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

var opcodeIdentifier = map[string]string{
	"000101":                           "B",
	"10001010000":                      "AND",
	"10001011000":                      "ADD",
	"1001000100":                       "ADDI",
	"10101010000":                      "ORR",
	"10110100":                         "CBZ",
	"10110101":                         "CBNZ",
	"11001011000":                      "SUB",
	"1101000100":                       "SUBI",
	"110100101":                        "MOVZ",
	"111100101":                        "MOVK",
	"11010011010":                      "LSR",
	"11010011011":                      "LSL",
	"11111000000":                      "STUR",
	"11111000010":                      "LDUR",
	"11010011100":                      "ASR",
	"00000000000000000000000000000000": "NOP",
	"11101010000":                      "EOR",
	"11111110110111101111111111100111": "BREAK",
	"000000":                           " ",
}
var instructionIdentifier = map[string]string{
	"B":     "B",
	"AND":   "R",
	"ADD":   "R",
	"ADDI":  "I",
	"ORR":   "R",
	"CBZ":   "CB",
	"CBNZ":  "CB",
	"SUB":   "R",
	"SUBI":  "I",
	"MOVZ":  "IM",
	"MOVK":  "IM",
	"LSR":   "SHIFT",
	"LSL":   "SHIFT",
	"STUR":  "D",
	"LDUR":  "D",
	"ASR":   "SHIFT",
	"NOP":   " ",
	"EOR":   "R",
	"BREAK": "BREAK",
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getRegisterNumber(registerBits string) int {
	registerNumber, _ := strconv.ParseInt(registerBits, 2, 64)
	return int(registerNumber)
}

func getShiftAmount(shiftBits string) uint {
	shiftAmount, _ := strconv.ParseInt(shiftBits, 2, 64)
	return uint(shiftAmount) * 16
}

func simulate(instructionType string, args []string, registers map[int]int64, memory map[int]int64) bool {
	//debug attempt 1
	fmt.Println("Arguments passed to simulate:", args)
	switch instructionType {
	case "B":
		offset, _ := strconv.ParseInt(args[0], 2, 64)
		offset *= 4
		registers[15] += offset

	case "CB":
		offset, _ := strconv.ParseInt(args[1], 2, 64)
		conditionMet := registers[getRegisterNumber(args[0])] == 0
		if (args[2] == "0" && conditionMet) || (args[2] == "1" && !conditionMet) {
			registers[15] += offset * 4
		}

	case "R":
		if len(args) != 4 {
			fmt.Println("Error: invalid number of arguments for R-format instruction")
		}

		rd := getRegisterNumber(args[2])
		rn := getRegisterNumber(args[1])
		rm := getRegisterNumber(args[0])

		switch args[3] {
		case "0000": // AND
			registers[rd] = registers[rn] & registers[rm]
		case "0100": // ADD
			registers[rd] = registers[rn] + registers[rm]
		case "1100": // SUB
			registers[rd] = registers[rn] - registers[rm]
		case "0001": // EOR
			registers[rd] = registers[rn] ^ registers[rm]
		default:
			fmt.Printf("Unknown R-format instruction: %s\n", args[3])
		}

	case "IM":
		rd := getRegisterNumber(args[2])
		immediate, _ := strconv.ParseInt(args[0], 2, 64)
		shiftAmount := getShiftAmount(args[1])
		registers[rd] = immediate << shiftAmount

	case "SHIFT":
		rd := getRegisterNumber(args[2])
		rn := getRegisterNumber(args[1])
		shiftAmount, _ := strconv.ParseInt(args[0], 2, 64)

		switch args[3] {
		case "0010": // LSR
			registers[rd] = registers[rn] >> shiftAmount
		case "0011": // LSL
			registers[rd] = registers[rn] << shiftAmount
		case "1010": // ASR
			registers[rd] = registers[rn] >> shiftAmount
		default:
			fmt.Printf("Unknown SHIFT instruction: %s\n", args[3])
		}

	case "D":
		rd := getRegisterNumber(args[2])
		rn := getRegisterNumber(args[0])
		offset, _ := strconv.ParseInt(args[1], 2, 64)
		address := registers[rn] + offset
		memory[int(address)] = registers[rd]

	case "I":
		rd := getRegisterNumber(args[2])
		rn := getRegisterNumber(args[1])
		immediate, _ := strconv.ParseInt(args[0], 2, 64)
		registers[rd] = registers[rn] + immediate

	case " ": // NOP
		// NOP instruction does not affect the state, so nothing to simulate

	case "BREAK":
		// Implement simulation logic for BREAK instruction
		fmt.Println("BREAK instruction encountered. Simulation stopped.")
		return true

	default:
		fmt.Println("Unknown instruction type:", instructionType)
	}

	return false
}

func printState(simOutputFile *os.File, cycle int, instructionAddr int, instruction string, registers map[int]int64, memory map[int]int64) {
	simOutputFile.WriteString("====================\n")
	simOutputFile.WriteString(fmt.Sprintf("cycle: %d\tinstruction address: %d\tinstruction string: %s\n\n", cycle, instructionAddr, instruction))
	simOutputFile.WriteString("registers:\n")
	for i := 0; i < 32; i += 8 {
		simOutputFile.WriteString(fmt.Sprintf("r%02d:\t", i))
		for j := i; j < i+8; j++ {
			simOutputFile.WriteString(fmt.Sprintf("%d\t", registers[j]))
		}
		simOutputFile.WriteString("\n")
	}
	simOutputFile.WriteString("\n")
	simOutputFile.WriteString("data:\n")
	// Assuming each data block is 8 words
	for address, value := range memory {
		if address%8 == 0 {
			simOutputFile.WriteString(fmt.Sprintf("%d:\t", address))
		}
		simOutputFile.WriteString(fmt.Sprintf("%d\t", value))
		if address%8 == 7 {
			simOutputFile.WriteString("\n")
		}
	}
	simOutputFile.WriteString("\n")
}

func main() {
	var inputFileName *string
	var outputFileName *string
	var simOutputFileName *string // New flag for simulation output file

	// Define flags
	inputFileName = flag.String("i", "", "Input file name")
	outputFileName = flag.String("o", "", "Output file name")
	simOutputFileName = flag.String("s", "", "Simulation output file name") // New flag for simulation output file

	// Parse the command-line arguments
	flag.Parse()

	// Check if both input and output files were not provided
	if *inputFileName == "" && *outputFileName == "" {
		fmt.Println("Both input and output file names are required. Use -i and -o flags.")
		return
	}

	*outputFileName = *outputFileName + "_dis.txt"
	if *simOutputFileName == "" {
		*simOutputFileName = *outputFileName + "_sim.txt"
	} else {
		*simOutputFileName = *simOutputFileName + "_sim.txt"
	}

	// Open input file
	inputFile, err := os.Open(*inputFileName)
	if err != nil {
		fmt.Println("Error opening input file:", err)
		return
	}
	defer inputFile.Close()

	// Open output file
	outputFile, err := os.Create(*outputFileName)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	// Open simulation output file
	simOutputFile, err := os.Create(*simOutputFileName)
	if err != nil {
		fmt.Println("Error creating simulation output file:", err)
		return
	}
	defer simOutputFile.Close()

	memLocation := 96
	breakFound := false
	cycle := 1 // New variable to track cycles

	registers := make(map[int]int64, 32)
	memory := make(map[int]int64)

	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		line := scanner.Text()
		binaryNumber := strings.TrimSpace(line)
		if len(binaryNumber) != 32 {
			invalidString := fmt.Sprintf("%.32s Invalid binary string! \n", binaryNumber)
			outputFile.WriteString(invalidString)
			continue
		}
	}

	//testing if statements
	if breakFound {
		decimalVal, _ := strconv.ParseInt(binaryNumber, 2, 64)
		if binaryNumber[0] == '1' {
			lead1 := int64(math.Pow(2, 32))
			decimalVal -= lead1
		}
		afterBreak := fmt.Sprintf("%.32s \t\t%d\t%d \n", binaryNumber, memLocation, decimalVal)
		outputFile.WriteString(afterBreak)
	} else if instructionIdentifier[opcodeIdentifier[binaryNumber[0:11]]] == "R" {
		opcode := binaryNumber[:11]
		rm := binaryNumber[11:16]
		shamt := binaryNumber[16:22]
		rn := binaryNumber[22:27]
		rd := binaryNumber[27:32]
		r2, _ := strconv.ParseInt(rm, 2, 64)
		r1, _ := strconv.ParseInt(rn, 2, 64)
		r3, _ := strconv.ParseInt(rd, 2, 64)

		// Update register values
		if opcodeIdentifier[opcode] == "ADD" {
			registers[r3] = registers[r1] + registers[r2]
		} else if opcodeIdentifier[opcode] == "SUB" {
			registers[r3] = registers[r1] - registers[r2]
		} else if opcodeIdentifier[opcode] == "AND" {
			registers[r3] = registers[r1] & registers[r2]
		} else if opcodeIdentifier[opcode] == "ORR" {
			registers[r3] = registers[r1] | registers[r2]
		} else if opcodeIdentifier[opcode] == "EOR" {
			registers[r3] = registers[r1] ^ registers[r2]

			// Print R-format
			rFormat := fmt.Sprintf("%.11s %.5s %.6s %.5s %.5s \t%d\t%s\tR%d, R%d, R%d \n",
				opcode, rm, shamt, rn, rd, memLocation, opcodeIdentifier[opcode], r3, r1, r2)
			outputFile.WriteString(rFormat)

		} else if instructionIdentifier[opcodeIdentifier[binaryNumber[0:11]]] == "SHIFT" {
			opcode := binaryNumber[:11]
			rm := binaryNumber[11:16]
			shamt := binaryNumber[16:22]
			rn := binaryNumber[22:27]
			rd := binaryNumber[27:32]
			r1, _ := strconv.ParseInt(rn, 2, 64)
			r3, _ := strconv.ParseInt(rd, 2, 64)
			sh, _ := strconv.ParseInt(shamt, 2, 64)

			// Update register values
			if opcodeIdentifier[opcode] == "LSL" {
				registers[r3] = registers[r1] << sh
			} else if opcodeIdentifier[opcode] == "LSR" {
				registers[r3] = registers[r1] >> sh
			} else if opcodeIdentifier[opcode] == "ASR" {
				registers[r3] = int64(int32(registers[r1]) >> sh)
			}

			// Print SHIFT-format
			lsFormat := fmt.Sprintf("%.11s %.5s %.6s %.5s %.5s \t%d\t%s\tR%d, R%d, #%d \n",
				opcode, rm, shamt, rn, rd, memLocation, opcodeIdentifier[opcode], r3, r1, sh)
			outputFile.WriteString(lsFormat)

		} else if instructionIdentifier[opcodeIdentifier[binaryNumber[0:11]]] == "D" {
			opcode := binaryNumber[:11]
			address := binaryNumber[11:20]
			op2 := binaryNumber[20:22]
			rn := binaryNumber[22:27]
			rt := binaryNumber[27:32]
			r2, _ := strconv.ParseInt(address, 2, 64)
			r1, _ := strconv.ParseInt(rn, 2, 64)
			r3, _ := strconv.ParseInt(rt, 2, 64)

			// Update register values
			if opcodeIdentifier[opcode] == "LDR" {
				registers[r3] = memory[r1+r2]
			} else if opcodeIdentifier[opcode] == "STR" {
				memory[r1+r2] = registers[r3]
			}

			// Print D-format
			dFormat := fmt.Sprintf("%.11s %.9s %.2s %.5s %.5s \t%d\t%s\tR%d, [R%d, #%d] \n",
				opcode, address, op2, rn, rt, memLocation, opcodeIdentifier[opcode], r3, r1, r2)
			outputFile.WriteString(dFormat)

		} else if instructionIdentifier[opcodeIdentifier[binaryNumber[0:10]]] == "I" {
			opcode := binaryNumber[:10]
			immediate := binaryNumber[10:22]
			rn := binaryNumber[22:27]
			rd := binaryNumber[27:32]
			r2, _ := strconv.ParseInt(immediate, 2, 64)
			r1, _ := strconv.ParseInt(rn, 2, 64)
			r3, _ := strconv.ParseInt(rd, 2, 64)

			//convert offset to decimal in 2's complement
			if binaryNumber[10] == '1' {
				lead1 := int64(math.Pow(2, 12))
				r2 -= lead1
			}

			// Update register values
			if opcodeIdentifier[opcode] == "ADDI" {
				registers[r3] = registers[r1] + r2
			} else if opcodeIdentifier[opcode] == "SUBI" {
				registers[r3] = registers[r1] - r2

				// Print I-format
				iFormat := fmt.Sprintf("%.10s %.12s %.5s %.5s \t%d\t%s\tR%d, R%d, #%d \n",
					opcode, immediate, rn, rd, memLocation, opcodeIdentifier[opcode], r3, r1, r2)
				outputFile.WriteString(iFormat)

			} else if instructionIdentifier[opcodeIdentifier[binaryNumber[0:6]]] == "B" {
				opcode := binaryNumber[:6]
				offset := binaryNumber[6:32]
				r2, _ := strconv.ParseInt(offset, 2, 64)

				//convert offset to decimal in 2's complement
				if binaryNumber[6] == '1' {
					lead1 := int64(math.Pow(2, 26))
					r2 -= lead1
				}

				// Update program counter
				if opcodeIdentifier[opcode] == "B" {
					cycle += r2
				}

				// Print B-format
				bFormat := fmt.Sprintf("%.6s %.26s \t%d\t%s\t#%d \n",
					opcode, offset, memLocation, opcodeIdentifier[opcode], r2)
				outputFile.WriteString(bFormat)

			} else if instructionIdentifier[opcodeIdentifier[binaryNumber[0:8]]] == "CB" {
				opcode := binaryNumber[:8]
				offset := binaryNumber[8:27]
				conditional := binaryNumber[27:32]
				r2, _ := strconv.ParseInt(offset, 2, 64)
				r1, _ := strconv.ParseInt(conditional, 2, 64)

				//convert offset to decimal in 2's complement
				if binaryNumber[8] == '1' {
					lead1 := int64(math.Pow(2, 19))
					r2 -= lead1
				}

				// Update cycle counter
				if opcodeIdentifier[opcode] == "CBNZ" && registers[r1] != 0 {
					cycle += r2
				} else if opcodeIdentifier[opcode] == "CBZ" && registers[r1] == 0 {
					cycle += r2
				}

				// Print CB-format
				cbFormat := fmt.Sprintf("%.8s %.19s %.5s \t%d\t%s\tR%d, #%d \n",
					opcode, offset, conditional, memLocation, opcodeIdentifier[opcode], r1, r2)
				outputFile.WriteString(cbFormat)

			} else if instructionIdentifier[opcodeIdentifier[binaryNumber[0:9]]] == "IM" {
				opcode := binaryNumber[:9]
				shiftCode := binaryNumber[9:11]
				field := binaryNumber[11:27]
				rd := binaryNumber[27:32]
				im_shift, _ := strconv.ParseInt(shiftCode, 2, 64)
				shifted := im_shift * 16
				im_field, _ := strconv.ParseInt(field, 2, 64)
				im_rd, _ := strconv.ParseInt(rd, 2, 64)

				// Update register values
				if opcodeIdentifier[opcode] == "MOVZ" {
					registers[im_rd] = im_field << shifted
				} else if opcodeIdentifier[opcode] == "MOVK" {
					mask := int64(0xFFFF) << shifted
					registers[im_rd] = (registers[im_rd] &^ mask) | (im_field << shifted)
				}

				// Print IM-format
				imFormat := fmt.Sprintf("%.9s %.2s %.16s %.5s \t%d\t%s\tR%d, %d, LSL %d \n",
					opcode, shiftCode, field, rd, memLocation, opcodeIdentifier[opcode], im_rd, im_field, shifted)
				outputFile.WriteString(imFormat)

			} else if opcodeIdentifier[binaryNumber] == "NOP" {
				// Print NOP-format
				breakFormat := fmt.Sprintf("%.32s \t\t%d\t%s \n", binaryNumber, memLocation, opcodeIdentifier[binaryNumber])
				outputFile.WriteString(breakFormat)

			} else if opcodeIdentifier[binaryNumber] == "BREAK" {
				// Print BREAK-format
				breakFormat := fmt.Sprintf("%.1s %.5s %.5s %.5s %.5s %.5s %.6s \t%d\t%s \n", binaryNumber[:1],
					binaryNumber[1:6], binaryNumber[6:11], binaryNumber[11:16], binaryNumber[16:21], binaryNumber[21:26],
					binaryNumber[26:32], memLocation, opcodeIdentifier[binaryNumber])
				breakFound = true
				outputFile.WriteString(breakFormat)

			} else {
				opcodeUnknown := fmt.Sprintf("%.8s %.3s %.5s %.5s %.5s %.6s \tUnknown Instruction!\n", binaryNumber[:8], binaryNumber[8:11],
					binaryNumber[11:16], binaryNumber[16:21], binaryNumber[21:26], binaryNumber[26:32])
				outputFile.WriteString(opcodeUnknown)
			}

			simOutput := fmt.Sprintf("=====================\ncycle:%d %d %s\nregisters:\n", cycle, memLocation, opcodeIdentifier[binaryNumber[0:11]])
			simOutputFile.WriteString(simOutput)

			// Print simulation state
			printState(simOutputFile, cycle, memLocation, opcodeIdentifier[binaryNumber[0:11]], registers, memory)

			memLocation += 4
			cycle++
		}

		// Close the simulation output file
		simOutputFile.Close()
	}
}
