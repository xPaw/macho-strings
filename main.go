package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
)

var (
	binaryOpt   = flag.String("binary", "", "the path to the binary you wish to parse")
	targetOpt   = flag.String("target", "", "the target type of the binary (macho/elf/pe)")
	demangleOpt = flag.Bool("demangle", true, "demangle C++ symbols into their original source identifiers")
	trimOpt     = flag.Bool("no-trim", false, "disable trimming whitespace and trailing newlines")
	humanOpt    = flag.Bool("no-human", false, "don't validate that it's a human readable string, this increases the amount of junk")
)

func ReadSection(reader *FileReader, section string) {
	sect := reader.ReaderParseSection(section)

	if sect != nil {
		nodes := reader.ReaderParseStrings(sect)

		for _, bytes := range nodes {
			str := string(bytes)

			if !*humanOpt {
				if !UtilIsNice(str) {
					continue
				}
			}

			if !*trimOpt {
				str = strings.TrimSpace(str)
				bad := []string{"\n", "\r"}
				for _, char := range bad {
					str = strings.Replace(str, char, "", -1)
				}
			}

			if *demangleOpt {
				demangled, err := UtilDemangle(&str)
				if err == nil {
					str = demangled
				}
			}

			fmt.Println(str)
		}
	}
}

func main() {
	flag.Parse()

	if *binaryOpt == "" {
		flag.PrintDefaults()
		return
	}

	r, err := NewFileReader(*binaryOpt, *targetOpt)
	if err != nil {
		log.Fatal(err.Error())
	}

	defer r.Close()

	var sections []string

	switch *targetOpt {
	case "macho":
		sections = []string{"__bss", "__const", "__cstring", "__cfstring", "__text", "__TEXT", "__objc_classname__TEXT"}
	case "elf":
		sections = []string{".dynstr", ".rodata", ".rdata", ".strtab", ".comment", ".note", ".stab", ".stabstr", ".note.ABI-tag", ".note.gnu.build-id"}
	case "pe":
		sections = []string{".data", ".rdata"}
	default:
		log.Fatal("Unknown target")
	}

	for _, section := range sections {
		ReadSection(r, section)
	}
}
