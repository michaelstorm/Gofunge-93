package main

import (
	"bytes";
	"container/vector";
	"flag";
	"fmt";
	"os";
	"rand";
	"strconv";
)

type IP interface {
	Dim(uint8) int32;
	Delta(uint8) int8;
	Go(uint8, int8);
	Tick();
}

type IP2d struct {
	x int32;
	y int32;
	length int32;
	height int32;
	delta [2]int8;
}

func NewIP2d(x int32, y int32, delta [2]int8, length int32, height int32) *IP2d {
	ip := new(IP2d);
	ip.x = x;
	ip.y = y;
	ip.delta = delta;
	ip.length = length;
	ip.height = height;
	return ip;
}

func (ip *IP2d) Dim(axis uint8) int32 {
	if axis == 0 {
		return ip.x;
	}
	return ip.y;
}

func (ip *IP2d) Delta(axis uint8) int8 {
	return ip.delta[axis];
}

func (ip *IP2d) Go(axis uint8, direction int8) {
	if axis == 0 {
		ip.delta = [2]int8{direction, 0};
	}
	else {
		ip.delta = [2]int8{0, direction};
	}
}

func (ip *IP2d) Tick() {
	ip.x += int32(ip.delta[0]);
	ip.y += int32(ip.delta[1]);

	if ip.x >= ip.length {
		ip.x = 0;
	}
	if ip.y >= ip.height {
		ip.y = 0;
	}
	if ip.x < 0 {
		ip.x = ip.length-1;
	}
	if ip.y < 0 {
		ip.y = ip.height-1;
	}
}

type Stack struct {
	v *vector.Vector;
}

func (stack *Stack) Push(val int64) {
	stack.v.Push(val);
}

func (stack *Stack) Pop() int64 {
	if stack.v.Len() == 0 {
		return 0;
	}
	
	return stack.v.Pop().(int64);
}

func (stack *Stack) Top() int64 {
	if stack.v.Len() == 0 {
		return 0;
	}
	
	return stack.v.Last().(int64);
}

type ChanStack struct {
	v *vector.Vector;
}

func (stack *ChanStack) Push(val <-chan int64) {
	stack.v.Push(val);
}

func (stack *ChanStack) Pop() <-chan int64 {
	if stack.v.Len() == 0 {
		c := make(chan int64, 1);
		c <- 0;
		return c;
	}
	
	return stack.v.Pop().(<-chan int64);
}

func (stack *ChanStack) Top() <-chan int64 {
	if stack.v.Len() == 0 {
		c := make(chan int64, 1);
		c <- 0;
		return c;
	}
	
	return stack.v.Last().(<-chan int64);
}

type Interpreter interface {
	Execute(IP) bool;
}

type Debuggable interface {
	Interpreter;
	
	Debug(IP);
	SetPause(bool);
	SetPrintCoords(bool);
	SetPrintStack(bool);
	SetPrintTrace(bool);
}

type Gofunge93 struct {
	code [][]byte;
	stack *Stack;
	stringMode bool;
}

func NewGofunge93(code [][]byte) *Gofunge93 {
	b := new(Gofunge93);
	b.code = code;
	b.stack = &Stack{vector.New(0)};
	return b;
}

func (b *Gofunge93) Execute(ip IP) bool {
	inst := b.code[ip.Dim(1)][ip.Dim(0)];
	
	switch {
	case inst == '"':
		b.stringMode = !b.stringMode;
	case b.stringMode:
		b.stack.Push(int64(inst));
	case '0' <= inst && inst <= '9':
		b.stack.Push(int64(inst-'0'));
	
	default:
		switch inst {
		case 'v':
			ip.Go(1, 1);
		case '^':
			ip.Go(1, -1);
		case '>':
			ip.Go(0, 1);
		case '<':
			ip.Go(0, -1);
		case '?':
			axis := uint8(rand.Intn(2));
			direction := int8(rand.Intn(2));
			if direction == 0 {
				direction = -1;
			}
			ip.Go(axis, direction);
		
		case '+':
			val1 := b.stack.Pop();
			val2 := b.stack.Pop();
			b.stack.Push(val2+val1);
		case '-':
			val1 := b.stack.Pop();
			val2 := b.stack.Pop();
			b.stack.Push(val2-val1);
		case '*':
			val1 := b.stack.Pop();
			val2 := b.stack.Pop();
			b.stack.Push(val2*val1);
		case '/':
			val1 := b.stack.Pop();
			val2 := b.stack.Pop();
			b.stack.Push(val2/val1);
		case '%':
			val1 := b.stack.Pop();
			val2 := b.stack.Pop();
			b.stack.Push(val2%val1);
		case '!':
			if b.stack.Pop() == 0 {
				b.stack.Push(1);
			} else {
				b.stack.Push(0);
			}
		case '`':
			val1 := b.stack.Pop();
			val2 := b.stack.Pop();
			if val2 > val1 {
				b.stack.Push(1);
			} else {
				b.stack.Push(0);
			}
		case '_':
			if b.stack.Pop() != 0 {
				ip.Go(0, -1);
			} else {
				ip.Go(0, 1);
			}
		case '|':
			if b.stack.Pop() != 0 {
				ip.Go(1, -1);
			} else {
				ip.Go(1, 1);
			}
		case ':':
			b.stack.Push(b.stack.Top());
		case '\\':
			val1 := b.stack.Pop();
			val2 := b.stack.Pop();
			b.stack.Push(val1);
			b.stack.Push(val2);
		case '$':
			b.stack.Pop();
		case '.':
			fmt.Printf("%d", b.stack.Pop());
		case ',':
			fmt.Printf("%c", b.stack.Pop());
		case '#':
			ip.Tick();
		case 'g':
			y := b.stack.Pop();
			x := b.stack.Pop();
			b.stack.Push(int64(b.code[y][x]));
		case 'p':
			y := b.stack.Pop();
			x := b.stack.Pop();
			val := b.stack.Pop();
			b.code[y][x] = byte(val);
		case '&':
			buf := make([]byte, 32);
			bufLen, err := os.Stdin.Read(buf);
			if err != nil {
				fmt.Fprintln(os.Stdout, "Could not read from standard input:", err.String());
				return false;
			}
			
			if i, err := strconv.Atoi64(string(buf[0:bufLen-1])); err == nil {
				b.stack.Push(i);
			}
			else {
				fmt.Fprintln(os.Stdout, "Bad int conversion:", err.String());
				return false;
			}
		case '~':
			buf := make([]byte, 1);
			if _, err := os.Stdin.Read(buf); err != nil {
				fmt.Fprintln(os.Stdout, "Could not read from standard input:", err.String());
				return false;
			}
			b.stack.Push(int64(buf[0]));
		case '@':
			return false;
		}
	}
	return true;
}

type Gofunge93Debugger struct {
	*Gofunge93;
	pause bool;
	printCoords bool;
	printStack bool;
	printTrace bool;
}

func NewGofunge93Debugger(g *Gofunge93) *Gofunge93Debugger {
	debugger := new(Gofunge93Debugger);
	debugger.Gofunge93 = g;
	return debugger;
}

func (d Gofunge93Debugger) Debug(ip IP) {
	inst := d.Gofunge93.code[ip.Dim(1)][ip.Dim(0)];
	if d.printCoords {
		fmt.Fprintf(os.Stderr, "\n%c (%v, %v)", inst, ip.Dim(0), ip.Dim(1));
	}
	if d.printStack {
		fmt.Fprintf(os.Stderr, "\n%v", d.Gofunge93.stack.v.Data());
	}
	if d.printTrace {
		fmt.Fprint(os.Stderr, "\n");
		for i, row := range d.Gofunge93.code {
			if int32(i) == ip.Dim(1) {
				fmt.Fprint(os.Stderr, string(row[0:ip.Dim(0)]));
				fmt.Fprint(os.Stderr, "█");
				fmt.Fprintln(os.Stderr, string(row[ip.Dim(0)+1:len(row)]));
			}
			else {
				fmt.Fprintln(os.Stderr, string(row));
			}
		}
	}
	if d.pause {
		os.Stdin.Read(make([]byte, 1));
	}
}

func (d *Gofunge93Debugger) SetPause(b bool) { d.pause = b }
func (d *Gofunge93Debugger) SetPrintCoords(b bool) { d.printCoords = b }
func (d *Gofunge93Debugger) SetPrintStack(b bool) { d.printStack = b }
func (d *Gofunge93Debugger) SetPrintTrace(b bool) { d.printTrace = b }

type AsyncGofunge93 struct {
	code [][]byte;
	stack *ChanStack;
	stringMode bool;
}

func NewAsyncGofunge93(code [][]byte) *AsyncGofunge93 {
	b := new(AsyncGofunge93);
	b.code = code;
	b.stack = &ChanStack{vector.New(0)};
	return b;
}

func (b *AsyncGofunge93) Execute(ip IP) bool {
	inst := b.code[ip.Dim(1)][ip.Dim(0)];
	
	switch {
	case inst == '"':
		b.stringMode = !b.stringMode;
	case b.stringMode:
		c := make(chan int64, 1);
		c <- int64(inst);
		b.stack.Push(c);
	case '0' <= inst && inst <= '9':
		c := make(chan int64, 1);
		c <- int64(inst-'0');
		b.stack.Push(c);
	
	default:
		switch inst {
		case 'v':
			ip.Go(1, 1);
		case '^':
			ip.Go(1, -1);
		case '>':
			ip.Go(0, 1);
		case '<':
			ip.Go(0, -1);
		case '?':
			axis := uint8(rand.Intn(2));
			direction := int8(rand.Intn(2));
			if direction == 0 {
				direction = -1;
			}
			ip.Go(axis, direction);
		case '_':
			if (<-b.stack.Pop()) != 0 {
				ip.Go(0, -1);
			} else {
				ip.Go(0, 1);
			}
		case '|':
			if (<-b.stack.Pop()) != 0 {
				ip.Go(1, -1);
			} else {
				ip.Go(1, 1);
			}
		
		case '+':
			in1 := b.stack.Pop();
			in2 := b.stack.Pop();
			c := make(chan int64, 1);
			b.stack.Push(c);
			
			go func(in1, in2 <-chan int64, out chan<- int64) {
				out <- (<-in2)+(<-in1);
			}(in1, in2, c);
		case '-':
			in1 := b.stack.Pop();
			in2 := b.stack.Pop();
			c := make(chan int64, 1);
			b.stack.Push(c);
			
			go func(in1, in2 <-chan int64, out chan<- int64) {
				out <- (<-in2)-(<-in1);
			}(in1, in2, c);
		case '*':
			in1 := b.stack.Pop();
			in2 := b.stack.Pop();
			c := make(chan int64, 1);
			b.stack.Push(c);
			
			go func(in1, in2 <-chan int64, out chan<- int64) {
				out <- (<-in2)*(<-in1);
			}(in1, in2, c);
		case '/':
			in1 := b.stack.Pop();
			in2 := b.stack.Pop();
			c := make(chan int64, 1);
			b.stack.Push(c);
			
			go func(in1, in2 <-chan int64, out chan<- int64) {
				out <- (<-in2)/(<-in1);
			}(in1, in2, c);
		case '%':
			in1 := b.stack.Pop();
			in2 := b.stack.Pop();
			c := make(chan int64, 1);
			b.stack.Push(c);
			
			go func(in1, in2 <-chan int64, out chan<- int64) {
				out <- (<-in2)%(<-in1);
			}(in1, in2, c);
		case '!':
			in := b.stack.Pop();
			c := make(chan int64, 1);
			b.stack.Push(c);
			
			go func(in <-chan int64, out chan<- int64) {
				if (<-in) == 0 {
					out <- 1;
				} else {
					out <- 0;
				}
			}(in, c);
		case '`':
			in2 := b.stack.Pop();
			in1 := b.stack.Pop();
			c := make(chan int64, 1);
			b.stack.Push(c);
			
			go func(in1, in2 <-chan int64, out chan<- int64) {
				if (<-in2) > (<-in1) {
					out <- 1;
				} else {
					out <- 0;
				}
			}(in1, in2, c);
		case ':':
			in := b.stack.Pop();
			c1 := make(chan int64, 1);
			c2 := make(chan int64, 1);
			b.stack.Push(c2);
			b.stack.Push(c1);
			
			go func(in <-chan int64, out1, out2 chan<- int64) {
				val := <-in;
				out1 <- val;
				out2 <- val;
			}(in, c1, c2);
		case '\\':
			c2 := b.stack.Pop();
			c1 := b.stack.Pop();
			b.stack.Push(c2);
			b.stack.Push(c1);
		case '$':
			b.stack.Pop();
		case '.':
			fmt.Printf("%d", <-b.stack.Pop());
		case ',':
			fmt.Printf("%c", <-b.stack.Pop());
		case '#':
			ip.Tick();
		case 'g':
			y := <-b.stack.Pop();
			x := <-b.stack.Pop();
			c := make(chan int64, 1);
			b.stack.Push(c);
			c <- int64(b.code[y][x]);
		case 'p':
			y := <-b.stack.Pop();
			x := <-b.stack.Pop();
			val := <-b.stack.Pop();
			b.code[y][x] = byte(val);
		case '&':
			buf := make([]byte, 256);
			bufLen, err := os.Stdin.Read(buf);
			if err != nil {
				fmt.Fprintln(os.Stdout, "Could not read from standard input:", err.String());
				return false;
			}
			
			if i, err := strconv.Atoi64(string(buf[0:bufLen-1])); err == nil {
				c := make(chan int64, 1);
				c <- i;
				b.stack.Push(c);
			}
			else {
				fmt.Fprintln(os.Stdout, "Bad int conversion:", err.String());
				return false;
			}
		case '~':
			buf := make([]byte, 1);
			if _, err := os.Stdin.Read(buf); err != nil {
				fmt.Fprintln(os.Stdout, "Could not read from standard input:", err.String());
				return false;
			}
			
			c := make(chan int64, 1);
			c <- int64(buf[0]);
			b.stack.Push(c);
		case '@':
			return false;
		}
	}
	return true;
}

type AsyncGofunge93Debugger struct {
	*AsyncGofunge93;
	pause bool;
	printCoords bool;
	printStack bool;
	printTrace bool;
}

func NewAsyncGofunge93Debugger(g *AsyncGofunge93) *AsyncGofunge93Debugger {
	debugger := new(AsyncGofunge93Debugger);
	debugger.AsyncGofunge93 = g;
	return debugger;
}

func (d AsyncGofunge93Debugger) Debug(ip IP) {
	inst := d.AsyncGofunge93.code[ip.Dim(1)][ip.Dim(0)];
	if d.printCoords {
		fmt.Fprintf(os.Stderr, "\n%c (%v, %v)", inst, ip.Dim(0), ip.Dim(1));
	}
	if d.printStack {
		fmt.Fprintf(os.Stderr, "\n%v", d.AsyncGofunge93.stack.v.Data());
	}
	if d.printTrace {
		fmt.Fprint(os.Stderr, "\n");
		for i, row := range d.AsyncGofunge93.code {
			if int32(i) == ip.Dim(1) {
				fmt.Fprint(os.Stderr, string(row[0:ip.Dim(0)]));
				fmt.Fprint(os.Stderr, "█");
				fmt.Fprintln(os.Stderr, string(row[ip.Dim(0)+1:len(row)]));
			}
			else {
				fmt.Fprintln(os.Stderr, string(row));
			}
		}
	}
	if d.pause {
		os.Stdin.Read(make([]byte, 1));
	}
}

func (d *AsyncGofunge93Debugger) SetPause(b bool) { d.pause = b }
func (d *AsyncGofunge93Debugger) SetPrintCoords(b bool) { d.printCoords = b }
func (d *AsyncGofunge93Debugger) SetPrintStack(b bool) { d.printStack = b }
func (d *AsyncGofunge93Debugger) SetPrintTrace(b bool) { d.printTrace = b }

func Start(interpreter Interpreter, ip IP) {
	for interpreter.Execute(ip) {
		ip.Tick();
	}
}

func StartDebug(debug Debuggable, ip IP) {
	debug.Debug(ip);
	for debug.Execute(ip) {
		ip.Tick();
		debug.Debug(ip);
	}
}

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
		befunge := NewAsyncGofunge93(space);
		debugger := NewAsyncGofunge93Debugger(befunge);
		debugger.SetPause(*pause);
		debugger.SetPrintCoords(*coords);
		debugger.SetPrintStack(*stack);
		debugger.SetPrintTrace(*trace);
		StartDebug(debugger, NewIP2d(0, 0, [2]int8{1, 0}, 80, 25));
	} else {
		befunge := NewGofunge93(space);
		debugger := NewGofunge93Debugger(befunge);
		debugger.SetPause(*pause);
		debugger.SetPrintCoords(*coords);
		debugger.SetPrintStack(*stack);
		debugger.SetPrintTrace(*trace);
		StartDebug(debugger, NewIP2d(0, 0, [2]int8{1, 0}, 80, 25));
	}
}
