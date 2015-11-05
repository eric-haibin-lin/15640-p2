package storageserver

import (
	"errors"
	"fmt"
	"github.com/cmu440/tribbler/rpc/storagerpc"
	"net"
	"net/http"
	"net/rpc"
)

type storageServer struct {
	numNodes             int
	nodeID               uint32
	port                 int
	masterServerHostPort string
	tribbleMap           map[string]string
	listMap              map[string][]string
	ackedSlaves          int
	serverList           []storagerpc.Node
	ackedSlavesMap       map[storagerpc.Node]bool
	leaseMap		map[string][]string
}

// NewStorageServer creates and starts a new StorageServer. masterServerHostPort
// is the master storage server's host:port address. If empty, then this server
// is the master; otherwise, this server is a slave. numNodes is the total number of
// servers in the ring. port is the port number that this server should listen on.
// nodeID is a random, unsigned 32-bit ID identifying this server.
//
// This function should return only once all storage servers have joined the ring,
// and should return a non-nil error if the storage server could not be started.
func NewStorageServer(masterServerHostPort string, numNodes, port int, nodeID uint32) (StorageServer, error) {
	defer fmt.Println("Leaving NewStorageServer")
	fmt.Println("Entered NewStorageServer")
	var a StorageServer

	server := storageServer{}

	server.numNodes = numNodes
	server.port = port
	server.masterServerHostPort = masterServerHostPort
	server.nodeID = nodeID
	server.tribbleMap = make(map[string]string)
	server.listMap = make(map[string][]string)
	server.ackedSlavesMap = make(map[storagerpc.Node]bool)
	//server.serverList = make([]storagerpc.Node, 32)

	a = &server

	if len(masterServerHostPort) == 0 {
		/*fmt.Println("This is the Master Speaking!")
		fmt.Println("Number of nodes in the master is ", numNodes, " and the port is ", port, " and the nodeID is ", nodeID)*/
		/* Now register for RPCs */
		server.ackedSlaves = 1
		self := storagerpc.Node{fmt.Sprintf("localhost:%d", port), nodeID}

		server.serverList = append(server.serverList, self)

		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return nil, err
		}

		err = rpc.RegisterName("StorageServer", storagerpc.Wrap(a))
		if err != nil {
			return nil, err
		}

		rpc.HandleHTTP()
		go http.Serve(listener, nil)

		/* Now wait until all nodes have joined */
		/*fmt.Println("Before spining, ackedSlaves=", server.ackedSlaves)
		fmt.Println("Sleeping for some time")*/
		//fmt.Println("Before sleeping, acked Slaves is ", server.ackedSlaves)
		//time.Sleep(5000 * time.Millisecond)
		//for server.ackedSlaves != numNodes {
		//fmt.Println("Spining :( with ackedSlaves = ", server.ackedSlaves)
		//}
	} else {
		/*fmt.Println("I am just a lowly Slave.")
		fmt.Println("Number of nodes in the client is ", numNodes, " and the port is ", port, " and the nodeID is ", nodeID, " and the masterserverhostport is ", masterServerHostPort)*/
		/* Now try connecting to the ring by calling the RegisterServer RPC */

		srvr, err := rpc.DialHTTP("tcp", masterServerHostPort)
		if err != nil {
			fmt.Println("Oops! Returning because couldn't dial master host port")
			return nil, errors.New("Couldn't Dial Master Host Port")
		}

		args := storagerpc.RegisterArgs{}
		args.ServerInfo.HostPort = fmt.Sprintf("localhost:%d", port)
		args.ServerInfo.NodeID = nodeID

		var reply storagerpc.RegisterReply

		err = srvr.Call("StorageServer.RegisterServer", args, &reply)
		if err != nil {
			return nil, err
		}

	}

	/* For master: Now wait until all other servers have joined the ring */

	/* For CP, since no slaves are gonna join, return directly */

	//fmt.Println("After sleeping, server.ackedSlaves = ", server.ackedSlaves, " and numNodes is ", server.numNodes)
	return a, nil
}

func (ss *storageServer) RegisterServer(args *storagerpc.RegisterArgs, reply *storagerpc.RegisterReply) error {
	defer fmt.Println("Leaving RegisterServer")
	fmt.Println("RegisterServer invoked!")
	/*fmt.Println("numNodes inside RegisterServer is ", ss.numNodes)
	fmt.Println("After adding slave to list of servers, ackedSlaves is ", ss.ackedSlaves)
	fmt.Println("HostPort of this slave server is ", args.ServerInfo.HostPort, " and the nodeID is ", args.ServerInfo.NodeID)*/

	if _, ok := ss.ackedSlavesMap[args.ServerInfo]; ok {
		if ss.ackedSlaves == ss.numNodes {
			reply.Status = storagerpc.OK
			//TODO: Sort this list!
			reply.Servers = ss.serverList
			return nil
		}
		reply.Status = storagerpc.NotReady
		return nil
	}

	ss.serverList = append(ss.serverList, args.ServerInfo)
	ss.ackedSlaves += 1
	ss.ackedSlavesMap[args.ServerInfo] = true

	if ss.ackedSlaves == ss.numNodes {
		reply.Status = storagerpc.OK
		reply.Servers = ss.serverList
	} else {
		reply.Status = storagerpc.NotReady
	}
	return nil
}

func (ss *storageServer) GetServers(args *storagerpc.GetServersArgs, reply *storagerpc.GetServersReply) error {
	defer fmt.Println("Leaving GetServers")
	fmt.Println("GetServers invoked!")

	if ss.ackedSlaves == ss.numNodes {
		reply.Status = storagerpc.OK
		reply.Servers = ss.serverList
		fmt.Println("Returning OK")
		return nil
	}
	reply.Status = storagerpc.NotReady
	fmt.Println("Returning NotReady")
	return nil
}

func (ss *storageServer) Get(args *storagerpc.GetArgs, reply *storagerpc.GetReply) error {
	fmt.Println("Get invoked!")
	fmt.Println("Key is ", args.Key, ", WantLease is ", args.WantLease, " and HostPort is ", args.HostPort)

	val, ok := ss.tribbleMap[args.Key]

	if !ok {
		reply.Status = storagerpc.KeyNotFound
		return nil
	}

	reply.Status = storagerpc.OK
	reply.Value = val

	if GetArgs.WantLease == true {
		
	return nil
}

func (ss *storageServer) Delete(args *storagerpc.DeleteArgs, reply *storagerpc.DeleteReply) error {
	//fmt.Println("Delete invoked!")

	//fmt.Println("Key is ", args.Key)

	_, ok := ss.tribbleMap[args.Key]

	if !ok {
		reply.Status = storagerpc.KeyNotFound
		return nil
	}

	delete(ss.tribbleMap, args.Key)
	reply.Status = storagerpc.OK

	return nil
}

func (ss *storageServer) GetList(args *storagerpc.GetArgs, reply *storagerpc.GetListReply) error {
	//fmt.Println("GetList invoked!")

	//fmt.Println("Key is ", args.Key)0

	reply.Status = storagerpc.OK
	reply.Value = ss.listMap[args.Key]

	return nil
}

func (ss *storageServer) Put(args *storagerpc.PutArgs, reply *storagerpc.PutReply) error {
	fmt.Println("Put invoked!")
	fmt.Println("Key is ", args.Key, " and value is ", args.Value)
	/* TODO: If key exists, revoke leases! */

	ss.tribbleMap[args.Key] = args.Value
	reply.Status = storagerpc.OK
	return nil
}

func (ss *storageServer) AppendToList(args *storagerpc.PutArgs, reply *storagerpc.PutReply) error {
	/*fmt.Println("AppendToList invoked!")
	fmt.Println("Key is ", args.Key, " and Value is ", args.Value)*/
	reply.Status = storagerpc.OK

	var templist []string

	_, ok := ss.listMap[args.Key]

	if !ok {
		//fmt.Println("Inside not ok")
		templist = append(templist, args.Value)
		ss.listMap[args.Key] = templist
	} else {
		//fmt.Println("Inside ok")
		for _, val := range ss.listMap[args.Key] {
			if val == args.Value {
				reply.Status = storagerpc.ItemExists
				return nil
			}
		}

		ss.listMap[args.Key] = append(ss.listMap[args.Key], args.Value)
	}

	reply.Status = storagerpc.OK
	return nil
}

func (ss *storageServer) RemoveFromList(args *storagerpc.PutArgs, reply *storagerpc.PutReply) error {
	//fmt.Println("RemoveFromList invoked!")

	i := 0
	for _, val := range ss.listMap[args.Key] {
		if val == args.Value {
			ss.listMap[args.Key] = append((ss.listMap[args.Key])[:i], (ss.listMap[args.Key])[i+1:]...)
			reply.Status = storagerpc.OK
			return nil
		}
	}
	reply.Status = storagerpc.ItemNotFound
	return nil
}

/*

// This file contains constants and arguments used to perform RPCs between
// a TribServer's local Libstore and the storage servers. DO NOT MODIFY!

package storagerpc

// Status represents the status of a RPC's reply.
type Status int

const (
	OK           Status = iota + 1 // The RPC was a success.
	KeyNotFound                    // The specified key does not exist.
	ItemNotFound                   // The specified item does not exist.
	WrongServer                    // The specified key does not fall in the server's hash range.
	ItemExists                     // The item already exists in the list.
	NotReady                       // The storage servers are still getting ready.
)

// Lease constants.
const (
	QueryCacheSeconds = 10 // Time period used for tracking queries/determining whether to request leases.
	QueryCacheThresh  = 3  // If QueryCacheThresh queries in last QueryCacheSeconds, then request a lease.
	LeaseSeconds      = 10 // Number of seconds a lease should remain valid.
	LeaseGuardSeconds = 2  // Additional seconds a server should wait before invalidating a lease.
)

// Lease stores information about a lease sent from the storage servers.
type Lease struct {
	Granted      bool
	ValidSeconds int
}

type Node struct {
	HostPort string // The host:port address of the storage server node.
	NodeID   uint32 // The ID identifying this storage server node.
}

type RegisterArgs struct {
	ServerInfo Node
}

type RegisterReply struct {
	Status  Status
	Servers []Node
}

type GetServersArgs struct {
	// Intentionally left empty.
}

type GetServersReply struct {
	Status  Status
	Servers []Node
}

type GetArgs struct {
	Key       string
	WantLease bool
	HostPort  string // The Libstore's callback host:port.
}

type GetReply struct {
	Status Status
	Value  string
	Lease  Lease
}

type GetListReply struct {
	Status Status
	Value  []string
	Lease  Lease
}

type PutArgs struct {
	Key   string
	Value string
}

type PutReply struct {
	Status Status
}

type DeleteArgs struct {
	Key string
}

type DeleteReply struct {
	Status Status
}

type RevokeLeaseArgs struct {
	Key string
}

type RevokeLeaseReply struct {
	Status Status
}
*/

/*func (tc *tribClient) CreateUser(userID string) (tribrpc.Status, error) {
	args := &tribrpc.CreateUserArgs{UserID: userID}
	var reply tribrpc.CreateUserReply
	if err := tc.client.Call("TribServer.CreateUser", args, &reply); err != nil {
		return 0, err
	}
	return reply.Status, nil
}*/
