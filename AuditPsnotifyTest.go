package main  

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"syscall"
	"unsafe"
	"errors"
	"os"
	"log"
)

const (
	MAX_AUDIT_MESSAGE_LENGTH = 8960
	AUDIT_GET = 1000
	AUDIT_SET = 1001 /* Set status (enable/disable/auditd) */
	AUDIT_LIST = 1002
	AUDIT_LIST_RULES = 1013
	AUDIT_ADD_RULE = 1011 /* Add syscall filtering rule */
	AUDIT_FIRST_USER_MSG = 1100 /* Userspace messages mostly uninteresting to kernel */
	AUDIT_MAX_FIELDS = 64
	AUDIT_BITMASK_SIZE = 64
	AUDIT_GET_FEATURE = 1019
	//Rule Flags
	AUDIT_FILTER_USER = 0x00 /* Apply rule to user-generated messages */
	AUDIT_FILTER_TASK = 0x01 /* Apply rule at task creation (not syscall) */
	AUDIT_FILTER_ENTRY = 0x02 /* Apply rule at syscall entry */
	AUDIT_FILTER_WATCH = 0x03 /* Apply rule to file system watches */
	AUDIT_FILTER_EXIT = 0x04 /* Apply rule at syscall exit */
	AUDIT_FILTER_TYPE = 0x05 /* Apply rule at audit_log_start */
	/* Rule actions */
	AUDIT_NEVER = 0 /* Do not build context if rule matches */
	AUDIT_POSSIBLE = 1 /* Build context if rule matches */
	AUDIT_ALWAYS = 2 /* Generate audit record if rule matches */
	/* Rule fields */
	/* These are useful when checking the
	* task structure at task creation time
	* (AUDIT_PER_TASK). */
	AUDIT_PID = 0
	AUDIT_UID = 1
	AUDIT_EUID = 2
	AUDIT_SUID = 3
	AUDIT_FSUID = 4
	AUDIT_GID = 5
	AUDIT_EGID = 6
	AUDIT_SGID = 7
	AUDIT_FSGID = 8
	AUDIT_LOGINUID = 9
	AUDIT_PERS = 10
	AUDIT_ARCH = 11
	AUDIT_MSGTYPE = 12
	AUDIT_SUBJ_USER = 13 /* security label user */
	AUDIT_SUBJ_ROLE = 14 /* security label role */
	AUDIT_SUBJ_TYPE = 15 /* security label type */
	AUDIT_SUBJ_SEN = 16 /* security label sensitivity label */
	AUDIT_SUBJ_CLR = 17 /* security label clearance label */
	AUDIT_PPID = 18
	AUDIT_OBJ_USER = 19
	AUDIT_OBJ_ROLE = 20
	AUDIT_OBJ_TYPE = 21
	AUDIT_OBJ_LEV_LOW = 22
	AUDIT_OBJ_LEV_HIGH = 23
	AUDIT_LOGINUID_SET = 24
	AUDIT_BIT_MASK = 0x08000000
	AUDIT_LESS_THAN = 0x10000000
	AUDIT_GREATER_THAN = 0x20000000
	AUDIT_NOT_EQUAL = 0x30000000
	AUDIT_EQUAL = 0x40000000
	AUDIT_BIT_TEST = (AUDIT_BIT_MASK | AUDIT_EQUAL)
	AUDIT_LESS_THAN_OR_EQUAL = (AUDIT_LESS_THAN | AUDIT_EQUAL)
	AUDIT_GREATER_THAN_OR_EQUAL = (AUDIT_GREATER_THAN | AUDIT_EQUAL)
	AUDIT_OPERATORS = (AUDIT_EQUAL | AUDIT_NOT_EQUAL | AUDIT_BIT_MASK)
	/* Status symbols */
	/* Mask values */
	AUDIT_STATUS_ENABLED = 0x0001
	AUDIT_STATUS_FAILURE = 0x0002
	AUDIT_STATUS_PID = 0x0004
	AUDIT_STATUS_RATE_LIMIT = 0x0008
	AUDIT_STATUS_BACKLOG_LIMIT = 0x0010
	/* Failure-to-log actions */
	AUDIT_FAIL_SILENT = 0
	AUDIT_FAIL_PRINTK = 1
	AUDIT_FAIL_PANIC = 2
	/* distinguish syscall tables */
	__AUDIT_ARCH_64BIT = 0x80000000
	__AUDIT_ARCH_LE = 0x40000000
	AUDIT_ARCH_ALPHA = (EM_ALPHA | __AUDIT_ARCH_64BIT | __AUDIT_ARCH_LE)
	AUDIT_ARCH_ARM = (EM_ARM | __AUDIT_ARCH_LE)
	AUDIT_ARCH_ARMEB = (EM_ARM)
	AUDIT_ARCH_CRIS = (EM_CRIS | __AUDIT_ARCH_LE)
	AUDIT_ARCH_FRV = (EM_FRV)
	AUDIT_ARCH_I386 = (EM_386 | __AUDIT_ARCH_LE)
	AUDIT_ARCH_IA64 = (EM_IA_64 | __AUDIT_ARCH_64BIT | __AUDIT_ARCH_LE)
	AUDIT_ARCH_M32R = (EM_M32R)
	AUDIT_ARCH_M68K = (EM_68K)
	AUDIT_ARCH_MIPS = (EM_MIPS)
	AUDIT_ARCH_MIPSEL = (EM_MIPS | __AUDIT_ARCH_LE)
	AUDIT_ARCH_MIPS64 = (EM_MIPS | __AUDIT_ARCH_64BIT)
	AUDIT_ARCH_MIPSEL64 = (EM_MIPS | __AUDIT_ARCH_64BIT | __AUDIT_ARCH_LE)
	// AUDIT_ARCH_OPENRISC = (EM_OPENRISC)
	// AUDIT_ARCH_PARISC = (EM_PARISC)
	// AUDIT_ARCH_PARISC64 = (EM_PARISC | __AUDIT_ARCH_64BIT)
	AUDIT_ARCH_PPC = (EM_PPC)
	AUDIT_ARCH_PPC64 = (EM_PPC64 | __AUDIT_ARCH_64BIT)
	AUDIT_ARCH_S390 = (EM_S390)
	AUDIT_ARCH_S390X = (EM_S390 | __AUDIT_ARCH_64BIT)
	AUDIT_ARCH_SH = (EM_SH)
	AUDIT_ARCH_SHEL = (EM_SH | __AUDIT_ARCH_LE)
	AUDIT_ARCH_SH64 = (EM_SH | __AUDIT_ARCH_64BIT)
	AUDIT_ARCH_SHEL64 = (EM_SH | __AUDIT_ARCH_64BIT | __AUDIT_ARCH_LE)
	AUDIT_ARCH_SPARC = (EM_SPARC)
	AUDIT_ARCH_SPARC64 = (EM_SPARCV9 | __AUDIT_ARCH_64BIT)
	AUDIT_ARCH_X86_64 = (EM_X86_64 | __AUDIT_ARCH_64BIT | __AUDIT_ARCH_LE)
	///Temporary Solution need to add linux/elf-em.h
	EM_NONE = 0
	EM_M32 = 1
	EM_SPARC = 2
	EM_386 = 3
	EM_68K = 4
	EM_88K = 5
	EM_486 = 6 /* Perhaps disused */
	EM_860 = 7
	EM_MIPS = 8 /* MIPS R3000 (officially, big-endian only) */
	/* Next two are historical and binaries and
	modules of these types will be rejected by
	Linux. */
	EM_MIPS_RS3_LE = 10 /* MIPS R3000 little-endian */
	EM_MIPS_RS4_BE = 10 /* MIPS R4000 big-endian */
	EM_PARISC = 15 /* HPPA */
	EM_SPARC32PLUS = 18 /* Sun's "v8plus" */
	EM_PPC = 20 /* PowerPC */
	EM_PPC64 = 21 /* PowerPC64 */
	EM_SPU = 23 /* Cell BE SPU */
	EM_ARM = 40 /* ARM 32 bit */
	EM_SH = 42 /* SuperH */
	EM_SPARCV9 = 43 /* SPARC v9 64-bit */
	EM_IA_64 = 50 /* HP/Intel IA-64 */
	EM_X86_64 = 62 /* AMD x86-64 */
	EM_S390 = 22 /* IBM S/390 */
	EM_CRIS = 76 /* Axis Communications 32-bit embedded processor */
	EM_V850 = 87 /* NEC v850 */
	EM_M32R = 88 /* Renesas M32R */
	EM_MN10300 = 89 /* Panasonic/MEI MN10300, AM33 */
	EM_BLACKFIN = 106 /* ADI Blackfin Processor */
	EM_TI_C6000 = 140 /* TI C6X DSPs */
	EM_AARCH64 = 183 /* ARM 64 bit */
	EM_FRV = 0x5441 /* Fujitsu FR-V */
	EM_AVR32 = 0x18ad /* Atmel AVR32 */
	/*
	* This is an interim value that we will use until the committee comes
	* up with a final number.
	*/
	EM_ALPHA = 0x9026
	/* Bogus old v850 magic number, used by old tools. */
	EM_CYGNUS_V850 = 0x9080
	/* Bogus old m32r magic number, used by old tools. */
	EM_CYGNUS_M32R = 0x9041
	/* This is the old interim value for S/390 architecture */
	EM_S390_OLD = 0xA390
	/* Also Panasonic/MEI MN10300, AM33 */
	EM_CYGNUS_MN10300 = 0xbeef
)

type AuditStatus struct {
	Mask          uint32 /* Bit mask for valid entries */
	Enabled       uint32 /* 1 = enabled, 0 = disabled */
	Failure       uint32 /* Failure-to-log action */
	Pid           uint32 /* pid of auditd process */
	Rate_limit    uint32 /* messages rate limit (per second) */
	Backlog_limit uint32 /* waiting messages limit */
	Lost          uint32 /* messages lost */
	Backlog       uint32 /* messages waiting in queue */
}

type AuditRuleData struct {
	Flags       uint32 /* AUDIT_PER_{TASK,CALL}, AUDIT_PREPEND */
	Action      uint32 /* AUDIT_NEVER, AUDIT_POSSIBLE, AUDIT_ALWAYS */
	Field_count uint32
	Mask        [AUDIT_BITMASK_SIZE]uint32 /* syscall(s) affected */
	Fields      [AUDIT_MAX_FIELDS]uint32
	Values      [AUDIT_MAX_FIELDS]uint32
	Fieldflags  [AUDIT_MAX_FIELDS]uint32
	Buflen      uint32  /* total length of string fields */
	Buf         [0]byte //[0]string /* string fields buffer */

}
type NetlinkSocket struct {
	fd  int
	lsa syscall.SockaddrNetlink
}

type NetlinkAuditRequest struct {
	Header syscall.NlMsghdr
	Data   []byte
}

var ParsedResult AuditStatus

func nativeEndian() binary.ByteOrder {
	var x uint32 = 0x01020304
	if *(*byte)(unsafe.Pointer(&x)) == 0x01 {
		return binary.BigEndian
	}
	return binary.LittleEndian
}

//The recvfrom in go takes only a byte [] to put the data recieved from the kernel that removes the need
//for having a separate audit_reply Struct for recieving data from kernel.
func (rr *NetlinkAuditRequest) ToWireFormat() []byte {
	b := make([]byte, rr.Header.Len)
	*(*uint32)(unsafe.Pointer(&b[0:4][0])) = rr.Header.Len
	*(*uint16)(unsafe.Pointer(&b[4:6][0])) = rr.Header.Type
	*(*uint16)(unsafe.Pointer(&b[6:8][0])) = rr.Header.Flags
	*(*uint32)(unsafe.Pointer(&b[8:12][0])) = rr.Header.Seq
	*(*uint32)(unsafe.Pointer(&b[12:16][0])) = rr.Header.Pid
	b = append(b[:16], rr.Data[:]...) //Important b[:16]
	return b
}

func newNetlinkAuditRequest(proto, seq, family, sizeofData int) *NetlinkAuditRequest {
	rr := &NetlinkAuditRequest{}

	rr.Header.Len = uint32(syscall.NLMSG_HDRLEN + sizeofData)
	rr.Header.Type = uint16(proto)
	rr.Header.Flags = syscall.NLM_F_REQUEST | syscall.NLM_F_ACK
	rr.Header.Seq = uint32(seq)
	return rr
	//	return rr.ToWireFormat()
}

// Round the length of a netlink message up to align it properly.
func nlmAlignOf(msglen int) int {
	return (msglen + syscall.NLMSG_ALIGNTO - 1) & ^(syscall.NLMSG_ALIGNTO - 1)
}

func ParseAuditNetlinkMessage(b []byte) ([]syscall.NetlinkMessage, error) {
	var msgs []syscall.NetlinkMessage
	for len(b) >= syscall.NLMSG_HDRLEN {
		h, dbuf, dlen, err := netlinkMessageHeaderAndData(b)
		if err != nil {
			fmt.Println("Error in parsing")
			return nil, err
		}
		m := syscall.NetlinkMessage{Header: *h, Data: dbuf[:int(h.Len)-syscall.NLMSG_HDRLEN]}
		msgs = append(msgs, m)
		b = b[dlen:]
	}
	return msgs, nil
}

func netlinkMessageHeaderAndData(b []byte) (*syscall.NlMsghdr, []byte, int, error) {

	h := (*syscall.NlMsghdr)(unsafe.Pointer(&b[0]))
	if int(h.Len) < syscall.NLMSG_HDRLEN || int(h.Len) > len(b) {
		fmt.Println("Error due to....HDRLEN:", syscall.NLMSG_HDRLEN, " Header Length:", h.Len, " Length of BYTE Array:", len(b))
		return nil, nil, 0, syscall.EINVAL
	}
	return h, b[syscall.NLMSG_HDRLEN:], nlmAlignOf(int(h.Len)), nil
}

// This function makes a conncetion with kernel space and is to be used for all further socket communication

func GetNetlinkSocket() (*NetlinkSocket, error) {
	fd, err := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_RAW, syscall.NETLINK_AUDIT) //connect to the socket of type RAW
	if err != nil {
		return nil, err
	}
	s := &NetlinkSocket{
		fd: fd,
	}
	s.lsa.Family = syscall.AF_NETLINK
	s.lsa.Groups = 0
	s.lsa.Pid = 0 //Kernel space pid is always set to be 0

	if err := syscall.Bind(fd, &s.lsa); err != nil {
		syscall.Close(fd)
		return nil, err
	}
	return s, nil
}

//To end the socket conncetion
func (s *NetlinkSocket) Close() {
	syscall.Close(s.fd)
}

func (s *NetlinkSocket) Send(request *NetlinkAuditRequest) error {
	if err := syscall.Sendto(s.fd, request.ToWireFormat(), 0, &s.lsa); err != nil {
		return err
	}
	return nil
}

func (s *NetlinkSocket) Receive(bytesize int, block int) ([]syscall.NetlinkMessage, error) {
	rb := make([]byte, bytesize)
	nr, _, err := syscall.Recvfrom(s.fd, rb, 0|block)
	//nr, _, err := syscall.Recvfrom(s, rb, syscall.MSG_PEEK|syscall.MSG_DONTWAIT)
	/*
		if err == syscall.EAGAIN {
			return nil, err
		}
	*/
	if err != nil {
		return nil, err
	}
	if nr < syscall.NLMSG_HDRLEN {
		return nil, syscall.EINVAL
	}
	rb = rb[:nr]
	return ParseAuditNetlinkMessage(rb) //Or syscall.ParseNetlinkMessage(rb)
}

func AuditSend(s *NetlinkSocket, proto int, data []byte, sizedata, seq int) error {

	wb := newNetlinkAuditRequest(proto, seq, syscall.AF_NETLINK, sizedata)
	wb.Data = append(wb.Data[:], data[:]...)
	if err := s.Send(wb); err != nil {
		return err
	}
	return nil
}

func AuditGetReply(s *NetlinkSocket, bytesize, block, seq int) error {
done:
	for {
		msgs, err := s.Receive(bytesize, block) //ParseAuditNetlinkMessage(rb)
		if err != nil {
			return err
		}
		for _, m := range msgs {
			lsa, err := syscall.Getsockname(s.fd)
			if err != nil {
				return err
			}
			switch v := lsa.(type) {
			case *syscall.SockaddrNetlink:

				if m.Header.Seq != uint32(seq) || m.Header.Pid != v.Pid {
					return syscall.EINVAL
				}
			default:
				return syscall.EINVAL

			}

			if m.Header.Type == syscall.NLMSG_DONE {
				fmt.Println("Done")
				break done
			}
			if m.Header.Type == syscall.NLMSG_ERROR {
				fmt.Println("NLMSG_ERROR")
				break done
				//return nil
			}
			if m.Header.Type == AUDIT_GET {
				fmt.Println("AUDIT_GET")
				//				break done
			}
			if m.Header.Type == AUDIT_FIRST_USER_MSG {
				fmt.Println("AUDIT_FIRST_USER_MS")
				//break done
			}
			if m.Header.Type == AUDIT_LIST_RULES {
				fmt.Println("AUDIT_LIST_RULES")
				//break done
			}
			if m.Header.Type == AUDIT_FIRST_USER_MSG {
				fmt.Println("AUDIT_FIRST_USER_MSG")
				//break done
			}
			if m.Header.Type == 1009 {
				fmt.Println("Watchlist")
			}

		}
	}
	return nil

}

func AuditSetEnabled(s *NetlinkSocket, seq int) error {
	var status AuditStatus
	status.Enabled = 1
	status.Mask = AUDIT_STATUS_ENABLED
	buff := new(bytes.Buffer)
	err := binary.Write(buff, nativeEndian(), status)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
		return err
	}

	err = AuditSend(s, AUDIT_SET, buff.Bytes(), int(unsafe.Sizeof(status)), seq)
	if err != nil {
		return err
	}
	// Receiving IN JUST ONE TRY
	err = AuditGetReply(s, syscall.Getpagesize(), 0, seq)
	if err != nil {
		return err
	}
	return nil
}

func AuditIsEnabled(s *NetlinkSocket, seq int) error {
	wb := newNetlinkAuditRequest(AUDIT_GET, seq, syscall.AF_NETLINK, 0)

	if err := s.Send(wb); err != nil {
		return err
	}

done:
	for {
		//Make the rb byte bigger because of large messages from Kernel doesn't fit in 4096
		msgs, err := s.Receive(MAX_AUDIT_MESSAGE_LENGTH, syscall.MSG_DONTWAIT) //ParseAuditNetlinkMessage(rb)
		if err != nil {
			return err
		}

		for _, m := range msgs {
			lsa, er := syscall.Getsockname(s.fd)
			if er != nil {
				return nil
			}
			switch v := lsa.(type) {
			case *syscall.SockaddrNetlink:

				if m.Header.Seq != uint32(seq) || m.Header.Pid != v.Pid {
					return syscall.EINVAL
				}
			default:
				return syscall.EINVAL
			}
			if m.Header.Type == syscall.NLMSG_DONE {
				fmt.Println("Done")
				break done

			}
			if m.Header.Type == syscall.NLMSG_ERROR {
				fmt.Println("NLMSG_ERROR\n")
			}
			if m.Header.Type == AUDIT_GET {
				//Conversion of the data part written to AuditStatus struct
				//Nil error : successfuly parsed
				b := m.Data[:]
				buf := bytes.NewBuffer(b)
				var dumm AuditStatus
				err = binary.Read(buf, nativeEndian(), &dumm)
				ParsedResult = dumm
				//fmt.Println("\nstruct :", dumm, err)
				//fmt.Println("\nStatus: ", dumm.Enabled)

				fmt.Println("ENABLED")
				break done
			}

		}

	}
	return nil

}
func AuditSetPid(s *NetlinkSocket, pid uint32 /*,Wait mode WAIT_YES | WAIT_NO */) error {
	var status AuditStatus
	status.Mask = AUDIT_STATUS_PID
	status.Pid = pid
	buff := new(bytes.Buffer)
	err := binary.Write(buff, nativeEndian(), status)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
		return err
	}

	err = AuditSend(s, AUDIT_SET, buff.Bytes(), int(unsafe.Sizeof(status)), 3)
	if err != nil {
		return err
	}

	err = AuditGetReply(s, syscall.Getpagesize(), 0, 3)
	if err != nil {
		return err
	}
	//Polling in GO

	return nil

}

func auditWord(nr int) uint32 {
	audit_word := (uint32)((nr) / 32)
	return (uint32)(audit_word)
}

func auditBit(nr int) uint32 {
	audit_bit := 1 << ((uint32)(nr) - auditWord(nr)*32)
	return (uint32)(audit_bit)
}

func AuditRuleSyscallData(rule *AuditRuleData, scall int) error {
	word := auditWord(scall)
	bit := auditBit(scall)

	if word >= AUDIT_BITMASK_SIZE-1 {
		fmt.Println("Some error occured")
	}
	rule.Mask[word] |= bit
	return nil
}

func AuditAddRuleData(s *NetlinkSocket, rule *AuditRuleData, flags int, action int) error {

	if flags == AUDIT_FILTER_ENTRY {
		fmt.Println("Use of entry filter is deprecated")
		return nil
	}

	rule.Flags = uint32(flags)
	rule.Action = uint32(action)

	buff := new(bytes.Buffer)
	err := binary.Write(buff, nativeEndian(), *rule)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
		return err
	}
	seq := 4 //Should be set accordingly
	err = AuditSend(s, AUDIT_ADD_RULE, buff.Bytes(), int(buff.Len())+int(rule.Buflen), seq)

	if err != nil {
		fmt.Println("Error sending add rule data request ()")
		return err
	}
	return nil
}


const (
// internal flags (from <linux/connector.h>)
_CN_IDX_PROC = 0x1
_CN_VAL_PROC = 0x1
// internal flags (from <linux/cn_proc.h>)
_PROC_CN_MCAST_LISTEN = 1
_PROC_CN_MCAST_IGNORE = 2
// Flags (from <linux/cn_proc.h>)
PROC_EVENT_FORK = 0x00000001 // fork() events
PROC_EVENT_EXEC = 0x00000002 // exec() events
PROC_EVENT_EXIT = 0x80000000 // exit() events
// Watch for all process events
PROC_EVENT_ALL = PROC_EVENT_FORK | PROC_EVENT_EXEC | PROC_EVENT_EXIT
)
var (
byteOrder = binary.LittleEndian
)
// linux/connector.h: struct cb_id
type cbId struct {
Idx uint32
Val uint32
}
// linux/connector.h: struct cb_msg
type cnMsg struct {
Id cbId
Seq uint32
Ack uint32
Len uint16
Flags uint16
}
// linux/cn_proc.h: struct proc_event.{what,cpu,timestamp_ns}
type procEventHeader struct {
What uint32
Cpu uint32
Timestamp uint64
}
// linux/cn_proc.h: struct proc_event.fork
type forkProcEvent struct {
ParentPid uint32
ParentTgid uint32
ChildPid uint32
ChildTgid uint32
}
// linux/cn_proc.h: struct proc_event.exec
type execProcEvent struct {
ProcessPid uint32
ProcessTgid uint32
}
// linux/cn_proc.h: struct proc_event.exit
type exitProcEvent struct {
ProcessPid uint32
ProcessTgid uint32
ExitCode uint32
ExitSignal uint32
}
// standard netlink header + connector header
type netlinkProcMessage struct {
Header syscall.NlMsghdr
Data cnMsg
}
type netlinkListener struct {
addr *syscall.SockaddrNetlink // Netlink socket address
sock int // The syscall.Socket() file descriptor
seq uint32 // struct cn_msg.seq
}
// Initialize linux implementation of the eventListener interface
func createListener() (eventListener, error) {
listener := &netlinkListener{}
err := listener.bind()
return listener, err
}
// noop on linux
func (w *Watcher) unregister(pid int) error {
return nil
}
// noop on linux
func (w *Watcher) register(pid int, flags uint32) error {
return nil
}
// Read events from the netlink socket
func (w *Watcher) readEvents() {
buf := make([]byte, syscall.Getpagesize())
listener, _ := w.listener.(*netlinkListener)
for {
if w.isDone() {
return
}
nr, _, err := syscall.Recvfrom(listener.sock, buf, 0)
if err != nil {
w.Error <- err
continue
}
if nr < syscall.NLMSG_HDRLEN {
w.Error <- syscall.EINVAL
continue
}
msgs, _ := syscall.ParseNetlinkMessage(buf[:nr])
for _, m := range msgs {
if m.Header.Type == syscall.NLMSG_DONE {
w.handleEvent(m.Data)
}
}
}
}
// Internal helper to check if pid && event is being watched
func (w *Watcher) isWatching(pid int, event uint32) bool {
if watch, ok := w.watches[pid]; ok {
return (watch.flags & event) == event
}
return false
}
// Dispatch events from the netlink socket to the Event channels.
// Unlike bsd kqueue, netlink receives events for all pids,
// so we apply filtering based on the watch table via isWatching()
func (w *Watcher) handleEvent(data []byte) {
buf := bytes.NewBuffer(data)
msg := &cnMsg{}
hdr := &procEventHeader{}
binary.Read(buf, byteOrder, msg)
binary.Read(buf, byteOrder, hdr)
switch hdr.What {
case PROC_EVENT_FORK:
event := &forkProcEvent{}
binary.Read(buf, byteOrder, event)
ppid := int(event.ParentTgid)
pid := int(event.ChildTgid)
if w.isWatching(ppid, PROC_EVENT_EXEC) {
// follow forks
watch, _ := w.watches[ppid]
w.Watch(pid, watch.flags)
}
if w.isWatching(ppid, PROC_EVENT_FORK) {
w.Fork <- &ProcEventFork{ParentPid: ppid, ChildPid: pid}
}
case PROC_EVENT_EXEC:
event := &execProcEvent{}
binary.Read(buf, byteOrder, event)
pid := int(event.ProcessTgid)
if w.isWatching(pid, PROC_EVENT_EXEC) {
w.Exec <- &ProcEventExec{Pid: pid}
}
case PROC_EVENT_EXIT:
event := &exitProcEvent{}
binary.Read(buf, byteOrder, event)
pid := int(event.ProcessTgid)
if w.isWatching(pid, PROC_EVENT_EXIT) {
w.RemoveWatch(pid)
w.Exit <- &ProcEventExit{Pid: pid}
}
}
}
// Bind our netlink socket and
// send a listen control message to the connector driver.
func (listener *netlinkListener) bind() error {
sock, err := syscall.Socket(
syscall.AF_NETLINK,
syscall.SOCK_DGRAM,
syscall.NETLINK_CONNECTOR)
if err != nil {
return err
}
listener.sock = sock
listener.addr = &syscall.SockaddrNetlink{
Family: syscall.AF_NETLINK,
Groups: _CN_IDX_PROC,
}
err = syscall.Bind(listener.sock, listener.addr)
if err != nil {
return err
}
return listener.send(_PROC_CN_MCAST_LISTEN)
}
// Send an ignore control message to the connector driver
// and close our netlink socket.
func (listener *netlinkListener) close() error {
err := listener.send(_PROC_CN_MCAST_IGNORE)
syscall.Close(listener.sock)
return err
}
// Generic method for sending control messages to the connector
// driver; where op is one of PROC_CN_MCAST_{LISTEN,IGNORE}
func (listener *netlinkListener) send(op uint32) error {
listener.seq++
pr := &netlinkProcMessage{}
plen := binary.Size(pr.Data) + binary.Size(op)
pr.Header.Len = syscall.NLMSG_HDRLEN + uint32(plen)
pr.Header.Type = uint16(syscall.NLMSG_DONE)
pr.Header.Flags = 0
pr.Header.Seq = listener.seq
pr.Header.Pid = uint32(os.Getpid())
pr.Data.Id.Idx = _CN_IDX_PROC
pr.Data.Id.Val = _CN_VAL_PROC
pr.Data.Len = uint16(binary.Size(op))
buf := bytes.NewBuffer(make([]byte, 0, pr.Header.Len))
binary.Write(buf, byteOrder, pr)
binary.Write(buf, byteOrder, op)
return syscall.Sendto(listener.sock, buf.Bytes(), 0, listener.addr)
}

type ProcEventFork struct {
ParentPid int // Pid of the process that called fork()
ChildPid int // Child process pid created by fork()
}
type ProcEventExec struct {
Pid int // Pid of the process that called exec()
}
type ProcEventExit struct {
Pid int // Pid of the process that called exit()
}
type watch struct {
flags uint32 // Saved value of Watch() flags param
}
type eventListener interface {
close() error // Watch.Close() closes the OS specific listener
}
type Watcher struct {
listener eventListener // OS specifics (kqueue or netlink)
watches map[int]*watch // Map of watched process ids
Error chan error // Errors are sent on this channel
Fork chan *ProcEventFork // Fork events are sent on this channel
Exec chan *ProcEventExec // Exec events are sent on this channel
Exit chan *ProcEventExit // Exit events are sent on this channel
done chan bool // Used to stop the readEvents() goroutine
isClosed bool // Set to true when Close() is first called
}
// Initialize event listener and channels
func NewWatcher() (*Watcher, error) {
listener, err := createListener()
if err != nil {
return nil, err
}
w := &Watcher{
listener: listener,
watches: make(map[int]*watch),
Fork: make(chan *ProcEventFork),
Exec: make(chan *ProcEventExec),
Exit: make(chan *ProcEventExit),
Error: make(chan error),
done: make(chan bool, 1),
}
go w.readEvents()
return w, nil
}
// Close event channels when done message is received
func (w *Watcher) finish() {
close(w.Fork)
close(w.Exec)
close(w.Exit)
close(w.Error)
}
// Closes the OS specific event listener,
// removes all watches and closes all event channels.
func (w *Watcher) Close() error {
if w.isClosed {
return nil
}
w.isClosed = true
for pid := range w.watches {
w.RemoveWatch(pid)
}
w.done <- true
w.listener.close()
return nil
}
// Add pid to the watched process set.
// The flags param is a bitmask of process events to capture,
// must be one or more of: PROC_EVENT_FORK, PROC_EVENT_EXEC, PROC_EVENT_EXIT
func (w *Watcher) Watch(pid int, flags uint32) error {
if w.isClosed {
return errors.New("psnotify watcher is closed")
}
watchEntry, found := w.watches[pid]
if found {
watchEntry.flags |= flags
} else {
if err := w.register(pid, flags); err != nil {
return err
}
w.watches[pid] = &watch{flags: flags}
}
return nil
}
// Remove pid from the watched process set.
func (w *Watcher) RemoveWatch(pid int) error {
_, ok := w.watches[pid]
if !ok {
msg := fmt.Sprintf("watch for pid=%d does not exist", pid)
return errors.New(msg)
}
delete(w.watches, pid)
return w.unregister(pid)
}
// Internal helper to check if there is a message on the "done" channel.
// The "done" message is sent by the Close() method; when received here,
// the Watcher.finish method is called to close all channels and return
// true - in which case the caller should break from the readEvents loop.
func (w *Watcher) isDone() bool {
var done bool
select {
case done = <-w.done:
w.finish()
default:
}
return done
}



/* How the file should look like
-- seprate constant, stuct to function
-- have a library function for different things like list all rules etc
-- have a main function like audit_send/get_reply
*/

/* Form of main function
package main

import (
	"fmt"
	"github.com/..../netlinkAudit"
)
func main() {
	s, err := netlinkAudit.GetNetlinkSocket()
	if err != nil {
		fmt.Println(err)
	}
	defer s.Close()

	netlinkAudit.AuditSetEnabled(s, 1)
	err = netlinkAudit.AuditIsEnabled(s, 2)
	fmt.Println("parsedResult")
	fmt.Println(netlinkAudit.ParsedResult)
	if err == nil {
		fmt.Println("Horrah")
	}

}

*/


func main() {
	s, err := GetNetlinkSocket()
	if err != nil {
	fmt.Println(err)
	}
	defer s.Close()
	AuditSetEnabled(s, 1)
	err = AuditIsEnabled(s, 2)
	fmt.Println("parsedResult")
	fmt.Println(ParsedResult)
	if err == nil {
	fmt.Println("Horrah")
	}
	AuditSetPid(s, uint32(syscall.Getpid()))
	var foo AuditRuleData
	// we need audit_name_to_field( ) && audit_rule_fieldpair_data
	//Syscall rmdir() is 84 on table
	//fmt.Println(unsafe.Sizeof(foo))
	AuditRuleSyscallData(&foo, 84)
	//fmt.Println(foo)
	foo.Fields[foo.Field_count] = AUDIT_ARCH
	foo.Fieldflags[foo.Field_count] = AUDIT_EQUAL
	foo.Values[foo.Field_count] = AUDIT_ARCH_X86_64
	foo.Field_count++
	//seq := 3
	AuditAddRuleData(s, &foo, AUDIT_FILTER_EXIT, AUDIT_ALWAYS)
	//TODO: Need to comeup with a method to generate atomic sequence numbers for sending the messages.
	//Listening in a while loop from kernel when some event goes down through Kernel
	//Creating Errors for now
	//recieved_value := make(chan int)
	//seq := make(chan int)
	seq := 3

	////
	for {
		
		//recieved_value <- seq
		err := AuditGetReply(s, syscall.Getpagesize(), syscall.MSG_DONTWAIT, seq)
		//reflect.TypeOf(AuditGetReply(s, syscall.Getpagesize(), syscall.MSG_DONTWAIT, seq))
		//fmt.Println(AuditGetReply(s, syscall.Getpagesize(), syscall.MSG_DONTWAIT, seq))
		if err != nil {
			continue
		}	
		seq++
	}	
	//////////////////////////////////
	//////////////////////////////
	watcher, err := NewWatcher()
    if err != nil {
        log.Fatal(err)
    }

    // Process events
    go func() {
        for {
            select {
            case ev := <-watcher.Fork:
                log.Println("fork event:", ev)
            case ev := <-watcher.Exec:
                log.Println("exec event:", ev)
            case ev := <-watcher.Exit:
                log.Println("exit event:", ev)
            case err := <-watcher.Error:
                log.Println("error:", err)
            }
        }
    }()

    err = watcher.Watch(os.Getpid(), PROC_EVENT_ALL)
    if err != nil {
        log.Fatal(err)
    }
    //fmt.Println("yo")

    /* ... do stuff ... */
    watcher.Close()

	//////////////////////////////
	///////////////////////////
	//auditctl -a rmdir exit,always
	//Flags are exit
	//Action is always
}