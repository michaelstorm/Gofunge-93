package main

import (
	"bytes";
	"flag";
	"fmt";
	"gofunge93";
	"os";
)

func ReadFile(name string) (b []byte, e os.Error) {
	var (
		raw []byte;
		err os.Error;
		file *os.File;
		dir *os.Dir;
	)
	
	file, err = os.Open(name, os.O_RDONLY, 0);
	if err != nil {
		return raw, err;
	}
	
	dir, err = file.Stat();
	if err != nil {
		return raw, err;
	}
	
	var size uint64 = dir.Size;
	raw = make([]byte, size);
	_, err = file.Read(raw);
	if err != nil {
		return raw, err;
	}
	
	return raw, err;
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] file\nOptions:\n", os.Args[0]);
		flag.PrintDefaults();
	};
	
	async := flag.Bool("async", false, "transparently execute instructions concurrently");
	coords := flag.Bool("coords", false, "print current instruction and coordinates to stderr");
	debug := flag.Bool("debug", false, "synonym for -coords -pause -stack -trace");
	help := flag.Bool("help", false, "show usage information");
	pause := flag.Bool("pause", false, "pause for return keypress between instructions");
	stack := flag.Bool("stack", false, "print stack contents to stderr ");
	trace := flag.Bool("trace", false, "print instruction pointer superimposed over code to stderr");
	flag.Parse();
	
	if *help {
		flag.Usage();
		return;
	}
	
	if *debug {
		*coords = true;
		*pause = true;
		*stack = true;
		*trace = true;
	}
	
	path := flag.Arg(0);
	if len(path) == 0 {
		fmt.Fprintln(os.Stderr, "No input file specified.");
		flag.Usage();
		return;
	}
	
	raw, err := ReadFile(path);
	if err != nil {
		fmt.Fprintln(os.Stderr, "File error:", err.String());
	}
	
	code := bytes.Split(raw, []byte{'\n'}, len(raw));
	
	space := make([][]byte, 25);
	for i, _ := range space {
		space[i] = make([]byte, 80);
		rowLen := 0;
		if i < len(code) {
			rowLen = len(code[i]);
			bytes.Copy(space[i], code[i]);
		}
		
		for j := 0; j < len(space[i])-rowLen; j++ {
			space[i][rowLen+j] = ' ';
		}
	}
	
	if *async {
		befunge := gofunge93.NewAsyncGofunge93(space);
		debugger := gofunge93.NewAsyncGofunge93Debugger(befunge);
		debugger.SetPause(*pause);
		debugger.SetPrintCoords(*coords);
		debugger.SetPrintStack(*stack);
		debugger.SetPrintTrace(*trace);
		gofunge93.StartDebug(debugger, gofunge93.NewIP2d(0, 0, [2]int8{1, 0}, 80, 25));
	} else {
		befunge := gofunge93.NewGofunge93(space);
		debugger := gofunge93.NewGofunge93Debugger(befunge);
		debugger.SetPause(*pause);
		debugger.SetPrintCoords(*coords);
		debugger.SetPrintStack(*stack);
		debugger.SetPrintTrace(*trace);
		gofunge93.StartDebug(debugger, gofunge93.NewIP2d(0, 0, [2]int8{1, 0}, 80, 25));
	}
}
