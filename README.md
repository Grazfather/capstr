capstr
--------

Capstone Go bindings facilitating highly optimized printing of disassembly.

Usage
--------

```
import "github.com/lunixbochs/capstr"

engine, err := capstr.New(capstr.ARCH_X86, capstr.MODE_32)
for _, ins := range engine.Dis(code, addr, insCount) {
    fmt.Printf("%#x: %s\n", ins.Addr, ins.Str)
}
```

Benchmarks
-------
```
BenchmarkX86-4            300000          5493 ns/op         784 B/op         22 allocs/op
BenchmarkGapstone-4       200000          9870 ns/op        9704 B/op         54 allocs/op
```
