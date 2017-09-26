package ergonode

import (
	"github.com/halturin/ergonode/etf"
	"log"
	"time"
)

// GenServer interface
type GenServer interface {
	Init(args ...interface{})
	HandleCast(message *etf.Term)
	HandleCall(from *etf.Tuple, message *etf.Term) (reply *etf.Term)
	HandleInfo(message *etf.Term)
	Terminate(reason interface{})
}

// GenServerImpl is implementation of GenServer interface
type GenServerImpl struct {
	Node *Node   // current node of process
	Self etf.Pid // Pid of process
}

// Options returns map of default process-related options
func (gs *GenServerImpl) Options() map[string]interface{} {
	return map[string]interface{}{
		"chan-size": 100, // size of channel for regular messages
	}
}

// ProcessLoop executes during whole time of process life.
// It receives incoming messages from channels and handle it using methods of behaviour implementation
func (gs *GenServerImpl) ProcessLoop(pcs procChannels, pd Process, args ...interface{}) {
	pd.(GenServer).Init(args...)
	pcs.init <- true
	defer func() {
		if r := recover(); r != nil {
			// TODO: send message to parent process
			log.Printf("GenServer recovered: %#v", r)
		}
	}()
	for {
		var message etf.Term
		var fromPid etf.Pid
		select {
		case msg := <-pcs.in:
			message = msg
		case msgFrom := <-pcs.inFrom:
			message = msgFrom[1]
			fromPid = msgFrom[0].(etf.Pid)
		}
		nLog("Message from %#v", fromPid)
		switch m := message.(type) {
		case etf.Tuple:
			switch mtag := m[0].(type) {
			case etf.Atom:
				switch mtag {
				case etf.Atom("$gen_call"):
					fromTuple := m[1].(etf.Tuple)
					reply := pd.(GenServer).HandleCall(&fromTuple, &m[2])
					if reply != nil {
						pid := fromTuple[0].(etf.Pid)
						ref := fromTuple[1]
						rep := etf.Term(etf.Tuple{ref, *reply})
						gs.Send(pid, &rep)
					}
				case etf.Atom("$gen_cast"):
					pd.(GenServer).HandleCast(&m[1])
				default:
					pd.(GenServer).HandleInfo(&message)
				}
			default:
				nLog("mtag: %#v", mtag)
				pd.(GenServer).HandleInfo(&message)
			}
		default:
			nLog("m: %#v", m)
			pd.(GenServer).HandleInfo(&message)
		}
	}
}

func (gs *GenServerImpl) setNode(node *Node) {
	gs.Node = node
}

func (gs *GenServerImpl) setPid(pid etf.Pid) {
	gs.Self = pid
}

func (gs *GenServerImpl) Call(to interface{}, message *etf.Term) (reply *etf.Term) {

	msg := etf.Term(etf.Tuple{etf.Atom("$gen_call"), message})
	if err := gs.Node.Send(gs.Self, to, &msg); err != nil {
		panic(err.Error())
	}

	replyTerm := etf.Term(etf.Atom("ok"))
	reply = &replyTerm

	return
}

func (gs *GenServerImpl) Cast(to interface{}, message *etf.Term) error {
	msg := etf.Term(etf.Tuple{etf.Atom("$gen_cast"), message})
	if err := gs.Node.Send(nil, to, &msg); err != nil {
		panic(err.Error())
	}

	return nil
}

func (gs *GenServerImpl) Send(to etf.Pid, reply *etf.Term) {
	gs.Node.Send(nil, to, reply)
}

func (gs *GenServerImpl) MakeRef() (ref etf.Ref) {
	ref.Node = etf.Atom(gs.Node.FullName)
	ref.Creation = 1

	nt := time.Now().UnixNano()
	id1 := uint32(uint64(nt) & ((2 << 17) - 1))
	id2 := uint32(uint64(nt) >> 46)
	ref.Id = []uint32{id1, id2, 0}

	return
}
