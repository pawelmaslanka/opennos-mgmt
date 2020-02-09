package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"strings"

	"github.com/abiosoft/ishell"
	log "github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"opennos-mgmt/gnmi"
	"opennos-mgmt/gnmi/modeldata"
	"opennos-mgmt/gnmi/modeldata/oc"

	"opennos-mgmt/utils/credentials"

	pb "github.com/openconfig/gnmi/proto/gnmi"
)

var validIfaces = [...]string{
	"eth-1", "eth-1/1", "eth-1/2", "eth-1/3", "eth-1/4",
	"eth-2", "eth-2/1", "eth-2/2", "eth-2/3", "eth-2/4",
	"eth-3", "eth-3/1", "eth-3/2", "eth-3/3", "eth-3/4",
	"eth-4", "eth-4/1", "eth-4/2", "eth-4/3", "eth-4/4",
	"eth-5", "eth-5/1", "eth-5/2", "eth-5/3", "eth-5/4",
	"eth-6", "eth-6/1", "eth-6/2", "eth-6/3", "eth-6/4",
	"eth-7", "eth-7/1", "eth-7/2", "eth-7/3", "eth-7/4",
	"eth-8", "eth-8/1", "eth-8/2", "eth-8/3", "eth-8/4",
	"eth-9", "eth-9/1", "eth-9/2", "eth-9/3", "eth-9/4",
	"eth-10", "eth-10/1", "eth-10/2", "eth-10/3", "eth-10/4",
	"eth-11", "eth-11/1", "eth-11/2", "eth-11/3", "eth-11/4",
	"eth-12", "eth-12/1", "eth-12/2", "eth-12/3", "eth-12/4",
	"eth-13", "eth-13/1", "eth-13/2", "eth-13/3", "eth-13/4",
	"eth-14", "eth-14/1", "eth-14/2", "eth-14/3", "eth-14/4",
	"eth-15", "eth-15/1", "eth-15/2", "eth-15/3", "eth-15/4",
	"eth-16", "eth-16/1", "eth-16/2", "eth-16/3", "eth-16/4",
	"eth-17", "eth-17/1", "eth-17/2", "eth-17/3", "eth-17/4",
	"eth-18", "eth-18/1", "eth-18/2", "eth-18/3", "eth-18/4",
	"eth-19", "eth-19/1", "eth-19/2", "eth-19/3", "eth-19/4",
	"eth-20", "eth-20/1", "eth-20/2", "eth-20/3", "eth-20/4",
	"eth-21", "eth-21/1", "eth-21/2", "eth-21/3", "eth-21/4",
	"eth-22", "eth-22/1", "eth-22/2", "eth-22/3", "eth-22/4",
	"eth-32", "eth-23/1", "eth-23/2", "eth-23/3", "eth-23/4",
	"eth-24", "eth-24/1", "eth-24/2", "eth-24/3", "eth-24/4",
	"eth-25", "eth-25/1", "eth-25/2", "eth-25/3", "eth-25/4",
	"eth-26", "eth-26/1", "eth-26/2", "eth-26/3", "eth-26/4",
	"eth-27", "eth-27/1", "eth-27/2", "eth-27/3", "eth-27/4",
	"eth-28", "eth-28/1", "eth-28/2", "eth-28/3", "eth-28/4",
	"eth-29", "eth-29/1", "eth-29/2", "eth-29/3", "eth-29/4",
	"eth-30", "eth-30/1", "eth-30/2", "eth-30/3", "eth-30/4",
	"eth-31", "eth-31/1", "eth-31/2", "eth-31/3", "eth-31/4",
	"eth-32", "eth-32/1", "eth-32/2", "eth-32/3", "eth-32/4",
}

var gEditIfaceCmdCompleterInvoked bool = false

type Iface struct {
	speed uint32
}

func NewIface() *Iface {
	return &Iface{}
}

type EditIfaceCmdCtx struct {
	completerInvoked bool
}

func NewEditIfaceCmdCtx() *EditIfaceCmdCtx {
	return &EditIfaceCmdCtx{
		completerInvoked: false,
	}
}

var editIfaceCmdCtx *EditIfaceCmdCtx = NewEditIfaceCmdCtx()

var (
	bindAddr   = flag.String("bind_address", ":10161", "Bind to address:port or just :port")
	configFile = flag.String("config", "", "IETF JSON file for target startup config")
)

type server struct {
	*gnmi.Server
}

func newServer(model *gnmi.Model, config []byte) (*server, error) {
	s, err := gnmi.NewServer(model, config, nil)
	if err != nil {
		return nil, err
	}
	return &server{Server: s}, nil
}

// Get overrides the Get func of gnmi.Target to provide user auth.
func (s *server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	msg, ok := credentials.AuthorizeUser(ctx)
	if !ok {
		log.Infof("denied a Get request: %v", msg)
		return nil, status.Error(codes.PermissionDenied, msg)
	}
	log.Infof("allowed a Get request: %v", msg)
	return s.Server.Get(ctx, req)
}

// Set overrides the Set func of gnmi.Target to provide user auth.
func (s *server) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	msg, ok := credentials.AuthorizeUser(ctx)
	if !ok {
		log.Infof("denied a Set request: %v", msg)
		return nil, status.Error(codes.PermissionDenied, msg)
	}
	log.Infof("allowed a Set request: %v", msg)
	return s.Server.Set(ctx, req)
}

func gNMIServerRun() {
	model := gnmi.NewModel(modeldata.ModelData,
		reflect.TypeOf((*oc.Device)(nil)),
		oc.SchemaTree["Device"],
		oc.Unmarshal,
		oc.ΛEnum)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Supported models:\n")
		for _, m := range model.SupportedModels() {
			fmt.Fprintf(os.Stderr, "  %s\n", m)
		}
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	opts := credentials.ServerCredentials()
	g := grpc.NewServer(opts...)

	var configData []byte
	if *configFile != "" {
		var err error
		configData, err = ioutil.ReadFile(*configFile)
		if err != nil {
			log.Exitf("error in reading config file: %v", err)
		}
	}
	s, err := newServer(model, configData)
	if err != nil {
		log.Exitf("error in creating gnmi target: %v", err)
	}
	pb.RegisterGNMIServer(g, s)
	reflection.Register(g)

	log.Infof("starting to listen on %s", *bindAddr)
	listen, err := net.Listen("tcp", *bindAddr)
	if err != nil {
		log.Exitf("failed to listen: %v", err)
	}

	log.Info("starting to serve")
	if err := g.Serve(listen); err != nil {
		log.Exitf("failed to serve: %v", err)
	}
}

func main() {
	go gNMIServerRun()
	shell := ishell.New()

	// display info.
	shell.Println("Welcome to OpenNOS CLI")

	//Consider the unicode characters supported by the users font
	//shell.SetMultiChoicePrompt(" >>"," - ")
	//shell.SetChecklistOptions("[ ] ","[X] ")

	// var vlans []string = make([]string, 1)
	// var editableIfaces = map[string]*Iface{}
	var editableIfaces = map[string]*Iface{
		"eth-1": NewIface(), "eth-2": NewIface(), "eth-3": NewIface(), "eth-4": NewIface(), "eth-5": NewIface(),
		"eth-6": NewIface(), "eth-7": NewIface(), "eth-8": NewIface(), "eth-9": NewIface(), "eth-10": NewIface(),
		"eth-11": NewIface(), "eth-12": NewIface(), "eth-13": NewIface(), "eth-14": NewIface(), "eth-15": NewIface(),
		"eth-16": NewIface(), "eth-17": NewIface(), "eth-18": NewIface(), "eth-19": NewIface(), "eth-20": NewIface(),
		"eth-21": NewIface(), "eth-22": NewIface(), "eth-23": NewIface(), "eth-24": NewIface(), "eth-25": NewIface(),
		"eth-26": NewIface(), "eth-27": NewIface(), "eth-28": NewIface(), "eth-29": NewIface(), "eth-30": NewIface(),
		"eth-31": NewIface(), "eth-32": NewIface(),
	}
	// var editableIfaces []string = []string{
	// 	"eth-1", "eth-2", "eth-3", "eth-4", "eth-5", "eth-6", "eth-7", "eth-8", "eth-9", "eth-10",
	// 	"eth-11", "eth-12", "eth-13", "eth-14", "eth-15", "eth-16", "eth-17", "eth-18", "eth-19", "eth-20",
	// 	"eth-21", "eth-22", "eth-23", "eth-24", "eth-25", "eth-26", "eth-27", "eth-28", "eth-29", "eth-30",
	// 	"eth-31", "eth-32",
	// }
	editCmd := &ishell.Cmd{
		Name:     "edit",
		Help:     "edit <interface> | <aggregate> | <vlan>",
		LongHelp: `Edit`,
	}
	editIfaceCmd := &ishell.Cmd{
		Name: "interface",
		Help: "edit interface <interface name>",
		Completer: func(args []string) []string {
			if editIfaceCmdCtx.completerInvoked {
				if len(args) > 1 {
					return nil
				}

				if len(args) > 0 {
					return []string{"help"}
					// return nil
				}
				// log.Println(args)
				// return nil
				// return []string{"\nPress enter to edit..."}
			}
			ifnames := make([]string, len(editableIfaces))
			i := 0
			for ifname := range editableIfaces {
				ifnames[i] = ifname
				i++
			}

			editIfaceCmdCtx.completerInvoked = true
			return ifnames
		},
		Func: func(c *ishell.Context) {
			log.Infof("Choosed interface %s", c.Args[0])
			if len(c.Args) == 0 {
				c.Err(errors.New("Missing interface name"))
				return
			}

			if len(c.Args) == 2 {
				if c.Args[1] != "help" {
					c.Err(errors.New("Invalid argument"))
					return
				}

				log.Infof("Enter to edit interface %s", c.Args[0])
				return
			}

			if len(c.Args) > 1 {
				c.Err(errors.New("Too many arguments"))
				return
			}

			foundIface := false
			for _, iface := range validIfaces {
				if c.Args[0] == iface {
					foundIface = true
					break
				}
			}

			if !foundIface {
				c.Err(errors.New("Invalid argument"))
				return
			}

			// editableIfaces[c.Args[0]] = NewIface()
			editIfaceCmdCtx.completerInvoked = false
			c.SetPrompt(fmt.Sprintf("[edit interface %s]# ", c.Args[0]))
			// editableIfaces[c.Args[2]] = NewIface()
			// vlans = append(vlans, c.Args...)
		},
	}

	executePromptCmd := &ishell.Cmd{
		Name: "execute_prompt",
		Help: "Press Enter to execute command",
		// Func: func(c *ishell.Context) {
		// 	log.Println("Press Enter to execute command")
		// },
	}
	// editIfaceCmd.AddCmd(&ishell.Cmd{
	// 	Name: "add_face_to_vlan",
	// 	Help: "add_face_to_vlan",
	// 	Func: func(c *ishell.Context) {
	// 		if len(c.Args) == 0 {
	// 			c.Err(errors.New("missing interface name"))
	// 			return
	// 		}
	// 		vlans = append(vlans, c.Args...)
	// 	},
	// })
	editIfaceCmd.AddCmd(executePromptCmd)
	editCmd.AddCmd(editIfaceCmd)
	// editCmd.AddCmd(executePromptCmd)

	addCmd := &ishell.Cmd{
		Name: "add",
		Help: "add",
		LongHelp: `Try dynamic autocomplete by adding and removing words.
	Then view the autocomplete by tabbing after "words" subcommand.

	This is an example of a long help.`,
	}
	addIfaceCmd := &ishell.Cmd{
		Name: "interfaces",
		Help: "Specify network interfaces to add into a VLAN",
		LongHelp: `Try dynamic autocomplete by adding and removing words.
	Then view the autocomplete by tabbing after "words" subcommand.

	This is an example of a long help.`,
	}
	addIfaceToCmd := &ishell.Cmd{
		Name: "to",
		Help: "to",
		LongHelp: `Try dynamic autocomplete by adding and removing words.
	Then view the autocomplete by tabbing after "words" subcommand.

	This is an example of a long help.`,
	}
	addIfaceToVlanCmd := &ishell.Cmd{
		Name:     "vlan",
		Help:     "<VLAN-ID> <IFACE-ID> [<IFACE-ID>...]",
		LongHelp: `add interfaces to vlan <VLAN ID> <PORT NAME>`,
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			defaultInput := "vlan-1 eth-1 eth-2"
			if len(c.Args) > 0 {
				defaultInput = strings.Join(c.Args, " ")
			}

			c.Print("input: ")
			read := c.ReadLineWithDefault(defaultInput)

			if read == defaultInput {
				c.Println("you left the default input intact")
			} else {
				c.Printf("you modified input to '%s'", read)
				c.Println()
			}
		},
	}

	addIfaceToCmd.AddCmd(addIfaceToVlanCmd)
	addIfaceCmd.AddCmd(addIfaceToCmd)
	addIfaceCmd.AddCmd(addIfaceCmd)
	addCmd.AddCmd(addIfaceCmd)
	shell.AddCmd(addCmd)

	shell.AddCmd(editCmd)

	// when started with "exit" as first argument, assume non-interactive execution
	if len(os.Args) > 1 && os.Args[1] == "exit" {
		shell.Process(os.Args[2:]...)
	} else {
		// start shell
		shell.Run()
		// teardown
		shell.Close()
	}
}
