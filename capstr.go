package capstr

import (
	"bytes"
	"unsafe"
)

// #cgo LDFLAGS: -lcapstone
// #cgo CFLAGS: -O3 -Wall -Werror
// #include <capstone/capstone.h>
import "C"

const (
	ARCH_ARM   = C.CS_ARCH_ARM
	ARCH_ARM64 = C.CS_ARCH_ARM64
	ARCH_MIPS  = C.CS_ARCH_MIPS
	ARCH_X86   = C.CS_ARCH_X86
	ARCH_PPC   = C.CS_ARCH_PPC
	ARCH_SPARC = C.CS_ARCH_SPARC
	ARCH_SYSZ  = C.CS_ARCH_SYSZ
	ARCH_XCORE = C.CS_ARCH_XCORE
)

const (
	MODE_LITTLE_ENDIAN = C.CS_MODE_LITTLE_ENDIAN
	MODE_ARM           = C.CS_MODE_ARM
	MODE_16            = C.CS_MODE_16
	MODE_32            = C.CS_MODE_32
	MODE_64            = C.CS_MODE_64
	MODE_THUMB         = C.CS_MODE_THUMB
	MODE_MCLASS        = C.CS_MODE_MCLASS
	MODE_V8            = C.CS_MODE_V8
	MODE_MICRO         = C.CS_MODE_MICRO
	MODE_MIPS3         = C.CS_MODE_MIPS3
	MODE_MIPS32R6      = C.CS_MODE_MIPS32R6
	MODE_MIPSGP64      = C.CS_MODE_MIPSGP64
	MODE_V9            = C.CS_MODE_V9
	MODE_BIG_ENDIAN    = C.CS_MODE_BIG_ENDIAN
	MODE_MIPS32        = C.CS_MODE_MIPS32
	MODE_MIPS64        = C.CS_MODE_MIPS64
)

type Engine struct {
	handle C.csh
}

type CsError C.cs_err

func (e CsError) Error() string {
	return C.GoString(C.cs_strerror(C.cs_err(e)))
}

func New(arch, mode int) (*Engine, error) {
	var handle C.csh
	cserr := C.cs_open(C.cs_arch(arch), C.cs_mode(mode), &handle)
	if cserr != C.CS_ERR_OK {
		return nil, CsError(cserr)
	}
	C.cs_option(handle, C.CS_OPT_DETAIL, C.CS_OPT_OFF)
	return &Engine{handle}, nil
}

func (e *Engine) Dis(code []byte, addr, count uint64) ([]Ins, error) {
	if len(code) == 0 {
		return nil, nil
	}
	ptr := (*C.uint8_t)(unsafe.Pointer(&code[0]))

	var disptr *C.cs_insn
	num := C.cs_disasm(e.handle, ptr, C.size_t(len(code)), C.uint64_t(addr), C.size_t(count), &disptr)
	if num > 0 {
		dis := (*[1 << 23]C.cs_insn)(unsafe.Pointer(disptr))[:num]
		ret := make([]Ins, num)
		for i, ins := range dis {
			outins := &ret[i]
			// byte array casts of cs_ins fields
			mnemonic := (*[32]byte)(unsafe.Pointer(&ins.mnemonic[0]))
			byteData := (*[16]byte)(unsafe.Pointer(&ins.bytes[0]))
			opstr := (*[160]byte)(unsafe.Pointer(&ins.op_str[0]))

			// populate the return ins fields
			outins.addr = uint64(ins.address)
			// this is faster than C.GoBytes()
			outins.dataSlice = outins.data[:ins.size]
			copy(outins.dataSlice, byteData[:])

			// populate the strings
			if off := bytes.IndexByte(mnemonic[:], 0); off > 0 {
				outins.mnemonic = string(mnemonic[:off])
			}
			if off := bytes.IndexByte(opstr[:], 0); off > 0 {
				outins.opstr = string(opstr[:off])
			}
		}
		C.free(unsafe.Pointer(disptr))
		return ret, nil
	}
	return nil, CsError(C.cs_errno(e.handle))
}

func (e *Engine) Close() error {
	return CsError(C.cs_close(&e.handle))
}

// conforms to usercorn/models.Ins interface
type Ins struct {
	addr      uint64
	dataSlice []byte
	mnemonic  string
	opstr     string
	data      [16]byte
}

func (i Ins) Addr() uint64 {
	return i.addr
}

func (i Ins) Bytes() []byte {
	return i.dataSlice
}

func (i Ins) Mnemonic() string {
	return i.mnemonic
}

func (i Ins) OpStr() string {
	return i.opstr
}
