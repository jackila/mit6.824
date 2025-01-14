package pbservice

import (
	"crypto/rand"
	"math/big"
	"net/rpc"
	"time"
	"viewservice"
)

type Clerk struct {
	vs *viewservice.Clerk
	// Your declarations here
	Primary string
}

// this may come in handy.
func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

func MakeClerk(vshost string, me string) *Clerk {
	ck := new(Clerk)
	ck.vs = viewservice.MakeClerk(me, vshost)
	// Your ck.* initializations here
	ck.Primary = ck.vs.Primary()
	return ck
}

//
// call() sends an RPC to the rpcname handler on server srv
// with arguments args, waits for the reply, and leaves the
// reply in reply. the reply argument should be a pointer
// to a reply structure.
//
// the return value is true if the server responded, and false
// if call() was not able to contact the server. in particular,
// the reply's contents are only valid if call() returned true.
//
// you should assume that call() will return an
// error after a while if the server is dead.
// don't provide your own time-out mechanism.
//
// please use call() to send all RPCs, in client.go and server.go.
// please don't change this function.
//
func call(srv string, rpcname string,
	args interface{}, reply interface{}) bool {
	c, errx := rpc.Dial("unix", srv)
	if errx != nil {
		return false
	}
	defer c.Close()

	err := c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	//fmt.Println("the caller return the err:")
	//fmt.Println(err)
	return false
}

//
// fetch a key's value from the current primary;
// if they key has never been set, return "".
// Get() must keep trying until it either the
// primary replies with the value or the primary
// says the key doesn't exist (has never been Put().
//
func (ck *Clerk) Get(key string) string {

	// Your code here.
	args := new(GetArgs)
	reply := new(GetReply)
	args.Key = key
	args.ClientRequest = true
	Primary := ck.Primary
	for Primary != "" {
		sucess := call(Primary, "PBServer.Get", &args, &reply)
		//log.Printf("after the proxy the value will get %t,and the vs is %+v", sucess, ck.vs)
		if sucess {
			//log.Printf("the result of the value is %+v", reply.Value)
			if reply.Err != "" {
				return ""
			}
			return reply.Value
		} else {
			//重新获取 primary sleep for a while
			time.Sleep(viewservice.PingInterval)
			Primary = ck.vs.Primary()
			//log.Printf("reget the primary url %s", Primary)
		}
	}

	return ""
}

//
// send a Put or Append RPC
//

func (ck *Clerk) PutAppend(key string, value string, op string) {

	// Your code here.
	args := new(PutAppendArgs)
	reply := new(PutAppendReply)
	args.XID = nrand()
	args.Key = key
	args.Value = value
	args.Type = op
	args.ClientRequest = true
	sucess := false
	for !sucess {
		for ck.Primary == "" {
			ck.Primary = ck.vs.Primary()
		}
		//log.Printf("the request XID is %d,and the pme is %s", args.XID, ck.Primary)
		sucess := call(ck.Primary, "PBServer.PutAppend", &args, &reply)
		//log.Printf("--------- return %t -----------the request XID %d ,the pb server is %s",sucess, args.XID,  ck.Primary)
		if sucess {
			break
		} else {
			time.Sleep(viewservice.PingInterval)
			ck.Primary = ck.vs.Primary()
		}
	}
}

//
// tell the primary to update key's value.
// must keep trying until it succeeds.
//
func (ck *Clerk) Put(key string, value string) {
	ck.PutAppend(key, value, "Put")
}

//
// tell the primary to append to key's value.
// must keep trying until it succeeds.
//
func (ck *Clerk) Append(key string, value string) {
	ck.PutAppend(key, value, "Append")
}

func (ck *Clerk) fetchPrimary() {
	ck.Primary = ck.vs.Primary()
}
